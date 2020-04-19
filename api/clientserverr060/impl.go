package clientserverr060

import (
	"context"
	"net/url"
)

func defineFilter(ctx context.Context, userID string, body DefineFilterBody) (DefineFilterResponse, error) {
	// TODO placeholder
	return DefineFilterResponse{}, nil
}

func getPushRules(ctx context.Context) (GetPushRulesResponse, error) {
	// TODO placeholder
	return GetPushRulesResponse{}, nil
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
