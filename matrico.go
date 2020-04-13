package matrico

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/casimir/matrico/api"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Server struct {
	*chi.Mux
}

func NewServer() Server {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	api.Register(r)

	return Server{r}
}

func (s *Server) ListRoutes() []string {
	var routes []string
	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		route = strings.Replace(route, "/*/", "/", -1)
		routes = append(routes, fmt.Sprintf("%s %s", method, route))
		return nil
	}
	chi.Walk(s, walkFunc)
	return routes
}
