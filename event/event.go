package event

// AppEvent is the base interface for all application events.
type AppEvent interface {
	// Order returns the order of the event.
	// Lower numbers indicate higher priority.
	Order() int
}
