package redirect

import (
	"fmt"
	"net/http"
)

func Conversation(w http.ResponseWriter, r *http.Request, id int64) {
	http.Redirect(w, r, fmt.Sprintf("/conversation?id=%d", id), http.StatusFound)
}
