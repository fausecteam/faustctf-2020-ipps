package postgres

import (
	"database/sql"

	"github.com/google/uuid"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/address"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/parcel"
)

const (
	installParcelTable = `CREATE TABLE IF NOT EXISTS ipps_parcel(
		id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
		destination_address uuid CONSTRAINT ipps_parcel_dest_addr_fkey
								 REFERENCES ipps_address (id) ON DELETE SET NULL ON UPDATE CASCADE,
		return_address      uuid CONSTRAINT ipps_parcel_return_addr_fkey
								 REFERENCES ipps_address (id) ON DELETE SET NULL ON UPDATE CASCADE
	);`
	insertParcelStmt = `INSERT INTO ipps_parcel(id, destination_address, return_address)
						VALUES ($1, $2, $3);`
	parcelByIDStmt = `SELECT (id, destination_address, return_address)
					  FROM ipps_parcel
					  WHERE id = $1;`
	parcelByDestinationStmt = `SELECT (id, destination_address, return_address)
					  FROM ipps_parcel
					  WHERE destination_address = $1;`
)

// ParcelStorage is the PostgreSQL based implementation of
// the parcel.Storage and parcel.EventStorage interfaces.
type ParcelStorage struct {
	insert        *sql.Stmt
	byID          *sql.Stmt
	byDestination *sql.Stmt
}

func NewParcelStorage(db *sql.DB) (*ParcelStorage, error) {
	ps := &ParcelStorage{}
	var err error
	ps.insert, err = db.Prepare(insertParcelStmt)
	if err != nil {
		return nil, err
	}
	ps.byID, err = db.Prepare(parcelByIDStmt)
	if err != nil {
		return nil, err
	}
	ps.byDestination, err = db.Prepare(parcelByDestinationStmt)
	if err != nil {
		return nil, err
	}

	return ps, nil
}

func (ps *ParcelStorage) Insert(p *parcel.Parcel) error {
	_, err := ps.insert.Exec(p.ID, p.DestinationAddress.ID, p.DestinationAddress.ID)
	return err
}

func (ps *ParcelStorage) ByID(id uuid.UUID) (*parcel.Parcel, error) {
	p := &parcel.Parcel{}
	err := ps.byID.QueryRow(id).Scan(&p.ID, &p.DestinationAddress, &p.ReturnAddress)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return p, nil
}

func (ps *ParcelStorage) ByDestination(a *address.Address) ([]*parcel.Parcel, error) {
	rows, err := ps.byDestination.Query(a.ID)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	pp := make([]*parcel.Parcel, 0)
	for rows.Next() {
		p := &parcel.Parcel{}
		err := rows.Scan(&p.ID, &p.DestinationAddress, &p.ReturnAddress)
		if err != nil {
			return nil, err
		}
		pp = append(pp, p)
	}

	return pp, nil
}

func (ps *ParcelStorage) Close() error {
	err := ps.insert.Close()
	if err != nil {
		return err
	}
	err = ps.byDestination.Close()
	if err != nil {
		return err
	}

	return ps.byID.Close()
}
