package clientserverr060

import (
	"context"
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/casimir/matrico/api/common"
)

func sync(ctx context.Context, query url.Values) (SyncResponse, error) {
	timeout := query.Get("timeout")
	if timeout == "" {
		// TODO init mode
		timeout = "0"
	}
	timeoutMs, err := strconv.Atoi(timeout)
	if err != nil {
		return SyncResponse{}, common.New("invalid timeout")
	}

	resp := SyncResponse{}
	d := common.Data(ctx)
	next := d.ListenNextEvent()
	select {
	case ev := <-next:
		log.Printf("got a new event: %s", ev)
	case <-time.After(time.Duration(timeoutMs) * time.Millisecond):
		d.CancelEventListener(next)
		resp.NextBatch = strconv.FormatInt(time.Now().Unix(), 10)
	}
	return resp, nil
}
