package common

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/casimir/matrico/data"
)

type CtxKey string

const (
	ctxDataKey  = CtxKey("data")
	CtxTokenKey = CtxKey("token")
	CtxUserKey  = CtxKey("user")
)

func Data(ctx context.Context) *data.DataGraph {
	return ctx.Value(ctxDataKey).(*data.DataGraph)
}

func ContextMiddleware(d *data.DataGraph) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), ctxDataKey, d)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
		return http.HandlerFunc(fn)
	}
}

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
			d := Data(r.Context())
			user, err := data.NewUserFromToken(d, token)
			if err != nil {
				log.Print(err)
				http.Error(w, ErrUnknown.Error(), ErrUnknown.Status)
			} else if user == nil {
				http.Error(w, ErrUnknownToken.Error(), ErrUnknownToken.Status)
			} else {
				ctx := r.Context()
				ctx = context.WithValue(ctx, CtxTokenKey, token)
				ctx = context.WithValue(ctx, CtxUserKey, user.Username)
				next.ServeHTTP(w, r.WithContext(ctx))
			}

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
		log.Print(err)
		if e, ok := err.(Error); ok {
			payload, em := json.Marshal(err)
			if em != nil {
				panic(em)
			}
			w.WriteHeader(e.Status)
			fmt.Fprintln(w, string(payload))
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
