package address

import (
	"github.com/google/uuid"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/user"
)

type Accesser interface {
	ByID(id uuid.UUID) (*Address, error)
	ByUser(u *user.User) ([]*Address, error)
}

type Inserter interface {
	Insert(a *Address) error
}

type Updater interface {
	Update(a *Address) error
}

type Storage interface {
	Accesser
	Inserter
	Updater
}
