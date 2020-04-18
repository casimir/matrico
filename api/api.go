package api

import (
	"net/http"

	"github.com/casimir/matrico/api/clientserverr060"
	"github.com/casimir/matrico/api/common"
	"github.com/casimir/matrico/data"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
)

func addHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func InitAndRegister(r chi.Router) {
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "X-Requested-With", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}).Handler

	r.Route("/", func(r chi.Router) {
		r.Use(corsHandler)
		r.Use(addHeaders)
		r.Use(common.ContextMiddleware(data.New("matrico")))
		clientserverr060.RegisterAPI(r)
	})
}
