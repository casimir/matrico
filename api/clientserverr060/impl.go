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
	resp := SyncResponse{
		NextBatch: "next123",
	}
	return resp, nil
}

func getVersions(ctx context.Context) (GetVersionsResponse, error) {
	return GetVersionsResponse{Versions: []string{"r0.6.0"}}, nil
}
