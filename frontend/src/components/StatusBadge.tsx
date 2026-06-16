import { useLang } from '../i18n'

export default function StatusBadge({ status }: { status: string }) {
  const { t } = useLang()
  const labels: Record<string, string> = {
    running: t('statusBadge.running'),
    pending: t('statusBadge.pending'),
    done: t('statusBadge.done'),
    error: t('statusBadge.error'),
  }
  return (
    <span className={`badge badge-${status}`}>
      {(status === 'running' || status === 'pending') && <span className="spinner" />}
      {labels[status] ?? status}
    </span>
  )
}
