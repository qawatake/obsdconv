package main

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

type YamlConverter interface {
	ConvertYAML(raw []byte, title string, alias string, newtags []string) (output []byte, err error)
}

type YamlConverterImpl struct {
	publishable bool
}

func NewYamlConverterImpl(publishable bool) *YamlConverterImpl {
	return &YamlConverterImpl{
		publishable: publishable,
	}
}

func (c *YamlConverterImpl) ConvertYAML(raw []byte, title string, alias string, newtags []string) (output []byte, err error) {

	m := make(map[interface{}]interface{})
	if err := yaml.Unmarshal(raw, m); err != nil {
		return nil, fmt.Errorf("failed to unmarshal front matter: %w", err)
	}
	if title != "" {
		m["title"] = title
	}

	if v, ok := m["aliases"]; !ok {
		if alias != "" {
			m["aliases"] = []string{alias}
		}
	} else {
		if vv, ok := v.([]interface{}); !ok {
			return nil, fmt.Errorf("aliases field found but its field type is not []interface{}: %T", v)
		} else {
			exists := false
			for _, a := range vv {
				aa, ok := a.(string)
				if !ok {
					return nil, fmt.Errorf("aliases field found but its field type is not string: %T", a)
				}
				if aa == alias {
					exists = true
				}
			}
			if !exists {
				vv = append(vv, alias)
			}
			m["aliases"] = vv
		}
	}

	if v, ok := m["tags"]; !ok {
		if len(newtags) > 0 {
			tags := make([]string, len(newtags))
			copy(tags, newtags)
			m["tags"] = tags
		}
	} else {
		if vv, ok := v.([]interface{}); !ok {
			return nil, fmt.Errorf("tags field found but its field type is not []interface{}: %T", v)
		} else {
			existingTag := make(map[string]bool)
			for _, a := range vv {
				aa, ok := a.(string)
				if !ok {
					return nil, fmt.Errorf("tags field found but its field type is not string: %T", a)
				}
				existingTag[aa] = true
			}
			for _, t := range newtags {
				if !existingTag[t] {
					vv = append(vv, t)
				}
			}
			m["tags"] = vv
		}
	}

	_, ok := m["draft"]
	if !ok && c.publishable {
		if p, ok := m["publish"]; !ok {
			m["draft"] = true
		} else {
			if publishable, ok := p.(bool); !ok {
				return nil, fmt.Errorf("publish field found but its field type is not bool: %T", p)
			} else {
				m["draft"] = !publishable
			}
		}
	}

	output, err = yaml.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output yaml: %w", err)
	}
	return output, nil
}
