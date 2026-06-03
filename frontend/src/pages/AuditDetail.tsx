import { useEffect, useState, useCallback, useRef } from 'react';
import { Link, useParams } from 'react-router-dom';
import Markdown from 'react-markdown';
import { api, Audit } from '../api';
import StatusBadge from '../components/StatusBadge';

export default function AuditDetail() {
  const { id } = useParams<{ id: string }>();
  const [audit, setAudit] = useState<Audit | null>(null);
  const [error, setError] = useState<string | null>(null);
  const prevStatus = useRef<string | null>(null);

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
                {audit.input_tokens ? audit.model || 'AI' : 'Static snapshot'}
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
            {audit.status === 'done' && (
              <button className="btn btn-primary" onClick={downloadReport} style={{ marginTop: 8 }}>
                Download .md
              </button>
            )}
          </div>

          {audit.status === 'done' && audit.report && (
            <div className="card audit-report">
              <Markdown>{audit.report}</Markdown>
            </div>
          )}
        </>
      )}
    </div>
  );
}
