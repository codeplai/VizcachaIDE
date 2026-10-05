package packageindex

import (
	"errors"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"golang.org/x/net/html"
)

var errUnexpectedPage = errors.New("the page does not look like the one that was expected")

// parseGoPage reads the results of https://pkg.go.dev/search?m=package: each result is a
// div.SearchSnippet with its import path in span.SearchSnippet-header-path (in parentheses),
// its synopsis in p.SearchSnippet-synopsis and its version in a <strong> inside the
// "go-textSubtle" span of div.SearchSnippet-infoLabel. A page without results still has the
// div.SearchResults-summary; a page with neither is not readable and gives
// ErrPackageIndexUnavailable. packageBase is the address of the page of a package.
func parseGoPage(body []byte, packageBase string) ([]domain.PackageInfo, error) {
	root, err := parseDocument(body)
	if err != nil {
		return nil, unavailable(err)
	}
	snippets := withClass(root, "SearchSnippet")
	if len(snippets) == 0 {
		if firstWithClass(root, "SearchResults-summary") != nil {
			return nil, nil
		}
		return nil, unavailable(errUnexpectedPage)
	}
	found := make([]domain.PackageInfo, 0, MaxResults)
	for _, snippet := range snippets {
		path := strings.Trim(textOf(firstWithClass(snippet, "SearchSnippet-header-path")), "() ")
		if path == "" || len(found) == MaxResults {
			continue
		}
		found = append(found, domain.PackageInfo{
			Name:        path,
			Version:     goSnippetVersion(snippet),
			Description: textOf(firstWithClass(snippet, "SearchSnippet-synopsis")),
			URL:         packageBase + path,
		})
	}
	if len(found) == 0 {
		return nil, unavailable(errUnexpectedPage)
	}
	return found, nil
}

// goSnippetVersion is the first <strong> whose parent is a span (the "Imported by" count is
// inside a link, the version is not).
func goSnippetVersion(snippet *html.Node) string {
	info := firstWithClass(snippet, "SearchSnippet-infoLabel")
	if info == nil {
		return ""
	}
	for _, strong := range findAll(info, func(node *html.Node) bool {
		return node.Type == html.ElementNode && node.Data == "strong" && node.Parent.Data == "span"
	}) {
		if text := textOf(strong); strings.HasPrefix(text, "v") {
			return text
		}
	}
	return ""
}
