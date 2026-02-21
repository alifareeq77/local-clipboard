<template>
  <section
    class="panel history-panel"
    :class="{ focused }"
    @focusin="$emit('focus')"
    @focusout="$emit('blur')"
  >
    <div class="panel-glow"></div>
    <div class="panel-inner">
      <div class="panel-header">
        <h2>History</h2>
        <span class="hint">{{ historyItems.length }} items</span>
      </div>
      <div class="controls-row">
        <div class="search-wrap">
          <Search class="search-icon" :size="18" :stroke-width="2" />
          <input
            ref="searchInputRef"
            :value="searchQuery"
            class="search-input"
            type="text"
            placeholder="Search…"
            @input="$emit('update:searchQuery', ($event.target).value); $emit('search-input')"
            @keydown.enter="$emit('search')"
          />
          <button
            v-if="searchQuery"
            class="search-clear"
            aria-label="Clear search"
            @click="$emit('update:searchQuery', ''); $emit('clear-search')"
          >
            <X :size="16" :stroke-width="2" />
          </button>
        </div>
        <button class="btn btn-ghost btn-sm" @click="$emit('search')">Search</button>
      </div>
      <div class="filter-row">
        <select :value="sortBy" class="select-sm" @change="$emit('update:sortBy', ($event.target).value)">
          <option value="recent">Recent first</option>
          <option value="oldest">Oldest first</option>
          <option value="pinned">Pinned only</option>
        </select>
        <select :value="limit" class="select-sm" @change="$emit('update:limit', Number(($event.target).value)); $emit('limit-change')">
          <option :value="50">50</option>
          <option :value="80">80</option>
          <option :value="120">120</option>
          <option :value="200">200</option>
        </select>
        <span class="view-toggle">
          <button
            class="view-btn"
            :class="{ active: viewMode === 'list' }"
            title="List"
            @click="$emit('update:viewMode', 'list')"
          >
            <List :size="16" :stroke-width="2" />
          </button>
          <button
            class="view-btn"
            :class="{ active: viewMode === 'compact' }"
            title="Compact"
            @click="$emit('update:viewMode', 'compact')"
          >
            <LayoutList :size="16" :stroke-width="2" />
          </button>
        </span>
      </div>
      <div class="history-list" ref="historyListRef">
        <template v-if="historyLoading">
          <div v-for="i in 5" :key="'sk-' + i" class="skeleton-item">
            <div class="skeleton-line"></div>
            <div class="skeleton-line short"></div>
          </div>
        </template>
        <template v-else-if="!filteredItems.length">
          <div class="empty-state">
            <div class="empty-icon">
              <ClipboardList :size="40" :stroke-width="1.5" />
            </div>
            <p>{{ searchQuery ? 'No matches.' : 'No history yet.' }}</p>
            <p class="empty-hint">{{ searchQuery ? 'Try another query' : 'Send something to get started' }}</p>
          </div>
        </template>
        <template v-else>
          <TransitionGroup name="list" tag="div" class="history-items" :class="viewMode">
            <article
              v-for="(item, index) in filteredItems"
              :key="item.id"
              class="history-item"
              :class="{ pinned: item.pinned, copied: copyId === item.id }"
              :style="{ '--stagger': index }"
            >
              <div class="item-body">
                <div class="item-text" v-html="highlightItem(item.text)"></div>
                <div class="item-meta">
                  <span class="source">{{ item.source }}</span>
                  <span class="time" :title="formatDate(item.updated_at)">
                    {{ relativeTime(item.updated_at) }}
                  </span>
                </div>
              </div>
              <div class="item-actions">
                <button
                  class="icon-btn"
                  :class="{ active: copyId === item.id }"
                  :title="copyId === item.id ? 'Copied!' : 'Copy'"
                  @click="$emit('copy-item', item)"
                >
                  <Check v-if="copyId === item.id" :size="15" :stroke-width="2.5" />
                  <Copy v-else :size="15" :stroke-width="2" />
                </button>
                <button
                  class="icon-btn pin-btn"
                  :class="{ pinned: item.pinned }"
                  :title="item.pinned ? 'Unpin' : 'Pin'"
                  @click="$emit('toggle-pin', item)"
                >
                  <Pin v-if="item.pinned" :size="15" :stroke-width="2" />
                  <PinOff v-else :size="15" :stroke-width="2" />
                </button>
                <button class="icon-btn delete-btn" title="Delete" @click="$emit('delete-item', item)">
                  <Trash2 :size="15" :stroke-width="2" />
                </button>
              </div>
            </article>
          </TransitionGroup>
        </template>
      </div>
    </div>
  </section>
