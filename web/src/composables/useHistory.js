import { ref, computed, onMounted, onUnmounted } from 'vue'
import { getHistory, setPin, deleteHistory } from '../api.js'
import { highlightSearch } from '../utils/text.js'

export function useHistory(showToast) {
  const historyItems = ref([])
  const historyLoading = ref(true)
  const searchQuery = ref('')
  const copyId = ref(null)
  const viewMode = ref('list')
  const sortBy = ref('recent')
  const limit = ref(80)
  let copyTimeout = null
  let searchDebounceTimer = null

  const filteredItems = computed(() => {
    let list = [...historyItems.value]
    if (sortBy.value === 'oldest') list.reverse()
    if (sortBy.value === 'pinned') list = list.filter((i) => i.pinned)
    return list
  })

  function highlightItem(text) {
    return highlightSearch(text, searchQuery.value)
  }

  function debouncedSearch() {
    clearTimeout(searchDebounceTimer)
    searchDebounceTimer = setTimeout(() => loadHistory(), 280)
  }

  async function loadHistory() {
    historyLoading.value = true
    try {
      const items = await getHistory(limit.value, searchQuery.value)
      historyItems.value = Array.isArray(items) ? items : []
    } catch {
      historyItems.value = []
    } finally {
      historyLoading.value = false
    }
  }

  async function togglePin(item) {
    try {
      await setPin(item.id, !item.pinned)
      await loadHistory()
      showToast(!item.pinned ? 'Pinned' : 'Unpinned', 'success')
    } catch {
      showToast('Pin failed', 'error')
    }
  }

  async function deleteItem(item) {
    try {
      await deleteHistory(item.id)
      await loadHistory()
      showToast('Deleted', 'success')
    } catch {
      showToast('Delete failed', 'error')
    }
  }

  async function copyItem(item, copyTextFn) {
    const text = item?.text || ''
    await copyTextFn(text)
    copyId.value = item.id
    clearTimeout(copyTimeout)
    copyTimeout = setTimeout(() => { copyId.value = null }, 1400)
  }

  onMounted(() => {
    loadHistory()
  })

  onUnmounted(() => {
    if (copyTimeout) clearTimeout(copyTimeout)
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
  })

  return {
    historyItems,
    historyLoading,
    searchQuery,
    copyId,
    viewMode,
    sortBy,
    limit,
    filteredItems,
    highlightItem,
    debouncedSearch,
    loadHistory,
    togglePin,
    deleteItem,
    copyItem,
  }
}
