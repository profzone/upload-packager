package enum

//go:generate eden generate enum --type-name=EventType
// api:enum
type EventType uint8

// event type
const (
	EVENT_TYPE_UNKNOWN    EventType = iota
	EVENT_TYPE__PARTITION           // partition
	EVENT_TYPE__MESSAAGE            // message
)
