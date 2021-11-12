package main

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type yamlExaminatorImpl struct {
	publishable bool
}

func newYamlExaminatorImpl(publishable bool) *yamlExaminatorImpl {
	return &yamlExaminatorImpl{
		publishable: publishable,
	}
}

func (examinator *yamlExaminatorImpl) ExamineYaml(yml []byte) (beProcessed bool, err error) {
	if !examinator.publishable {
		return true, nil
	}

	m := make(map[interface{}]interface{})
	if err := yaml.Unmarshal(yml, m); err != nil {
		return false, errors.Wrap(err, "failed to unmarshal front matter")
	}

	if vdraft, ok := m["draft"]; ok {
		if isdraft, ok := vdraft.(bool); ok && !isdraft {
			return true, nil
		}
	}

	if vpublishable, ok := m["publish"]; ok {
		if ispublishable, ok := vpublishable.(bool); ok && ispublishable {
			return true, nil
		}
	}

	return false, nil
}

