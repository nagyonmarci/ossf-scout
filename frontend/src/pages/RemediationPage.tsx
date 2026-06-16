import { useEffect, useState, useCallback } from 'react';
import { useSearchParams, Link } from 'react-router-dom';
import { api, RemediationItem } from '../api';
import { useLang } from '../i18n';

const SEVERITY_ORDER = ['critical', 'high', 'medium', 'low', 'info'];
const SEVERITY_COLORS: Record<string, string> = {
  critical: '#f44336',
  high: '#ff5722',
  medium: '#ff9800',
  low: '#4caf50',
  info: '#9e9e9e',
};

const STATUS_COLS: Array<'open' | 'in_progress' | 'resolved'> = ['open', 'in_progress', 'resolved'];

function statusLabel(status: string, t: (key: any, params?: Record<string, string | number>) => string): string {
  if (status === 'open') return t('remediation.statusOpen');
  if (status === 'in_progress') return t('remediation.statusInProgress');
  return t('remediation.statusResolved');
}

function SeverityBadge({ sev }: { sev: string }) {
  return (
    <span style={{
      fontSize: 10, fontWeight: 700, padding: '1px 6px', borderRadius: 3, textTransform: 'uppercase',
      background: `${SEVERITY_COLORS[sev] || '#9e9e9e'}22`,
      color: SEVERITY_COLORS[sev] || '#9e9e9e',
      border: `1px solid ${SEVERITY_COLORS[sev] || '#9e9e9e'}44`,
    }}>
      {sev}
    </span>
  );
}

