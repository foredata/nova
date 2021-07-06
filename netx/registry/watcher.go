package registry

// Watcher is an interface that returns updates
// about services within the registry.
type Watcher interface {
	// Next is a blocking call
	Next() (*Event, error)
	Stop()
}

// Event is registry event
type Event struct {
	Id      string    // delete时只需使用id
	Type    EventType //
	Service *Service
}

type EventType uint8

const (
	EventCreate EventType = iota
	EventDelete
	EventUpdate
)

func (t EventType) String() string {
	switch t {
	case EventCreate:
		return "create"
	case EventDelete:
		return "delete"
	case EventUpdate:
		return "update"
	default:
		return "unknown"
	}
}
