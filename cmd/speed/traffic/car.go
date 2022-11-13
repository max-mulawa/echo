package traffic

import (
	"max-mulawa/echo/cmd/speed/tracking"
	"sort"
	"sync"
)

type MeasurementsRegistry struct {
	cars sync.Map
	pub  FeedPublisher
}

type Car struct {
	Plate   string
	PerRoad map[uint16][]tracking.Measurement
	lock    sync.Mutex
}

func NewMeasurementsRegistry(pub FeedPublisher) *MeasurementsRegistry {
	return &MeasurementsRegistry{
		cars: sync.Map{},
		pub:  pub,
	}
}

func (r *MeasurementsRegistry) Register(m tracking.Measurement) {
	plate := m.Time.Plate
	road := m.Device.Road
	car, _ := r.cars.LoadOrStore(plate, &Car{
		Plate:   plate,
		PerRoad: map[uint16][]tracking.Measurement{},
		lock:    sync.Mutex{},
	})

	go car.(*Car).registerMeasurement(r, road, m)
}

func (c *Car) registerMeasurement(r *MeasurementsRegistry, road uint16, m tracking.Measurement) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.PerRoad[road] = append(c.PerRoad[road], m)
	roadMeasures := c.PerRoad[road]
	sort.Slice(roadMeasures, func(i, j int) bool {
		return roadMeasures[i].Time.Timestamp.Before(roadMeasures[j].Time.Timestamp)
	})
	r.onRegisterMeasurement(c, road, roadMeasures)
}

func (r *MeasurementsRegistry) onRegisterMeasurement(car *Car, road uint16, roadMeasures []tracking.Measurement) {
	if len(roadMeasures) == 0 {
		return
	}

	first := roadMeasures[0]
	roadLimit := float64(first.Device.Limit)

	// discover the offence
	for i := 0; i < len(roadMeasures); i++ {
		for j := i + 1; j < len(roadMeasures); j++ {
			mile1 := roadMeasures[i].Device.Mile
			timestamp1 := roadMeasures[i].Time.Timestamp
			mile2 := roadMeasures[j].Device.Mile
			timestamp2 := roadMeasures[j].Time.Timestamp
			distance := float64(mile2 - mile1)
			duration := timestamp2.Sub(timestamp1)

			if distance > 0 {
				speed := distance / duration.Hours()
				if speed >= (roadLimit + 0.5) {
					// publish offense
					r.pub.Publish(Offense{
						Plate:      car.Plate,
						Road:       road,
						Mile1:      mile1,
						Timestamp1: timestamp1,
						Mile2:      mile2,
						Timestamp2: timestamp2,
						Speed:      uint16(speed) * 100,
					})
				}
			}
		}
	}
}
