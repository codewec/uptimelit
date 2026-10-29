import toast from 'svelte-french-toast'
export { toast }
export function toastError(error: unknown, fallback = 'Something went wrong') {
  if (error instanceof DOMException && error.name === 'AbortError') return
  toast.error(error instanceof Error ? error.message : fallback)
}
