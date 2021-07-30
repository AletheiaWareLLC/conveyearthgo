package redirect

import (
	"net/http"
)

func Publish(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/publish", http.StatusFound)
}
