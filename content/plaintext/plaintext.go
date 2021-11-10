package plaintext

import (
	"html/template"
	"io"
	"io/ioutil"
	"regexp"
)

var (
	newlines = regexp.MustCompile(`\r?\n\r?\n`)
	anchors  = regexp.MustCompile(`\b(file|ftp|http|https?):\/\/\S+[\/\w]`)
)

func ToHTML(reader io.Reader) (template.HTML, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}
	safe := template.HTMLEscapeString(string(data))
	safe = anchors.ReplaceAllString(safe, `<a class="ucc" href="$0">$0</a>`)
	safe = newlines.ReplaceAllString(safe, `</p><p class="ucc">`)
	return template.HTML(`<p class="ucc">` + safe + `</p>`), nil
}
