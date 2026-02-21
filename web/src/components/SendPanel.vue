<template>
  <section
    class="panel send-panel"
    :class="{ focused: focused, 'has-content': inputText.trim() }"
    @focusin="$emit('focus')"
    @focusout="$emit('blur')"
  >
    <div class="panel-glow"></div>
    <div class="panel-inner">
      <div class="panel-header">
        <h2>Send text</h2>
        <span class="hint">Paste to sync</span>
        <span v-if="inputText.trim()" class="char-count">{{ inputText.length }} chars</span>
      </div>
      <textarea
        ref="sendTextareaRef"
        :value="inputText"
        class="textarea"
        placeholder="Paste or type to sync across devices…"
        rows="4"
        @input="$emit('update:inputText', ($event.target).value)"
        @keydown.meta.enter.exact.prevent="onSend"
        @keydown.ctrl.enter.exact.prevent="onSend"
      />
      <div class="toolbar">
        <button class="btn btn-ghost" type="button" title="Paste from clipboard" @click="$emit('paste')">
          <ClipboardPaste :size="18" :stroke-width="2" />
          Paste
        </button>
        <button
          class="btn btn-primary"
          :class="{ sending }"
          :disabled="sending || !inputText.trim()"
          @click="onSend"
        >
          <span class="btn-content">
            <span v-if="sending" class="btn-loader"></span>
            {{ sending ? 'Sending…' : 'Send' }}
          </span>
        </button>
        <button class="btn btn-ghost" @click="$emit('clear')">
          Clear
        </button>
      </div>
      <p v-if="sendStatus" class="status" :class="{ error: sendError }">{{ sendStatus }}</p>

      <div class="panel-header latest-header">
        <h2>
          Latest
          <span v-if="latestUpdated" class="pulse-dot" title="Just updated"></span>
        </h2>
        <span class="hint">Current clipboard</span>
        <div class="panel-header-actions">
          <button
            v-if="latest?.text"
            class="icon-btn-small"
            title="Copy latest"
            @click="$emit('copy-latest', latest.text)"
          >
            <Copy :size="16" :stroke-width="2" />
          </button>
        </div>
      </div>
      <div
        class="latest-box"
        :class="{ empty: !latest?.text, 'latest-loading': latestLoading, pulse: latestUpdated }"
      >
        <template v-if="latestLoading">
          <div class="skeleton latest-skeleton">
            <div class="skeleton-line"></div>
            <div class="skeleton-line short"></div>
          </div>
        </template>
        <template v-else-if="latest?.text">{{ latest.text }}</template>
        <template v-else>
          <span class="empty-placeholder">Clipboard is empty</span>
        </template>
      </div>
    </div>
  </section>
</template>

<script setup>
import { ref } from 'vue'
import { ClipboardPaste, Copy } from 'lucide-vue-next'

defineProps({
  inputText: { type: String, default: '' },
  sending: { type: Boolean, default: false },
  sendStatus: { type: String, default: '' },
  sendError: { type: Boolean, default: false },
  latest: { type: Object, default: null },
  latestLoading: { type: Boolean, default: false },
  latestUpdated: { type: Boolean, default: false },
  focused: { type: Boolean, default: false },
})
const emit = defineEmits(['update:inputText', 'send', 'paste', 'clear', 'copy-latest', 'focus', 'blur'])

const sendTextareaRef = ref(null)
function onSend() {
  emit('send', sendTextareaRef.value)
}
defineExpose({ sendTextareaRef })
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
.send-panel { animation-delay: 0.08s; }
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
.panel-header h2 {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 0.35rem;
  color: var(--headline);
}
.hint { color: var(--text-muted); font-size: 0.8rem; margin-left: auto; }
.panel-header-actions { display: flex; align-items: center; gap: 0.35rem; margin-left: auto; }
.char-count { font-size: 0.75rem; color: var(--text-muted); }
.latest-header { margin-top: 1.35rem; }
.icon-btn-small {
  padding: 0.35rem 0.5rem;
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
.icon-btn-small:hover { color: var(--accent); border-color: var(--accent-dim); }
.textarea {
  width: 100%;
  min-height: 108px;
  padding: 0.85rem 1rem;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  color: var(--text);
  font: inherit;
  resize: vertical;
  transition: border-color 0.25s var(--ease-out), box-shadow 0.25s var(--ease-out);
}
.textarea:focus {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 0 3px var(--accent-soft);
}
.textarea::placeholder { color: var(--text-muted); }
.toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 0.6rem;
  margin-top: 0.75rem;
}
.btn {
  padding: 0.6rem 1.1rem;
  border: none;
  border-radius: var(--radius-sm);
  font: inherit;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: transform 0.15s var(--ease-out), background 0.2s, box-shadow 0.2s;
  position: relative;
}
.btn:active { transform: scale(0.98); }
.btn:disabled { opacity: 0.55; cursor: not-allowed; transform: none; }
.btn-primary { background: var(--accent); color: #000; }
.btn-primary:not(:disabled):hover { background: var(--text-muted); color: #000; }
.btn-primary .btn-content { display: inline-flex; align-items: center; gap: 0.4rem; }
.btn-loader {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(0,0,0,0.2);
  border-top-color: #000;
  border-radius: 50%;
  animation: btnSpin 0.7s linear infinite;
}
@keyframes btnSpin { to { transform: rotate(360deg); } }
.btn-ghost { background: var(--accent-soft); color: var(--accent); }
.btn-ghost:hover { background: var(--accent-dim); }
.toolbar .btn-ghost:first-child { display: inline-flex; align-items: center; gap: 0.4rem; }
.status { margin: 0.5rem 0 0; font-size: 0.85rem; color: var(--text-muted); }
.status.error { color: var(--danger); }
.latest-box {
  padding: 0.9rem 1rem;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 0.9rem;
  line-height: 1.45;
  transition: border-color 0.3s, box-shadow 0.3s;
}
.latest-box.empty .empty-placeholder { color: var(--text-muted); font-style: italic; }
.latest-box.pulse { border-color: var(--accent); animation: latestPulse 2s var(--ease-out); }
@keyframes latestPulse {
  0% { box-shadow: 0 0 0 1px var(--accent); }
  100% { box-shadow: none; }
}
.pulse-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--accent);
  animation: dotPulse 1.5s var(--ease-out) infinite;
  margin-left: 2px;
}
@keyframes dotPulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.6; transform: scale(0.9); }
}
.skeleton-line {
  height: 12px;
  background: var(--border);
  border-radius: 6px;
  opacity: 0.5;
  animation: shimmer 1.2s ease-in-out infinite;
}
@keyframes shimmer { 0%, 100% { opacity: 0.4; } 50% { opacity: 0.7; } }
.skeleton-line.short { width: 60%; margin-top: 0.5rem; }
.latest-skeleton { padding: 0; }
</style>
