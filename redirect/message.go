package redirect

import (
	"fmt"
	"net/http"
)

func Message(w http.ResponseWriter, r *http.Request, id int64) {
	http.Redirect(w, r, fmt.Sprintf("/message?id=%d", id), http.StatusFound)
}
