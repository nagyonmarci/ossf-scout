import { ScanStatus } from '../api'

export default function StatusBadge({ status }: { status: ScanStatus }) {
  return (
    <span className={`badge badge-${status}`}>
      {status === 'running' && <span className="spinner" />}
      {status}
    </span>
  )
}
