package clientserverr060

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/casimir/matrico/api/common"
	"github.com/casimir/matrico/data"
	"github.com/google/uuid"
	rg "github.com/redislabs/redisgraph-go"
	"golang.org/x/crypto/bcrypt"
)

var FlowPassword = "m.login.password"

func getLoginFlows(ctx context.Context) (GetLoginFlowsResponse, error) {
	resp := GetLoginFlowsResponse{}
	resp.Flows = append(resp.Flows, GetLoginFlowsResponseFlows{&FlowPassword})
	return resp, nil
}

func makeUsername(name string) string {
	return "@" + strings.ToLower(name) + ":homeserver.local"
}

func makeDeviceID(initial *string, username string) string {
	if initial != nil && *initial != "" {
		return *initial
	}
	return fmt.Sprintf("%s_%d", username, time.Now().Unix())
}

func register(ctx context.Context, body RegisterBody, query url.Values) (RegisterResponse, error) {
	if query.Get("kind") == "guest" {
		return RegisterResponse{}, common.ErrUnknown
	}

	if body.Username == nil && body.Password == nil {
		log.Print("? empty register request")
		return RegisterResponse{}, nil
	}

	username := makeUsername(*body.Username)
	if username == "" {
		log.Printf("invalid username: %v", body.Username)
		return RegisterResponse{}, common.New("invalid username")
	}

	d := common.Data(ctx)
	exists, err := d.NodeExists(data.NewUser(username))
	if err != nil {
		log.Print(err)
		return RegisterResponse{}, common.ErrUnknown
	}
	if exists {
		log.Print("user already exists")
		// TODO specific error
		return RegisterResponse{}, common.ErrUnknown
	}

	if body.Password == nil {
		return RegisterResponse{}, common.New("invalid password")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(*body.Password), -1)
	if err != nil {
		panic(err)
	}

	deviceID := makeDeviceID(body.DeviceID, *body.Username)
	token := uuid.New().String()
	displayName := ""
	if body.InitialDeviceDisplayName != nil {
		displayName = *body.InitialDeviceDisplayName
	}

	user := rg.Node{
		Label: "User",
		Properties: map[string]interface{}{
			"username": username,
			"password": string(hashed),
		},
	}
	device := rg.Node{
		Label: "Device",
		Properties: map[string]interface{}{
			"deviceId":    deviceID,
			"displayName": displayName,
			"token":       token,
		},
	}
	connectedWith := rg.Edge{
		Source:      &user,
		Relation:    "USES",
		Destination: &device,
	}
	graph := d.DELETEME()
	defer graph.Conn.Close()
	graph.AddNode(&user)
	graph.AddNode(&device)
	graph.AddEdge(&connectedWith)
	if _, err := graph.Commit(); err != nil {
		panic(err)
	}

	resp := RegisterResponse{
		UserID: username,
	}
	if body.InhibitLogin == nil || !*body.InhibitLogin {
		resp.AccessToken = &token
		resp.DeviceID = &deviceID
	}

	return resp, nil
}

func login(ctx context.Context, body LoginBody) (LoginResponse, error) {
	if body.Type != FlowPassword {
		return LoginResponse{}, common.ErrUnknown
	}
	typ := body.Identifier["type"].(string)
	if typ != "m.id.user" {
		return LoginResponse{}, common.ErrUnknown
	}
	username := makeUsername(body.Identifier["user"].(string))

	d := common.Data(ctx)
	user := data.NewUser(username)
	ok, err := user.CheckPassword(d, *body.Password)
	if err != nil || !ok {
		return LoginResponse{}, common.ErrForbidden
	}
	device := data.NewDevice(makeDeviceID(body.DeviceID, body.Identifier["user"].(string)))
	token := uuid.New().String()
	if err := user.ActivateDevice(d, device, token, *body.InitialDeviceDisplayName); err != nil {
		log.Print(err)
		return LoginResponse{}, common.ErrUnknown
	}

	resp := LoginResponse{
		UserID:      &username,
		AccessToken: &token,
		DeviceID:    &device.DeviceID,
	}
	return resp, nil
}

func logout(ctx context.Context) (LogoutResponse, error) {
	d := common.Data(ctx)
	token := ctx.Value(common.CtxTokenKey).(string)
	device, err := data.NewDeviceFromToken(d, token)
	if err != nil {
		log.Print(err)
		return LogoutResponse{}, common.ErrUnknown
	}
	if _, err := d.NodeDelete(device); err != nil {
		log.Print(err)
		return LogoutResponse{}, common.ErrUnknown
	}
	return LogoutResponse{}, nil
}

func getTokenOwner(ctx context.Context) (GetTokenOwnerResponse, error) {
	user := ctx.Value(common.CtxUserKey).(string)
	return GetTokenOwnerResponse{user}, nil
}
