package data

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username   string
	Properties map[string]interface{}
}

func (n *User) Label() string                 { return "User" }
func (n *User) Key() string                   { return "username" }
func (n *User) KeyVal() interface{}           { return n.Username }
func (n *User) Props() map[string]interface{} { return n.Properties }

func NewUser(username string) *User {
	return &User{Username: username}
}

func NewUserFromToken(d *DataGraph, token string) (*User, error) {
	q := fmt.Sprintf(
		`MATCH (u:User)-[:USES]->(d:Device) WHERE d.token = '%s' RETURN u.username`,
		token,
	)
	res, err := d.Query(q)
	defer res.Close()
	if err != nil {
		return nil, err
	}
	if !res.Next() {
		return nil, nil
	}

	r := res.Record()
	username, _ := r.Get("u.username")
	return &User{Username: username.(string)}, nil
}

func (n *User) CheckPassword(d *DataGraph, password string) (bool, error) {
	q := fmt.Sprintf(
		`MATCH (u:User) WHERE u.username = '%s' return u.password`,
		n.Username,
	)
	res, err := d.Query(q)
	defer res.Close()
	if err != nil {
		return false, err
	}
	if !res.Next() {
		return false, err
	}

	hashed := res.Record().GetByIndex(0).(string)
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)); err != nil {
		return false, err
	}
	return true, nil
}

func (n *User) ActivateDevice(d *DataGraph, device *Device, token, displayName string) error {
	res, err := d.Query(fmt.Sprintf(
		`MERGE (d:Device {deviceId: '%s'})
		SET d.token = '%s', d.displayName = COALESCE(d.displayName, '%s')`,
		device.DeviceID, token, displayName,
	))
	defer res.Close()
	if err != nil {
		return err
	}
	return d.LinkNodes(n, device, "USES")
}

func (n *User) SetActiveNow(d *DataGraph) error {
	props := map[string]interface{}{
		"lastActiveAgo": NowMs(),
	}
	if err := d.NodeSet(n, props); err != nil {
		return fmt.Errorf("set %q active: %w", n.Username, err)
	}
	return nil
}

func (n *User) GetPresence(d *DataGraph) (string, error) {
	u, ok, err := d.NodeGet(n)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", nil
	}
	if v, ok := u["presence"]; ok {
		return v.(string), nil
	}
	return "", nil
}

type Presence string

const (
	Online      Presence = "online"
	Offline     Presence = "offline"
	Unavailable Presence = "unavailable"
)

func (p Presence) String() string {
	return string(p)
}

func ToPresence(s string) (Presence, bool) {
	switch Presence(s) {
	case Online:
		return Online, true
	case Offline:
		return Offline, true
	case Unavailable:
		return Unavailable, true
	default:
		return "", false
	}
}

func (n *User) MarkAs(d *DataGraph, presence Presence, status *string) error {
	if presence == Online {
		if err := n.SetActiveNow(d); err != nil {
			return err
		}
	}
	oldPresence, err := n.GetPresence(d)
	if err != nil {
		return fmt.Errorf("fetch %q presence: %v", n.Username, err)
	}
	if oldPresence == string(presence) {
		return nil
	}
	props := map[string]interface{}{
		"presence": presence.String(),
	}
	if status != nil {
		props["statusMessage"] = *status
	}
	if err := d.NodeSet(n, props); err != nil {
		return fmt.Errorf("update %q: %v", n.Username, err)
	}
	u, _, err := d.NodeGet(n)
	if err != nil {
		return fmt.Errorf("refresh %q data: %v", n.Username, err)
	}
	event := NewEvent(EvPresence)
	// TODO content["content_avatar_url"]
	// TODO content["content_currently_active"]
	if v, ok := u["lastActiveAgo"]; ok {
		ago := NowMs() - int64(v.(int))
		event.Properties["content_last_active_ago"] = ago
	}
	if v, ok := u["presence"]; ok {
		event.Properties["content_presence"] = v.(string)
	}
	if v, ok := u["statusMessage"]; ok && v.(string) != "" {
		event.Properties["content_status_msg"] = v.(string)
	}
	if err := event.LinkTo(d, n); err != nil {
		return err
	}
	d.BroadcastEvent(event)
	return nil
}
