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

func appendEvent(r *SyncResponse, ev data.Event) {
	switch ev.Type() {
	case data.EvPresence:
		if r.Presence == nil {
			r.Presence = &SyncResponsePresence{[]Events{}}
		}
		r.Presence.Events = append(r.Presence.Events, Events(ev))
	default:
		log.Printf("unknown event type: %s", ev.Type())
	}
}

func sync(ctx context.Context, query url.Values) (SyncResponse, error) {
	// TODO handle `filter` param
	timeout := query.Get("timeout")
	if timeout == "" {
		// TODO init mode
		timeout = "0"
	}
	timeoutMs, err := strconv.Atoi(timeout)
	if err != nil {
		return SyncResponse{}, common.New("invalid timeout")
	}

	d := common.Data(ctx)
	username := ctx.Value(common.CtxUserKey).(string)
	user := data.NewUser(username)
	switch query.Get("set_presence") {
	case "online", "":
		if err := user.MarkAs(d, data.Online, nil); err != nil {
			log.Print(err)
			return SyncResponse{}, common.ErrUnknown
		}
	case "unavailable":
		if err := user.MarkAs(d, data.Unavailable, nil); err != nil {
			log.Print(err)
			return SyncResponse{}, common.ErrUnknown
		}
	}

	// sinceP := query.Get("since")
	// since, err := strconv.Atoi(sinceP)
	// if err != nil {
	// 	return SyncResponse{}, common.ErrUnknown
	// }
	// fullState := query.Get("full_state") == "true"

	resp := SyncResponse{}
	next := d.ListenNextEvent(func(ev *data.Event) bool {
		// skip presence events from current user
        // TODO only users sharing rooms
		return !(ev.Type() == data.EvPresence && ev.Properties["sender"] == username)
	})
	select {
	case ev := <-next:
		log.Printf("got a new event: %s", ev)
		appendEvent(&resp, ev)
	case <-time.After(time.Duration(timeoutMs) * time.Millisecond):
		d.CancelEventListener(next)
		resp.NextBatch = strconv.FormatInt(time.Now().Unix(), 10)
	}
	return resp, nil
}
