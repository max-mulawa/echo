package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

const (
	serverPort     = 9999
	ProductVersion = "Maks Key-Value Store 0.1"
)

type storage map[string]string

type Operation string

type Store struct {
	s    storage
	lock sync.RWMutex
}

type Result struct {
	Op      Operation
	Payload string
}

type InsertCmd struct {
	Key   string
	Value string
}

type ReadQuery struct {
	Key string
}

type VersionQuery struct {
}

type Action interface {
	Run(*Store) (*Result, error)
}

var (
	errNotFound = errors.New("key not found")

	NoOp Operation = "NoOp"
	Send Operation = "Send"
)

func main() {
	startServer()
}

func startServer() {
	s, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", serverPort))
	if err != nil {
		log.Fatalf("udp resolution failed: %v", err)
	}

	conn, err := net.ListenUDP("udp4", s)
	if err != nil {
		log.Fatalf("udp listen failed: %v", err)
	}

	defer conn.Close()

	store := NewStore()

	for {
		buffer := make([]byte, 1000)
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("error occured reading: %v", err)
		}

		data := string(buffer[0:n])
		action := getAction(data)

		res, err := action.Run(store)
		if err != nil {
			fmt.Printf("error occured running action: %v", err)
		}

		switch res.Op {
		case NoOp:
			continue
		case Send:
			_, err = conn.WriteToUDP([]byte(res.Payload), addr)
			if err != nil {
				fmt.Printf("error occured writing: %v", err)
			}
		default:
			log.Fatalf("%s opp is not supported", res.Op)
		}

	}

}

func NewStore() *Store {
	return &Store{
		lock: sync.RWMutex{},
		s:    make(storage),
	}
}

func (s *Store) Insert(c *InsertCmd) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.s[c.Key] = c.Value
}

func (s *Store) Read(q *ReadQuery) (string, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var v string
	var ok bool
	if v, ok = s.s[q.Key]; !ok {
		return "", errNotFound
	}

	return v, nil
}

func (s *storage) Version(v *VersionQuery) string {
	return ProductVersion
}

func getAction(input string) Action {
	if strings.Contains(input, "=") {
		result := strings.SplitAfterN(input, "=", 2)

		key := result[0][:len(result[0])-1]

		return &InsertCmd{Key: key, Value: result[1]}
	} else if input == "version" {
		return &VersionQuery{}
	} else {
		return &ReadQuery{Key: input}
	}
}

func (c *InsertCmd) Run(s *Store) (*Result, error) {
	s.Insert(c)
	return &Result{Op: NoOp}, nil //noop
}

func (r *ReadQuery) Run(s *Store) (*Result, error) {
	val, err := s.Read(r)

	if err != nil {
		if errors.Is(err, errNotFound) {
			return &Result{Op: Send, Payload: fmt.Sprintf("%s=%s", r.Key, "")}, nil
		}
		return nil, err
	}

	return &Result{Op: Send, Payload: fmt.Sprintf("%s=%s", r.Key, val)}, nil
}

func (v *VersionQuery) Run(s *Store) (*Result, error) {
	version := s.s.Version(v)
	return &Result{Op: Send, Payload: fmt.Sprintf("version=%s", version)}, nil
}
