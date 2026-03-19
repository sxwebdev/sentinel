package servers

import (
	"net/http"
	"strings"
)

type pathRewriter struct {
	next       http.Handler
	rpcBases   []string // "/pkg.Service/"
	reflection []string // reflection base paths
}

func (w *pathRewriter) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	p := r.URL.Path

	if strings.HasPrefix(p, "/api") {
		w.next.ServeHTTP(rw, r)
		return
	}

	for _, base := range w.reflection {
		if p == base || strings.HasPrefix(p, base) {
			r2 := r.Clone(r.Context())
			r2.URL = r.URL
			r2.URL.Path = "/api" + p
			w.next.ServeHTTP(rw, r2)
			return
		}
	}

	for _, base := range w.rpcBases {
		if p == base || strings.HasPrefix(p, base) {
			r2 := r.Clone(r.Context())
			r2.URL = r.URL
			r2.URL.Path = "/api" + p
			w.next.ServeHTTP(rw, r2)
			return
		}
	}

	w.next.ServeHTTP(rw, r)
}
