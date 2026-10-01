// Command po2json builds the translation catalogs of the Wails app.
//
// Inputs: the hand-written semantic catalog (ux_copy.json, from docs/wails/UX_COPY.md)
// and the gettext catalogs of the 1.0 (vizcacha/i18n/locale/{en,es}/LC_MESSAGES/vizcacha.po).
// Outputs, for each language:
//   - frontend/src/lib/i18n/locales/<lang>.json (svelte-i18n, ICU MessageFormat);
//   - internal/i18n/locales/<lang>.json (go-i18n v2, embedded in the binary).
//
// Run it from the wails/ folder: go run ./tools/po2json
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	semantic := flag.String("semantic", "tools/po2json/ux_copy.json", "semantic catalog")
	poDir := flag.String("po", "../vizcacha/i18n/locale", "folder with <lang>/LC_MESSAGES/vizcacha.po")
	frontendOut := flag.String("frontend", "frontend/src/lib/i18n/locales", "output folder for svelte-i18n")
	backendOut := flag.String("backend", "internal/i18n/locales", "output folder for go-i18n")
	flag.Parse()

	if err := run(*semantic, *poDir, *frontendOut, *backendOut); err != nil {
		log.Fatalf("po2json: %v", err)
	}
}

func run(semanticPath, poDir, frontendOut, backendOut string) error {
	all, err := loadSemantic(semanticPath)
	if err != nil {
		return err
	}
	legacy, err := loadLegacy(
		filepath.Join(poDir, "en", "LC_MESSAGES", "vizcacha.po"),
		filepath.Join(poDir, "es", "LC_MESSAGES", "vizcacha.po"),
	)
	if err != nil {
		return err
	}
	for key, value := range legacy {
		if _, clash := all[key]; clash {
			return fmt.Errorf("key %q exists in both catalogs", key)
		}
		all[key] = value
	}
	for _, lang := range []string{"en", "es"} {
		if err := writeLanguage(all, lang, frontendOut, backendOut); err != nil {
			return err
		}
	}
	log.Printf("wrote %d messages per language", len(all))
	return nil
}

func writeLanguage(all catalog, lang, frontendOut, backendOut string) error {
	icu := map[string]string{}
	templates := map[string]any{}
	for key, value := range all {
		text := value.EN
		if lang == "es" {
			text = value.ES
		}
		icu[key] = text
		templates[key] = goTemplate(text)
	}
	if err := writeJSON(filepath.Join(frontendOut, lang+".json"), icu); err != nil {
		return err
	}
	return writeJSON(filepath.Join(backendOut, lang+".json"), templates)
}

func writeJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create folder: %w", err)
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil { //nolint:gosec // generated, public file
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
