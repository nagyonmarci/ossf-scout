import { useState, useCallback } from 'react'
import { ToastItem } from '../components/Toast'
import { Scan } from '../api'
import { useLang } from '../i18n'

let nextId = 0

export function useToast() {
  const { t } = useLang()
  const [toasts, setToasts] = useState<ToastItem[]>([])

  const dismiss = useCallback((id: number) => {
    setToasts(prev => prev.filter(toast => toast.id !== id))
  }, [])

  const notify = useCallback((scan: Scan) => {
    const label = scan.language ? `[${scan.language}] ` : ''
    const message = scan.status === 'done'
      ? t('useToast.scanDone', { id: scan.id, label, count: scan.result_count ?? 0 })
      : t('useToast.scanError', { id: scan.id, label })

    setToasts(prev => [...prev, { id: nextId++, message, type: scan.status === 'done' ? 'success' : 'error' }])

    if ('Notification' in window && Notification.permission === 'granted') {
      new Notification('ossf-scout', { body: message, icon: '/favicon.ico' })
    }
  }, [t])

  const requestPermission = useCallback(() => {
    if ('Notification' in window && Notification.permission === 'default') {
      Notification.requestPermission()
    }
  }, [])

  return { toasts, notify, dismiss, requestPermission }
}
