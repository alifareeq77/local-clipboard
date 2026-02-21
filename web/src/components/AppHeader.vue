<template>
  <header class="header">
    <div class="header-inner">
      <div class="logo">
        <ClipboardList class="logo-icon" :size="26" :stroke-width="1.8" />
        <h1 class="title">Clipboard Bridge</h1>
      </div>
      <p class="tagline">Sync across devices · Search · Pin · Premium</p>
      <nav class="header-nav">
        <button
          class="nav-link"
          :class="{ active: currentPage === 'main' }"
          @click="$emit('navigate', 'main')"
        >
          Clipboard
        </button>
        <button
          class="nav-link"
          :class="{ active: currentPage === 'logs' }"
          @click="$emit('navigate', 'logs')"
        >
          Request logs
        </button>
      </nav>
      <div class="header-actions">
        <button
          v-if="currentPage === 'main'"
          class="icon-btn-header"
          :class="{ spinning: refreshing }"
          title="Refresh"
          @click="$emit('refresh')"
        >
          <RefreshCw :size="18" :stroke-width="2" />
        </button>
        <span v-if="currentPage === 'main'" class="sync-badge" :class="{ live: latest?.text }">
          {{ latest?.text ? 'Live' : 'Empty' }}
        </span>
      </div>
    </div>
  </header>
</template>

<script setup>
import { ClipboardList, RefreshCw } from 'lucide-vue-next'

defineProps({
  currentPage: { type: String, default: 'main' },
  refreshing: { type: Boolean, default: false },
  latest: { type: Object, default: null },
})
defineEmits(['navigate', 'refresh'])
</script>

<style scoped>
.header {
  position: relative;
  z-index: 1;
  max-width: 1120px;
  margin: 0 auto;
  padding: 1.5rem 1.25rem 1rem;
  animation: slideDown 0.6s var(--ease-out);
}
@keyframes slideDown {
  from { opacity: 0; transform: translateY(-12px); }
  to { opacity: 1; transform: translateY(0); }
}
.header-inner {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem 1rem;
}
.logo {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.logo-icon {
  display: flex;
  flex-shrink: 0;
  color: var(--headline);
  opacity: 0.9;
}
.title {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 700;
  letter-spacing: -0.03em;
  color: var(--headline);
}
.tagline {
  margin: 0;
  color: var(--text-muted);
  font-size: 0.875rem;
  flex: 1;
}
.header-nav {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}
.nav-link {
  padding: 0.4rem 0.75rem;
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--text-muted);
  background: transparent;
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: color 0.2s, background 0.2s, border-color 0.2s;
}
.nav-link:hover { color: var(--text); }
.nav-link.active {
  color: var(--accent);
  background: var(--accent-soft);
  border-color: var(--accent-dim);
}
.header-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.icon-btn-header {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  color: var(--text-muted);
  cursor: pointer;
  transition: transform 0.2s var(--ease-out), color 0.2s, border-color 0.2s;
}
.icon-btn-header:hover {
  color: var(--text);
  border-color: var(--border-strong);
}
.icon-btn-header.spinning {
  animation: spin 0.5s var(--ease-out);
}
@keyframes spin { to { transform: rotate(360deg); } }
.sync-badge {
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  padding: 0.25rem 0.5rem;
  border-radius: 6px;
  background: var(--bg-elevated);
  color: var(--text-muted);
  border: 1px solid var(--border);
}
.sync-badge.live {
  background: var(--accent-soft);
  color: var(--accent);
  border-color: var(--accent-dim);
}
</style>
