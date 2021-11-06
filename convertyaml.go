package main

import (
	"errors"
	"fmt"

	"github.com/qawatake/obsdconv/process"
	"gopkg.in/yaml.v2"
)

type yamlConvAuxInImpl struct {
	title   string
	alias   string
	newtags []string
}

func newYamlConvAuxInImpl(title string, alias string, newtags []string) *yamlConvAuxInImpl {
	return &yamlConvAuxInImpl{
		title:   title,
		alias:   alias,
		newtags: newtags,
	}
}

type yamlConverterImpl struct {
	publishable bool
}

func newYamlConverterImpl(publishable bool) *yamlConverterImpl {
	return &yamlConverterImpl{
		publishable: publishable,
	}
}

func (c *yamlConverterImpl) ConvertYAML(raw []byte, aux process.YamlConvAuxIn) (output []byte, err error) {
	title := ""
	alias := ""
	var newtags []string

	if v, ok := aux.(*yamlConvAuxInImpl); !ok {
		return nil, errors.New("input (YamlConverterInput) cannot be converted to yamlConverterInputImpl")
	} else {
		title = v.title
		alias = v.alias
		newtags = v.newtags
	}

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
