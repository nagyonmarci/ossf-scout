import { useState } from 'react'
import { api, ScanResult } from '../api'
import ResultsTable from '../components/ResultsTable'
import { useLang } from '../i18n'

export default function TrendingPage() {
  const { t } = useLang()
  const SINCE_OPTIONS = [
    { label: t('trending.sinceToday'), value: 'daily' },
    { label: t('trending.sinceWeek'), value: 'weekly' },
    { label: t('trending.sinceMonth'), value: 'monthly' },
  ]
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
        <h2>{t('trending.heading')}</h2>
        <p style={{ color: 'var(--muted)', fontSize: 13, marginBottom: 16 }}>
          {t('trending.descriptionPrefix')} <a href="https://github.com/trending" target="_blank" rel="noopener noreferrer">github.com/trending</a> {t('trending.descriptionSuffix')}
        </p>
        <div className="form-grid">
          <div className="form-field">
            <label>{t('trending.languageLabel')}</label>
            <input
              type="text"
              placeholder="go, python, typescript…"
              value={language}
              onChange={e => setLanguage(e.target.value)}
              onKeyDown={e => e.key === 'Enter' && fetchTrending()}
            />
          </div>
          <div className="form-field full-width">
            <label>{t('trending.sinceLabel')}</label>
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
            {loading ? t('common.loading') : t('trending.fetchButton')}
          </button>
          {error && <span className="error-msg">{error}</span>}
        </div>
      </div>

      {results !== null && (
        <div className="card">
          <h2>{t('trending.resultsHeading', { count: results.length })}</h2>
          {results.length === 0
            ? <p className="empty">{t('trending.noResults')}</p>
            : <ResultsTable results={results} />
          }
        </div>
      )}
    </div>
  )
}
