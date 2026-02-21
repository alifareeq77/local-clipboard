<template>
  <div class="app" :class="{ 'panel-focus': focusedPanel }">
    <div class="bg" aria-hidden="true"></div>

    <AppHeader
      :current-page="currentPage"
      :refreshing="refreshing"
      :latest="clipboard.latest.value"
      @navigate="onNavigate"
      @refresh="refreshAll"
      @copy-url="() => toasts.showToast('URL copied', 'success')"
    />

    <main class="main" v-show="currentPage === 'main'">
      <SendPanel
        ref="sendPanelRef"
        :input-text="clipboard.inputText.value"
        @update:input-text="clipboard.inputText.value = $event"
        :sending="clipboard.sending.value"
        :send-status="clipboard.sendStatus.value"
        :send-error="clipboard.sendError.value"
        :latest="clipboard.latest.value"
        :latest-loading="clipboard.latestLoading.value"
        :latest-updated="clipboard.latestUpdated.value"
        :focused="focusedPanel === 'send'"
        @send="(el) => clipboard.send(el)"
        @paste="clipboard.pasteFromClipboard()"
        @clear="onSendClear"
        @copy-latest="clipboard.copyText($event)"
        @focus="focusedPanel = 'send'"
        @blur="onPanelBlur('send')"
      />
      <HistoryPanel
        ref="historyPanelRef"
        :search-query="history.searchQuery.value"
        @update:search-query="history.searchQuery.value = $event"
        :sort-by="history.sortBy.value"
        @update:sort-by="history.sortBy.value = $event"
        :limit="history.limit.value"
        @update:limit="history.limit.value = $event"
        :view-mode="history.viewMode.value"
        @update:view-mode="history.viewMode.value = $event"
        :history-items="history.historyItems.value"
        :history-loading="history.historyLoading.value"
        :copy-id="history.copyId.value"
        :filtered-items="history.filteredItems.value"
        :highlight-item="history.highlightItem"
        :focused="focusedPanel === 'history'"
        @search-input="history.debouncedSearch()"
        @search="history.loadHistory()"
        @clear-search="history.searchQuery.value = ''; history.loadHistory()"
        @limit-change="history.loadHistory()"
        @copy-item="(item) => history.copyItem(item, clipboard.copyText)"
        @toggle-pin="(item) => history.togglePin(item)"
        @delete-item="(item) => history.deleteItem(item)"
        @focus="focusedPanel = 'history'"
        @blur="onPanelBlur('history')"
      />
    </main>

    <LogsPage
      v-show="currentPage === 'logs'"
      :request-logs="logs.requestLogs.value"
      :logs-loading="logs.logsLoading.value"
      :selected-log-entry="logs.selectedLogEntry.value"
      :format-log-time="logs.formatLogTime"
      :format-log-body="logs.formatLogBody"
      @refresh="logs.loadLogs()"
      @select="logs.selectedLogEntry.value = $event"
      @close-detail="logs.selectedLogEntry.value = null"
    />

    <ToastContainer
      :toasts="toasts.toasts.value"
      @remove="toasts.removeToast"
    />

    <ShortcutsModal v-if="showShortcuts" @close="showShortcuts = false" />

    <footer class="credit">
      Created with ❤️ by
      <a href="https://github.com/alifareeq77" target="_blank" rel="noopener noreferrer">alifareeq</a>
      ·
      <a href="https://www.linkedin.com/in/ali-fareeq-1390351b0/" target="_blank" rel="noopener noreferrer">LinkedIn</a>
    </footer>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import AppHeader from './components/AppHeader.vue'
import SendPanel from './components/SendPanel.vue'
import HistoryPanel from './components/HistoryPanel.vue'
import LogsPage from './components/LogsPage.vue'
import ToastContainer from './components/ToastContainer.vue'
import ShortcutsModal from './components/ShortcutsModal.vue'
import { useToasts } from './composables/useToasts.js'
import { useClipboard } from './composables/useClipboard.js'
import { useHistory } from './composables/useHistory.js'
import { useLogs } from './composables/useLogs.js'
import './styles/variables.css'

const currentPage = ref('main')
const focusedPanel = ref(null)
const showShortcuts = ref(false)
const refreshing = ref(false)
const sendPanelRef = ref(null)
const historyPanelRef = ref(null)

