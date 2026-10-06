/**
 * A snippet: a template the student gets by typing its label. `template` uses CodeMirror's
 * snippet syntax: `${name}` is a stop for Tab (the same name twice is typed once), `${}` an
 * unnamed one, and a tab at the start of a line becomes the file's indentation.
 */
export interface SnippetDef {
  /** What the student types ("for", "def", "println"). */
  label: string
  template: string
  /** i18n key of the short description (it is shown in the student's language). */
  description: string
}
