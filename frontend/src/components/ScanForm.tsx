import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, CreateScanParams } from '../api'
import { useLang } from '../i18n'

function firstDayOfMonth(): string {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-01`
}

function daysAgo(n: number): string {
  const d = new Date()
  d.setDate(d.getDate() - n)
  return d.toISOString().slice(0, 10)
}

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

export default function ScanForm({ onScanStarted }: { onScanStarted?: () => void } = {}) {
  const { t } = useLang()
  const navigate = useNavigate()

  const DATE_PRESETS = [
    { label: t('scanForm.anyTime'), value: '' },
    { label: t('scanForm.thisMonth'), value: firstDayOfMonth() },
    { label: t('scanForm.last30Days'), value: daysAgo(30) },
    { label: t('scanForm.last90Days'), value: daysAgo(90) },
  ]

  const PRESETS: { label: string; hint: string; params: Partial<CreateScanParams> }[] = [
    {
      label: t('scanForm.presetDevSecOps'),
      hint: t('scanForm.presetDevSecOpsHint'),
      params: { ...BASE_DEVSECOPS },
    },
    {
      label: t('scanForm.presetAiLlm'),
      hint: t('scanForm.presetAiLlmHint'),
      params: { ...BASE_DEVSECOPS, topic: 'llm', keyword: '' },
    },
    {
      label: t('scanForm.presetMcpAgent'),
      hint: t('scanForm.presetMcpAgentHint'),
      params: { ...BASE_DEVSECOPS, topic: 'mcp', keyword: '' },
    },
    {
      label: t('scanForm.presetCloudNative'),
      hint: t('scanForm.presetCloudNativeHint'),
      params: { ...BASE_DEVSECOPS, topic: 'kubernetes', keyword: '' },
    },
    {
      label: t('scanForm.presetSecurityTooling'),
      hint: t('scanForm.presetSecurityToolingHint'),
      params: { ...BASE_DEVSECOPS, topic: 'security', min_stars: 200 },
    },
  ]

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
        <div className="preset-label">{t('scanForm.quickPresets')}</div>
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
        <label>{t('scanForm.singleRepoLabel')}</label>
        <input
          type="text"
          placeholder="e.g. torvalds/linux"
          value={params.single_repo}
          onChange={e => set('single_repo', e.target.value)}
        />
      </div>

      <div className="form-grid">
        <div className="form-field">
          <label>{t('scanForm.languageLabel')}</label>
          <input
            type="text"
            placeholder="go, python, java…"
            value={params.language}
            onChange={e => set('language', e.target.value)}
          />
        </div>

        <div className="form-field">
          <label>{t('scanForm.topicLabel')}</label>
          <input
            type="text"
            placeholder="ai, machine-learning…"
            value={params.topic}
            onChange={e => set('topic', e.target.value)}
          />
        </div>

        <div className="form-field">
          <label>{t('scanForm.keywordLabel')}</label>
          <input
            type="text"
            placeholder="llm, kubernetes…"
            value={params.keyword}
            onChange={e => set('keyword', e.target.value)}
          />
        </div>

        <div className="form-field">
          <label>{t('scanForm.minStarsLabel')}</label>
          <input
            type="number"
            min={0}
            value={params.min_stars}
            onChange={e => set('min_stars', Number(e.target.value))}
          />
          <small>{t('scanForm.minStarsHint')}</small>
        </div>

        <div className="form-field">
          <label>{t('scanForm.maxScoreLabel')}</label>
          <input
            type="number"
            min={0}
            max={10}
            step={0.1}
            value={params.max_score}
            onChange={e => set('max_score', Number(e.target.value))}
          />
          <small>{t('scanForm.maxScoreHint')}</small>
        </div>

        <div className="form-field">
          <label>{t('scanForm.repoLimitLabel')}</label>
          <input
            type="number"
            min={1}
            max={1000}
            value={params.limit}
            onChange={e => set('limit', Number(e.target.value))}
          />
          <small>{t('scanForm.repoLimitHint')}</small>
        </div>

        <div className="form-field">
          <label>{t('scanForm.workersLabel')}</label>
          <input
            type="number"
            min={1}
            max={20}
            value={params.workers}
            onChange={e => set('workers', Number(e.target.value))}
          />
          <small>{t('scanForm.workersHint')}</small>
        </div>

        <div className="form-field">
          <label>{t('scanForm.minMaintainedLabel')}</label>
          <input
            type="number"
            min={0}
            max={10}
            value={params.min_maintained}
            onChange={e => set('min_maintained', Number(e.target.value))}
          />
          <small>{t('scanForm.minMaintainedHint')}</small>
        </div>

        <div className="form-field">
          <label>{t('scanForm.checkFilterLabel')}</label>
          <input
            type="text"
            placeholder="SAST,Code-Review,…"
            value={params.check_filter}
            onChange={e => set('check_filter', e.target.value)}
          />
          <small>{t('scanForm.checkFilterHint')}</small>
        </div>

        <div className="form-field full-width">
          <label>{t('scanForm.githubTokenLabel')}</label>
          <input
            type="password"
            placeholder="ghp_…"
            value={params.github_token}
            onChange={e => set('github_token', e.target.value)}
            autoComplete="off"
          />
        </div>

        <div className="form-field full-width">
          <label>{t('scanForm.pushedAfterLabel')}</label>
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
            {t('scanForm.useCliFallback')} <code>scorecard</code> {t('scanForm.useCliFallbackSuffix')}
          </label>
        </div>
      </div>

      <div className="form-actions">
        <button className="btn" type="submit" disabled={loading}>
          {loading ? t('common.starting') : t('scanForm.runScan')}
        </button>
        <button className="btn btn-danger" type="button" onClick={resetDefaults} disabled={loading}>
          {t('common.reset')}
        </button>
        {error && <span className="error-msg">{error}</span>}
      </div>
    </form>
  )
}
