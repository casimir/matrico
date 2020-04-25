package clientserverr060

import (
	"context"
	"log"
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
	presence, ok := data.ToPresence(body.Presence)
	if !ok {
		return SetPresenceResponse{}, common.New("invalid presence")
	}
	user := data.NewUser(username)
	if err := user.MarkAs(d, presence, body.StatusMsg); err != nil {
		return SetPresenceResponse{}, err
	}
	return SetPresenceResponse{}, nil
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
	p, _ := props["presence"]
	presence, _ := data.ToPresence(p.(string))
	resp := GetPresenceResponse{
		Presence: presence.String(),
	}
	if v, ok := props["lastActiveAgo"]; ok {
		ago := data.NowMs() - int64(v.(int))
		threshold := 5 * time.Minute
		idle := time.Duration(ago)*time.Millisecond > threshold
		if presence == data.Online && idle {
			resp.Presence = data.Unavailable.String()
			if err := user.MarkAs(d, data.Unavailable, nil); err != nil {
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

func getVersions(ctx context.Context) (GetVersionsResponse, error) {
	return GetVersionsResponse{Versions: []string{"r0.6.0"}}, nil
}
