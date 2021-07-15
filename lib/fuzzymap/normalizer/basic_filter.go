package normalizer

import "strings"

type BasicFilter struct{}

func NewBasicFilter() *BasicFilter {
	return &BasicFilter{}
}

func (n *BasicFilter) Normalize(key string) string {
	return strings.ToLower(strings.Join(strings.Fields(key), " "))
}
