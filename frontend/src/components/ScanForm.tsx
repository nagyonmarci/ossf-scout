import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, CreateScanParams } from '../api'

function firstDayOfMonth(): string {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-01`
}

function daysAgo(n: number): string {
  const d = new Date()
  d.setDate(d.getDate() - n)
  return d.toISOString().slice(0, 10)
}

const DATE_PRESETS = [
  { label: 'Any time', value: '' },
  { label: 'This month', value: firstDayOfMonth() },
  { label: 'Last 30 days', value: daysAgo(30) },
  { label: 'Last 90 days', value: daysAgo(90) },
]

const defaults: CreateScanParams = {
  language: '',
  min_stars: 500,
  max_score: 5.0,
  limit: 100,
  workers: 5,
  check_filter: '',
  github_token: '',
  use_cli_fallback: false,
  pushed_after: '',
}

export default function ScanForm() {
  const navigate = useNavigate()
  const [params, setParams] = useState<CreateScanParams>(defaults)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function set<K extends keyof CreateScanParams>(key: K, value: CreateScanParams[K]) {
    setParams(p => ({ ...p, [key]: value }))
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError(null)
    try {
      const scan = await api.createScan(params)
      navigate(`/scans/${scan.id}`)
    } catch (err) {
      setError(String(err))
      setLoading(false)
    }
  }

  return (
    <form onSubmit={submit}>
      <div className="form-grid">
        <div className="form-field">
          <label>Language</label>
          <input
            type="text"
            placeholder="go, python, java…"
            value={params.language}
            onChange={e => set('language', e.target.value)}
          />
        </div>

        <div className="form-field">
          <label>Min Stars</label>
          <input
            type="number"
            min={0}
            value={params.min_stars}
            onChange={e => set('min_stars', Number(e.target.value))}
          />
        </div>

        <div className="form-field">
          <label>Max Score (0–10)</label>
          <input
            type="number"
            min={0}
            max={10}
            step={0.1}
            value={params.max_score}
            onChange={e => set('max_score', Number(e.target.value))}
          />
        </div>

        <div className="form-field">
          <label>Repo Limit</label>
          <input
            type="number"
            min={1}
            max={1000}
            value={params.limit}
            onChange={e => set('limit', Number(e.target.value))}
          />
        </div>

        <div className="form-field">
          <label>Workers</label>
          <input
            type="number"
            min={1}
            max={20}
            value={params.workers}
            onChange={e => set('workers', Number(e.target.value))}
          />
        </div>

        <div className="form-field">
          <label>Check Filter</label>
          <input
            type="text"
            placeholder="SAST,Code-Review,…"
            value={params.check_filter}
            onChange={e => set('check_filter', e.target.value)}
          />
        </div>

        <div className="form-field full-width">
          <label>GitHub Token (optional — uses server GITHUB_TOKEN if empty)</label>
          <input
            type="password"
            placeholder="ghp_…"
            value={params.github_token}
            onChange={e => set('github_token', e.target.value)}
            autoComplete="off"
          />
        </div>

        <div className="form-field full-width">
          <label>Pushed after</label>
          <div className="preset-btns">
            {DATE_PRESETS.map(p => (
              <button
                key={p.label}
                type="button"
                className={params.pushed_after === p.value ? 'preset-btn active' : 'preset-btn'}
                onClick={() => set('pushed_after', p.value)}
              >
                {p.label}
              </button>
            ))}
          </div>
          <input
            type="date"
            value={params.pushed_after}
            onChange={e => set('pushed_after', e.target.value)}
          />
        </div>

        <div className="form-field full-width">
          <label style={{ flexDirection: 'row', alignItems: 'center', gap: 8, cursor: 'pointer' }}>
            <input
              type="checkbox"
              checked={params.use_cli_fallback}
              onChange={e => set('use_cli_fallback', e.target.checked)}
              style={{ width: 'auto' }}
            />
            Use scorecard CLI for unscanned repos (slower — requires <code>scorecard</code> in PATH)
          </label>
        </div>
      </div>

      <div className="form-actions">
        <button className="btn" type="submit" disabled={loading}>
          {loading ? 'Starting…' : 'Run Scan'}
        </button>
        {error && <span className="error-msg">{error}</span>}
      </div>
    </form>
  )
}
