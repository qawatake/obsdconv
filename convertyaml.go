package main

import (
	"errors"
	"fmt"
	"strings"

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
	synctag     bool
	synctlal    bool
	publishable bool
	remap       map[string]string
}

func newYamlConverterImpl(synctag bool, synctlal bool, publishable bool, remap map[string]string) *yamlConverterImpl {
	return &yamlConverterImpl{
		synctag:     synctag,
		synctlal:    synctlal,
		publishable: publishable,
		remap:       remap,
	}
}

func parseRemap(input string) (remap map[string]string, err error) {
	remap = make(map[string]string)
	if input == "" {
		return nil, nil
	}
	entries := strings.Split(input, ",")
	for _, entry := range entries {
		pair := strings.Split(entry, ":")
		if len(pair) != 2 {
			return nil, newMainErrf(MAIN_ERR_KIND_INVALID_REMAP_FORMAT, "invalid format of %s: \"%s\"", FLAG_REMAP_META_KEYS, input)
		}
		remap[pair[0]] = pair[1]
	}
	return remap, nil
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

	// synctlal
	existingTitle := ""
	if c.synctlal {
		if v, ok := m["title"]; ok {
			if vv, ok := v.(string); !ok {
				return nil, fmt.Errorf("aliases field found but its field type is not string: %T", v)
			} else {
				existingTitle = vv
			}
		}
	}

	// title
	if title != "" {
		m["title"] = title
	}

	// alias
	if alias != "" {
		if v, ok := m["aliases"]; !ok {
			if alias != "" {
				m["aliases"] = []string{alias}
			}
		} else {
			if vv, ok := v.([]interface{}); !ok {
				return nil, fmt.Errorf("aliases field found but its field type is not []interface{}: %T", v)
			} else {
				exists := false
				aliases := make([]string, 0, len(vv))
				for _, a := range vv {
					aa, ok := a.(string)
					if !ok {
						return nil, fmt.Errorf("aliases field found but its field type is not string: %T", a)
					}
					if c.synctlal && aa == existingTitle {
						continue
					}
					if aa == alias {
						exists = true
					}
					aliases = append(aliases, aa)
				}
				if !exists {
					aliases = append(aliases, alias)
				}
				m["aliases"] = aliases
			}
		}
	}

	// tags
	if c.synctag {
		delete(m, "tags")
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

	// publishable -> draft
	// if draft field already exists, then keep it as is.
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

	// remap keys in front matter
	if len(c.remap) > 0 {
		for oldKey, newKey := range c.remap {
			v, ok := m[oldKey]
			if !ok {
				continue
			}
			delete(m, oldKey)
			if newKey == "" {
				continue
			}
			m[newKey] = v
		}
	}

	// for empty front matters
	if len(m) == 0 {
		return nil, nil
	}
	output, err = yaml.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output yaml: %w", err)
	}
	return output, nil
}
