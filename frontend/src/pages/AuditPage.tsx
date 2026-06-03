import { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { api, Audit, AuditStatus } from '../api';
import StatusBadge from '../components/StatusBadge';

const MODELS = [
  { id: 'claude-opus-4-8',           label: 'Opus 4',   hint: '~$0.50–$1.50/run', inputRate: 15,   outputRate: 75  },
  { id: 'claude-sonnet-4-6',         label: 'Sonnet 4', hint: '~$0.10–$0.30/run', inputRate: 3,    outputRate: 15  },
  { id: 'claude-haiku-4-5-20251001', label: 'Haiku 4',  hint: '~$0.02–$0.05/run', inputRate: 0.80, outputRate: 4   },
];

function formatDate(s: string) {
  return new Date(s).toLocaleString();
}

function costEstimate(model: string, inputTokens: number | null, outputTokens: number | null): string {
  if (!inputTokens && !outputTokens) return '—';
  const m = MODELS.find((x) => x.id === model);
  const inputRate  = m?.inputRate  ?? 15;
  const outputRate = m?.outputRate ?? 75;
  const cost = ((inputTokens ?? 0) / 1_000_000) * inputRate
             + ((outputTokens ?? 0) / 1_000_000) * outputRate;
  return `~$${cost.toFixed(3)}`;
}

function modelLabel(model: string): string {
  return MODELS.find((x) => x.id === model)?.label ?? (model || 'Snapshot');
}

export default function AuditPage() {
  const navigate = useNavigate();
  const [audits, setAudits] = useState<Audit[]>([]);
  const [repo, setRepo] = useState('');
  const [token, setToken] = useState('');
  const [anthropicKey, setAnthropicKey] = useState('');
  const [model, setModel] = useState(MODELS[0].id);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    try {
      const data = await api.listAudits();
      setAudits(data);
    } catch (e) {
      setError(String(e));
    }
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
        anthropic_key: anthropicKey || undefined,
        model: anthropicKey ? model : undefined,
      });
      setRepo('');
      setToken('');
      setAnthropicKey('');
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

  const selectedModel = MODELS.find((m) => m.id === model);

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
          <label>
            <span style={{ fontSize: 13, color: 'var(--muted)' }}>
              Anthropic API key{' '}
              <span style={{ fontWeight: 400 }}>(optional — without it a free static snapshot is generated)</span>
            </span>
            <input
              type="password"
              className="input"
              placeholder="sk-ant-…"
              value={anthropicKey}
              onChange={(e) => setAnthropicKey(e.target.value)}
              style={{ marginTop: 4 }}
            />
          </label>
          <label style={{ opacity: anthropicKey ? 1 : 0.45 }}>
            <span style={{ fontSize: 13, color: 'var(--muted)' }}>
              Model{' '}
              {selectedModel && anthropicKey && (
                <span style={{ fontWeight: 400 }}>{selectedModel.hint}</span>
              )}
            </span>
            <select
              className="input"
              value={model}
              onChange={(e) => setModel(e.target.value)}
              disabled={!anthropicKey}
              style={{ marginTop: 4 }}
            >
              {MODELS.map((m) => (
                <option key={m.id} value={m.id}>
                  {m.label} — {m.hint}
                </option>
              ))}
            </select>
          </label>
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
              onChange={(e) => setToken(e.target.value)}
              style={{ marginTop: 4 }}
            />
          </label>
          {error && <p className="error-msg">{error}</p>}
          <button type="submit" className="btn btn-primary" disabled={submitting || !repo.trim()}>
            {submitting ? 'Starting…' : 'Run Audit'}
          </button>
        </form>
        <p style={{ fontSize: 12, color: 'var(--muted)', marginTop: 8 }}>
          Clones the repo and runs static analysis. With an Anthropic API key the selected model synthesises a full
          DevSecOps report. Without a key a free raw data snapshot is generated.
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
                  <th>Model</th>
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
                    <td style={{ color: 'var(--muted)', fontSize: 12 }}>{modelLabel(a.model)}</td>
                    <td>{costEstimate(a.model, a.input_tokens, a.output_tokens)}</td>
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
