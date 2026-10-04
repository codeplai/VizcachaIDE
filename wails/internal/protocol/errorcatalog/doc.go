// Package errorcatalog is the engine that explains errors to beginners, shared by every language.
//
// A language gives it a catalog (JSON, embedded in its adapter) and an OutputParser. The catalog
// is data, not code: for every stable id (E-*, P-*, V-*) it holds regular expressions and the
// EN/ES texts. The texts live there and not in internal/i18n because go-i18n parses "{{" as a
// template action and the fix hints contain code with braces ("func main() { ... }").
//
// ResolvePath, LocationFrom, NamedGroups and SplitLines help the language parsers.
package errorcatalog
