package markdown

import (
	"fmt"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"html/template"
	"io"
	"io/ioutil"
	"log"
)

func ToHTML(reader io.Reader) (template.HTML, error) {
	s, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}
	p := goldmark.DefaultParser()
	p.AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(extension.NewLinkifyParser(extension.WithLinkifyAllowedProtocols([][]byte{
				[]byte("file:"),
				[]byte("ftp:"),
				[]byte("http:"),
				[]byte("https:"),
			})), 999),
		),
	)
	r := p.Parse(text.NewReader(s))
	var result string
	if err := ast.Walk(r, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch n.Kind().String() {
			case "Document":
				// Do nothing
			case "Paragraph":
				result += `<p class="ucc">`
			case "Text":
				text := string(n.Text(s))
				result += template.HTMLEscapeString(text)
				if t := n.(*ast.Text); t.HardLineBreak() {
					result += `<br />`
				} else if t.SoftLineBreak() {
					result += ` `
				}
			case "TextBlock":
				// Do nothing
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
				result += `<pre class="ucc"><code class="ucc">
`
				lines := n.Lines()
				for i := 0; i < lines.Len(); i++ {
					line := lines.At(i)
					result += string(line.Value(s))
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
				if p := n.Parent().(*ast.List); !p.IsTight {
					result += `
`
				}
			case "Link":
				l := n.(*ast.Link)
				result += fmt.Sprintf(`<a class="ucc" href="%s" title="%s">`, l.Destination, l.Title)
			case "AutoLink":
				l := n.(*ast.AutoLink)
				switch l.AutoLinkType {
				case ast.AutoLinkEmail:
					result += fmt.Sprintf(`<a class="ucc" href="mailto:%s">`, string(l.URL(s)))
				case ast.AutoLinkURL:
					result += fmt.Sprintf(`<a class="ucc" href="%s">`, string(l.URL(s)))
				}
				result += string(l.Label(s))
			default:
				log.Println("Entering Unhandled Node:", n.Kind().String())
			}
		} else {
			switch n.Kind().String() {
			case "Document":
				// Do nothing
			case "Paragraph":
				result += `</p>
`
			case "Text":
				// Do nothing
			case "TextBlock":
				// Do nothing
			case "ThematicBreak":
				// Do nothing
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
