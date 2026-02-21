import { ref, onMounted, onUnmounted } from 'vue'
import { getClipboard, postClipboard } from '../api.js'
import { normalizeLineEndings } from '../utils/text.js'

export function useClipboard(showToast) {
  const inputText = ref('')
  const sending = ref(false)
  const sendStatus = ref('')
  const sendError = ref(false)
  const latest = ref(null)
  const latestLoading = ref(true)
  const latestUpdated = ref(false)
  let latestTimer = null
  let latestUpdatedTimeout = null

  /** @param {boolean} [skipHistoryRefresh] - true when we're about to refresh history ourselves (e.g. after send from this tab) */
  async function loadLatest(skipHistoryRefresh = false) {
    latestLoading.value = true
    try {
      const data = await getClipboard()
      const hadText = latest.value?.text
      latest.value = data
      if (data?.text && data.text !== hadText) {
        latestUpdated.value = true
        clearTimeout(latestUpdatedTimeout)
        latestUpdatedTimeout = setTimeout(() => { latestUpdated.value = false }, 2000)
        // Update history when a new push is detected (from any device). Skip when we just sent from this tab (caller will refresh).
        if (loadHistoryRef && !skipHistoryRefresh) {
          await loadHistoryRef()
        }
      }
    } catch {
      latest.value = null
    } finally {
      latestLoading.value = false
    }
  }

  async function send(sendTextareaRef) {
    const raw = sendTextareaRef?.value?.value !== undefined
      ? sendTextareaRef.value.value
      : inputText.value
    const text = normalizeLineEndings(raw).trim()
    if (!text || sending.value) return
    sending.value = true
    sendStatus.value = ''
    sendError.value = false
    try {
      await postClipboard(text, 'web')
      sendStatus.value = 'Saved.'
      showToast('Sent to clipboard', 'success')
      inputText.value = ''
      await loadLatest(true)
      if (loadHistoryRef) await loadHistoryRef()
    } catch (e) {
      sendStatus.value = 'Failed: ' + (e.message || 'network error')
      sendError.value = true
      showToast('Send failed', 'error')
    } finally {
      sending.value = false
    }
  }

  let loadHistoryRef = null
  function setLoadHistory(fn) {
    loadHistoryRef = fn
  }

  async function pasteFromClipboard() {
    const appendPasted = (pasted) => {
      const normalized = normalizeLineEndings(pasted)
      inputText.value = inputText.value ? inputText.value + normalized : normalized
      showToast('Pasted', 'success')
    }
    try {
      if (navigator.clipboard && typeof navigator.clipboard.readText === 'function') {
        const pasted = await navigator.clipboard.readText()
        appendPasted(pasted)
        return
      }
    } catch (e) {
      console.warn('Clipboard API read failed:', e)
    }
    try {
      const ta = document.createElement('textarea')
      ta.value = ''
      ta.setAttribute('readonly', '')
      ta.style.position = 'fixed'
      ta.style.left = '-9999px'
      ta.style.top = '0'
      ta.style.opacity = '0'
      document.body.appendChild(ta)
      ta.focus()
      const ok = document.execCommand('paste')
      const pasted = ta.value || ''
      document.body.removeChild(ta)
      if (ok && pasted) {
        appendPasted(pasted)
        return
      }
    } catch (e) {
      console.warn('execCommand paste fallback failed:', e)
    }
    showToast('Use Ctrl+V / Cmd+V in the text area', 'info')
  }

  async function copyText(text) {
    if (!text) return
    try {
      if (navigator.clipboard && window.isSecureContext) {
        await navigator.clipboard.writeText(text)
      } else {
        const ta = document.createElement('textarea')
        ta.value = text
        ta.style.position = 'fixed'
        ta.style.opacity = '0'
        document.body.appendChild(ta)
        ta.select()
        document.execCommand('copy')
        document.body.removeChild(ta)
      }
      showToast('Copied to clipboard', 'success')
    } catch {
      showToast('Copy failed', 'error')
    }
  }

  onMounted(() => {
    loadLatest()
    latestTimer = setInterval(loadLatest, 3500)
  })

  onUnmounted(() => {
    if (latestTimer) clearInterval(latestTimer)
    if (latestUpdatedTimeout) clearTimeout(latestUpdatedTimeout)
  })

  return {
    inputText,
    sending,
    sendStatus,
    sendError,
    latest,
    latestLoading,
    latestUpdated,
    loadLatest,
    send,
    setLoadHistory,
    pasteFromClipboard,
    copyText,
  }
}
