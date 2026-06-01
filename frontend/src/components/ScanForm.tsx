import { useState, useEffect } from 'react'
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
  min_maintained: 0,
  topic: '',
  keyword: '',
  single_repo: '',
}

const BASE_DEVSECOPS: Partial<CreateScanParams> = {
  check_filter: 'CI-Tests,SAST,Dependency-Update-Tool,Pinned-Dependencies,Branch-Protection,Code-Review',
  min_maintained: 3,
  max_score: 7.0,
  language: '',
  keyword: '',
  topic: '',
}

const PRESETS: { label: string; hint: string; params: Partial<CreateScanParams> }[] = [
  {
    label: 'DevSecOps opportunities',
    hint: 'Actively maintained repos missing key CI/CD security checks',
    params: { ...BASE_DEVSECOPS },
  },
  {
    label: 'AI / LLM ecosystem',
    hint: 'AI and LLM projects with weak pipelines — high visibility contributions',
    params: { ...BASE_DEVSECOPS, topic: 'llm', keyword: '' },
  },
  {
    label: 'MCP / Agent tools',
    hint: 'MCP and agent infrastructure projects — fast-growing ecosystem',
    params: { ...BASE_DEVSECOPS, topic: 'mcp', keyword: '' },
  },
  {
    label: 'Cloud Native',
    hint: 'Kubernetes and cloud native projects with security gaps',
    params: { ...BASE_DEVSECOPS, topic: 'kubernetes', keyword: '' },
  },
  {
    label: 'Security tooling',
    hint: 'Security tools that ironically have weak pipelines themselves',
    params: { ...BASE_DEVSECOPS, topic: 'security', min_stars: 200 },
  },
]

export default function ScanForm({ onScanStarted }: { onScanStarted?: () => void } = {}) {
  const navigate = useNavigate()
  const [params, setParams] = useState<CreateScanParams>(() => ({
    ...defaults,
    github_token: localStorage.getItem('ossf_scout_token') ?? '',
  }))
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (params.github_token) {
      localStorage.setItem('ossf_scout_token', params.github_token)
    } else {
      localStorage.removeItem('ossf_scout_token')
    }
  }, [params.github_token])

  function set<K extends keyof CreateScanParams>(key: K, value: CreateScanParams[K]) {
    setParams(p => ({ ...p, [key]: value }))
  }

  const [activePreset, setActivePreset] = useState<string | null>(null)

  function applyPreset(preset: typeof PRESETS[number]) {
    setParams(p => ({ ...p, ...preset.params }))
    setActivePreset(preset.label)
  }

  function resetDefaults() {
    setParams(p => ({ ...defaults, github_token: p.github_token }))
    setActivePreset(null)
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError(null)
    try {
      onScanStarted?.()
      const scan = await api.createScan(params)
      navigate(`/scans/${scan.id}`)
    } catch (err) {
      setError(String(err))
      setLoading(false)
    }
  }

  return (
    <form onSubmit={submit}>
      <div className="preset-section">
        <div className="preset-label">Quick presets</div>
        <div className="preset-list">
          {PRESETS.map(p => (
            <button
              key={p.label}
              type="button"
              className={activePreset === p.label ? 'btn btn-preset active' : 'btn btn-preset'}
              onClick={() => applyPreset(p)}
              title={p.hint}
            >
              {p.label}
            </button>
          ))}
        </div>
        {activePreset && (
          <p className="preset-hint">
            {PRESETS.find(p => p.label === activePreset)?.hint}
          </p>
        )}
      </div>

      <div className="form-field" style={{ marginBottom: 16 }}>
        <label>Single repo (owner/repo) — skips search if filled</label>
        <input
          type="text"
          placeholder="e.g. torvalds/linux"
          value={params.single_repo}
          onChange={e => set('single_repo', e.target.value)}
        />
      </div>

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
          <label>Topic</label>
          <input
            type="text"
            placeholder="ai, machine-learning…"
            value={params.topic}
            onChange={e => set('topic', e.target.value)}
          />
        </div>

        <div className="form-field">
          <label>Keyword</label>
          <input
            type="text"
            placeholder="llm, kubernetes…"
            value={params.keyword}
            onChange={e => set('keyword', e.target.value)}
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
          <small>Minimum GitHub star count — higher = more popular repos</small>
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
          <small>Maximum OpenSSF total score to include — lower is stricter, 7.0 catches partial gaps</small>
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
          <small>Max repos fetched from GitHub Search (100 per page)</small>
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
          <small>Concurrent Scorecard API queries — 2–5 recommended to avoid rate limits</small>
        </div>

        <div className="form-field">
          <label>Min Maintained score</label>
          <input
            type="number"
            min={0}
            max={10}
            value={params.min_maintained}
            onChange={e => set('min_maintained', Number(e.target.value))}
          />
          <small>0 = no filter; excludes abandoned repos (Scorecard Maintained check, 3–7 is a practical range)</small>
        </div>

        <div className="form-field">
          <label>Check Filter</label>
          <input
            type="text"
            placeholder="SAST,Code-Review,…"
            value={params.check_filter}
            onChange={e => set('check_filter', e.target.value)}
          />
          <small>Comma-separated check names to highlight — leave empty to use all security checks</small>
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
            type="text"
            placeholder="YYYY-MM-DD"
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
        <button className="btn btn-danger" type="button" onClick={resetDefaults} disabled={loading}>
          Reset
        </button>
        {error && <span className="error-msg">{error}</span>}
      </div>
    </form>
  )
}
