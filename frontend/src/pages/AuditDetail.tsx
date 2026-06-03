import { useEffect, useState, useCallback, useRef } from 'react';
import { Link, useParams, useNavigate } from 'react-router-dom';
import Markdown from 'react-markdown';
import { api, Audit } from '../api';
import StatusBadge from '../components/StatusBadge';

const ANTHROPIC_MODELS = [
  { id: 'claude-opus-4-8',           label: 'Opus 4'   },
  { id: 'claude-sonnet-4-6',         label: 'Sonnet 4' },
  { id: 'claude-haiku-4-5-20251001', label: 'Haiku 4'  },
];

const LS = {
  get: (k: string, fallback: string) => localStorage.getItem(k) ?? fallback,
  set: (k: string, v: string) => localStorage.setItem(k, v),
};

export default function AuditDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [audit, setAudit] = useState<Audit | null>(null);
  const [error, setError] = useState<string | null>(null);
  const prevStatus = useRef<string | null>(null);

  // generate-form state
  const [showGenForm, setShowGenForm] = useState(false);
  const [genProvider, setGenProvider] = useState(() => LS.get('audit.provider', 'ollama'));
  const [genApiKey, setGenApiKey] = useState(() => LS.get('audit.anthropicKey', ''));
  const [genModel, setGenModel] = useState(() => LS.get('audit.model', ANTHROPIC_MODELS[0].id));
  const [genOllamaURL, setGenOllamaURL] = useState(() => LS.get('audit.ollamaURL', ''));
  const [genOllamaModel, setGenOllamaModel] = useState(() => LS.get('audit.ollamaModel', ''));
  const [genSplitGeneration, setGenSplitGeneration] = useState(() => LS.get('audit.splitGeneration', 'false') === 'true');
  const [genAnalysisModel, setGenAnalysisModel] = useState(() => LS.get('audit.analysisModel', ''));
  const [genSubmitting, setGenSubmitting] = useState(false);
  const [genError, setGenError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!id) return;
    try {
      const a = await api.getAudit(id);
      prevStatus.current = a.status;
      setAudit(a);
    } catch (e) {
      setError(String(e));
    }
  }, [id]);

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    if (!audit || (audit.status !== 'pending' && audit.status !== 'running')) return;
    const timer = setInterval(load, 3000);
    return () => clearInterval(timer);
  }, [audit, load]);

  function downloadReport() {
    if (!audit?.report) return;
    const blob = new Blob([audit.report], { type: 'text/markdown' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `audit-${audit.repo.replace('/', '-')}-${audit.created_at.slice(0, 10)}.md`;
    a.click();
    URL.revokeObjectURL(url);
  }

  async function submitGenerate() {
    if (!audit?.id) return;
    setGenSubmitting(true);
    setGenError(null);
    try {
      await api.generateAudit(audit.id, {
        provider: genProvider || undefined,
        anthropic_key: genProvider === 'anthropic' ? genApiKey || undefined : undefined,
        model: genProvider === 'anthropic' ? genModel || undefined
             : genProvider === 'ollama'     ? genOllamaModel || undefined
             : undefined,
        split_generation: genProvider === 'ollama' ? genSplitGeneration : undefined,
        analysis_model: genProvider === 'ollama' && genSplitGeneration ? genAnalysisModel || undefined : undefined,
        ollama_url: genProvider === 'ollama' ? genOllamaURL || undefined : undefined,
      });
      setShowGenForm(false);
      load();
    } catch (e) {
      setGenError(String(e));
    } finally {
      setGenSubmitting(false);
    }
  }

  const canRunWithAI = audit &&
    ((audit.status === 'done' && (!audit.provider || audit.provider === '')) || audit.status === 'error');

  return (
    <div className="container">
      <Link to="/audits" className="back-link">
        ← Back to audits
      </Link>

      {error && <p className="error-msg">{error}</p>}

      {audit && (
        <>
          <div className="card">
            <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 16 }}>
              <h2 style={{ margin: 0 }}>
                Audit: <code>{audit.repo}</code>
              </h2>
              <StatusBadge status={audit.status} />
            </div>

            <div className="detail-meta">
              <span className="meta-item">
                <span className="meta-label">Started:</span>
                {new Date(audit.created_at).toLocaleString()}
              </span>
              {audit.completed_at && (
                <span className="meta-item">
                  <span className="meta-label">Completed:</span>
                  {new Date(audit.completed_at).toLocaleString()}
                </span>
              )}
              <span className="meta-item">
                <span className="meta-label">Mode:</span>
                {audit.provider === 'ollama'
                  ? `Ollama · ${audit.model || '?'}`
                  : audit.provider === 'anthropic'
                  ? audit.model || 'Anthropic'
                  : 'Static snapshot'}
              </span>
              {(audit.input_tokens ?? 0) > 0 && (
                <span className="meta-item">
                  <span className="meta-label">Tokens in/out:</span>
                  {audit.input_tokens!.toLocaleString()} / {(audit.output_tokens ?? 0).toLocaleString()}
                </span>
              )}
            </div>

            {(audit.status === 'pending' || audit.status === 'running') && (
              <p style={{ color: 'var(--muted)', fontSize: 13 }}>
                {audit.status === 'pending'
                  ? 'Waiting to start…'
                  : 'Cloning repo and running analysis — this takes 1–3 minutes…'}
              </p>
            )}
            {audit.status === 'error' && <p className="error-msg">Error: {audit.error}</p>}
            <div style={{ display: 'flex', gap: 8, marginTop: 8, flexWrap: 'wrap' }}>
              {(audit.status === 'done' || (audit.status === 'error' && audit.report)) && (
                <button className="btn btn-primary" onClick={downloadReport}>
                  Download .md
                </button>
              )}
              {audit.has_context && (
                <a className="btn" href={`/api/audits/${audit.id}/context.md`} download>
                  Download AI context
                </a>
              )}
              {canRunWithAI && (
                audit.has_context ? (
                  <button
                    className="btn"
                    style={{ background: 'transparent', border: '1px solid var(--accent)', color: 'var(--accent)' }}
                    onClick={() => setShowGenForm(v => !v)}
                  >
                    Run with AI
                  </button>
                ) : (
                  <button
                    className="btn"
                    style={{ background: 'transparent', border: '1px solid var(--accent)', color: 'var(--accent)' }}
                    onClick={() => navigate(`/audits?repo=${encodeURIComponent(audit!.repo)}`)}
                  >
                    Run with AI
                  </button>
                )
              )}
            </div>

            {showGenForm && audit.has_context && (
              <div style={{ marginTop: 12, padding: 12, border: '1px solid var(--border)', borderRadius: 6 }}>
                <p style={{ margin: '0 0 8px', fontSize: 13, color: 'var(--muted)' }}>
                  Re-generates the report using saved context — no re-cloning needed.
                </p>
                <div style={{ display: 'flex', gap: 6, marginBottom: 10 }}>
                  {(['ollama', 'anthropic', ''] as const).map(p => (
                    <button
                      key={p}
                      className={`btn${genProvider === p ? ' btn-primary' : ''}`}
                      style={{ padding: '4px 12px', fontSize: 13, background: genProvider === p ? undefined : 'transparent', border: '1px solid var(--border)', color: genProvider === p ? undefined : 'var(--muted)' }}
                      onClick={() => { setGenProvider(p); LS.set('audit.provider', p); }}
                    >
                      {p === '' ? 'Static snapshot' : p === 'anthropic' ? 'Anthropic' : 'Ollama'}
                    </button>
                  ))}
                </div>
                {genProvider === 'anthropic' && (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: 8 }}>
                    <input
                      type="password"
                      className="input"
                      placeholder="Anthropic API key"
                      value={genApiKey}
                      onChange={e => { setGenApiKey(e.target.value); LS.set('audit.anthropicKey', e.target.value); }}
                    />
                    <select
                      className="input"
                      value={genModel}
                      onChange={e => { setGenModel(e.target.value); LS.set('audit.model', e.target.value); }}
                    >
                      {ANTHROPIC_MODELS.map(m => (
                        <option key={m.id} value={m.id}>{m.label}</option>
                      ))}
                    </select>
                  </div>
                )}
                {genProvider === 'ollama' && (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: 8 }}>
                    <input
                      type="text"
                      className="input"
                      placeholder="Ollama URL (leave blank for server default)"
                      value={genOllamaURL}
                      onChange={e => { setGenOllamaURL(e.target.value); LS.set('audit.ollamaURL', e.target.value); }}
                    />
                    <input
                      type="text"
                      className="input"
                      placeholder="Final report model (e.g. llama3, qwen2.5)"
                      value={genOllamaModel}
                      onChange={e => { setGenOllamaModel(e.target.value); LS.set('audit.ollamaModel', e.target.value); }}
                    />
                    <label style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                      <input
                        type="checkbox"
                        checked={genSplitGeneration}
                        onChange={e => {
                          setGenSplitGeneration(e.target.checked);
                          LS.set('audit.splitGeneration', String(e.target.checked));
                        }}
                      />
                      <span style={{ fontSize: 13, color: 'var(--muted)' }}>
                        Split generation into section summaries before the final report
                      </span>
                    </label>
                    {genSplitGeneration && (
                      <input
                        type="text"
                        className="input"
                        placeholder="Analysis model (optional — empty uses final model)"
                        value={genAnalysisModel}
                        onChange={e => { setGenAnalysisModel(e.target.value); LS.set('audit.analysisModel', e.target.value); }}
                      />
                    )}
                  </div>
                )}
                <div style={{ display: 'flex', gap: 8 }}>
                  <button className="btn btn-primary" onClick={submitGenerate} disabled={genSubmitting}>
                    {genSubmitting ? 'Starting…' : 'Generate report'}
                  </button>
                  <button
                    className="btn"
                    style={{ background: 'transparent', border: '1px solid var(--border)' }}
                    onClick={() => setShowGenForm(false)}
                  >
                    Cancel
                  </button>
                </div>
                {genError && <p className="error-msg" style={{ marginTop: 6 }}>{genError}</p>}
              </div>
            )}
          </div>

          {audit.report && (audit.status === 'done' || audit.status === 'error') && (
            <div className="card audit-report">
              <Markdown>{audit.report}</Markdown>
            </div>
          )}
        </>
      )}
    </div>
  );
}
