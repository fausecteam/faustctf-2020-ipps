// Package feedback defines interfaces and data
// structures for handling customer feedback.
package feedback

import (
	"errors"
	"html/template"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmptyFeedback = errors.New("feedback text is empty")
)

// Feedback is the representation of a customer's feedback message.
type Feedback struct {
	ID     uuid.UUID `json:"id"`
	Author string    `json:"author"`
	Rating uint8     `json:"rating"`
	Text   string    `json:"text"`
	Date   time.Time `json:"datePosted"`
}

func New(author string, rating uint8, text string) (*Feedback, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	if text == "" {
		return nil, ErrEmptyFeedback
	}

	return &Feedback{
		ID:     id,
		Author: author,
		Rating: rating,
		Text:   text,
		Date:   time.Now().Local(),
	}, nil
}

func (f *Feedback) Stars() template.HTML {
	var b strings.Builder
	for i := uint8(0); i < f.Rating; i++ {
		b.WriteString(`<span class="material-icons rating-star">stare_rate</span>`)
	}

	return template.HTML(b.String())
}
