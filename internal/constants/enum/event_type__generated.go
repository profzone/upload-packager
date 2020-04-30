package enum

import (
	"bytes"
	"encoding"
	"errors"

	github_com_profzone_eden_framework_pkg_enumeration "github.com/profzone/eden-framework/pkg/enumeration"
)

var InvalidEventType = errors.New("invalid EventType")

func init() {
	github_com_profzone_eden_framework_pkg_enumeration.RegisterEnums("EventType", map[string]string{
		"MESSAAGE":  "message",
		"PARTITION": "partition",
	})
}

func ParseEventTypeFromString(s string) (EventType, error) {
	switch s {
	case "":
		return EVENT_TYPE_UNKNOWN, nil
	case "MESSAAGE":
		return EVENT_TYPE__MESSAAGE, nil
	case "PARTITION":
		return EVENT_TYPE__PARTITION, nil
	}
	return EVENT_TYPE_UNKNOWN, InvalidEventType
}

func ParseEventTypeFromLabelString(s string) (EventType, error) {
	switch s {
	case "":
		return EVENT_TYPE_UNKNOWN, nil
	case "message":
		return EVENT_TYPE__MESSAAGE, nil
	case "partition":
		return EVENT_TYPE__PARTITION, nil
	}
	return EVENT_TYPE_UNKNOWN, InvalidEventType
}

func (EventType) EnumType() string {
	return "EventType"
}

func (EventType) Enums() map[int][]string {
	return map[int][]string{
		int(EVENT_TYPE__MESSAAGE):  {"MESSAAGE", "message"},
		int(EVENT_TYPE__PARTITION): {"PARTITION", "partition"},
	}
}

func (v EventType) String() string {
	switch v {
	case EVENT_TYPE_UNKNOWN:
		return ""
	case EVENT_TYPE__MESSAAGE:
		return "MESSAAGE"
	case EVENT_TYPE__PARTITION:
		return "PARTITION"
	}
	return "UNKNOWN"
}

func (v EventType) Label() string {
	switch v {
	case EVENT_TYPE_UNKNOWN:
		return ""
	case EVENT_TYPE__MESSAAGE:
		return "message"
	case EVENT_TYPE__PARTITION:
		return "partition"
	}
	return "UNKNOWN"
}

var _ interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
} = (*EventType)(nil)

func (v EventType) MarshalText() ([]byte, error) {
	str := v.String()
	if str == "UNKNOWN" {
		return nil, InvalidEventType
	}
	return []byte(str), nil
}

func (v *EventType) UnmarshalText(data []byte) (err error) {
	*v, err = ParseEventTypeFromString(string(bytes.ToUpper(data)))
	return
}
