import { useEffect, useState, useCallback, useRef } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api, Scan, ScanResult } from '../api'
import StatusBadge from '../components/StatusBadge'
import ResultsTable from '../components/ResultsTable'
import Toast from '../components/Toast'
import { useToast } from '../hooks/useToast'
import { useLang } from '../i18n'

export default function ScanDetail() {
  const { t } = useLang()
  const { id } = useParams<{ id: string }>()
  const scanId = Number(id)

  const [scan, setScan] = useState<Scan | null>(null)
  const [results, setResults] = useState<ScanResult[]>([])
  const [error, setError] = useState<string | null>(null)
  const { toasts, notify, dismiss } = useToast()
  const prevStatus = useRef<string | null>(null)

  const loadScan = useCallback(async () => {
    try {
      const s = await api.getScan(scanId)
      if (prevStatus.current === 'running' && s.status !== 'running') notify(s)
      prevStatus.current = s.status
      setScan(s)
      if (s.status === 'done') {
        const r = await api.getResults(scanId)
        setResults(r)
      }
    } catch (e) {
      setError(String(e))
    }
  }, [scanId, notify])

  useEffect(() => {
    loadScan()
  }, [loadScan])

  // Poll while running
  useEffect(() => {
    if (!scan || scan.status !== 'running') return
    const timer = setInterval(loadScan, 3000)
    return () => clearInterval(timer)
  }, [scan, loadScan])

  return (
    <div className="container">
      <Link to="/" className="back-link">{t('scanDetail.backToScans')}</Link>

      {error && <p className="error-msg">{error}</p>}

      {scan && (
        <>
          <div className="card">
            <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 16 }}>
              <h2 style={{ margin: 0 }}>{t('scanDetail.scanHeading', { id: scan.id })}</h2>
              <StatusBadge status={scan.status} />
            </div>
            <div className="detail-meta">
              <span className="meta-item">
                <span className="meta-label">{t('scanDetail.language')}</span>
                {scan.language || t('common.any')}
              </span>
              <span className="meta-item">
                <span className="meta-label">{t('scanDetail.minStars')}</span>
                {scan.min_stars.toLocaleString()}
              </span>
              <span className="meta-item">
                <span className="meta-label">{t('scanDetail.maxScore')}</span>
                {scan.max_score}
              </span>
              <span className="meta-item">
                <span className="meta-label">{t('scanDetail.limit')}</span>
                {scan.limit}
              </span>
              <span className="meta-item">
                <span className="meta-label">{t('scanDetail.workers')}</span>
                {scan.workers}
              </span>
              {scan.topic && (
                <span className="meta-item">
                  <span className="meta-label">{t('scanDetail.topic')}</span>
                  {scan.topic}
                </span>
              )}
              {scan.keyword && (
                <span className="meta-item">
                  <span className="meta-label">{t('scanDetail.keyword')}</span>
                  {scan.keyword}
                </span>
              )}
              {scan.check_filter && (
                <span className="meta-item">
                  <span className="meta-label">{t('scanDetail.checks')}</span>
                  {scan.check_filter}
                </span>
              )}
              <span className="meta-item">
                <span className="meta-label">{t('scanDetail.started')}</span>
                {new Date(scan.created_at).toLocaleString()}
              </span>
              {scan.finished_at && (
                <span className="meta-item">
                  <span className="meta-label">{t('scanDetail.finished')}</span>
                  {new Date(scan.finished_at).toLocaleString()}
                </span>
              )}
            </div>
            {scan.status === 'running' && (
              <p style={{ color: 'var(--muted)', fontSize: 13 }}>
                {t('scanDetail.inProgress')}
              </p>
            )}
            {scan.status === 'error' && (
              <p className="error-msg">{t('scanDetail.error', { msg: scan.error_msg ?? '' })}</p>
            )}
            {scan.status === 'done' && (
              <p style={{ fontSize: 13, color: 'var(--muted)' }}>
                {t('scanDetail.foundSummary', { found: scan.result_count ?? 0, total: scan.total_repos ?? 0 })}
              </p>
            )}
          </div>

          {scan.status === 'done' && results.length > 0 && (
            <div className="card">
              <h2>{t('scanDetail.resultsHeading', { count: results.length })}</h2>
              <ResultsTable results={results} />
            </div>
          )}
        </>
      )}
      <Toast toasts={toasts} onDismiss={dismiss} />
    </div>
  )
}
