package generate

import (
	"log"
	"sort"
	"strings"
)

type Type struct {
	string
	nullable bool
}

func (t *Type) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var array []string
	if unmarshal(&array) == nil && len(array) == 2 {
		typ := ""
		nullable := false
		for _, it := range array {
			if it == "null" {
				nullable = true
			} else {
				typ = it
			}
		}
		if nullable {
			t.string = typ
			t.nullable = nullable
			return nil
		}
	}
	return unmarshal(&t.string)
}

func (t *Type) GoType() string {
	typ := ""
	switch t.string {
	case "boolean":
		typ = "bool"
	case "integer":
		typ = "int"
	case "string":
		typ = "string"
	default:
		log.Printf("unkwnown swagger type: %q", t.string)
		return "interface{}"
	}
	if t.nullable {
		return "*" + typ
	}
	return typ
}

type AdditionalProperties struct {
	Schema
}

func (ap *AdditionalProperties) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var b bool
	if unmarshal(&b) == nil && b {
		ap.Schema.Type = Type{"object", false}
	} else if err := unmarshal(&ap.Schema); err != nil {
		return err
	}
	return nil
}

func isStruct(typ string) bool {
	if typ == "interface{}" {
		return false
	}
	if strings.HasPrefix(typ, "map") {
		return false
	}
	if strings.HasPrefix(typ, "[]") {
		return false
	}
	if strings.HasPrefix(typ, "*") {
		return false
	}
	return true
}

type Schema struct {
	Type                 Type
	Description          string
	Properties           map[string]*Schema
	AdditionalProperties *AdditionalProperties `yaml:"additionalProperties"`
	Items                *Schema
	Required             []string
	Ref                  string `yaml:"$ref"`

	Identifier string    `yaml:"-"`
	Attributes []GoAttr  `yaml:"-"`
	Nested     []*Schema `yaml:"-"`
}

func (s *Schema) syncAttributes() {
	if len(s.Attributes) == len(s.Properties) {
		return
	}

	keys := make([]string, 0, len(s.Properties))
	for k := range s.Properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, name := range keys {
		prop := s.Properties[name]
		attr := NewGoAttribute(name, "", prop.Description, !s.isRequired(name))
		switch prop.Type.string {
		case "array":
			if prop.Items.Type.string == "object" {
				prop.Items.Identifier = s.Identifier + attr.ID
				prop.Items.syncAttributes()
				s.Nested = append(s.Nested, prop.Items)
				attr.Type = "[]" + prop.Items.Identifier
			} else {
				attr.Type = "[]" + prop.Items.Type.GoType()
			}
		case "object":
			if prop.AdditionalProperties != nil {
				attr.Type = "map[string]" + prop.AdditionalProperties.Type.GoType()
			} else {
				prop.Identifier = s.Identifier + attr.ID
				prop.syncAttributes()
				s.Nested = append(s.Nested, prop)
				attr.Type = prop.Identifier
			}
		default:
			attr.Type = prop.Type.GoType()
		}
		if attr.Opt && isStruct(attr.Type) {
			attr.Type = "*" + attr.Type
		}
		s.Attributes = append(s.Attributes, attr)
	}
}

func (s *Schema) FollowRef(defs map[string]*Schema) {
	if def, ok := defs[s.Ref]; ok {
		s.Type = def.Type
		s.Properties = def.Properties
		s.AdditionalProperties = def.AdditionalProperties
		s.Required = def.Required
	}

	for _, it := range s.Properties {
		it.FollowRef(defs)
		it.syncAttributes()
	}

	s.syncAttributes()
}

func (s *Schema) isRequired(name string) bool {
	for _, it := range s.Required {
		if name == it {
			return true
		}
	}
	return false
}

func (s *Schema) Doc() string {
	return toComment(s.Description)
}
