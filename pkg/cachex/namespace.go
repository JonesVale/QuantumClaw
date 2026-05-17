package cachex

import "strings"

type Namespace struct {
	prefix string
}

func NewNamespace(prefix string) Namespace {
	return Namespace{prefix: prefix}
}

func (n Namespace) FullKey(key string) string {
	if key == "" {
		return ""
	}
	if n.prefix == "" {
		return key
	}
	return n.prefix + ":" + key
}

func (n Namespace) MatchPattern() string {
	if n.prefix == "" {
		return "*"
	}
	return n.prefix + ":*"
}

func (n Namespace) Prefix() string {
	return n.prefix
}

func (n Namespace) HasPrefix(key string) bool {
	if n.prefix == "" {
		return true
	}
	return strings.HasPrefix(key, n.prefix+":")
}