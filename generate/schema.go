package generate

import (
	"log"
	"sort"
)

func getGoType(swaggerType string) string {
	switch swaggerType {
	case "boolean":
		return "bool"
	case "integer":
		return "int"
	case "string":
		return "string"
	default:
		log.Printf("unkwnown swagger type: %q", swaggerType)
		return "interface{}"
	}
}

type AdditionalProperties struct {
	Schema
}

func (ap *AdditionalProperties) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var b bool
	if unmarshal(&b) == nil && b {
		ap.Schema.Type = "object"
	} else if err := unmarshal(&ap.Schema); err != nil {
		return err
	}
	return nil
}

type Schema struct {
	Type                 string
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
		switch prop.Type {
		case "array":
			if prop.Items.Type == "object" {
				prop.Items.Identifier = s.Identifier + attr.ID
				prop.Items.syncAttributes()
				s.Nested = append(s.Nested, prop.Items)
				attr.Type = "[]" + prop.Items.Identifier
			} else {
				attr.Type = "[]" + getGoType(prop.Items.Type)
			}
		case "object":
			if prop.AdditionalProperties != nil {
				attr.Type = "map[string]" + getGoType(prop.AdditionalProperties.Type)
			} else {
				prop.Identifier = s.Identifier + attr.ID
				prop.syncAttributes()
				s.Nested = append(s.Nested, prop)
				attr.Type = prop.Identifier
			}
		default:
			attr.Type = getGoType(prop.Type)
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
