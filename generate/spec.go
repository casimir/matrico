package generate

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
)

func toID(name string, exportable bool) string {
	var ID strings.Builder
	upper := exportable
	for i, c := range name {
		if upper {
			ID.WriteString(strings.ToUpper(string(c)))
			upper = false
			continue
		} else if i == 0 {
			ID.WriteString(strings.ToLower(string(c)))
			continue
		}
		if strings.ContainsRune("_.:", c) {
			upper = true
			continue
		}
		ID.WriteRune(c)

	}

	s := ID.String()
	if strings.HasSuffix(s, "Id") {
		return s[:len(s)-1] + "D"
	}
	return s
}

func toComment(s string) string {
	return strings.ReplaceAll("// "+s, "\n", "\n// ")
}

type GoAttr struct {
	ID   string
	Type string
	Doc  string
	Tag  string
	Opt  bool
}

func NewGoAttribute(name, typ, doc string, opt bool) GoAttr {
	jopts := name
	if opt {
		jopts += ",omitempty"
	}
	return GoAttr{
		ID:   toID(name, true),
		Type: typ,
		Doc:  toComment(doc),
		Tag:  `json:"` + jopts + `"`,
		Opt:  opt,
	}
}

func (ga *GoAttr) Decl() string {
	s := strings.Join([]string{ga.ID, ga.Type}, " ")
	if ga.Tag != "" {
		s += " `" + ga.Tag + "`"
	}
	return s
}

type Parameter struct {
	In     string
	Name   string
	Schema Schema
}

type ResponseSchema struct {
	Description string
	Schema      Schema
}

type Method struct {
	Summary     string
	Description string
	OperationID string `yaml:"operationId"`
	Security    interface{}
	Parameters  []Parameter
	Responses   map[int]ResponseSchema
}

type Spec struct {
	BasePath string `yaml:"basePath"`
	Paths    map[string]map[string]Method
}

func (s *Spec) extractHandlers(version string, defs map[string]*Schema, skips map[string]bool) []APIHandler {
	var handlers []APIHandler
	for path, methods := range s.Paths {
		for methodName, method := range methods {
			if _, ok := skips[method.OperationID]; ok {
				continue
			}
			log.Printf("> %s", method.OperationID)
			handlerName := toID(method.OperationID, true)
			h := APIHandler{
				method:   methodName,
				Auth:     method.Security != nil,
				Doc:      toComment(method.Description),
				path:     strings.ReplaceAll(s.BasePath+path, "%CLIENT_MAJOR_VERSION%", version),
				Func:     handlerName,
				ImplFunc: toID(handlerName, false),
			}
			for _, it := range method.Parameters {
				if it.In == "body" {
					h.Body = &it.Schema
					h.Body.Identifier = handlerName + "Body"
					h.Body.FollowRef(defs)
				} else if it.In == "path" {
					h.Params = append(h.Params, it.Name)
				} else if it.In == "query" {
					h.Query = append(h.Query, it.Name)
				}
			}
			for status, schema := range method.Responses {
				if status < 300 {
					h.Response = schema.Schema
					h.Response.Identifier = handlerName + "Response"
					h.Response.FollowRef(defs)
					break
				}
			}
			handlers = append(handlers, h)
		}
	}
	return handlers
}

func parseSpecFile(path string) (Spec, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return Spec{}, err
	}
	var spec Spec
	if err := yaml.Unmarshal(raw, &spec); err != nil {
		return spec, err
	}
	return spec, nil
}

func parseDefinitionFiles(root string) (map[string]*Schema, error) {
	basePath := filepath.Dir(root)
	defs := make(map[string]*Schema)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			log.Printf("@ %s", path)
			if strings.HasSuffix(path, "push_rule.yaml") {
				log.Print("TODO handle later, multiple types")
				return nil
			}
			raw, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			var schema Schema
			if err := yaml.Unmarshal(raw, &schema); err != nil {
				return err
			}
			rel, _ := filepath.Rel(basePath, path)
			key := filepath.ToSlash(rel)
			defs[key] = &schema
		}
		return nil
	})
	for _, it := range defs {
		it.FollowRef(defs)
	}
	return defs, err
}

func parseSpecFiles(root string) ([]Spec, error) {
	var specs []Spec
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && info.Name() == "definitions" {
			return filepath.SkipDir
		}
		if !info.IsDir() {
			log.Printf("@ %s", path)
			if strings.HasSuffix(path, "keys.yaml") {
				log.Print("TODO handle later, multiple types")
				return nil
			}
			if strings.HasSuffix(path, "notifications.yaml") {
				log.Print("TODO handle later, multiple types")
				return nil
			}
			raw, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			var spec Spec
			if err := yaml.Unmarshal(raw, &spec); err != nil {
				return err
			}
			specs = append(specs, spec)
		}
		return nil
	})
	return specs, err
}

type APISpec struct {
	Package   string
	version   string
	overrides map[string]string
	defs      map[string]*Schema
	specs     []Spec
	common    map[string]Schema
	Handlers  []APIHandler
}

func ParseAPISpec(root, version, pkg string, overrides map[string]string, skipOperationIDs []string) (*APISpec, error) {
	defs, err := parseDefinitionFiles(path.Join(root, "definitions"))
	if err != nil {
		return nil, err
	}
	specs, err := parseSpecFiles(root)
	if err != nil {
		return nil, err
	}

	skips := make(map[string]bool, len(skipOperationIDs))
	for _, it := range skipOperationIDs {
		skips[it] = true
	}
	common := make(map[string]Schema)
	var handlers []APIHandler
	for _, s := range specs {
		hs := s.extractHandlers(version, defs, skips)
		handlers = append(handlers, hs...)
		for _, h := range hs {
			if h.Body != nil {
				for _, it := range h.Body.NestedDef {
					common[it.Identifier] = it
				}
			}
			for _, it := range h.Response.NestedDef {
				common[it.Identifier] = it
			}
		}
	}
	return &APISpec{pkg, version, overrides, defs, specs, common, handlers}, nil
}

type orderedSchemas []Schema

func (os orderedSchemas) Len() int           { return len(os) }
func (os orderedSchemas) Less(i, j int) bool { return os[i].Identifier < os[j].Identifier }
func (os orderedSchemas) Swap(i, j int)      { os[i], os[j] = os[j], os[i] }

func (s *APISpec) CommonDefs() []Schema {
	ss := orderedSchemas(make([]Schema, 0, len(s.common)))
	for _, v := range s.common {
		ss = append(ss, v)
	}
	sort.Sort(ss)
	return ss
}

func (s *APISpec) Source() []byte {
	tmpl := template.Must(
		template.New("api").
			Funcs(template.FuncMap{
				"getOverride": func(typ string) string { return s.overrides[typ] },
			}).
			ParseGlob("generate/templates/*"),
	)
	var content bytes.Buffer
	if err := tmpl.Execute(&content, s); err != nil {
		panic(err)
	}
	return content.Bytes()
}
