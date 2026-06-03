import { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { api, Audit, AuditStatus } from '../api';
import StatusBadge from '../components/StatusBadge';

const ANTHROPIC_MODELS = [
  { id: 'claude-opus-4-8',           label: 'Opus 4',   hint: '~$0.50–$1.50/run', inputRate: 15,   outputRate: 75  },
  { id: 'claude-sonnet-4-6',         label: 'Sonnet 4', hint: '~$0.10–$0.30/run', inputRate: 3,    outputRate: 15  },
  { id: 'claude-haiku-4-5-20251001', label: 'Haiku 4',  hint: '~$0.02–$0.05/run', inputRate: 0.80, outputRate: 4   },
];

function formatDate(s: string) {
  return new Date(s).toLocaleString();
}

function costEstimate(provider: string, model: string, inputTokens: number | null, outputTokens: number | null): string {
  if (provider === 'ollama') return 'free (local)';
  if (!inputTokens && !outputTokens) return '—';
  const m = ANTHROPIC_MODELS.find((x) => x.id === model);
  const inputRate  = m?.inputRate  ?? 15;
  const outputRate = m?.outputRate ?? 75;
  const cost = ((inputTokens ?? 0) / 1_000_000) * inputRate
             + ((outputTokens ?? 0) / 1_000_000) * outputRate;
  return `~$${cost.toFixed(3)}`;
}

function modeLabel(provider: string, model: string): string {
  if (provider === 'ollama') return model ? `Ollama · ${model}` : 'Ollama';
  if (provider === 'anthropic') {
    const m = ANTHROPIC_MODELS.find((x) => x.id === model);
    return m?.label ?? model ?? 'Anthropic';
  }
  return 'Snapshot';
}

const LS = {
  get: (k: string, fallback: string) => localStorage.getItem(k) ?? fallback,
  set: (k: string, v: string) => localStorage.setItem(k, v),
};

export default function AuditPage() {
  const navigate = useNavigate();
  const [audits, setAudits] = useState<Audit[]>([]);
  const [repo, setRepo] = useState('');
  const [token, setToken] = useState(() => LS.get('audit.ghToken', ''));
  const [provider, setProvider] = useState(() => LS.get('audit.provider', ''));
  const [anthropicKey, setAnthropicKey] = useState(() => LS.get('audit.anthropicKey', ''));
  const [anthropicModel, setAnthropicModel] = useState(() => LS.get('audit.model', ANTHROPIC_MODELS[0].id));
  const [ollamaURL, setOllamaURL] = useState(() => LS.get('audit.ollamaURL', 'http://localhost:11434'));
  const [ollamaModel, setOllamaModel] = useState(() => LS.get('audit.ollamaModel', ''));
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    try { setAudits(await api.listAudits()); }
    catch (e) { setError(String(e)); }
  }, []);

  useEffect(() => { load(); }, [load]);

  useEffect(() => {
    const active = audits.some((a) => a.status === 'pending' || a.status === 'running');
    if (!active) return;
    const id = setInterval(load, 5000);
    return () => clearInterval(id);
  }, [audits, load]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!repo.trim()) return;
    setSubmitting(true);
    setError(null);
    try {
      await api.createAudit({
        repo: repo.trim(),
        github_token: token || undefined,
        provider: provider || undefined,
        anthropic_key: provider === 'anthropic' ? (anthropicKey || undefined) : undefined,
        model: provider === 'anthropic' ? anthropicModel : provider === 'ollama' ? ollamaModel : undefined,
        ollama_url: provider === 'ollama' ? (ollamaURL || undefined) : undefined,
      });
      setRepo('');
      load();
    } catch (e) {
      setError(String(e));
    } finally {
      setSubmitting(false);
    }
  }

  async function deleteAudit(e: React.MouseEvent, id: string) {
    e.stopPropagation();
    await api.deleteAudit(id);
    setAudits((prev) => prev.filter((a) => a.id !== id));
  }

  const selectedAnthropicModel = ANTHROPIC_MODELS.find((m) => m.id === anthropicModel);

  return (
    <div className="container">
      <div className="card">
        <h2>New Audit</h2>
        <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <label>
            <span style={{ fontSize: 13, color: 'var(--muted)' }}>Target repository (owner/name)</span>
            <input
              type="text"
              className="input"
              placeholder="e.g. directus/directus"
              value={repo}
              onChange={(e) => setRepo(e.target.value)}
              required
              style={{ marginTop: 4 }}
            />
          </label>

          {/* Provider */}
          <div>
            <span style={{ fontSize: 13, color: 'var(--muted)' }}>AI provider</span>
            <div style={{ display: 'flex', gap: 8, marginTop: 4 }}>
              {(['', 'anthropic', 'ollama'] as const).map((p) => (
                <button
                  key={p}
                  type="button"
                  className={`btn${provider === p ? ' btn-primary' : ''}`}
                  style={{ padding: '5px 14px', fontSize: 13, background: provider === p ? undefined : 'transparent', border: '1px solid var(--border)', color: provider === p ? undefined : 'var(--muted)' }}
                  onClick={() => { setProvider(p); LS.set('audit.provider', p); }}
                >
                  {p === '' ? 'Static snapshot' : p === 'anthropic' ? 'Anthropic' : 'Ollama'}
                </button>
              ))}
            </div>
          </div>

          {/* Anthropic fields */}
          {provider === 'anthropic' && (
            <>
              <label>
                <span style={{ fontSize: 13, color: 'var(--muted)' }}>
                  Anthropic API key{' '}
                  <span style={{ fontWeight: 400 }}>(or set <code>ANTHROPIC_API_KEY</code> on the server)</span>
                </span>
                <input
                  type="password"
                  className="input"
                  placeholder="sk-ant-…"
                  value={anthropicKey}
                  onChange={(e) => { setAnthropicKey(e.target.value); LS.set('audit.anthropicKey', e.target.value); }}
                  style={{ marginTop: 4 }}
                />
              </label>
              <label>
                <span style={{ fontSize: 13, color: 'var(--muted)' }}>
                  Model{' '}
                  {selectedAnthropicModel && <span style={{ fontWeight: 400 }}>{selectedAnthropicModel.hint}</span>}
                </span>
                <select
                  className="input"
                  value={anthropicModel}
                  onChange={(e) => { setAnthropicModel(e.target.value); LS.set('audit.model', e.target.value); }}
                  style={{ marginTop: 4 }}
                >
                  {ANTHROPIC_MODELS.map((m) => (
                    <option key={m.id} value={m.id}>{m.label} — {m.hint}</option>
                  ))}
                </select>
              </label>
            </>
          )}

          {/* Ollama fields */}
          {provider === 'ollama' && (
            <>
              <label>
                <span style={{ fontSize: 13, color: 'var(--muted)' }}>
                  Ollama base URL{' '}
                  <span style={{ fontWeight: 400 }}>
                    (Docker: use <code>host.docker.internal</code> instead of <code>localhost</code>)
                  </span>
                </span>
                <input
                  type="text"
                  className="input"
                  placeholder="http://host.docker.internal:11434"
                  value={ollamaURL}
                  onChange={(e) => { setOllamaURL(e.target.value); LS.set('audit.ollamaURL', e.target.value); }}
                  style={{ marginTop: 4 }}
                />
              </label>
              <label>
                <span style={{ fontSize: 13, color: 'var(--muted)' }}>
                  Model{' '}
                  <span style={{ fontWeight: 400 }}>(<code>ollama list</code> to see available models)</span>
                </span>
                <input
                  type="text"
                  className="input"
                  placeholder="llama3.2, qwen2.5, deepseek-r1:8b, …"
                  value={ollamaModel}
                  onChange={(e) => { setOllamaModel(e.target.value); LS.set('audit.ollamaModel', e.target.value); }}
                  style={{ marginTop: 4 }}
                />
              </label>
            </>
          )}

          <label>
            <span style={{ fontSize: 13, color: 'var(--muted)' }}>
              GitHub token{' '}
              <span style={{ fontWeight: 400 }}>(optional — needed for secret-scanning alerts and private repos)</span>
            </span>
            <input
              type="password"
              className="input"
              placeholder="ghp_…"
              value={token}
              onChange={(e) => { setToken(e.target.value); LS.set('audit.ghToken', e.target.value); }}
              style={{ marginTop: 4 }}
            />
          </label>

          {error && <p className="error-msg">{error}</p>}
          <button type="submit" className="btn btn-primary" disabled={submitting || !repo.trim()}>
            {submitting ? 'Starting…' : 'Run Audit'}
          </button>
        </form>
        <p style={{ fontSize: 12, color: 'var(--muted)', marginTop: 8 }}>
          {provider === 'ollama'
            ? 'Clones the repo, runs static analysis, generates a report with your local Ollama model — free, no API key needed.'
            : provider === 'anthropic'
            ? 'Clones the repo, runs static analysis, calls the Anthropic API. Typical cost: $0.02–$1.50/run.'
            : 'Clones the repo and runs static analysis. Returns a structured raw-data snapshot — free, no AI.'}
        </p>
      </div>

      <div className="card">
        <h2>Audit History</h2>
        {audits.length === 0 ? (
          <p className="empty">No audits yet. Run one above.</p>
        ) : (
          <div className="table-wrap">
            <table className="scans-table">
              <thead>
                <tr>
                  <th>Repository</th>
                  <th>Started</th>
                  <th>Completed</th>
                  <th>Status</th>
                  <th>Mode</th>
                  <th>Cost</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {audits.map((a) => (
                  <tr key={a.id} className="scan-row" onClick={() => navigate(`/audits/${a.id}`)}>
                    <td><code>{a.repo}</code></td>
                    <td>{formatDate(a.created_at)}</td>
                    <td>{a.completed_at ? formatDate(a.completed_at) : '—'}</td>
                    <td><StatusBadge status={a.status as AuditStatus} /></td>
                    <td style={{ color: 'var(--muted)', fontSize: 12 }}>{modeLabel(a.provider, a.model)}</td>
                    <td>{costEstimate(a.provider, a.model, a.input_tokens, a.output_tokens)}</td>
                    <td>
                      <button className="btn btn-danger" onClick={(e) => deleteAudit(e, a.id)}>
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
  );
}