function AddItemForm({ auditId, onAdded }: { auditId: string; onAdded: (item: RemediationItem) => void }) {
  const { t } = useLang();
  const [title, setTitle] = useState('');
  const [severity, setSeverity] = useState('medium');
  const [submitting, setSubmitting] = useState(false);

  async function submit() {
    if (!title.trim()) return;
    setSubmitting(true);
    try {
      const item = await api.createRemediationItem(auditId, '', title.trim(), severity);
      setTitle('');
      onAdded(item);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div style={{ display: 'flex', gap: 6, marginBottom: 12, flexWrap: 'wrap' }}>
      <input
        className="input"
        style={{ flex: 1, minWidth: 200 }}
        placeholder={t('remediation.newFindingPlaceholder')}
        value={title}
        onChange={e => setTitle(e.target.value)}
        onKeyDown={e => { if (e.key === 'Enter') submit(); }}
      />
      <select className="input" style={{ width: 110 }} value={severity} onChange={e => setSeverity(e.target.value)}>
        {SEVERITY_ORDER.map(s => <option key={s} value={s}>{s}</option>)}
      </select>
      <button className="btn btn-primary" onClick={submit} disabled={submitting || !title.trim()}>
        {submitting ? t('remediation.adding') : t('remediation.addButton')}
      </button>
    </div>
  );
}

function ItemCard({
  item,
  onMove,
  onDelete,
  onUpdate,
}: {
  item: RemediationItem;
  onMove: (id: string, status: RemediationItem['status']) => void;
  onDelete: (id: string) => void;
  onUpdate: (id: string, patch: Partial<RemediationItem>) => void;
}) {
  const { t } = useLang();
  const [editing, setEditing] = useState(false);
  const [notes, setNotes] = useState(item.notes);
  const [title, setTitle] = useState(item.title);
  const [severity, setSeverity] = useState(item.severity);

  async function save() {
    await api.updateRemediationItem(item.id, { title, severity, status: item.status, notes });
    onUpdate(item.id, { title, severity, notes });
    setEditing(false);
  }

  return (
    <div style={{
      background: 'var(--surface, #1e1e2e)',
      border: '1px solid var(--border)',
      borderRadius: 6,
      padding: 10,
      marginBottom: 8,
    }}>
      {editing ? (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <input className="input" style={{ fontSize: 13 }} value={title} onChange={e => setTitle(e.target.value)} />
          <select className="input" style={{ fontSize: 12, width: 110 }} value={severity}
            onChange={e => setSeverity(e.target.value as RemediationItem['severity'])}>
            {SEVERITY_ORDER.map(s => <option key={s} value={s}>{s}</option>)}
          </select>
          <textarea
            className="input"
            style={{ fontSize: 12, minHeight: 60, resize: 'vertical' }}
            placeholder={t('remediation.notesPlaceholder')}
            value={notes}
            onChange={e => setNotes(e.target.value)}
          />
          <div style={{ display: 'flex', gap: 6 }}>
            <button className="btn btn-primary" style={{ padding: '3px 10px', fontSize: 12 }} onClick={save}>{t('common.save')}</button>
            <button className="btn" style={{ padding: '3px 10px', fontSize: 12, background: 'transparent', border: '1px solid var(--border)' }}
              onClick={() => { setEditing(false); setTitle(item.title); setSeverity(item.severity); setNotes(item.notes); }}>
              {t('common.cancel')}
            </button>
          </div>
        </div>
      ) : (
        <>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 6, marginBottom: 6 }}>
            <Link
              to={`/audits/${item.audit_id}`}
              className="remediation-title-link"
              style={{ fontSize: 13, lineHeight: 1.4, flex: 1, color: 'inherit', textDecoration: 'none' }}
            >
              {item.title}
            </Link>
            <SeverityBadge sev={item.severity} />
          </div>
          {item.repo && (
            <div style={{ fontSize: 11, color: 'var(--muted)', marginBottom: 2, fontFamily: 'monospace' }}>{item.repo}</div>
          )}
          {item.audit_id && (
            <div style={{ fontSize: 10, color: 'var(--muted)', marginBottom: 4, fontFamily: 'monospace', opacity: 0.6 }}>
              {item.audit_id.slice(0, 8)}
            </div>
          )}
          {item.notes && (
            <div style={{ fontSize: 12, color: 'var(--muted)', marginBottom: 6, whiteSpace: 'pre-wrap' }}>{item.notes}</div>
          )}
          <div style={{ display: 'flex', gap: 4, flexWrap: 'wrap' }}>
            {STATUS_COLS.filter(s => s !== item.status).map(s => (
              <button
                key={s}
                style={{ padding: '2px 8px', fontSize: 11, background: 'transparent', border: '1px solid var(--border)', color: 'var(--muted)', borderRadius: 4, cursor: 'pointer' }}
                onClick={() => onMove(item.id, s)}
              >
                → {statusLabel(s, t)}
              </button>
            ))}
            <button
              style={{ padding: '2px 8px', fontSize: 11, background: 'transparent', border: '1px solid var(--border)', color: 'var(--muted)', borderRadius: 4, cursor: 'pointer', marginLeft: 'auto' }}
              onClick={() => setEditing(true)}
            >
              {t('common.edit')}
            </button>
            <button
              style={{ padding: '2px 8px', fontSize: 11, background: 'transparent', border: '1px solid rgba(244,67,54,0.4)', color: '#ef5350', borderRadius: 4, cursor: 'pointer' }}
              onClick={() => onDelete(item.id)}
            >
              ✕
            </button>
          </div>
        </>
      )}
    </div>
  );
}

