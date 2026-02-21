/**
 * Text utilities: truncation, escaping, search highlight, line endings.
 */

export function shortText(text, max = 320) {
  if (!text) return ''
  const t = String(text)
  return t.length <= max ? t : t.slice(0, max) + '…'
}

export function escapeHtml(s) {
  const div = document.createElement('div')
  div.textContent = s
  return div.innerHTML
}

export function highlightSearch(text, query, shortTextFn = shortText, escapeHtmlFn = escapeHtml) {
  const q = (query || '').trim().toLowerCase()
  if (!q || !text) return escapeHtmlFn(shortTextFn(text))
  const t = shortTextFn(text)
  const lower = t.toLowerCase()
  const idx = lower.indexOf(q)
  if (idx === -1) return escapeHtmlFn(t)
  const before = escapeHtmlFn(t.slice(0, idx))
  const match = escapeHtmlFn(t.slice(idx, idx + q.length))
  const after = escapeHtmlFn(t.slice(idx + q.length))
  return `${before}<mark class="search-highlight">${match}</mark>${after}`
}

/** Normalize line endings for multi-line text (e.g. \r\n, \r, Unicode separators -> \n). */
export function normalizeLineEndings(s) {
  if (typeof s !== 'string') return s
  return s
    .replace(/\r\n/g, '\n')
    .replace(/\r/g, '\n')
    .replace(/\u2028/g, '\n')
    .replace(/\u2029/g, '\n')
}
