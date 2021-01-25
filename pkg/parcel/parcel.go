// Package parcel implements data structures and
// interfaces for working with parcel information.
package parcel

import (
	"time"

	"github.com/google/uuid"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/address"
)

// Parcel is the data type representing a single parcel.
type Parcel struct {
	// ID is the parcel's unique identifier. This is also its tracking id.
	ID                 uuid.UUID
	ReturnAddress      *address.Address
	DestinationAddress *address.Address
}

type EventType int

const (
	DataReceived EventType = iota
	DeliveredToIPPS
	DeliveredToProcessing
	LoadedIntoRocket
	LoadedIntoVehicle
	DeliveredToDestination
)

func (t EventType) String() string {
	switch t {
	case DataReceived:
		return "We have received parcel processing information from the sender"
	case DeliveredToIPPS:
		return "The sender has delivered the parcel to one of our shops"
	case DeliveredToProcessing:
		return "The parcel has been delivered to one of our logistics centers"
	case LoadedIntoRocket:
		return "The parcel has been loaded into one of our delivery rockets"
	case LoadedIntoVehicle:
		return "The parcel has been loaded into a vehicle and is going to be delivered to its final destination"
	case DeliveredToDestination:
		return "The package has been delivered to its destination"
	default:
		return "Unknown event"
	}
}

// Event is the type representing tracking events for parcels.
type Event struct {
	ID     uuid.UUID
	Parcel *Parcel
	Type   EventType
	Time   time.Time
}
