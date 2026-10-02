/**
 * Convierte el markdown en línea que usamos en los textos (negrita, cursiva,
 * `código` y [enlaces](url)) a HTML seguro. Primero escapa el HTML.
 */
const escapeHtml = (s) =>
  s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');

export function inlineMarkdown(text) {
  return escapeHtml(text)
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
    .replace(/\*([^*]+)\*/g, '<em>$1</em>')
    .replace(/\[([^\]]+)\]\(([^)\s]+)\)/g, (_m, label, href) => {
      const external = /^https?:/.test(href) && !href.includes('vizcacha.codeplai.pe');
      return `<a href="${href}"${external ? ' rel="noopener"' : ''}>${label}</a>`;
    });
}