export default function RemediationPage() {
  const { t } = useLang();
  const [searchParams] = useSearchParams();
  const [items, setItems] = useState<RemediationItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filterRepo, setFilterRepo] = useState('');
  const [selectedAuditId, setSelectedAuditId] = useState(searchParams.get('audit_id') ?? '');

  const load = useCallback(() => {
    api.listRemediation()
      .then(setItems)
      .catch(e => setError(String(e)))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => { load(); }, [load]);

  const repos = Array.from(new Set(items.map(i => i.repo).filter(Boolean))).sort();

  const filtered = items.filter(i =>
    (!filterRepo || i.repo === filterRepo)
  );

  const byStatus = (status: string) =>
    filtered
      .filter(i => i.status === status)
      .sort((a, b) => SEVERITY_ORDER.indexOf(a.severity) - SEVERITY_ORDER.indexOf(b.severity));

  async function handleMove(id: string, status: RemediationItem['status']) {
    const item = items.find(i => i.id === id);
    if (!item) return;
    await api.updateRemediationItem(id, { title: item.title, severity: item.severity, status, notes: item.notes });
    setItems(prev => prev.map(i => i.id === id ? { ...i, status } : i));
  }

  async function handleDelete(id: string) {
    await api.deleteRemediationItem(id);
    setItems(prev => prev.filter(i => i.id !== id));
  }

  function handleUpdate(id: string, patch: Partial<RemediationItem>) {
    setItems(prev => prev.map(i => i.id === id ? { ...i, ...patch } : i));
  }

  function handleAdded(item: RemediationItem) {
    setItems(prev => [...prev, item]);
  }

  const openCount = items.filter(i => i.status === 'open').length;
  const inProgressCount = items.filter(i => i.status === 'in_progress').length;
  const resolvedCount = items.filter(i => i.status === 'resolved').length;

  return (
    <div className="container">
      <h2>{t('remediation.heading')}</h2>
      <p style={{ color: 'var(--muted)', fontSize: 13, marginTop: -8 }}>
        {t('remediation.description')}
      </p>

      {error && <p className="error-msg">{error}</p>}

      <div style={{ display: 'flex', gap: 8, marginBottom: 16, flexWrap: 'wrap', alignItems: 'center' }}>
        {repos.length > 0 && (
          <select className="input" style={{ maxWidth: 260 }} value={filterRepo} onChange={e => setFilterRepo(e.target.value)}>
            <option value="">{t('remediation.allRepos')}</option>
            {repos.map(r => <option key={r} value={r}>{r}</option>)}
          </select>
        )}
        <div style={{ display: 'flex', gap: 12, fontSize: 13, color: 'var(--muted)' }}>
          <span><strong style={{ color: '#ef5350' }}>{openCount}</strong> {t('remediation.statusOpen').toLowerCase()}</span>
          <span><strong style={{ color: '#ff9800' }}>{inProgressCount}</strong> {t('remediation.statusInProgress').toLowerCase()}</span>
          <span><strong style={{ color: '#4caf50' }}>{resolvedCount}</strong> {t('remediation.statusResolved').toLowerCase()}</span>
        </div>
      </div>

      {loading && <p style={{ color: 'var(--muted)' }}>{t('common.loading')}</p>}

      {!loading && (
        <>
          <div style={{ marginBottom: 16 }}>
            <div style={{ fontSize: 13, fontWeight: 600, marginBottom: 8, color: 'var(--muted)' }}>
              {t('remediation.addFindingLabel')}
            </div>
            <div style={{ display: 'flex', gap: 6, marginBottom: 6 }}>
              <input
                className="input"
                style={{ maxWidth: 340 }}
                placeholder={t('remediation.auditIdPlaceholder')}
                value={selectedAuditId}
                onChange={e => setSelectedAuditId(e.target.value)}
              />
            </div>
            {selectedAuditId && (
              <AddItemForm
                auditId={selectedAuditId}
                onAdded={handleAdded}
              />
            )}
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 12 }}>
            {STATUS_COLS.map(col => (
              <div key={col}>
                <div style={{
                  padding: '6px 10px', borderRadius: '6px 6px 0 0',
                  background: col === 'open' ? 'rgba(244,67,54,0.12)' : col === 'in_progress' ? 'rgba(255,152,0,0.12)' : 'rgba(76,175,80,0.12)',
                  border: '1px solid var(--border)',
                  borderBottom: 'none',
                  fontWeight: 600, fontSize: 13,
                  display: 'flex', justifyContent: 'space-between',
                }}>
                  <span>{statusLabel(col, t)}</span>
                  <span style={{ color: 'var(--muted)', fontWeight: 400 }}>{byStatus(col).length}</span>
                </div>
                <div style={{
                  border: '1px solid var(--border)',
                  borderRadius: '0 0 6px 6px',
                  padding: 8,
                  minHeight: 120,
                  background: 'var(--surface-2, rgba(255,255,255,0.02))',
                }}>
                  {byStatus(col).map(item => (
                    <ItemCard
                      key={item.id}
                      item={item}
                      onMove={handleMove}
                      onDelete={handleDelete}
                      onUpdate={handleUpdate}
                    />
                  ))}
                  {byStatus(col).length === 0 && (
                    <div style={{ color: 'var(--muted)', fontSize: 12, textAlign: 'center', marginTop: 16 }}>{t('remediation.empty')}</div>
                  )}
                </div>
              </div>
            ))}
          </div>
        </>
      )}
    </div>
  );
}
