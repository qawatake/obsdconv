package main

import (
	"fmt"
	"strings"

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

func checkFilter(fm map[interface{}]interface{}, filter string) (value bool, err error) {
	token, err := tokenizeFilter(filter)
	if err != nil {
		return false, err
	}
	nd, token, err := parseTokens(token)
	if err != nil || token.kind != TOKEN_EOS {
		return false, newMainErr(MAIN_ERR_KIND_INVALID_FILTER_FORMAT)
	}

	return evaluateNode(nd, fm)
}

func evaluateNode(nd *nodeImpl, fm map[interface{}]interface{}) (value bool, err error) {
	if nd.kind == NODE_IDENT {
		name := nd.name
		if v, ok := fm[name]; ok {
			if value, ok := v.(bool); ok {
				return value, nil
			}
		} else {
			return false, nil
		}
		return false, newMainErr(MAIN_ERR_KIND_INVALID_FILTER_FORMAT)
	} else if nd.kind == NODE_NOT {
		if v, err := evaluateNode(nd.left, fm); err != nil {
			return false, err
		} else {
			return !v, nil
		}
	}

	left, err := evaluateNode(nd.left, fm)
	if err != nil {
		return false, err
	}
	right, err := evaluateNode(nd.right, fm)
	if err != nil {
		return false, err
	}

	if nd.kind == NODE_AND {
		return left && right, nil
	} else {
		return left || right, nil
	}

}

type tokenKind = uint

const (
	TOKEN_IDENT = iota + 1
	TOKEN_AND
	TOKEN_OR
	TOKEN_NOT
	TOKEN_RESERVED
	TOKEN_EOS
)

type tokenImpl struct {
	kind tokenKind
	name string
	next *tokenImpl
}

type nodeKind uint

const (
	NODE_IDENT nodeKind = iota + 1
	NODE_AND
	NODE_OR
	NODE_NOT
)

type nodeImpl struct {
	kind  nodeKind
	name  string
	left  *nodeImpl
	right *nodeImpl
}

func parseTokens(cur *tokenImpl) (nd *nodeImpl, next *tokenImpl, err error) {
	left, cur, err := andNode(cur)
	if err != nil {
		return nil, nil, err
	}
	if cur.kind == TOKEN_OR {
		right, next, err := parseTokens(cur.next)
		if err != nil {
			return nil, nil, err
		}
		return newParentNode(NODE_OR, left, right), next, nil
	}
	return left, cur, nil
}

func andNode(cur *tokenImpl) (nd *nodeImpl, next *tokenImpl, err error) {
	left, cur, err := unaryNode(cur)
	if err != nil {
		return nil, nil, err
	}
	if cur.kind == TOKEN_AND {
		right, next, err := andNode(cur.next)
		if err != nil {
			return nil, nil, err
		}
		return newParentNode(NODE_AND, left, right), next, nil
	}
	return left, cur, err
}

func unaryNode(cur *tokenImpl) (nd *nodeImpl, next *tokenImpl, err error) {
	if cur == nil {
		return nil, nil, newMainErr(MAIN_ERR_KIND_INVALID_FILTER_FORMAT)
	}
	if cur.kind == TOKEN_NOT {
		left, next, err := unaryNode(cur.next)
		if err != nil {
			return nil, nil, err
		}
		nd = newParentNode(NODE_NOT, left, nil)
		return nd, next, nil
	}
	return primaryNode(cur)
}

func primaryNode(cur *tokenImpl) (nd *nodeImpl, next *tokenImpl, err error) {
	if cur == nil {
		return nil, nil, newMainErr(MAIN_ERR_KIND_INVALID_FILTER_FORMAT)
	}
	if cur.kind == TOKEN_IDENT {
		next = cur.next
		return newIdentNode(cur.name), next, nil
	}
	if cur.kind == TOKEN_RESERVED && cur.name == "(" {
		nd, cur, err = parseTokens(cur.next)
		if err != nil {
			return nil, nil, err
		}
		if cur.kind == TOKEN_RESERVED && cur.name == ")" {
			next = cur.next
			return nd, next, nil
		}
	}
	return nil, nil, newMainErr(MAIN_ERR_KIND_INVALID_FILTER_FORMAT)
}

func newParentNode(kind nodeKind, left *nodeImpl, right *nodeImpl) *nodeImpl {
	return &nodeImpl{
		kind:  kind,
		left:  left,
		right: right,
	}
}

func newIdentNode(name string) *nodeImpl {
	return &nodeImpl{
		kind: NODE_IDENT,
		name: name,
	}
}

func newTokenImpl(kind tokenKind, name string, pre *tokenImpl) *tokenImpl {
	child := &tokenImpl{
		kind: kind,
		name: name,
	}
	pre.next = child
	return child
}

func tokenizeFilter(input string) (token *tokenImpl, err error) {
	head := new(tokenImpl)
	runes := []rune(input)
	cur := head
	p := 0
	for i := 0; i < len(runes); i++ {
		if p >= len(runes) {
			break
		}

		if p < len(runes)-1 {
			if string(runes[p:p+2]) == "&&" {
				cur = newTokenImpl(TOKEN_AND, "", cur)
				p += 2
				continue
			}

			if string(runes[p:p+2]) == "||" {
				cur = newTokenImpl(TOKEN_OR, "", cur)
				p += 2
				continue
			}

		}

		if string(runes[p:p+1]) == "!" {
			cur = newTokenImpl(TOKEN_NOT, "", cur)
			p++
			continue
		}

		if runes[p] == '(' || runes[p] == ')' {
			cur = newTokenImpl(TOKEN_RESERVED, string(runes[p:p+1]), cur)
			p++
			continue
		}

		if length := consumeIdent(runes[p:]); length > 0 {
			cur = newTokenImpl(TOKEN_IDENT, string(runes[p:p+length]), cur)
			p += length
			continue
		}

		return nil, newTokinizeErr(string(runes), p)

	}
	cur = newTokenImpl(TOKEN_EOS, "", cur)
	return head.next, nil
}

func consumeIdent(runes []rune) (length int) {
	for i := 0; i < len(runes); i++ {
		char := runes[i]
		if 'a' <= char && char <= 'z' {
			continue
		}
		if 'A' <= char && char <= 'Z' {
			continue
		}
		if '0' <= char && char <= '9' {
			continue
		}
		if char == '-' || char == '_' {
			continue
		}
		return i
	}
	return len(runes)
}

type FilterErrorKind uint

const (
	FILTER_ERROR_TOKENIZE FilterErrorKind = iota + 1
	FILTER_ERROR_PARSE
	FILTER_ERROR_EVALUATE
)

type FilterErr struct {
	kind    FilterErrorKind
	message string
}

func newFilterErrf(kind FilterErrorKind, format string, a ...interface{}) *FilterErr {
	err := new(FilterErr)
	err.kind = kind
	err.message = fmt.Sprintf(format, a...)
	return err
}

func (err *FilterErr) Error() string {
	return err.message
}

func newTokinizeErr(filter string, pos int) *FilterErr {
	errHere := strings.Repeat(" ", pos) + "↑"
	return newFilterErrf(FILTER_ERROR_TOKENIZE, "-filter=%s\n        %s", filter, errHere)
}
