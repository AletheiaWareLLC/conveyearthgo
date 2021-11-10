package plaintext_test

import (
	"aletheiaware.com/conveyearthgo/content/plaintext"
	"github.com/stretchr/testify/assert"
	"html/template"
	"strings"
	"testing"
)

func TestPlainTextToHTML(t *testing.T) {
	reader := strings.NewReader(`This is a

test with a link; https://example.com`)
	expected := template.HTML(`<p class="ucc">This is a</p><p class="ucc">test with a link; <a class="ucc" href="https://example.com">https://example.com</a></p>`)
	actual, err := plaintext.ToHTML(reader)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
