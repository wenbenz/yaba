package helper

import "net/http"

type FuncHandler struct {
	HandlerFunc func(http.ResponseWriter, *http.Request)
}

func (h FuncHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.HandlerFunc(w, r)
}
