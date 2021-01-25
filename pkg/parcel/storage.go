package parcel

import (
	"github.com/google/uuid"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/address"
)

type Inserter interface {
	Insert(p *Parcel) error
}

type Accesser interface {
	ByID(id uuid.UUID) (*Parcel, error)
	ByDestination(a *address.Address) ([]*Parcel, error)
}

type Storage interface {
	Inserter
	Accesser
}

type EventInserter interface {
	Insert(e *Event) error
}

type EventAccesser interface {
	ByParcel(p *Parcel) ([]*Event, error)
}

type EventStorage interface {
	EventInserter
	EventAccesser
}
