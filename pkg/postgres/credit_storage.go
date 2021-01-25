package postgres

import (
	"database/sql"

	"github.com/google/uuid"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/credit"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/user"
)

const (
	installCardTable = `CREATE TABLE IF NOT EXISTS ipps_card (
							id      uuid PRIMARY KEY DEFAULT gen_random_uuid(),
							num     text NOT NULL,
							user_id uuid NOT NULL CONSTRAINT ipps_card_user_fkey REFERENCES ipps_user
										ON UPDATE CASCADE
										ON DELETE CASCADE
						);`
	insertCardStmt = `INSERT INTO ipps_card (id, num, user_id)
					  VALUES ($1, $2, $3);`
	cardByUserStmt = `SELECT id, num, user_id
					  FROM ipps_card
					  WHERE user_id = $1;`
	updateCardStmt = `UPDATE ipps_card
					  SET num = $2
					  WHERE id = $1;`
	deleteCardStmt = `DELETE
					  FROM ipps_card
					  WHERE id = $1;`
)

// CreditCardStorage is an implementation of the credit.Storage interface
// using a PostgreSQL database as its underlying storage.
type CreditCardStorage struct {
	insert *sql.Stmt
	byUser *sql.Stmt
	update *sql.Stmt
	delete *sql.Stmt
}

// New CreditCardStorage returns
func NewCreditCardStorage(db *sql.DB) (*CreditCardStorage, error) {
	cs := &CreditCardStorage{}
	var err error
	cs.insert, err = db.Prepare(insertCardStmt)
	if err != nil {
		return nil, err
	}
	cs.update, err = db.Prepare(updateCardStmt)
	if err != nil {
		return nil, err
	}
	cs.byUser, err = db.Prepare(cardByUserStmt)
	if err != nil {
		return nil, err
	}
	cs.delete, err = db.Prepare(deleteCardStmt)
	if err != nil {
		return nil, err
	}

	return cs, nil
}

func (cs *CreditCardStorage) Insert(c *credit.Card) error {
	_, err := cs.insert.Exec(c.ID, c.Number, c.User.ID)
	return err
}

func (cs *CreditCardStorage) ByUser(u *user.User) ([]*credit.Card, error) {
	rows, err := cs.byUser.Query(u.ID)
	if err == sql.ErrNoRows {
		return nil, credit.ErrNoCards
	} else if err != nil {
		return nil, err
	}
	cc := make([]*credit.Card, 0)
	for rows.Next() {
		var uid uuid.UUID
		c := &credit.Card{User: u}
		err := rows.Scan(&c.ID, &c.Number, &uid)
		if err != nil {
			return nil, err
		}
		cc = append(cc, c)
	}

	return cc, nil
}

func (cs *CreditCardStorage) Update(c *credit.Card) error {
	_, err := cs.update.Exec(c.ID, c.Number, c.User.ID)
	return err
}

func (cs *CreditCardStorage) Delete(c *credit.Card) error {
	_, err := cs.delete.Exec(c.ID)
	return err
}

func (cs *CreditCardStorage) Close() error {
	err := cs.insert.Close()
	if err != nil {
		return err
	}
	err = cs.byUser.Close()
	if err != nil {
		return err
	}
	err = cs.update.Close()
	if err != nil {
		return err
	}
	err = cs.delete.Close()
	if err != nil {
		return err
	}

	return nil
}
