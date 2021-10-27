package main

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

type frontMatter struct {
	title string
	tags  []string
	alias string
}

func convertYAML(raw []byte, frontmatter frontMatter, flags *flagBundle) (output []byte, err error) {
	m := make(map[interface{}]interface{})
	if err := yaml.Unmarshal(raw, m); err != nil {
		return nil, fmt.Errorf("failed to unmarshal front matter: %w", err)
	}
	m["title"] = frontmatter.title

	if v, ok := m["aliases"]; !ok {
		m["aliases"] = []string{frontmatter.alias}
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
				if aa == frontmatter.alias {
					exists = true
				}
			}
			if !exists {
				vv = append(vv, frontmatter.alias)
			}
			m["aliases"] = vv
		}
	}

	if v, ok := m["tags"]; !ok {
		tags := make([]string, len(frontmatter.tags))
		copy(tags, frontmatter.tags)
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
			for _, t := range frontmatter.tags {
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
