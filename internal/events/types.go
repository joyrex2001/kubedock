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
	// Image defines the event/filter type image
	Image = "image"
	// Container defines the event/filter type container
	Container = "container"
	// Type defines the filter type Type
	Type = "type"
	// Create defines the event action create (container)
	Create = "create"
	// Start defines the event action start (container)
	Start = "start"
	// Die defines the event action die (container)
	Die = "die"
	// Detach defines the event action detach (container)
	Detach = "detach"
	// Pull defines the event action image (container)
	Pull = "pull"
)
