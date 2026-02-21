<template>
  <TransitionGroup name="toast" tag="div" class="toast-container">
    <div
      v-for="t in toasts"
      :key="t.id"
      class="toast"
      :class="t.type"
      @click="$emit('remove', t.id)"
    >
      <span class="toast-icon">
        <Check v-if="t.type === 'success'" :size="18" :stroke-width="2.5" />
        <AlertCircle v-else-if="t.type === 'error'" :size="18" :stroke-width="2" />
        <Info v-else :size="18" :stroke-width="2" />
      </span>
      <span class="toast-msg">{{ t.message }}</span>
    </div>
  </TransitionGroup>
</template>

<script setup>
import { Check, AlertCircle, Info } from 'lucide-vue-next'

defineProps({
  toasts: { type: Array, default: () => [] },
})
defineEmits(['remove'])
</script>

<style scoped>
.toast-container {
  position: fixed;
  top: 1rem;
  right: 1rem;
  z-index: 100;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  max-width: 320px;
  pointer-events: none;
}
.toast-container > * { pointer-events: auto; }
.toast {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.65rem 1rem;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
  background: var(--bg-card);
  backdrop-filter: blur(16px);
  font-size: 0.9rem;
  cursor: pointer;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.35);
  animation: toastIn 0.35s var(--ease-spring);
}
.toast.success { border-color: var(--accent-dim); background: var(--bg-card); }
.toast.error { border-color: rgba(255, 68, 102, 0.4); background: var(--bg-card); }
.toast-icon {
  width: 22px;
  height: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  flex-shrink: 0;
}
.toast.success .toast-icon { background: var(--accent-soft); color: var(--accent); }
.toast.error .toast-icon { background: rgba(255, 68, 102, 0.2); color: var(--danger); }
.toast-msg { flex: 1; }
@keyframes toastIn {
  from { opacity: 0; transform: translateX(24px); }
  to { opacity: 1; transform: translateX(0); }
}
.toast-leave-active { transition: all 0.25s var(--ease-out); }
.toast-leave-to { opacity: 0; transform: translateX(24px); }
</style>
