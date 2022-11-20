package ticketing

import (
	"fmt"
	"sync"
)

type TicketEntry struct {
	Plate string
	Day   int32
}

type TicketLedger struct {
	store map[TicketEntry]bool
	lock  sync.Mutex
}

func NewTicketLedger() *TicketLedger {
	return &TicketLedger{
		store: make(map[TicketEntry]bool),
		lock:  sync.Mutex{},
	}
}

func (l *TicketLedger) Add(t TicketMsg) bool {
	// check in the ledger if there is already a ticket for this day, for this car
	// floor(timestamp / 86400) = day
	startDay := int32(float64(t.Timestamp1.Unix() / 86400))
	endDay := int32(float64(t.Timestamp2.Unix() / 86400))

	days := []int32{startDay}
	if startDay < endDay {
		days = append(days, endDay)
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	entries := make(map[TicketEntry]bool)
	for _, day := range days {
		entry := TicketEntry{
			Plate: t.Plate,
			Day:   day,
		}
		_, existsInLedger := l.store[entry]
		if !existsInLedger {
			entries[entry] = true
			fmt.Println("ticket added to ledger ", entry)
		} else {
			entries = make(map[TicketEntry]bool)
			fmt.Println("ticket already stored in the ledger ", entry)
			break
		}
	}

	if len(entries) > 0 {
		for k, v := range entries {
			l.store[k] = v
		}
	}

	return len(entries) > 0
}
