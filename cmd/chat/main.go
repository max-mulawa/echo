package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	serverPort = 8888
)

var (
	errInvalidUsername    = errors.New("username should be between 1 and 16 alphanumeric characters")
	errUniqueUsername     = errors.New("username already taken")
	errChatMessageTooLong = errors.New("chat message too long, 1100 characters allowed")
	isAlphanumeric        = regexp.MustCompile(`^[a-zA-Z0-9]{1,16}$`).MatchString
)

func main() {
	startServer()
}

type Message struct {
	from        string
	body        string
	excludeFrom bool
}

type MemberNet struct {
	conn net.Conn
	w    *bufio.Writer
	r    *bufio.Reader
}

type Member struct {
	name  string
	input chan Message
	conn  *MemberNet
}

type ChatRoom struct {
	members map[string]*Member
	lock    *sync.RWMutex
}

func (r *ChatRoom) initMember(c net.Conn) (*Member, error) {
	mnet := newMemberNetwork(c)
	_, err := mnet.w.WriteString("Welcome to budgetchat! What shall I call you?\n")
	if err != nil {
		return nil, fmt.Errorf("failed to write welcome message: %w", err)
	}
	err = mnet.w.Flush()
	if err != nil {
		return nil, fmt.Errorf("failed to flush write: %w", err)
	}

	username, err := mnet.r.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read username: %w", err)
	}
	username = normalizeReadLine(username)

	if !isAlphanumeric(username) {
		return nil, errInvalidUsername
	}

	if r.memberNameTaken(username) {
		return nil, errUniqueUsername
	}

	u := &Member{}
	u.input = make(chan Message, 1)
	u.name = username
	u.conn = mnet

	return u, nil
}

func normalizeReadLine(txt string) string {
	return strings.TrimSuffix(txt, "\n")
}

func newMemberNetwork(c net.Conn) *MemberNet {
	return &MemberNet{
		conn: c,
		w:    bufio.NewWriter(c),
		r:    bufio.NewReader(c),
	}
}

func (r *ChatRoom) memberNameTaken(name string) bool {
	r.lock.RLock()
	defer r.lock.RUnlock()
	_, ok := r.members[name]
	return ok
}

func (r *ChatRoom) registerMember(m *Member) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.members[m.name] = m
}

func (r *ChatRoom) unregisterMember(u *Member) {
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.members, u.name)
}

func (r *ChatRoom) onRegisteredUser(u *Member) error {
	msg := Message{
		from:        u.name,
		body:        fmt.Sprintf("* %s has entered the room", u.name),
		excludeFrom: true,
	}
	r.publish(msg)
	members := r.getOtherMembers(u.name)
	err := u.SendTxt(fmt.Sprintf("* The room contains: %s\n", strings.Join(members, ", ")))
	if err != nil {
		return fmt.Errorf("cannot list members on join: %w", err)
	}
	return nil
}

func (m *Member) SendTxt(txt string) error {
	_, err := m.conn.w.Write([]byte(txt))
	if err != nil {
		return fmt.Errorf("cannot send txt: %w", err)
	}

	err = m.conn.w.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush write: %w", err)
	}

	return nil
}

func (m *Member) Send(msg Message) error {
	if msg.excludeFrom {
		return m.SendTxt(fmt.Sprintf("%s\n", msg.body))
	} else {
		return m.SendTxt(fmt.Sprintf("[%s] %s\n", msg.from, msg.body))
	}
}

func (m *Member) ReadMemberMessage() (*Message, error) {
	txt, err := m.conn.r.ReadString('\n')
	if err != nil {
		return nil, err
	}

	if len(txt) > 1100 {
		return nil, errChatMessageTooLong
	}

	return &Message{from: m.name, body: normalizeReadLine(txt)}, nil
}

func (r *ChatRoom) getMembers() []string {
	r.lock.RLock()
	defer r.lock.RUnlock()

	var members []string
	for name := range r.members {
		members = append(members, name)
	}
	return members
}

func (r *ChatRoom) getOtherMembers(me string) []string {
	members := r.getMembers()
	var others []string
	for _, name := range members {
		if name == me {
			continue
		}
		others = append(others, name)
	}

	return others
}

func (r *ChatRoom) publish(msg Message) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	for username, v := range r.members {
		if msg.from != username {
			v.input <- msg
		}
	}
}

func startServer() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", serverPort))
	if err != nil {
		os.Exit(1)
	}

	r := newRoom()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}

		conn.SetDeadline(time.Now().Add(time.Second * 120))
		go handleConnection(conn, r)
	}
}

func newRoom() *ChatRoom {
	return &ChatRoom{
		members: make(map[string]*Member),
		lock:    &sync.RWMutex{},
	}
}

func handleConnection(c net.Conn, r *ChatRoom) {
	defer func() {
		fmt.Print("Closing connection on server\n")
		c.Close()
	}()

	// write to provide username
	// read username
	member, err := r.initMember(c)
	if err != nil {
		if errors.Is(err, errInvalidUsername) || errors.Is(err, errUniqueUsername) {
			c.Write([]byte(err.Error()))
		} else {
			log.Printf("member init failed: %v", err)
		}
		return
	}
	r.registerMember(member)
	defer r.unregisterMember(member)
	r.onRegisteredUser(member)

	go r.readMember(member)

	for {
		select {
		case msg := <-member.input:
			member.Send(msg)
		}
	}
}

func (r *ChatRoom) readMember(m *Member) {
	for {
		msg, err := m.ReadMemberMessage()
		if err != nil {
			if err == io.EOF {
				log.Printf("Client closed connection")
				r.publish(Message{
					from:        m.name,
					body:        fmt.Sprintf("* %s has left the room", m.name),
					excludeFrom: true,
				})
				r.unregisterMember(m)
				return
			} else if opError, ok := err.(*net.OpError); ok {
				if opError.Op == "abc" {
					log.Printf("abc reading message failed: %v", err)
				}
				continue
			} else {
				log.Printf("reading message failed: %v", err)
				continue
			}
		}
		r.publish(*msg)
	}
}
