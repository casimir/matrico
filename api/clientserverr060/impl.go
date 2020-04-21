package clientserverr060

import (
	"context"
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/casimir/matrico/api/common"
	"github.com/casimir/matrico/data"
)

func defineFilter(ctx context.Context, userID string, body DefineFilterBody) (DefineFilterResponse, error) {
	// TODO placeholder
	return DefineFilterResponse{}, nil
}

func getPushRules(ctx context.Context) (GetPushRulesResponse, error) {
	// TODO placeholder
	return GetPushRulesResponse{}, nil
}

func setPresence(ctx context.Context, userID string, body SetPresenceBody) (SetPresenceResponse, error) {
	username := ctx.Value(common.CtxUserKey).(string)
	if userID != username {
		log.Printf("%q != %q", userID, username)
		return SetPresenceResponse{}, common.ErrForbidden
	}
	d := common.Data(ctx)
	props := map[string]interface{}{
		"presence": body.Presence,
	}
	if body.StatusMsg != nil {
		props["statusMessage"] = body.StatusMsg
	}
	if body.Presence == "online" {
		props["lastActiveAgo"] = int(data.NowMs())
	}
	err := d.NodeSet(data.NewUser(username), props)
	// TODO create event
	return SetPresenceResponse{}, err
}

func getPresence(ctx context.Context, userID string) (GetPresenceResponse, error) {
	// TODO 403
	// Presence information is shared with all users who share a room with the target user. In large public rooms this could be undesirable.
	d := common.Data(ctx)
	user := data.NewUser(userID)
	props, ok, err := d.NodeGet(user)
	if err != nil {
		panic(err)
	}
	if !ok {
		return GetPresenceResponse{}, common.ErrNotFound
	}
	presence, _ := props["presence"]
	resp := GetPresenceResponse{
		Presence: presence.(string),
	}
	if v, ok := props["lastActiveAgo"]; ok {
		ago := data.NowMs() - int64(v.(int))
		threshold := 5 * time.Minute
		tooLong := time.Duration(ago)*time.Millisecond > threshold
		if resp.Presence == "online" && tooLong {
			resp.Presence = "unavailable"
			props := map[string]interface{}{"presence": resp.Presence}
			if err := d.NodeSet(user, props); err != nil {
				panic(err)
			}
		}
		val := int(ago)
		resp.LastActiveAgo = &val
	}
	if v, ok := props["displayMessage"]; ok {
		msg := v.(string)
		resp.StatusMsg = &msg
	}
	return resp, nil
}

func sync(ctx context.Context, query url.Values) (SyncResponse, error) {
	timeout := query.Get("timeout")
	if timeout == "" {
		timeout = "0"
	}
	timeoutMs, err := strconv.Atoi(timeout)
	if err != nil {
		return SyncResponse{}, common.New("invalid timeout")
	}

	done := make(chan bool, 1)
	resp := SyncResponse{}
	go func() {
		time.Sleep(1000 * time.Second)
		done <- true
	}()
	select {
	case <-done:
		log.Print("got a new event")
	case <-time.After(time.Duration(timeoutMs) * time.Millisecond):
		resp.NextBatch = strconv.FormatInt(time.Now().Unix(), 10)
	}
	return resp, nil
}

func getVersions(ctx context.Context) (GetVersionsResponse, error) {
	return GetVersionsResponse{Versions: []string{"r0.6.0"}}, nil
}
