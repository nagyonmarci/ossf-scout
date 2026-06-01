import { useEffect, useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, Scan } from '../api'
import ScanForm from '../components/ScanForm'
import StatusBadge from '../components/StatusBadge'

function formatDate(s: string) {
  return new Date(s).toLocaleString()
}

export default function ScanList() {
  const navigate = useNavigate()
  const [scans, setScans] = useState<Scan[]>([])
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    try {
      const data = await api.listScans()
      setScans(data)
    } catch (e) {
      setError(String(e))
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  // Poll while any scan is running
  useEffect(() => {
    const hasRunning = scans.some(s => s.status === 'running')
    if (!hasRunning) return
    const id = setInterval(load, 5000)
    return () => clearInterval(id)
  }, [scans, load])

  async function deleteScan(e: React.MouseEvent, id: number) {
    e.stopPropagation()
    await api.deleteScan(id)
    setScans(prev => prev.filter(s => s.id !== id))
  }

  return (
    <div className="container">
      <header>
        <h1>ossf-scout</h1>
        <p>Find GitHub repos with weak OpenSSF Scorecard scores</p>
      </header>

      <div className="card">
        <h2>New Scan</h2>
        <ScanForm />
      </div>

      <div className="card">
        <h2>Scan History</h2>
        {error && <p className="error-msg">{error}</p>}
        {scans.length === 0 ? (
          <p className="empty">No scans yet. Run one above.</p>
        ) : (
          <div className="table-wrap">
            <table className="scans-table">
              <thead>
                <tr>
                  <th>#</th>
                  <th>Date</th>
                  <th>Language</th>
                  <th>Min Stars</th>
                  <th>Max Score</th>
                  <th>Limit</th>
                  <th>Status</th>
                  <th>Results</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {scans.map(s => (
                  <tr
                    key={s.id}
                    className="scan-row"
                    onClick={() => navigate(`/scans/${s.id}`)}
                  >
                    <td>{s.id}</td>
                    <td>{formatDate(s.created_at)}</td>
                    <td>{s.language || '(any)'}</td>
                    <td>{s.min_stars.toLocaleString()}</td>
                    <td>{s.max_score}</td>
                    <td>{s.limit}</td>
                    <td><StatusBadge status={s.status} /></td>
                    <td>{s.result_count ?? '—'}</td>
                    <td>
                      <button
                        className="btn btn-danger"
                        onClick={e => deleteScan(e, s.id)}
                      >
                        Delete
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
