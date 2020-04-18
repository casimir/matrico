package generate

import (
	"fmt"
	"strings"
)

// APIHandler contains the data to generate the code of an handler.
type APIHandler struct {
	method   string
	Auth     bool
	Doc      string
	path     string
	Func     string
	ImplFunc string
	Body     *Schema
	Params   []string
	Query    []string
	Response Schema
}

func (h *APIHandler) ImplFuncArgs() string {
	all := []string{"r.Context()"}
	all = append(all, h.Params...)
	if h.Body != nil {
		all = append(all, "body")
	}
	if len(h.Query) > 0 {
		all = append(all, "r.URL.Query()")
	}
	return strings.Join(all, ", ")
}

func (h *APIHandler) RouteSource(routerName string) string {
	return fmt.Sprintf(
		"%s.%s(\"%s\", %s)",
		routerName, strings.Title(h.method), h.path, h.Func,
	)
}
