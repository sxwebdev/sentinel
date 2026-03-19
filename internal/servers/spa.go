package servers

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/sxwebdev/sentinel/frontend"
)

func spaFileServer(index string) http.Handler {
	sub, err := fs.Sub(frontend.StaticFS, "dist")
	if err != nil {
		return http.NotFoundHandler()
	}
	static := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// не перехватываем API-префикс
		if r.URL.Path == "/api" || strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		upath := path.Clean(r.URL.Path)
		if upath == "/" {
			upath = "/" + index
		}
		// отдать файл, если есть
		if f, err := sub.Open(strings.TrimPrefix(upath, "/")); err == nil {
			_ = f.Close()
			static.ServeHTTP(w, r)
			return
		}
		// иначе — SPA fallback на index.html
		r2 := *r
		r2.URL = r.URL
		r2.URL.Path = "/" + index
		static.ServeHTTP(w, &r2)
	})
}
