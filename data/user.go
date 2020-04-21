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
	if err := d.LinkNodes(n, device, "USES"); err != nil {
		return err
	}
	return nil
}

func (n *User) MarkOnline(d *DataGraph) error {
	props := map[string]interface{}{
		"presence":      "online",
		"lastActiveAgo": NowMs(),
	}
	return d.NodeSet(n, props)
}
