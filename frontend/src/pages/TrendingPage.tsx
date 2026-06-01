import { useState } from 'react'
import { api, ScanResult } from '../api'
import ResultsTable from '../components/ResultsTable'

const SINCE_OPTIONS = [
  { label: 'Today', value: 'daily' },
  { label: 'This week', value: 'weekly' },
  { label: 'This month', value: 'monthly' },
]

export default function TrendingPage() {
  const [language, setLanguage] = useState('')
  const [since, setSince] = useState('daily')
  const [loading, setLoading] = useState(false)
  const [results, setResults] = useState<ScanResult[] | null>(null)
  const [error, setError] = useState<string | null>(null)

  async function fetchTrending() {
    setLoading(true)
    setError(null)
    try {
      const token = localStorage.getItem('ossf_scout_token') ?? ''
      const data = await api.getTrending({ language: language.trim(), since, token })
      setResults(data)
    } catch (e) {
      setError(String(e))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="container">
      <div className="card">
        <h2>GitHub Trending</h2>
        <p style={{ color: 'var(--muted)', fontSize: 13, marginBottom: 16 }}>
          Scores trending repos from <a href="https://github.com/trending" target="_blank" rel="noopener noreferrer">github.com/trending</a> against the OpenSSF Scorecard API.
          Takes ~20–40s for 25 repos.
        </p>
        <div className="form-grid">
          <div className="form-field">
            <label>Language (optional)</label>
            <input
              type="text"
              placeholder="go, python, typescript…"
              value={language}
              onChange={e => setLanguage(e.target.value)}
              onKeyDown={e => e.key === 'Enter' && fetchTrending()}
            />
          </div>
          <div className="form-field full-width">
            <label>Since</label>
            <div className="preset-btns">
              {SINCE_OPTIONS.map(o => (
                <button
                  key={o.value}
                  type="button"
                  className={since === o.value ? 'preset-btn active' : 'preset-btn'}
                  onClick={() => setSince(o.value)}
                >
                  {o.label}
                </button>
              ))}
            </div>
          </div>
        </div>
        <div className="form-actions">
          <button className="btn" disabled={loading} onClick={fetchTrending}>
            {loading ? 'Loading…' : 'Fetch Trending'}
          </button>
          {error && <span className="error-msg">{error}</span>}
        </div>
      </div>

      {results !== null && (
        <div className="card">
          <h2>Results ({results.length})</h2>
          {results.length === 0
            ? <p className="empty">No results — try a different language or time range.</p>
            : <ResultsTable results={results} />
          }
        </div>
      )}
    </div>
  )
}