</template>

<script setup>
import { ref } from 'vue'
import { ClipboardList, Search, X, List, LayoutList, Copy, Check, Pin, PinOff, Trash2 } from 'lucide-vue-next'
import { formatDate, relativeTime } from '../utils/format.js'

const props = defineProps({
  historyItems: { type: Array, default: () => [] },
  historyLoading: { type: Boolean, default: false },
  searchQuery: { type: String, default: '' },
  copyId: { type: [Number, String], default: null },
  viewMode: { type: String, default: 'list' },
  sortBy: { type: String, default: 'recent' },
  limit: { type: Number, default: 80 },
  filteredItems: { type: Array, default: () => [] },
  highlightItem: { type: Function, required: true },
  focused: { type: Boolean, default: false },
})
defineEmits([
  'update:searchQuery', 'update:sortBy', 'update:limit', 'update:viewMode',
  'clear-search', 'limit-change', 'search-input', 'search',
  'copy-item', 'toggle-pin', 'delete-item', 'focus', 'blur',
])

const searchInputRef = ref(null)
const historyListRef = ref(null)
defineExpose({ searchInputRef, historyListRef })
</script>

<style scoped>
.panel {
  position: relative;
  border-radius: var(--radius);
  border: 1px solid var(--border);
  background: var(--bg-card);
  overflow: hidden;
  transition: transform 0.35s var(--ease-out), box-shadow 0.35s var(--ease-out), border-color 0.25s;
  animation: panelIn 0.6s var(--ease-out) backwards;
}
.history-panel { animation-delay: 0.16s; }
.panel.focused {
  transform: translateY(-2px);
  box-shadow: 0 24px 48px rgba(0, 0, 0, 0.35), 0 0 0 1px var(--border-strong);
  border-color: var(--accent-dim);
}
.panel-glow { position: absolute; inset: 0; pointer-events: none; }
.panel-inner { position: relative; padding: 1.25rem; }
@keyframes panelIn {
  from { opacity: 0; transform: translateY(16px); }
  to { opacity: 1; transform: translateY(0); }
}
.panel-header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.35rem 0.75rem;
  margin-bottom: 0.85rem;
}
.panel-header h2 { margin: 0; font-size: 1rem; font-weight: 600; color: var(--headline); }
.hint { color: var(--text-muted); font-size: 0.8rem; margin-left: auto; }
.controls-row { display: grid; grid-template-columns: 1fr auto; gap: 0.5rem; margin-bottom: 0.65rem; }
.search-wrap { position: relative; display: flex; align-items: center; }
.search-icon {
  position: absolute;
  left: 0.85rem;
  display: flex;
  flex-shrink: 0;
  color: var(--text-muted);
  pointer-events: none;
}
.search-input {
  width: 100%;
  padding: 0.6rem 2rem 0.6rem 2.25rem;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  color: var(--text);
  font: inherit;
  transition: border-color 0.2s, box-shadow 0.2s;
}
.search-input:focus {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 0 2px var(--accent-soft);
}
.search-clear {
  position: absolute;
  right: 0.5rem;
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  border-radius: 6px;
  color: var(--text-muted);
  cursor: pointer;
  transition: background 0.2s, color 0.2s;
}
.search-clear:hover { background: var(--border); color: var(--text); }
.filter-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 0.6rem;
  margin-bottom: 0.85rem;
}
.select-sm {
  padding: 0.5rem 0.85rem;
  min-width: 0;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 10px;
  color: var(--text);
  font: inherit;
  font-size: 0.875rem;
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%23b0b0b0' stroke-width='2'%3E%3Cpath d='m6 9 6 6 6-6'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 0.6rem center;
  padding-right: 1.75rem;
  transition: border-color 0.2s, color 0.2s;
}
.view-toggle { display: flex; gap: 2px; margin-left: auto; }
.view-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: 1px solid var(--border);
  border-radius: 8px;
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.2s;
}
.view-btn:hover, .view-btn.active {
  background: var(--accent-soft);
  color: var(--accent);
  border-color: var(--accent-dim);
}
.btn { padding: 0.6rem 1.1rem; border: none; border-radius: var(--radius-sm); font: inherit; font-size: 0.9rem; font-weight: 500; cursor: pointer; }
.btn-ghost { background: var(--accent-soft); color: var(--accent); }
.btn-sm { padding: 0.45rem 0.75rem; font-size: 0.85rem; }
.history-list {
  max-height: 56vh;
  overflow: auto;
  padding-right: 6px;
  scrollbar-width: thin;
  scrollbar-color: var(--border) transparent;
}
.history-list::-webkit-scrollbar { width: 8px; }
.history-list::-webkit-scrollbar-thumb { background: var(--border); border-radius: 4px; }
.history-items { display: flex; flex-direction: column; gap: 0.5rem; }
.history-items.compact .history-item { padding: 0.55rem 0.85rem; }
.history-items.compact .item-text { font-size: 0.85rem; }
.history-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem 1rem;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  transition: transform 0.2s var(--ease-out), border-color 0.2s, box-shadow 0.2s;
}
.history-item:hover { border-color: var(--border-strong); box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2); }
.history-item.pinned { border-color: var(--accent-dim); box-shadow: 0 0 0 1px var(--accent-soft); }
.history-item.copied { border-color: var(--accent); box-shadow: 0 0 0 1px var(--accent-soft); }
.item-body { flex: 1; min-width: 0; }
.item-actions { display: flex; align-items: center; gap: 0.3rem; flex-shrink: 0; }
.icon-btn {
  padding: 0.38rem 0.55rem;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 8px;
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.2s var(--ease-out);
}
.icon-btn:hover { color: var(--text); border-color: var(--border-strong); }
.icon-btn.active { background: var(--accent-soft); color: var(--accent); border-color: var(--accent); }
.pin-btn.pinned { background: var(--accent-soft); color: var(--accent); border-color: var(--accent); }
.delete-btn:hover { color: var(--danger); border-color: rgba(255, 68, 102, 0.35); }
.item-text { font-size: 0.9rem; line-height: 1.45; white-space: pre-wrap; word-break: break-word; }
.item-text :deep(.search-highlight) {
  background: var(--accent-soft);
  color: var(--accent);
  padding: 0.1em 0.2em;
  border-radius: 3px;
}
.item-meta { margin-top: 0.5rem; font-size: 0.75rem; color: var(--text-muted); display: flex; gap: 1rem; flex-wrap: wrap; }
.source { font-weight: 500; }
.empty-state { padding: 2.5rem 1.5rem; text-align: center; color: var(--text-muted); }
.empty-icon { display: flex; justify-content: center; margin-bottom: 0.75rem; color: var(--border); }
.empty-hint { font-size: 0.85rem; opacity: 0.85; }
.skeleton-line {
  height: 12px;
  background: var(--border);
  border-radius: 6px;
  opacity: 0.5;
  animation: shimmer 1.2s ease-in-out infinite;
}
.skeleton-line.short { width: 60%; margin-top: 0.5rem; }
.skeleton-item {
  padding: 0.85rem 1rem;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  margin-bottom: 0.5rem;
}
.list-enter-active { transition: all 0.4s var(--ease-out); transition-delay: calc(var(--stagger, 0) * 0.03s); }
.list-leave-active { transition: all 0.3s var(--ease-out); position: absolute; width: 100%; }
.list-enter-from, .list-leave-to { opacity: 0; transform: translateY(12px) scale(0.98); }
.list-move { transition: transform 0.45s var(--ease-spring); }
</style>
