import { useEffect } from 'react'

export interface ToastItem {
  id: number
  message: string
  type: 'success' | 'error'
}

interface Props {
  toasts: ToastItem[]
  onDismiss: (id: number) => void
}

export default function Toast({ toasts, onDismiss }: Props) {
  return (
    <div className="toast-container">
      {toasts.map(t => (
        <ToastEntry key={t.id} toast={t} onDismiss={onDismiss} />
      ))}
    </div>
  )
}

function ToastEntry({ toast, onDismiss }: { toast: ToastItem; onDismiss: (id: number) => void }) {
  useEffect(() => {
    const timer = setTimeout(() => onDismiss(toast.id), 5000)
    return () => clearTimeout(timer)
  }, [toast.id, onDismiss])

  return (
    <div className={`toast toast-${toast.type}`} onClick={() => onDismiss(toast.id)}>
      {toast.message}
    </div>
  )
}
