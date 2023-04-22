package events

// Message is the structure that defines the details of the event.
type Message struct {
	ID       string
	Type     string
	Action   string
	Time     int64
	TimeNano int64
}

const (
	// Image defines the event type image
	Image = "image"
	// Container defines the event type container
	Container = "container"
	// Create defines the event action create (container)
	Create = "create"
	// Start defines the event action start (container)
	Start = "start"
	// Die defines the event action die (container)
	Die = "die"
	// Pull defines the event action image (container)
	Pull = "pull"
)
