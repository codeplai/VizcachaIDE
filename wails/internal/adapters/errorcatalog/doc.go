// Package errorcatalog implements app.ErrorExplainer: a parser of the output of go
// build, go run and go vet (and of panics) plus a catalog of 25 explanations.
//
// The catalog is data, not code: data/catalog.json is embedded in the binary and holds,
// for every stable id (E-*, P-*, V-*), its regular expressions and its EN/ES texts.
// The texts live here and not in internal/i18n because go-i18n parses "{{" as a template
// action and the fix hints contain Go code with braces ("func main() { ... }").
package errorcatalog
