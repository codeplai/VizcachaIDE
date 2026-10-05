package packageindex

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

// The search pages are read with a real HTML parser, never with regular expressions: these are
// the few questions the two pages ask of the tree.

func parseDocument(body []byte) (*html.Node, error) {
	return html.Parse(bytes.NewReader(body))
}

// hasClass tells whether an element has the class among its classes.
func hasClass(node *html.Node, class string) bool {
	if node.Type != html.ElementNode {
		return false
	}
	for _, attribute := range node.Attr {
		if attribute.Key != "class" {
			continue
		}
		for _, candidate := range strings.Fields(attribute.Val) {
			if candidate == class {
				return true
			}
		}
	}
	return false
}

// findAll returns, in document order, the elements for which match is true.
func findAll(root *html.Node, match func(*html.Node) bool) []*html.Node {
	var found []*html.Node
	var visit func(*html.Node)
	visit = func(node *html.Node) {
		if match(node) {
			found = append(found, node)
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(root)
	return found
}

// withClass returns the elements below root that have the class.
func withClass(root *html.Node, class string) []*html.Node {
	return findAll(root, func(node *html.Node) bool { return hasClass(node, class) })
}

// firstWithClass returns the first element below root with the class, or nil.
func firstWithClass(root *html.Node, class string) *html.Node {
	if found := withClass(root, class); len(found) > 0 {
		return found[0]
	}
	return nil
}

// textOf returns the text inside a node on one line; "" for a nil node.
func textOf(node *html.Node) string {
	if node == nil {
		return ""
	}
	var text strings.Builder
	var visit func(*html.Node)
	visit = func(current *html.Node) {
		if current.Type == html.TextNode {
			text.WriteString(current.Data)
			text.WriteByte(' ')
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(node)
	return oneLine(text.String())
}

// mentions tells whether the visible text of the page contains the phrase (any case).
func mentions(root *html.Node, phrase string) bool {
	return strings.Contains(strings.ToLower(textOf(root)), phrase)
}
