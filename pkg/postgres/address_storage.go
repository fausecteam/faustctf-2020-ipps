package postgres

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/address"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/user"
)

const (
	installAddressTable = `CREATE TABLE IF NOT EXISTS ipps_address (
							id     uuid PRIMARY KEY DEFAULT gen_random_uuid(),
							street  text NOT NULL,
							zip     text NOT NULL,
							city    text NOT NULL,
							country text NOT NULL,
							planet  text NOT NULL DEFAULT 'Mars',
							user_id  uuid NOT NULL CONSTRAINT ipps_address_user_fkey
								REFERENCES ipps_user ON DELETE CASCADE ON UPDATE CASCADE,
		CONSTRAINT ipps_address_unique_per_user
			UNIQUE (street, zip, city, country, planet, user_id)
	);`
	addressByID = `SELECT id, street, zip, city, country, planet
					 FROM ipps_address
					 WHERE id = $1;`
	addressByUser = `SELECT id, street, zip, city, country, planet
					 FROM ipps_address
					 WHERE user_id = $1;`
	insertAddress = `INSERT INTO ipps_address (id, street, zip, city, country, planet, user_id)
					 VALUES ($1, $2, $3, $4, $5, $6, $7);`
	updateAddress = `UPDATE ipps_address
					 SET (street, zip, city, country, planet) = ($2, $3, $4, $5, $6)
					 WHERE id = $1;`
)

// AddressStorage is the type implemented the address.Storage interface.
type AddressStorage struct {
	byID   *sql.Stmt
	byUser *sql.Stmt
	insert *sql.Stmt
	update *sql.Stmt
}

func NewAddressStorage(db *sql.DB) (*AddressStorage, error) {
	s := &AddressStorage{}
	var err error

	s.byID, err = db.Prepare(addressByID)
	if err != nil {
		return nil, err
	}
	s.byUser, err = db.Prepare(addressByUser)
	if err != nil {
		return nil, err
	}
	s.insert, err = db.Prepare(insertAddress)
	if err != nil {
		return nil, err
	}
	s.update, err = db.Prepare(updateAddress)
	if err != nil {
		return nil, err
	}

	return s, nil
}
func (s *AddressStorage) ByID(id uuid.UUID) (*address.Address, error) {
	a := &address.Address{}
	err := s.byID.QueryRow(id).Scan(&a.ID, &a.Street, &a.Zip, &a.City, &a.Country, &a.Planet)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return a, nil
}

func (s *AddressStorage) ByUser(u *user.User) ([]*address.Address, error) {
	rr, err := s.byUser.Query(u.ID)
	if err != nil {
		return nil, err
	}

	var aa []*address.Address
	for rr.Next() {
		a := &address.Address{User: u}
		err := rr.Scan(&a.ID, &a.Street, &a.Zip, &a.City, &a.Country, &a.Planet)
		if err != nil {
			return nil, err
		}
		aa = append(aa, a)
	}

	return aa, nil
}

func (s *AddressStorage) Insert(a *address.Address) error {
	_, err := s.insert.Exec(a.ID, a.Street, a.Zip, a.City, a.Country, a.Planet, a.User.ID)
	if err != nil {
		pgErr, ok := err.(*pq.Error)
		if ok && pgErr.Constraint == "ipps_address_unique_per_user" {
			return address.ErrAddressAlreadyAdded
		}
	}

	return err
}

func (s *AddressStorage) Update(a *address.Address) error {
	_, err := s.update.Exec(a.ID, a.Street, a.Zip, a.City, a.Country, a.Planet)
	if err != nil {
		pgErr, ok := err.(*pq.Error)
		if ok && pgErr.Constraint == "ipps_address_unique_per_user" {
			return address.ErrAddressAlreadyAdded
		}
	}

	return nil
}

func (s *AddressStorage) Close() error {
	err := s.byUser.Close()
	if err != nil {
		return err
	}
	err = s.insert.Close()
	if err != nil {
		return err
	}

	return s.update.Close()
}
