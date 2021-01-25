package feedback

type Accesser interface {
	// Recent returns all feedback from the last 24 hours.
	Recent() ([]Feedback, error)
	// Multiple returns up to n feedback recent posts, skipping
	// offset posts.
	Multiple(n, offset uint) ([]Feedback, error)
}

type Inserter interface {
	Insert(feedback *Feedback) error
}

type Storage interface {
	Accesser
	Inserter
}
