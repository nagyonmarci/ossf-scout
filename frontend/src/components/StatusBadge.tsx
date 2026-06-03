export default function StatusBadge({ status }: { status: string }) {
  return (
    <span className={`badge badge-${status}`}>
      {(status === 'running' || status === 'pending') && <span className="spinner" />}
      {status}
    </span>
  )
}
