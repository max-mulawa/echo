package ticketing_test

import (
	"max-mulawa/echo/cmd/speed/ticketing"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAddingToLedger(t *testing.T) {
	ledger := ticketing.NewTicketLedger()

	// 13-14.09.1971
	added := ledger.Add(ticketing.TicketMsg{
		Plate:      "ABC",
		Road:       1,
		Speed:      10,
		Timestamp1: time.Date(1971, 9, 14, 14, 0, 0, 0, time.Local),
		Timestamp2: time.Date(1971, 9, 14, 19, 0, 0, 0, time.Local),
	})
	require.True(t, added)
	added = ledger.Add(ticketing.TicketMsg{
		Plate:      "ABC",
		Road:       1,
		Speed:      10,
		Timestamp1: time.Date(1971, 9, 13, 22, 0, 0, 0, time.Local),
		Timestamp2: time.Date(1971, 9, 14, 3, 0, 0, 0, time.Local),
	})
	require.False(t, added)

	// 01-02.08.1970
	added = ledger.Add(ticketing.TicketMsg{
		Plate:      "ABC",
		Road:       1,
		Speed:      10,
		Timestamp1: time.Date(1971, 8, 1, 20, 0, 0, 0, time.Local),
		Timestamp2: time.Date(1971, 8, 1, 22, 0, 0, 0, time.Local),
	})
	require.True(t, added)

	added = ledger.Add(ticketing.TicketMsg{
		Plate:      "ABC",
		Road:       1,
		Speed:      10,
		Timestamp1: time.Date(1971, 8, 1, 19, 0, 0, 0, time.Local),
		Timestamp2: time.Date(1971, 8, 2, 5, 0, 0, 0, time.Local),
	})
	require.False(t, added)

	added = ledger.Add(ticketing.TicketMsg{
		Plate:      "ABC",
		Road:       1,
		Speed:      10,
		Timestamp1: time.Date(1971, 8, 2, 15, 0, 0, 0, time.Local),
		Timestamp2: time.Date(1971, 8, 2, 21, 0, 0, 0, time.Local),
	})
	require.True(t, added)

	// 14-15.01.1971
	added = ledger.Add(ticketing.TicketMsg{
		Plate:      "ABC",
		Road:       1,
		Speed:      10,
		Timestamp1: time.Date(1971, 1, 15, 15, 0, 0, 0, time.Local),
		Timestamp2: time.Date(1971, 1, 15, 21, 0, 0, 0, time.Local),
	})
	require.True(t, added)

	added = ledger.Add(ticketing.TicketMsg{
		Plate:      "ABC",
		Road:       1,
		Speed:      10,
		Timestamp1: time.Date(1971, 1, 14, 21, 0, 0, 0, time.Local),
		Timestamp2: time.Date(1971, 1, 15, 1, 0, 0, 0, time.Local),
	})
	require.False(t, added)

	added = ledger.Add(ticketing.TicketMsg{
		Plate:      "ABC",
		Road:       1,
		Speed:      10,
		Timestamp1: time.Date(1971, 1, 14, 21, 0, 0, 0, time.Local),
		Timestamp2: time.Date(1971, 1, 14, 22, 0, 0, 0, time.Local),
	})
	require.True(t, added)
}

// ticket was dispatched (road: 58592) ticket ({GR07HYG 58592 401 1970-08-01 20:13:25 +0000 UTC 198 1970-08-01 22:51:16 +0000 UTC 7716})
