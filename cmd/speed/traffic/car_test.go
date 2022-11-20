package traffic_test

import (
	"max-mulawa/echo/cmd/speed/tracking"
	"max-mulawa/echo/cmd/speed/traffic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type MockFeed struct {
	Offense []traffic.Offense
}

func (f *MockFeed) Publish(o traffic.Offense) {
	f.Offense = append(f.Offense, o)
}

func TestOffenseCalculation(t *testing.T) {
	feed := &MockFeed{}
	registry := traffic.NewMeasurementsRegistry(feed)
	registry.Register(tracking.Measurement{
		Device: tracking.IAmCameraMsg{
			Road: 8945,
			Mile: 1011,
		},
		Time: tracking.MeasurementTimeMsg{
			Plate:     "RB84TGF",
			Timestamp: time.Unix(60350885, 0),
		},
	})
	registry.Register(tracking.Measurement{
		Device: tracking.IAmCameraMsg{
			Road: 8945,
			Mile: 10,
		},
		Time: tracking.MeasurementTimeMsg{
			Plate:     "RB84TGF",
			Timestamp: time.Unix(60397346, 0),
		},
	})

	for len(feed.Offense) == 0 {
		time.Sleep(time.Millisecond * 10)
	}
	require.Equal(t, uint16(7756), feed.Offense[0].Speed)

}
