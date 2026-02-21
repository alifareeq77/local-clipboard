<template>
  <div class="shortcuts-overlay" @click="$emit('close')">
    <div class="shortcuts-modal" @click.stop>
      <h3>Shortcuts</h3>
      <ul>
        <li><kbd>Ctrl</kbd>+<kbd>Enter</kbd> Send</li>
        <li><kbd>Esc</kbd> Clear focus</li>
        <li><kbd>/</kbd> Focus search</li>
        <li><kbd>?</kbd> This help</li>
      </ul>
      <button class="btn btn-ghost" @click="$emit('close')">Close</button>
    </div>
  </div>
</template>

<script setup>
defineEmits(['close'])
</script>

<style scoped>
.shortcuts-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  z-index: 200;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  animation: fadeIn 0.2s;
}
.shortcuts-modal {
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 1.5rem;
  max-width: 320px;
  box-shadow: 0 24px 48px rgba(0, 0, 0, 0.4);
  animation: panelIn 0.3s var(--ease-spring);
}
.shortcuts-modal h3 { margin: 0 0 1rem; font-size: 1rem; color: var(--headline); }
.shortcuts-modal ul { margin: 0 0 1rem; padding: 0; list-style: none; }
.shortcuts-modal li {
  padding: 0.35rem 0;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.9rem;
}
.shortcuts-modal kbd {
  padding: 0.2rem 0.4rem;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 4px;
  font-size: 0.8rem;
  font-family: inherit;
}
.btn { padding: 0.6rem 1.1rem; border: none; border-radius: var(--radius-sm); font: inherit; cursor: pointer; }
.btn-ghost { background: var(--accent-soft); color: var(--accent); }
@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
@keyframes panelIn {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
