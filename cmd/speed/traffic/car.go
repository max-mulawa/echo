package traffic

import (
	"fmt"
	"max-mulawa/echo/cmd/speed/tracking"
	"regexp"
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

var carPlateRegEx = regexp.MustCompile("^[A-Z0-9]*$")

func NewMeasurementsRegistry(pub FeedPublisher) *MeasurementsRegistry {
	return &MeasurementsRegistry{
		cars: sync.Map{},
		pub:  pub,
	}
}

func (r *MeasurementsRegistry) Register(m tracking.Measurement) error {
	plate := m.Time.Plate
	road := m.Device.Road

	if !carPlateRegEx.Match([]byte(plate)) {
		return fmt.Errorf("invalid format of car plate %q", plate)
	}

	car, _ := r.cars.LoadOrStore(plate, &Car{
		Plate:   plate,
		PerRoad: map[uint16][]tracking.Measurement{},
		lock:    sync.Mutex{},
	})

	go car.(*Car).registerMeasurement(r, road, m)
	return nil
}

func (c *Car) registerMeasurement(r *MeasurementsRegistry, road uint16, m tracking.Measurement) {
	c.lock.Lock()
	defer c.lock.Unlock()

	fmt.Printf("registering measurement for car (%s) on road (%d) at (%s), camera (%d , %d) \n", m.Time.Plate, road, m.Time.Timestamp, road, m.Device.Mile)

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
					o := Offense{
						Plate:      car.Plate,
						Road:       road,
						Mile1:      mile1,
						Timestamp1: timestamp1,
						Mile2:      mile2,
						Timestamp2: timestamp2,
						Speed:      uint16(speed) * 100,
					}
					r.pub.Publish(o)
				}
			}
		}
	}
}
