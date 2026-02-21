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
    <div v-if="serverUrls.length" class="exposed-urls">
      <span class="exposed-label">Open from phone:</span>
      <span v-for="(url, i) in serverUrls" :key="url" class="exposed-url-wrap">
        <a :href="url" target="_blank" rel="noopener noreferrer" class="exposed-url">{{ url }}</a>
        <button type="button" class="copy-url-btn" title="Copy URL" @click="copyUrl(url)">
          <Copy :size="14" :stroke-width="2" />
        </button>
      </span>
    </div>
  </header>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ClipboardList, RefreshCw, Copy } from 'lucide-vue-next'
import { getServerInfo } from '../api.js'

defineProps({
  currentPage: { type: String, default: 'main' },
  refreshing: { type: Boolean, default: false },
  latest: { type: Object, default: null },
})
const emit = defineEmits(['navigate', 'refresh', 'copy-url'])

const serverUrls = ref([])

onMounted(async () => {
  try {
    const info = await getServerInfo()
    if (info?.urls?.length) serverUrls.value = info.urls
  } catch {
    // ignore
  }
})

async function copyUrl(url) {
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(url)
      emit('copy-url', url)
    }
  } catch {
    // ignore
  }
}
</script>

<style scoped>
.header {
  position: relative;
  z-index: 1;
  max-width: 1120px;
  width: 100%;
  margin: 0 auto;
  padding: 0.75rem 0.75rem 0.5rem;
  padding-top: max(0.75rem, env(safe-area-inset-top));
  padding-left: max(0.75rem, env(safe-area-inset-left));
  padding-right: max(0.75rem, env(safe-area-inset-right));
  animation: slideDown 0.6s var(--ease-out);
}
@media (min-width: 600px) {
  .header { padding: 1rem 1rem 0.75rem; }
}
@media (min-width: 900px) {
  .header { padding: 1.5rem 1.25rem 1rem; }
}
@keyframes slideDown {
  from { opacity: 0; transform: translateY(-12px); }
  to { opacity: 1; transform: translateY(0); }
}
.header-inner {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.4rem 0.75rem;
}
@media (min-width: 600px) {
  .header-inner { gap: 0.5rem 1rem; }
}
.logo {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}
.logo-icon {
  display: flex;
  flex-shrink: 0;
  color: var(--headline);
  opacity: 0.9;
}
.title {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 700;
  letter-spacing: -0.03em;
  color: var(--headline);
}
@media (min-width: 600px) {
  .title { font-size: 1.5rem; }
}
.tagline {
  margin: 0;
  color: var(--text-muted);
  font-size: 0.75rem;
  flex: 1;
  min-width: 0;
}
@media (min-width: 600px) {
  .tagline { font-size: 0.875rem; }
}
@media (max-width: 380px) {
  .tagline { display: none; }
}
.header-nav {
  display: flex;
  align-items: center;
  gap: 0.2rem;
}
.nav-link {
  padding: 0.45rem 0.65rem;
  font-size: 0.8rem;
  font-weight: 500;
  color: var(--text-muted);
  background: transparent;
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: color 0.2s, background 0.2s, border-color 0.2s;
  min-height: 44px;
  min-width: 44px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
@media (min-width: 600px) {
  .nav-link { padding: 0.4rem 0.75rem; font-size: 0.85rem; min-height: 0; min-width: 0; }
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
  gap: 0.4rem;
}
.icon-btn-header {
  width: 40px;
  height: 40px;
  min-width: 40px;
  min-height: 40px;
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
@media (min-width: 600px) {
  .icon-btn-header { width: 36px; height: 36px; min-width: 36px; min-height: 36px; }
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
  font-size: 0.65rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  padding: 0.2rem 0.4rem;
  border-radius: 6px;
  background: var(--bg-elevated);
  color: var(--text-muted);
  border: 1px solid var(--border);
}
@media (min-width: 600px) {
  .sync-badge { font-size: 0.7rem; padding: 0.25rem 0.5rem; }
}
.sync-badge.live {
  background: var(--accent-soft);
  color: var(--accent);
  border-color: var(--accent-dim);
}

.exposed-urls {
  margin-top: 0.5rem;
  padding-top: 0.5rem;
  border-top: 1px solid var(--border);
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.4rem 0.75rem;
  font-size: 0.7rem;
}
@media (min-width: 600px) {
  .exposed-urls { margin-top: 0.75rem; padding-top: 0.75rem; gap: 0.5rem 1rem; font-size: 0.8rem; }
}
.exposed-label {
  color: var(--text-muted);
  flex-shrink: 0;
  width: 100%;
}
@media (min-width: 480px) {
  .exposed-label { width: auto; }
}
.exposed-url-wrap {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
}
.exposed-url {
  color: var(--accent);
  text-decoration: none;
  word-break: break-all;
}
.exposed-url:hover {
  text-decoration: underline;
}
.copy-url-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.35rem;
  min-width: 36px;
  min-height: 36px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 6px;
  color: var(--text-muted);
  cursor: pointer;
  transition: color 0.2s, border-color 0.2s;
}
.copy-url-btn:hover {
  color: var(--accent);
  border-color: var(--accent-dim);
}
</style>
