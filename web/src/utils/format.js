/**
 * Date and time formatting utilities.
 */

export function formatDate(iso) {
  if (!iso) return ''
  try {
    return new Date(iso).toLocaleString(undefined, {
      dateStyle: 'medium',
      timeStyle: 'short',
    })
  } catch {
    return iso
  }
}

export function relativeTime(iso) {
  if (!iso) return ''
  try {
    const d = new Date(iso)
    const now = new Date()
    const s = Math.floor((now - d) / 1000)
    if (s < 10) return 'just now'
    if (s < 60) return `${s}s ago`
    const m = Math.floor(s / 60)
    if (m < 60) return `${m}m ago`
    const h = Math.floor(m / 60)
    if (h < 24) return `${h}h ago`
    const day = Math.floor(h / 24)
    if (day < 7) return `${day}d ago`
    return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
  } catch {
    return iso
  }
}

export function formatLogTime(iso) {
  if (!iso) return '—'
  try {
    return new Date(iso).toLocaleString(undefined, {
      dateStyle: 'short',
      timeStyle: 'medium',
    })
  } catch {
    return String(iso)
  }
}

/** Pretty-print JSON when possible, otherwise return as-is. */
export function formatLogBody(raw) {
  if (raw == null || raw === '') return ''
  const s = typeof raw === 'string' ? raw : String(raw)
  const t = s.trim()
  if ((t.startsWith('{') && t.endsWith('}')) || (t.startsWith('[') && t.endsWith(']'))) {
    try {
      const parsed = JSON.parse(s)
      return JSON.stringify(parsed, null, 2)
    } catch {
      return s
    }
  }
  return s
}
