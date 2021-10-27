package main

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

type frontMatter struct {
	Title   string   `yaml:"title,omitempty"`
	Aliases []string `yaml:"aliases,omitempty"`
	Tags    []string `yaml:"tags,omitempty"`
}

func convertYAML(raw []byte, frontmatter frontMatter) (output []byte, err error) {
	m := make(map[interface{}]interface{})
	if err := yaml.Unmarshal(raw, m); err != nil {
		return nil, fmt.Errorf("failed to unmarshal front matter: %w", err)
	}
	m["title"] = frontmatter.Title

	if v, ok := m["aliases"]; !ok {
		m["aliases"] = make([]string, len(frontmatter.Aliases))
		if vv, ok := m["aliases"].([]string); ok {
			copy(vv, frontmatter.Aliases)
		} else {
			return nil, fmt.Errorf("aliases field not found and failed to add aliases: %w", err)
		}
	} else {
		if vv, ok := v.([]string); ok {
			vv = append(vv, frontmatter.Aliases...)
		} else {
			return nil, fmt.Errorf("aliases field found but failed to add aliases: %w", err)
		}
	}

	if v, ok := m["tags"]; !ok {
		m["tags"] = make([]string, len(frontmatter.Tags))
		if vv, ok := m["tags"].([]string); ok {
			copy(vv, frontmatter.Tags)
		} else {
			return nil, fmt.Errorf("tags field not found and failed to add tags: %w", err)
		}
	} else {
		if vv, ok := v.([]string); ok {
			vv = append(vv, frontmatter.Tags...)
		} else {
			return nil, fmt.Errorf("tags field found but failed to add tags: %w", err)
		}
	}

	output, err = yaml.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output yaml: %w", err)
	}
	return output, nil
}
