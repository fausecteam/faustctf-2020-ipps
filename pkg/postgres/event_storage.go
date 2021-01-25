package postgres

import (
	"database/sql"

	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/parcel"
)

const (
	installParcelEventTable = `CREATE TABLE IF NOT EXISTS ipps_parcel_event(
		id         uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
		event_type integer     NOT NULL,
		event_time timestamptz NOT NULL,
		parcel     uuid 	   NOT NULL CONSTRAINT ipps_parcel_event_parcel_fkey
                               REFERENCES ipps_parcel (id) ON DELETE CASCADE ON UPDATE CASCADE
	);`
	insertParcelEventStmt = `INSERT INTO ipps_parcel_event (id, event_type, event_time, parcel)
                             VALUES ($1, $2, $3, $4);`
	parcelEventByParcelStmt = `SELECT (id, event_type, event_time)
                               FROM ipps_parcel_event
                               WHERE parcel = $1
                               ORDER BY event_time ASC;`
)

type EventStorage struct {
	insert   *sql.Stmt
	byParcel *sql.Stmt
}

func NewEventStorage(db *sql.DB) (*EventStorage, error) {
	s := &EventStorage{}
	var err error
	s.insert, err = db.Prepare(insertParcelEventStmt)
	if err != nil {
		return nil, err
	}
	s.byParcel, err = db.Prepare(parcelEventByParcelStmt)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (es *EventStorage) Insert(e *parcel.Event) error {
	_, err := es.insert.Exec(e.ID, e.Type, e.Time, e.Parcel.ID)

	return err
}

func (es *EventStorage) ByParcel(p *parcel.Parcel) ([]*parcel.Event, error) {
	rows, err := es.byParcel.Query(p.ID)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var ee []*parcel.Event
	for rows.Next() {
		e := &parcel.Event{Parcel: p}
		err := rows.Scan(&e.ID, &e.Type, &e.Time)
		if err != nil {
			return nil, err
		}
		ee = append(ee, e)
	}

	return ee, nil
}

func (es *EventStorage) Close() error {
	err := es.insert.Close()
	if err != nil {
		return err
	}

	return es.byParcel.Close()
}
