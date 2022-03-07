package markdown

import (
	"fmt"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"html/template"
	"io"
	"log"
	"strings"
)

func ToHTML(reader io.Reader) (template.HTML, error) {
	s, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	root := goldmark.DefaultParser().Parse(text.NewReader(s))
	var result string
	if err := ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch n.Kind().String() {
			case "Document":
				// Do Nothing
			case "Paragraph":
				result += `<p class="ucc">`
			case "Text":
				result += template.HTMLEscapeString(string(n.Text(s)))
				if t := n.(*ast.Text); t.HardLineBreak() {
					result += `<br />`
				} else if t.SoftLineBreak() {
					result += ` `
				}
			case "TextBlock":
				// Do Nothing
			case "ThematicBreak":
				result += `<hr class="ucc" />
`
			case "Heading":
				result += fmt.Sprintf(`<h%d class="ucc">`, n.(*ast.Heading).Level)
			case "Emphasis":
				switch n.(*ast.Emphasis).Level {
				case 1:
					result += `<em class="ucc">`
				case 2:
					result += `<strong class="ucc">`
				default:
					log.Println("Unhandled Emphasis Level:", n.(*ast.Emphasis).Level)
				}
			case "Blockquote":
				result += `<blockquote class="ucc">
`
			case "CodeSpan":
				result += `<code class="ucc">`
			case "CodeBlock", "FencedCodeBlock":
				result += `<pre class="ucc"><code class="ucc">`
				lines := n.Lines()
				for i := 0; i < lines.Len(); i++ {
					line := lines.At(i)
					result += template.HTMLEscapeString(string(line.Value(s)))
				}
			case "List":
				l := n.(*ast.List)
				if l.IsOrdered() {
					if s := l.Start; s == 1 {
						result += `<ol class="ucc">
`
					} else {
						result += fmt.Sprintf(`<ol class="ucc" start="%d">
`, s)
					}
				} else {
					result += `<ul class="ucc">
`
				}
			case "ListItem":
				result += `<li class="ucc">`
				if p, ok := n.Parent().(*ast.List); ok && !p.IsTight {
					result += `
`
				}
			case "Link":
				l := n.(*ast.Link)
				u := string(util.EscapeHTML(util.URLEscape(l.Destination, true)))
				result += `<a class="ucc" href="`
				result += u
				if t := template.HTMLEscapeString(string(l.Title)); t != "" {
					result += `" title="`
					result += t
				}
				result += `">`
			case "AutoLink":
				l := n.(*ast.AutoLink)
				u := string(util.EscapeHTML(util.URLEscape(l.URL(s), true)))
				result += `<a class="ucc" href="`
				if l.AutoLinkType == ast.AutoLinkEmail && !strings.HasPrefix(strings.ToLower(u), `mailto:`) {
					result += `mailto:`
				}
				result += u
				result += `">`
				result += template.HTMLEscapeString(string(l.Label(s)))
			case "RawHTML":
				// Not Supported
			default:
				log.Println("Entering Unhandled Node:", n.Kind().String())
			}
		} else {
			switch n.Kind().String() {
			case "Document":
				// Do Nothing
			case "Paragraph":
				result += `</p>
`
			case "Text":
				// Do Nothing
			case "TextBlock":
				// Do Nothing
			case "ThematicBreak":
				// Do Nothing
			case "Heading":
				result += fmt.Sprintf(`</h%d>
`, n.(*ast.Heading).Level)
			case "Emphasis":
				switch n.(*ast.Emphasis).Level {
				case 1:
					result += `</em>`
				case 2:
					result += `</strong>`
				default:
					log.Println("Unhandled Emphasis Level:", n.(*ast.Emphasis).Level)
				}
			case "Blockquote":
				result += `</blockquote>
`
			case "CodeSpan":
				result += `</code>`
			case "CodeBlock", "FencedCodeBlock":
				result += `</code></pre>
`
			case "List":
				l := n.(*ast.List)
				if l.IsOrdered() {
					result += `</ol>
`
				} else {
					result += `</ul>
`
				}
			case "ListItem":
				result += `</li>
`
			case "Link", "AutoLink":
				result += `</a>`
			case "RawHTML":
				// Not Supported
			default:
				log.Println("Exiting Unhandled Node:", n.Kind().String())
			}
		}
		return ast.WalkContinue, nil
	}); err != nil {
		return "", err
	}
	return template.HTML(result), nil
}
