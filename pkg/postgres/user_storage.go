package postgres

import (
	"database/sql"
	"net/mail"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/user"
)

const (
	installUserTable = `CREATE TABLE IF NOT EXISTS ipps_user (
							id                 uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
							username           varchar(128) NOT NULL CONSTRAINT ipps_user_username_key UNIQUE,
							email              varchar(128) NOT NULL CONSTRAINT ipps_user_email_key UNIQUE,
							full_name          varchar(128) NOT NULL,
                            password           varchar(64)  NOT NULL
						);`
	insertUserStmt = `INSERT INTO ipps_user (id, username, email, password, full_name)
                      VALUES ($1, $2, $3, $4, $5);`
	updateUserStmt = `UPDATE ipps_user
                      SET (email, password, full_name) = ($2, $3, $4)
                      WHERE id = $1;`
	deleteUserStmt = `DELETE FROM ipps_user WHERE id = $1;`
	userByIDStmt   = `SELECT id, username, email, password, full_name
                      FROM ipps_user
					  WHERE id = $1;`
	userByNameStmt = `SELECT id, username, email, password, full_name
                      FROM ipps_user
                      WHERE username = $1;`
	userByEmailStmt = `SELECT id, username, email, password, full_name
                      FROM ipps_user
                      WHERE email = $1;`
)

// UserStorage implements the user.Storage interface for a postgres
// database.
type UserStorage struct {
	insert     *sql.Stmt
	update     *sql.Stmt
	delete     *sql.Stmt
	byID       *sql.Stmt
	byUsername *sql.Stmt
	byEmail    *sql.Stmt
}

// NewUserStorage returns a new user storage that runs its database
// queries on db.
func NewUserStorage(db *sql.DB) (*UserStorage, error) {
	us := &UserStorage{}
	var err error
	us.insert, err = db.Prepare(insertUserStmt)
	if err != nil {
		return nil, err
	}
	us.update, err = db.Prepare(updateUserStmt)
	if err != nil {
		return nil, err
	}
	us.delete, err = db.Prepare(deleteUserStmt)
	if err != nil {
		return nil, err
	}
	us.byID, err = db.Prepare(userByIDStmt)
	if err != nil {
		return nil, err
	}
	us.byUsername, err = db.Prepare(userByNameStmt)
	if err != nil {
		return nil, err
	}
	us.byEmail, err = db.Prepare(userByEmailStmt)
	if err != nil {
		return nil, err
	}

	return us, nil
}

func (us *UserStorage) Insert(u *user.User) error {
	_, err := us.insert.Exec(u.ID, u.Username, u.Email.Address, u.Password, u.Name)
	if err == nil {
		return nil
	}

	pgErr, ok := err.(*pq.Error)
	if !ok {
		return err
	}
	if pgErr.Constraint == "ipps_user_email_key" && pgErr.Code.Name() == "unique_violation" {
		return user.ErrEmailExists
	} else if pgErr.Constraint == "ipps_user_username_key" &&
		pgErr.Code.Name() == "unique_violation" {
		return user.ErrUserExists
	}

	return err
}

func (us *UserStorage) ByID(id uuid.UUID) (*user.User, error) {
	return userFromRow(us.byID.QueryRow(id))
}

func (us *UserStorage) ByEmail(address *mail.Address) (*user.User, error) {
	return userFromRow(us.byEmail.QueryRow(address.Address))
}

func (us *UserStorage) ByUsername(username string) (*user.User, error) {
	return userFromRow(us.byUsername.QueryRow(username))
}

func (us *UserStorage) Update(user *user.User) error {
	_, err := us.update.Exec(user.ID, user.Email.Address, user.Password, user.Name)
	return err
}

func (us *UserStorage) Delete(user *user.User) error {
	_, err := us.update.Exec(user.ID)
	return err
}

// Close closes the us's underlying database connection.
func (us *UserStorage) Close() error {
	err := us.insert.Close()
	if err != nil {
		return err
	}
	err = us.byID.Close()
	if err != nil {
		return err
	}
	err = us.byEmail.Close()
	if err != nil {
		return err
	}
	err = us.update.Close()
	if err != nil {
		return err
	}
	err = us.delete.Close()
	if err != nil {
		return err
	}

	return nil
}

func userFromRow(row *sql.Row) (*user.User, error) {
	u := &user.User{}
	var email string
	err := row.Scan(&u.ID, &u.Username, &email, &u.Password, &u.Name)
	if err == sql.ErrNoRows {
		return nil, user.ErrUserNotExists
	} else if err != nil {
		return nil, err
	}
	u.Email, err = mail.ParseAddress(email)
	if err != nil {
		return nil, err
	}

	return u, nil
}
