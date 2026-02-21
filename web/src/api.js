const API = '/api'

export async function getClipboard() {
  const res = await fetch(`${API}/clipboard`, { cache: 'no-store' })
  if (!res.ok) {
    if (res.status === 404) return null
    throw new Error(res.statusText)
  }
  return res.json()
}

/** Normalize line endings for multi-line text (iOS can use \r, \r\n, or Unicode separators). */
function normalizeLineEndings(s) {
  if (typeof s !== 'string') return s
  return s
    .replace(/\r\n/g, '\n')
    .replace(/\r/g, '\n')
    .replace(/\u2028/g, '\n')
    .replace(/\u2029/g, '\n')
}

export async function postClipboard(text, source = 'web') {
  const normalized = normalizeLineEndings(text).trim()
  const tryJson = async () => {
    const payload = JSON.stringify({ text: normalized, source })
    const body = new Blob([payload], { type: 'application/json; charset=utf-8' })
    const res = await fetch(`${API}/clipboard`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json; charset=utf-8' },
      body,
    })
    if (!res.ok) {
      const msg = await res.text()
      throw new Error(msg || res.statusText)
    }
    return res.json()
  }
  const tryForm = async () => {
    const body = new URLSearchParams()
    body.set('text', normalized)
    body.set('source', source)
    const res = await fetch(`${API}/clipboard`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded; charset=utf-8' },
      body: body.toString(),
    })
    if (!res.ok) {
      const msg = await res.text()
      throw new Error(msg || res.statusText)
    }
    return res.json()
  }
  try {
    return await tryJson()
  } catch (e) {
    try {
      return await tryForm()
    } catch (e2) {
      throw e
    }
  }
}

export async function getHistory(limit = 80, search = '') {
  const params = new URLSearchParams({ limit: String(limit) })
  if (search) params.set('q', search)
  const res = await fetch(`${API}/history?${params}`, { cache: 'no-store' })
  if (!res.ok) throw new Error(res.statusText)
  return res.json()
}

export async function setPin(id, pinned) {
  const res = await fetch(`${API}/history/pin`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ id, pinned }),
  })
  if (!res.ok) throw new Error(res.statusText)
  return res.json()
}

export async function deleteHistory(id) {
  const res = await fetch(`${API}/history/delete`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ id }),
  })
  if (!res.ok) throw new Error(res.statusText)
}

export async function getLogs() {
  const res = await fetch(`${API}/logs`, { cache: 'no-store' })
  if (!res.ok) throw new Error(res.statusText)
  return res.json()
}
