package traffic

import (
	"fmt"
	"sync"
	"time"
)

type Offense struct {
	Plate      string
	Road       uint16
	Mile1      uint16
	Timestamp1 time.Time
	Mile2      uint16
	Timestamp2 time.Time
	Speed      uint16
}

type OffenseFeed struct {
	published sync.Map
	Sub       chan<- Offense
	lock      sync.RWMutex
}

type FeedPublisher interface {
	Publish(o Offense)
}

func NewOffenseFeed(sub chan<- Offense) *OffenseFeed {
	return &OffenseFeed{
		published: sync.Map{},
		Sub:       sub,
		lock:      sync.RWMutex{},
	}
}

func (p *OffenseFeed) Publish(o Offense) {
	if _, exists := p.published.LoadOrStore(o, true); !exists {
		fmt.Printf("publishing offense: %v\n", o)
		p.Sub <- o
	}
}
