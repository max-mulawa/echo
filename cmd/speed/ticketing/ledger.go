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
	store sync.Map
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

	added := false
	for _, day := range days {
		entry := TicketEntry{
			Plate: t.Plate,
			Day:   day,
		}
		_, existsInLedger := l.store.LoadOrStore(entry, true)
		if !existsInLedger {
			added = true
			fmt.Println("ticket added to ledger ", entry)
		} else {
			fmt.Println("ticket already stored in the ledger ", entry)
		}
	}

	return added
}
