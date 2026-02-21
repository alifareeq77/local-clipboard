import { ref } from 'vue'

export function useToasts() {
  const toasts = ref([])

  function showToast(message, type = 'success') {
    const id = Math.random().toString(36).slice(2)
    toasts.value.push({ id, message, type })
    setTimeout(() => removeToast(id), 4000)
  }

  function removeToast(id) {
    toasts.value = toasts.value.filter((t) => t.id !== id)
  }

  return { toasts, showToast, removeToast }
}
