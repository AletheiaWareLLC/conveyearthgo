package markdown_test

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	v8 "rogchap.com/v8go"
	"strings"
	"testing"
)

func Test_JSMarkdownToHTML(t *testing.T) {
	wd, err := os.Getwd()
	assert.NoError(t, err)
	javascript := filepath.Join(wd, "..", "..", "cmd", "server", "assets", "html", "static")
	commonmark, err := ioutil.ReadFile(filepath.Join(javascript, "commonmark.min.js"))
	assert.NoError(t, err)
	editor, err := ioutil.ReadFile(filepath.Join(javascript, "editor.js"))
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
			markdown, err := ioutil.ReadFile(filepath.Join(directory, tt.given))
			assert.NoError(t, err)
			iso, err := v8.NewIsolate()
			assert.NoError(t, err)
			defer iso.Dispose()
			ctx, err := v8.NewContext(iso)
			assert.NoError(t, err)
			defer ctx.Close()

			/* Enable Logging
			global := ctx.Global()
			console, err := v8.NewObjectTemplate(iso)
			assert.NoError(t, err)
			logfn, err := v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
				log.Println(info.Args()[0])
				return nil
			})
			assert.NoError(t, err)
			console.Set("log", logfn)
			consoleObj, _ := console.NewInstance(ctx)
			global.Set("console", consoleObj)
			*/

			_, err = ctx.RunScript(string(commonmark), "commonmark.js")
			assert.NoError(t, err)
			_, err = ctx.RunScript(string(editor), "editor.js")
			assert.NoError(t, err)
			_, err = ctx.RunScript("const parser = new commonmark.Parser();", "test.js")
			assert.NoError(t, err)
			_, err = ctx.RunScript("const markdown = `"+strings.ReplaceAll(string(markdown), "`", "\\`")+"`;", "test.js")
			assert.NoError(t, err)
			_, err = ctx.RunScript("const html = markdownToHTML(parser, markdown);", "test.js")
			assert.NoError(t, err)
			v, err := ctx.RunScript("html", "test.js")
			assert.NoError(t, err)
			html := v.String()
			assertMatchesMaster(t, directory, tt.expected, []byte(html))
		})
	}
}
