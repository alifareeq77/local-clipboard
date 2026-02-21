<template>
  <main class="main logs-page">
    <section class="logs-panel-full">
      <div class="logs-panel-header">
        <h2>Request logs</h2>
        <span class="hint">{{ requestLogs.length }} entries</span>
        <button
          class="btn btn-ghost btn-sm"
          :class="{ spinning: logsLoading }"
          :disabled="logsLoading"
          title="Refresh logs"
          @click="$emit('refresh')"
        >
          <RefreshCw :size="16" :stroke-width="2" />
          Refresh
        </button>
      </div>
      <div class="logs-table-wrap">
        <template v-if="logsLoading && !requestLogs.length">
          <div class="skeleton-item log-skeleton" v-for="i in 10" :key="'log-sk-' + i">
            <div class="skeleton-line"></div>
            <div class="skeleton-line short"></div>
          </div>
        </template>
        <template v-else-if="!requestLogs.length">
          <div class="empty-state logs-empty">
            <p>No requests logged yet.</p>
            <p class="empty-hint">Use the app or refresh to capture requests.</p>
          </div>
        </template>
        <div v-else class="logs-list">
          <button
            v-for="(entry, idx) in requestLogs"
            :key="entry.timestamp + entry.path + entry.method + String(idx)"
            type="button"
            class="log-card"
            :class="{ 'status-error': entry.status >= 400, 'status-success': entry.status < 400 }"
            @click="$emit('select', entry)"
          >
            <div class="log-card-main">
              <span class="log-badge-method" :class="entry.method">{{ entry.method }}</span>
              <span class="log-path">{{ entry.path }}</span>
              <span class="log-status" :class="{ error: entry.status >= 400 }">{{ entry.status }}</span>
            </div>
            <div class="log-card-meta">
              <span class="log-time">{{ formatLogTime(entry.timestamp) }}</span>
              <span class="log-ip">{{ entry.remote_addr }}</span>
            </div>
          </button>
        </div>
      </div>
    </section>

    <LogDetailOverlay
      :entry="selectedLogEntry"
      :format-log-time="formatLogTime"
      :format-log-body="formatLogBody"
      @close="$emit('close-detail')"
    />
  </main>
</template>

<script setup>
import { RefreshCw } from 'lucide-vue-next'
import LogDetailOverlay from './LogDetailOverlay.vue'

defineProps({
  requestLogs: { type: Array, default: () => [] },
  logsLoading: { type: Boolean, default: false },
  selectedLogEntry: { type: Object, default: null },
  formatLogTime: { type: Function, required: true },
  formatLogBody: { type: Function, required: true },
})
defineEmits(['refresh', 'select', 'close-detail'])
</script>

<style scoped>
.logs-page {
  max-width: none;
  width: 100%;
  padding: 0;
  margin: 0;
  min-height: calc(100dvh - 5rem);
  min-height: calc(100vh - 5rem);
  display: flex;
  flex-direction: column;
}
.logs-panel-full {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: var(--bg-card);
  border-radius: var(--radius);
  border: 1px solid var(--border);
  margin: 0 0.75rem 1.5rem;
  min-height: 0;
}
@media (min-width: 600px) {
  .logs-panel-full { margin: 0 1rem 1.75rem; }
}
@media (min-width: 900px) {
  .logs-panel-full { margin: 0 1.25rem 2rem; }
}
.logs-panel-header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem 0.75rem;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}
@media (min-width: 600px) {
  .logs-panel-header { padding: 1rem 1.25rem; gap: 0.75rem; }
}
.logs-panel-header h2 { margin: 0; font-size: 1.25rem; font-weight: 700; color: var(--headline); }
.logs-panel-header .hint { color: var(--text-muted); font-size: 0.9rem; }
.logs-panel-header .btn { margin-left: auto; }
.logs-table-wrap {
  overflow: auto;
  overflow-x: hidden;
  flex: 1;
  padding: 0.5rem 0.75rem;
  min-height: 200px;
}
@media (min-width: 600px) {
  .logs-table-wrap { padding: 0.75rem; }
}
.logs-list { display: flex; flex-direction: column; gap: 0.5rem; }
.log-card {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem 1rem;
  padding: 0.75rem 0.85rem;
  text-align: left;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: border-color 0.2s, background 0.2s, box-shadow 0.2s;
  font: inherit;
  color: inherit;
  min-height: 44px;
}
@media (min-width: 600px) {
  .log-card { padding: 0.85rem 1rem; min-height: 0; }
}
.log-card:hover {
  border-color: var(--border-strong);
  background: var(--bg-elevated);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}
.log-card-main { display: flex; align-items: center; gap: 0.75rem; flex: 1; min-width: 0; }
.log-badge-method {
  flex-shrink: 0;
  font-size: 0.7rem;
  font-weight: 700;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
.log-badge-method.GET { background: rgba(45, 212, 138, 0.2); color: var(--accent); }
.log-badge-method.POST { background: rgba(99, 102, 241, 0.2); color: #818cf8; }
.log-badge-method.PUT, .log-badge-method.PATCH { background: rgba(251, 191, 36, 0.2); color: #fbbf24; }
.log-badge-method.DELETE { background: rgba(255, 68, 102, 0.2); color: var(--danger); }
.log-card .log-path {
  font-family: ui-monospace, monospace;
  font-size: 0.9rem;
  word-break: break-all;
  color: var(--text);
}
.log-card .log-status { flex-shrink: 0; font-weight: 700; font-size: 0.9rem; }
.log-card.status-success .log-status { color: var(--accent); }
.log-card.status-error .log-status { color: var(--danger); }
.log-card-meta {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 1rem;
  font-size: 0.8rem;
  color: var(--text-muted);
}
.log-skeleton, .logs-empty { margin: 0.5rem 0; }
.skeleton-line {
  height: 12px;
  background: var(--border);
  border-radius: 6px;
  opacity: 0.5;
  animation: shimmer 1.2s ease-in-out infinite;
}
.skeleton-line.short { width: 60%; margin-top: 0.5rem; }
.empty-state { padding: 2.5rem 1.5rem; text-align: center; color: var(--text-muted); }
.empty-hint { font-size: 0.85rem; opacity: 0.85; }
</style>
