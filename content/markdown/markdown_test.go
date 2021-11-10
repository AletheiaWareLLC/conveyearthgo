package markdown_test

import (
	"aletheiaware.com/conveyearthgo/content/markdown"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestMarkdownToHTML(t *testing.T) {
	wd, err := os.Getwd()
	assert.NoError(t, err)
	directory := filepath.Join(wd, "testdata")
	for name, tt := range map[string]struct {
		given    string
		expected string
	}{
		"rule": {
			given:    "rule.md",
			expected: "rule.html",
		},
		"heading": {
			given:    "heading.md",
			expected: "heading.html",
		},
		"code": {
			given:    "code.md",
			expected: "code.html",
		},
		"paragraph": {
			given:    "paragraph.md",
			expected: "paragraph.html",
		},
		"emphasis": {
			given:    "emphasis.md",
			expected: "emphasis.html",
		},
		"list": {
			given:    "list.md",
			expected: "list.html",
		},
		"quote": {
			given:    "quote.md",
			expected: "quote.html",
		},
		"link": {
			given:    "link.md",
			expected: "link.html",
		},
	} {
		t.Run(name, func(t *testing.T) {
			log.Println("********", name, "********")
			given, err := os.Open(filepath.Join(directory, tt.given))
			assert.NoError(t, err)
			html, err := markdown.ToHTML(given)
			assert.NoError(t, err)
			assertMatchesMaster(t, directory, tt.expected, []byte(html))
		})
	}
}

func assertMatchesMaster(t *testing.T, directory, name string, html []byte) {
	t.Helper()
	masterPath := filepath.Join(directory, name)
	failedPath := filepath.Join(directory, "failed", name)
	_, err := os.Stat(masterPath)
	if os.IsNotExist(err) {
		assert.Nil(t, writeHTML(t, failedPath, html))
		t.Fatalf("Master not found at %s. HTML written to %s might be used as master.", masterPath, failedPath)
	}

	master, err := ioutil.ReadFile(masterPath)
	assert.NoError(t, err)

	if !assert.Equal(t, master, html, "HTML did not match master. Actual HTML written to file://%s.", failedPath) {
		assert.Nil(t, writeHTML(t, failedPath, html))
	}
}

func writeHTML(t *testing.T, path string, html []byte) error {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(path, html, 0644)
}
