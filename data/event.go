package data

import (
	"encoding/json"
	"fmt"
	"strings"
)

type EventType string

const (
	EvPresence EventType = "m.presence"
)

type Event struct {
	ID         string
	Properties map[string]interface{}
}

func (n *Event) Label() string                 { return "Event" }
func (n *Event) Key() string                   { return "event_id" }
func (n *Event) KeyVal() interface{}           { return n.ID }
func (n *Event) Props() map[string]interface{} { return n.Properties }

func NewEvent(typ EventType) *Event {
	return &Event{
		Properties: map[string]interface{}{
			"type": string(typ),
		},
	}
}

func NewEventWithID(id string, typ EventType) *Event {
	e := NewEvent(typ)
	e.ID = id
	return e
}

func (n *Event) LinkTo(d *DataGraph, m Noder) error {
	n.Properties["sender"] = m.KeyVal()
	if _, ok := n.Properties["when"]; !ok {
		n.Properties["when"] = NowMs()
	}
	if n.ID == "" {
		n.ID = fmt.Sprintf("%s_%d", m.KeyVal(), n.Properties["when"])
	}
	if err := d.NodeCreate(n); err != nil {
		return fmt.Errorf("link event to %q: %v", m.KeyVal(), err)
	}
	return d.LinkNodes(n, m, "SENT_BY")
}

func (n *Event) Type() EventType {
	return EventType(n.Properties["type"].(string))
}

func (n Event) MarshalJSON() ([]byte, error) {
	data := make(map[string]interface{})
	content := make(map[string]interface{})
	for k, v := range n.Properties {
		if strings.HasPrefix(k, "content_") {
			content[k[8:]] = v
		} else {
			data[k] = v
		}
	}
	data["content"] = content
	return json.Marshal(data)
}
