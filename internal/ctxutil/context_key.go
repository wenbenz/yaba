package ctxutil

type ContextKey uint

const (
	CTXUser ContextKey = iota
	CTXSID
)
