<template>
  <Transition name="detail-overlay">
    <div
      v-if="entry"
      class="log-detail-overlay"
      @click.self="$emit('close')"
    >
      <div class="log-detail-inner">
        <header class="log-detail-header">
          <div class="log-detail-title">
            <span class="log-badge-method" :class="entry.method">{{ entry.method }}</span>
            <span class="log-detail-path">{{ entry.path }}</span>
            <span class="log-detail-status" :class="{ error: entry.status >= 400 }">{{ entry.status }}</span>
            <span class="log-detail-meta">{{ formatLogTime(entry.timestamp) }} · {{ entry.remote_addr }}</span>
          </div>
          <button class="icon-btn-close" aria-label="Close" @click="$emit('close')">
            <X :size="22" :stroke-width="2" />
          </button>
        </header>
        <div class="log-detail-body">
          <div class="log-detail-panel">
            <h3 class="log-detail-panel-title">Request body</h3>
            <div class="log-detail-content">
              <pre v-if="entry.request_body" class="log-detail-pre"><code>{{ formatLogBody(entry.request_body) }}</code></pre>
              <p v-else class="log-detail-empty">No body</p>
            </div>
          </div>
          <div class="log-detail-panel">
            <h3 class="log-detail-panel-title">Response body</h3>
            <div class="log-detail-content">
              <pre v-if="entry.response_body" class="log-detail-pre"><code>{{ formatLogBody(entry.response_body) }}</code></pre>
              <p v-else class="log-detail-empty">Empty response</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { X } from 'lucide-vue-next'

defineProps({
  entry: { type: Object, default: null },
  formatLogTime: { type: Function, required: true },
  formatLogBody: { type: Function, required: true },
})
defineEmits(['close'])
</script>

<style scoped>
.log-detail-overlay {
  position: fixed;
  inset: 0;
  z-index: 100;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: stretch;
  justify-content: center;
  padding: 1.5rem;
  overflow: hidden;
}
.log-detail-inner {
  flex: 1;
  max-width: 1400px;
  display: flex;
  flex-direction: column;
  background: var(--bg-card);
  border-radius: var(--radius);
  border: 1px solid var(--border-strong);
  box-shadow: 0 24px 48px rgba(0, 0, 0, 0.4);
  overflow: hidden;
}
.log-detail-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--border);
  background: var(--bg-elevated);
  flex-shrink: 0;
}
.log-detail-title {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex: 1;
  min-width: 0;
  flex-wrap: wrap;
}
.log-detail-path {
  font-family: ui-monospace, monospace;
  font-size: 1rem;
  word-break: break-all;
  color: var(--text);
}
.log-detail-status { font-weight: 700; color: var(--accent); }
.log-detail-status.error { color: var(--danger); }
.log-detail-meta { font-size: 0.85rem; color: var(--text-muted); }
.icon-btn-close {
  flex-shrink: 0;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  color: var(--text-muted);
  cursor: pointer;
  transition: color 0.2s, border-color 0.2s, background 0.2s;
}
.icon-btn-close:hover {
  color: var(--text);
  border-color: var(--border-strong);
  background: var(--bg);
}
.log-detail-body {
  flex: 1;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0;
  min-height: 0;
  overflow: hidden;
}
@media (max-width: 900px) {
  .log-detail-body { grid-template-columns: 1fr; }
}
.log-detail-panel {
  display: flex;
  flex-direction: column;
  min-height: 0;
  border-right: 1px solid var(--border);
}
.log-detail-body .log-detail-panel:last-child { border-right: none; }
@media (max-width: 900px) {
  .log-detail-panel {
    border-right: none;
    border-bottom: 1px solid var(--border);
  }
}
.log-detail-panel-title {
  margin: 0;
  padding: 0.65rem 1rem;
  font-size: 0.8rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
  background: var(--bg-elevated);
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}
.log-detail-content {
  flex: 1;
  overflow: auto;
  padding: 1rem;
  min-height: 120px;
}
.log-detail-pre {
  margin: 0;
  font-family: ui-monospace, 'SF Mono', 'Cascadia Code', monospace;
  font-size: 0.85rem;
  line-height: 1.5;
  color: var(--text);
  white-space: pre-wrap;
  word-break: break-word;
}
.log-detail-empty { margin: 0; font-size: 0.9rem; color: var(--text-muted); font-style: italic; }
.detail-overlay-enter-active, .detail-overlay-leave-active { transition: opacity 0.2s ease; }
.detail-overlay-enter-active .log-detail-inner,
.detail-overlay-leave-active .log-detail-inner {
  transition: transform 0.25s var(--ease-out);
}
.detail-overlay-enter-from, .detail-overlay-leave-to { opacity: 0; }
.detail-overlay-enter-from .log-detail-inner,
.detail-overlay-leave-to .log-detail-inner {
  transform: scale(0.97) translateY(8px);
}
</style>
