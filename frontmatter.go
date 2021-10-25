package main

import (
	_ "gopkg.in/yaml.v2"
)

type frontMatter struct {
	Title   string   `yaml:"title,omitempty"`
	Aliases []string `yaml:"aliases,omitempty"`
	Tags    []string `yaml:"tags,omitempty"`
}
