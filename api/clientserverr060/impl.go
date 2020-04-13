package clientserverr060

import (
	"net/url"

	"github.com/casimir/matrico/api/common"
)

const (
	FlowPassword = "m.login.password"
)

func defineFilter(userID string, body DefineFilterBody) (DefineFilterResponse, error) {
	// TODO placeholder
	return DefineFilterResponse{}, nil
}

func getLoginFlows() (GetLoginFlowsResponse, error) {
	resp := GetLoginFlowsResponse{}
	resp.Flows = append(resp.Flows, GetLoginFlowsResponseFlows{FlowPassword})
	return resp, nil
}

func login(body LoginBody) (LoginResponse, error) {
	if body.Type != FlowPassword {
		return LoginResponse{}, common.ErrUnknown
	}
	typ := body.Identifier["type"].(string)
	if typ != "m.id.user" {
		return LoginResponse{}, common.ErrUnknown
	}
	user := body.Identifier["user"].(string)
	deviceID := body.DeviceID
	if deviceID == "" {
		deviceID = "device_" + user
	}
	resp := LoginResponse{
		UserID:      "@" + user + ":server.tld",
		AccessToken: "token_" + user,
		DeviceID:    deviceID,
	}
	return resp, nil
}

func logout() (LogoutResponse, error) {
	// TODO placeholder
	return LogoutResponse{}, nil
}

func getPushRules() (GetPushRulesResponse, error) {
	// TODO placeholder
	return GetPushRulesResponse{}, nil
}

func sync(query url.Values) (SyncResponse, error) {
	resp := SyncResponse{
		NextBatch: "next123",
	}
	return resp, nil
}

func getVersions() (GetVersionsResponse, error) {
	return GetVersionsResponse{Versions: []string{"r0.6.0"}}, nil
}
