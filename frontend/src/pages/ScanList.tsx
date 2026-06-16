import { useEffect, useState, useCallback, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, Scan, ScanStatus } from '../api'
import ScanForm from '../components/ScanForm'
import StatusBadge from '../components/StatusBadge'
import Toast from '../components/Toast'
import { useToast } from '../hooks/useToast'
import { useLang } from '../i18n'

function formatDate(s: string) {
  return new Date(s).toLocaleString()
}

export default function ScanList() {
  const { t } = useLang()
  const navigate = useNavigate()
  const [scans, setScans] = useState<Scan[]>([])
  const [error, setError] = useState<string | null>(null)
  const { toasts, notify, dismiss, requestPermission } = useToast()
  const prevStatuses = useRef<Map<number, ScanStatus>>(new Map())

  const load = useCallback(async () => {
    try {
      const data = await api.listScans()
      // detect running → done/error transitions
      data.forEach(scan => {
        const prev = prevStatuses.current.get(scan.id)
        if (prev === 'running' && scan.status !== 'running') notify(scan)
      })
      prevStatuses.current = new Map(data.map(s => [s.id, s.status]))
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
      <div className="card">
        <h2>{t('scanList.newScan')}</h2>
        <ScanForm onScanStarted={requestPermission} />
      </div>

      <div className="card">
        <h2>{t('scanList.scanHistory')}</h2>
        {error && <p className="error-msg">{error}</p>}
        {scans.length === 0 ? (
          <p className="empty">{t('scanList.noScans')}</p>
        ) : (
          <div className="table-wrap">
            <table className="scans-table">
              <thead>
                <tr>
                  <th>#</th>
                  <th>{t('scanList.colDate')}</th>
                  <th>{t('scanList.colLanguage')}</th>
                  <th>{t('scanList.colMinStars')}</th>
                  <th>{t('scanList.colMaxScore')}</th>
                  <th>{t('scanList.colLimit')}</th>
                  <th>{t('scanList.colStatus')}</th>
                  <th>{t('scanList.colResults')}</th>
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
                    <td>{s.language || t('common.any')}</td>
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
                        {t('common.delete')}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
      <Toast toasts={toasts} onDismiss={dismiss} />
    </div>
  )
}