const toasts = useToasts()
const clipboard = useClipboard(toasts.showToast)
const history = useHistory(toasts.showToast)
const logs = useLogs(toasts.showToast)

clipboard.setLoadHistory(history.loadHistory)

function onNavigate(page) {
  currentPage.value = page
  if (page === 'logs') logs.loadLogs()
}

async function refreshAll() {
  refreshing.value = true
  await Promise.all([clipboard.loadLatest(), history.loadHistory()])
  toasts.showToast('Refreshed', 'success')
  setTimeout(() => { refreshing.value = false }, 400)
}

function onSendClear() {
  clipboard.inputText.value = ''
  toasts.showToast('Cleared', 'info')
}

function onPanelBlur(panel) {
  setTimeout(() => {
    if (focusedPanel.value === panel) focusedPanel.value = null
  }, 0)
}

function onKeydown(e) {
  if (e.key === 'Escape') {
    if (logs.selectedLogEntry.value) {
      logs.selectedLogEntry.value = null
      return
    }
    focusedPanel.value = null
    if (document.activeElement?.blur) document.activeElement.blur()
  }
  if (e.key === '/' && !e.ctrlKey && !e.metaKey && !e.altKey) {
    const target = e.target
    if (!target || !target.closest?.('.search-input')) {
      e.preventDefault()
      historyPanelRef.value?.searchInputRef?.focus()
    }
  }
  if (e.key === '?' && !e.ctrlKey && !e.metaKey) {
    showShortcuts.value = true
  }
}

onMounted(() => {
  window.addEventListener('keydown', onKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKeydown)
})
</script>

<style>
* { box-sizing: border-box; }

/* Minimal scrollbar: thin and only when needed; no horizontal scroll */
html {
  overflow-x: hidden;
  scrollbar-width: thin;
  scrollbar-color: rgba(255, 255, 255, 0.15) transparent;
}
body {
  overflow-x: hidden;
  margin: 0;
  min-height: 100dvh;
  min-height: 100vh;
}
/* WebKit: thin, subtle scrollbar */
html::-webkit-scrollbar,
.app::-webkit-scrollbar,
[class*="-list"]::-webkit-scrollbar,
[class*="-wrap"]::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}
html::-webkit-scrollbar-track,
.app::-webkit-scrollbar-track,
[class*="-list"]::-webkit-scrollbar-track,
[class*="-wrap"]::-webkit-scrollbar-track {
  background: transparent;
}
html::-webkit-scrollbar-thumb,
.app::-webkit-scrollbar-thumb,
[class*="-list"]::-webkit-scrollbar-thumb,
[class*="-wrap"]::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.12);
  border-radius: 3px;
}
html::-webkit-scrollbar-thumb:hover,
.app::-webkit-scrollbar-thumb:hover,
[class*="-list"]::-webkit-scrollbar-thumb:hover,
[class*="-wrap"]::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.2);
}
</style>

<style scoped>
.app {
  min-height: 100dvh;
  min-height: 100vh;
  position: relative;
  font-family: var(--font);
  font-size: 15px;
  line-height: 1.5;
  -webkit-font-smoothing: antialiased;
  color: var(--text);
  overflow-x: hidden;
  padding-left: env(safe-area-inset-left);
  padding-right: env(safe-area-inset-right);
  padding-bottom: env(safe-area-inset-bottom);
}

.bg {
  position: fixed;
  inset: 0;
  background: var(--bg);
  pointer-events: none;
  z-index: 0;
}

.main {
  position: relative;
  z-index: 1;
  max-width: 1120px;
  width: 100%;
  margin: 0 auto;
  padding: 0 0.75rem 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

@media (min-width: 600px) {
  .main {
    padding: 0 1rem 1.75rem;
    gap: 1.25rem;
  }
}

@media (min-width: 900px) {
  .main {
    gap: 1.5rem;
    padding: 0 1.25rem 2rem;
  }
}

.credit {
  position: relative;
  z-index: 1;
  text-align: center;
  padding: 0.75rem 1rem;
  font-size: 0.8rem;
  color: var(--text-muted);
}
@media (min-width: 600px) {
  .credit { padding: 1rem 1rem; font-size: 0.85rem; }
}
.credit a {
  color: var(--accent);
  text-decoration: none;
}
.credit a:hover {
  text-decoration: underline;
}
</style>
