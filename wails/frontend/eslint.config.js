import js from '@eslint/js'
import globals from 'globals'
import svelte from 'eslint-plugin-svelte'
import tseslint from 'typescript-eslint'

const wailsjsPatterns = [
  {
    group: ['**/wailsjs/**', 'wailsjs/**'],
    message:
      'Import wailsjs only inside src/lib/bridge. Everything else goes through the typed bridge (PLAN_WAILS.md section 2.1).'
  }
]

export default tseslint.config(
  { ignores: ['dist/**', 'wailsjs/**', 'node_modules/**'] },
  js.configs.recommended,
  ...tseslint.configs.recommended,
  ...svelte.configs['flat/recommended'],
  {
    languageOptions: { globals: { ...globals.browser, ...globals.node } }
  },
  {
    files: ['**/*.svelte', '**/*.svelte.ts'],
    languageOptions: { parserOptions: { parser: tseslint.parser } }
  },
  {
    files: ['src/**/*.{ts,svelte}'],
    ignores: ['src/lib/bridge/**'],
    rules: { 'no-restricted-imports': ['error', { patterns: wailsjsPatterns }] }
  },
  {
    rules: {
      '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
      'svelte/no-at-html-tags': 'error',
      'max-lines': ['error', { max: 200, skipBlankLines: true, skipComments: true }],
      'max-depth': ['error', 3],
      'max-lines-per-function': ['error', { max: 50, skipBlankLines: true, skipComments: true }]
    }
  },
  {
    // Tests group many small cases under one describe.
    files: ['**/*.test.ts'],
    rules: { 'max-lines-per-function': 'off' }
  }
)
