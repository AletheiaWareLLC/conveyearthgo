package redirect

import (
	"fmt"
	"net/http"
)

func Conversation(w http.ResponseWriter, r *http.Request, conversation, message int64) {
	var destination string
	if message == 0 {
		destination = fmt.Sprintf("/conversation?id=%d", conversation)
	} else {
		destination = fmt.Sprintf("/conversation?id=%d#message%d", conversation, message)
	}
	http.Redirect(w, r, destination, http.StatusFound)
}
