package main

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

type frontMatter struct {
	Title   string   `yaml:"title,omitempty"`
	Tags    []string `yaml:"tags,omitempty"`
	Alias   string
}

func convertYAML(raw []byte, frontmatter frontMatter) (output []byte, err error) {
	m := make(map[interface{}]interface{})
	if err := yaml.Unmarshal(raw, m); err != nil {
		return nil, fmt.Errorf("failed to unmarshal front matter: %w", err)
	}
	m["title"] = frontmatter.Title

	if v, ok := m["aliases"]; !ok {
		m["aliases"] = []string{frontmatter.Alias}
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
				if aa == frontmatter.Alias {
					exists = true
				}
			}
			if !exists {
				vv = append(vv, frontmatter.Alias)
			}
			m["aliases"] = vv
		}
	}

	if v, ok := m["tags"]; !ok {
		tags := make([]string, len(frontmatter.Tags))
		copy(tags, frontmatter.Tags)
		m["tags"] = tags
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
			for _, t := range frontmatter.Tags {
				if !existingTag[t] {
					vv = append(vv, t)
				}
			}
			m["tags"] = vv
		}
	}

	output, err = yaml.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output yaml: %w", err)
	}
	return output, nil
}
