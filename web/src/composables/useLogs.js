import { ref, onUnmounted } from 'vue'
import { getLogs } from '../api.js'
import { formatLogTime, formatLogBody } from '../utils/format.js'

export function useLogs(showToast) {
  const requestLogs = ref([])
  const logsLoading = ref(false)
  const selectedLogEntry = ref(null)

  async function loadLogs() {
    logsLoading.value = true
    try {
      const list = await getLogs()
      requestLogs.value = Array.isArray(list) ? list : []
    } catch {
      requestLogs.value = []
      showToast('Failed to load logs', 'error')
    } finally {
      logsLoading.value = false
    }
  }

  onUnmounted(() => {
    selectedLogEntry.value = null
  })

  return {
    requestLogs,
    logsLoading,
    selectedLogEntry,
    loadLogs,
    formatLogTime,
    formatLogBody,
  }
}
