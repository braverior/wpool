package wpool


type Job interface {
	Do() error
	ID() string
}

