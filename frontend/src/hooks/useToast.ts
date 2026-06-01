import { useState, useCallback } from 'react'
import { ToastItem } from '../components/Toast'
import { Scan } from '../api'

let nextId = 0

export function useToast() {
  const [toasts, setToasts] = useState<ToastItem[]>([])

  const dismiss = useCallback((id: number) => {
    setToasts(prev => prev.filter(t => t.id !== id))
  }, [])

  const notify = useCallback((scan: Scan) => {
    const label = scan.language ? `[${scan.language}] ` : ''
    const message = scan.status === 'done'
      ? `✓ Scan #${scan.id} ${label}kész — ${scan.result_count ?? 0} találat`
      : `✗ Scan #${scan.id} ${label}hibával végzett`

    setToasts(prev => [...prev, { id: nextId++, message, type: scan.status === 'done' ? 'success' : 'error' }])

    if ('Notification' in window && Notification.permission === 'granted') {
      new Notification('ossf-scout', { body: message, icon: '/favicon.ico' })
    }
  }, [])

  const requestPermission = useCallback(() => {
    if ('Notification' in window && Notification.permission === 'default') {
      Notification.requestPermission()
    }
  }, [])

  return { toasts, notify, dismiss, requestPermission }
}
