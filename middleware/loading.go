package middleware

import (
	"net/http"
	"reflect"
)

type Loading struct {
	next   http.Handler
	wait   bool
	object IsLoading
}

type IsLoading interface {
	Loading() bool
}

func NewLoading(next http.Handler, wait bool, object IsLoading) *Loading {
	hasObject := !(object == nil || reflect.ValueOf(object).IsNil())

	return &Loading{
		next:   next,
		wait:   wait && hasObject,
		object: object,
	}
}

func (m *Loading) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if m.wait && m.object.Loading() {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	m.next.ServeHTTP(w, r)
}
