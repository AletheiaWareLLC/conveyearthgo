package handler

import (
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/netgo/handler"
	"net/http"
)

func AttachContentHandler(m *http.ServeMux, cm conveyearthgo.ContentManager, cache string) {
	m.Handle("/content", http.NotFoundHandler())
	m.Handle("/content/", handler.Log(handler.CacheControl(http.StripPrefix("/content/", Content(cm)), cache)))
}

func Content(cm conveyearthgo.ContentManager) http.Handler {
	fs := http.FileServer(http.FS(cm))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if results, ok := r.URL.Query()["mime"]; ok && len(results) > 0 {
			w.Header().Add("Content-Type", results[0])
		}
		fs.ServeHTTP(w, r)
	})
}
