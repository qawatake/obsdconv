package main

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

type YamlConverter interface {
	ConvertYAML(raw []byte) (output []byte, err error)
}

type YamlConverterImpl struct {
	title string
	tags  []string
	alias string
	flags *flagBundle
}

func NewYamlConverterImpl(flags *flagBundle) *YamlConverterImpl {
	return &YamlConverterImpl{
		flags: flags,
	}
}

func (c *YamlConverterImpl) convertYAML(raw []byte) (output []byte, err error) {
	if c.flags == nil {
		return nil, fmt.Errorf("pointer to flagBundle is nil")
	}
	m := make(map[interface{}]interface{})
	if err := yaml.Unmarshal(raw, m); err != nil {
		return nil, fmt.Errorf("failed to unmarshal front matter: %w", err)
	}
	if c.title != "" {
		m["title"] = c.title
	}

	if v, ok := m["aliases"]; !ok {
		if c.alias != "" {
			m["aliases"] = []string{c.alias}
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
				if aa == c.alias {
					exists = true
				}
			}
			if !exists {
				vv = append(vv, c.alias)
			}
			m["aliases"] = vv
		}
	}

	if v, ok := m["tags"]; !ok {
		if len(c.tags) > 0 {
			tags := make([]string, len(c.tags))
			copy(tags, c.tags)
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
			for _, t := range c.tags {
				if !existingTag[t] {
					vv = append(vv, t)
				}
			}
			m["tags"] = vv
		}
	}

	_, ok := m["draft"]
	if !ok && c.flags.publishable {
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
