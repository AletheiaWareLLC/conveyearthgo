package redirect

import (
	"net/http"
)

func Reply(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/reply", http.StatusFound)
}
