package postgres

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/feedback"
)

const (
	installFeedbackTable = `CREATE TABLE IF NOT EXISTS ipps_feedback(
		id              uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
		author          varchar(128) NOT NULL CONSTRAINT ipps_feedback_author_fkey
							REFERENCES ipps_user (username) ON DELETE CASCADE ON UPDATE CASCADE,
		rating          integer      NOT NULL CHECK (rating > 0 and rating <= 5),
		feedback        text         NOT NULL,
		date_posted     timestamptz  NOT NULL
	);`
	multipleFeedbackStmt = `SELECT id, author, rating, feedback, date_posted
							 FROM ipps_feedback
							 WHERE date_posted >= NOW() - INTERVAL '1 hour'
							 ORDER BY date_posted DESC
							 OFFSET $1 ROWS
							 FETCH FIRST $2 ROWS ONLY;`
	recentFeedbackStmt = `SELECT id, author, rating, feedback, date_posted
						   FROM ipps_feedback
						   WHERE date_posted >= NOW() - INTERVAL '1 hour'
						   ORDER BY date_posted DESC;`
	insertFeedbackStmt = `INSERT INTO ipps_feedback (id, author, rating, feedback, date_posted)
						  VALUES ($1, $2, $3, $4, $5);`
)

// FeedbackStorage is the postgres implementation of the feedback.Storage interface.
type FeedbackStorage struct {
	multiple *sql.Stmt
	recent   *sql.Stmt
	insert   *sql.Stmt
}

func NewFeedbackStorage(db *sql.DB) (*FeedbackStorage, error) {
	ms, err := db.Prepare(multipleFeedbackStmt)
	if err != nil {
		return nil, err
	}
	rs, err := db.Prepare(recentFeedbackStmt)
	if err != nil {
		return nil, err
	}
	is, err := db.Prepare(insertFeedbackStmt)
	if err != nil {
		return nil, err
	}

	return &FeedbackStorage{
		multiple: ms,
		recent:   rs,
		insert:   is,
	}, nil
}

func (fs *FeedbackStorage) Multiple(n, offset uint) ([]feedback.Feedback, error) {
	rows, err := fs.multiple.Query(offset, n)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	ff := make([]feedback.Feedback, n)
	i := 0
	for rows.Next() {
		f := &ff[i]
		err := rows.Scan(&f.ID, &f.Author, &f.Rating, &f.Text, &f.Date)
		if err != nil {
			return nil, err
		}
		i++
	}
	ff = ff[:i]

	return ff, nil
}

func (fs *FeedbackStorage) Recent() ([]feedback.Feedback, error) {
	rows, err := fs.recent.Query()
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	ff := make([]feedback.Feedback, 0, 10)
	for rows.Next() {
		var id uuid.UUID
		var author string
		var rating uint8
		var text string
		var datePosted time.Time
		err := rows.Scan(&id, &author, &rating, &text, &datePosted)
		if err != nil {
			return nil, err
		}

		ff = append(ff, feedback.Feedback{
			ID:     id,
			Author: author,
			Rating: rating,
			Text:   text,
			Date:   datePosted,
		})
	}

	return ff, nil
}

func (fs *FeedbackStorage) Insert(f *feedback.Feedback) error {
	_, err := fs.insert.Exec(&f.ID, &f.Author, &f.Rating, &f.Text, &f.Date)
	return err
}

func (fs *FeedbackStorage) Close() error {
	err := fs.insert.Close()
	if err != nil {
		return err
	}
	err = fs.recent.Close()
	if err != nil {
		return err
	}

	return fs.multiple.Close()
}
