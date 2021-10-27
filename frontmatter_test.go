package main

import "testing"

func TestConvertYAML(t *testing.T) {
	raw := []byte(`cssclass: index-page
publish: true
`)
	frontmatter := frontMatter{
		Title:   "211027",
		Aliases: []string{"59ndb0zo"},
		Tags:    []string{"todo", "math"},
	}
	want := []byte(`aliases:
- 59ndb0zo
cssclass: index-page
publish: true
tags:
- todo
- math
title: "211027"
`)
	got, err := convertYAML(raw, frontmatter)
	if err != nil {
		t.Fatalf("[FATAL] unexpected error occurred: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("[ERROR]\ngot:\n%s\nwant:\n%s", string(got), string(want))
	}

}
