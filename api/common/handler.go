package common

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := ""

		if header := r.Header.Get("Authorization"); header != "" {
			if strings.HasPrefix(header, "Bearer ") {
				token = strings.Fields(header)[1]
			}
		} else {
			param, ok := r.URL.Query()["access_token"]
			if ok {
				token = param[0]
			}
		}

		if token == "" {
			http.Error(w, ErrMissingToken.Error(), ErrMissingToken.Status)
		} else {
			ctx := context.WithValue(r.Context(), "authToken", token)
			// TODO add user in context
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func UnmarshalBody(r *http.Request, v interface{}) error {
	err := json.NewDecoder(r.Body).Decode(&v)
	if err != nil {
		return ErrBadJSON
	}
	return nil
}

func ResponseHandler(w http.ResponseWriter, data interface{}, err error) {
	if err != nil {
		if e, ok := err.(Error); ok {
			http.Error(w, err.Error(), e.Status)
		} else {
			panic(e)
		}
	} else {
		raw, err := json.Marshal(data)
		if err != nil {
			panic(err)
		}
		w.Write(raw)
	}
}
