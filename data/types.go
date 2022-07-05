package data

// EventLevel represents the alarm's event level
type EventLevel string

const (
	// NoEvent specify that no event was triggered
	NoEvent EventLevel = "No event"
	// Info specify that an info event was triggered
	Info EventLevel = "Info"
	// Error specify that an error event was triggered
	Error EventLevel = "Error"
)
