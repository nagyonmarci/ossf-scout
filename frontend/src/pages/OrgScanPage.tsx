import { useState } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api';
import { useLang } from '../i18n';

interface OrgRepo {
  full_name: string;
  description: string;
  stargazers_count: number;
  language: string;
  archived: boolean;
  fork: boolean;
}

export default function OrgScanPage() {
  const { t } = useLang();
  const [org, setOrg] = useState('');
  const [minStars, setMinStars] = useState(100);
  const [excludeForks, setExcludeForks] = useState(true);
  const [excludeArchived, setExcludeArchived] = useState(true);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [repos, setRepos] = useState<OrgRepo[]>([]);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [queuing, setQueuing] = useState(false);
  const [queueResult, setQueueResult] = useState<string | null>(null);

  async function fetchRepos() {
    if (!org.trim()) return;
    setLoading(true);
    setError(null);
    setRepos([]);
    setSelected(new Set());
    setQueueResult(null);
    try {
      const data = await api.listOrgRepos(org.trim(), {
        min_stars: minStars,
        exclude_forks: excludeForks,
        exclude_archived: excludeArchived,
      });
      setRepos(data);
      setSelected(new Set(data.map(r => r.full_name)));
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  }

  function toggleAll() {
    if (selected.size === repos.length) {
      setSelected(new Set());
    } else {
      setSelected(new Set(repos.map(r => r.full_name)));
    }
  }

  function toggle(name: string) {
    setSelected(prev => {
      const next = new Set(prev);
      if (next.has(name)) next.delete(name);
      else next.add(name);
      return next;
    });
  }

  async function queueAudits() {
    if (selected.size === 0) return;
    setQueuing(true);
    setQueueResult(null);
    let ok = 0, fail = 0;
    for (const repo of selected) {
      try {
        await api.createAudit({ repo, provider: '', model: '' });
        ok++;
      } catch {
        fail++;
      }
    }
    setQueuing(false);
    setQueueResult(
      fail
        ? t('orgScan.queuedWithFailures', { ok, fail })
        : t('orgScan.queued', { ok })
    );
  }

  return (
    <div className="container">
      <h2>{t('orgScan.heading')}</h2>
      <p style={{ color: 'var(--muted)', fontSize: 13, marginTop: -8 }}>
        {t('orgScan.description')}
      </p>

      <div className="card">
        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', alignItems: 'flex-end' }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
            <label style={{ fontSize: 12, color: 'var(--muted)' }}>{t('orgScan.githubOrgLabel')}</label>
            <input
              className="input"
              style={{ width: 200 }}
              placeholder="e.g. directus"
              value={org}
              onChange={e => setOrg(e.target.value)}
              onKeyDown={e => { if (e.key === 'Enter') fetchRepos(); }}
            />
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
            <label style={{ fontSize: 12, color: 'var(--muted)' }}>{t('orgScan.minStarsLabel')}</label>
            <input
              className="input"
              style={{ width: 90 }}
              type="number"
              min={0}
              value={minStars}
              onChange={e => setMinStars(Number(e.target.value))}
            />
          </div>
          <label style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 13 }}>
            <input type="checkbox" checked={excludeForks} onChange={e => setExcludeForks(e.target.checked)} />
            {t('orgScan.excludeForks')}
          </label>
          <label style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 13 }}>
            <input type="checkbox" checked={excludeArchived} onChange={e => setExcludeArchived(e.target.checked)} />
            {t('orgScan.excludeArchived')}
          </label>
          <button className="btn btn-primary" onClick={fetchRepos} disabled={loading || !org.trim()}>
            {loading ? t('orgScan.fetching') : t('orgScan.fetchRepos')}
          </button>
        </div>
        {error && <p className="error-msg" style={{ marginTop: 8 }}>{error}</p>}
      </div>

      {repos.length > 0 && (
        <div className="card" style={{ marginTop: 12 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 10 }}>
            <span style={{ fontSize: 13, color: 'var(--muted)' }}>
              {t('orgScan.reposFoundSelected', { count: repos.length, selected: selected.size })}
            </span>
            <div style={{ display: 'flex', gap: 8 }}>
              <button
                className="btn"
                style={{ padding: '4px 10px', fontSize: 12, background: 'transparent', border: '1px solid var(--border)' }}
                onClick={toggleAll}
              >
                {selected.size === repos.length ? t('orgScan.deselectAll') : t('orgScan.selectAll')}
              </button>
              <button
                className="btn btn-primary"
                style={{ padding: '4px 12px', fontSize: 12 }}
                onClick={queueAudits}
                disabled={queuing || selected.size === 0}
              >
                {queuing ? t('orgScan.queuing') : t('orgScan.queueAudits', { count: selected.size })}
              </button>
            </div>
          </div>

          {queueResult && (
            <div style={{ padding: '8px 12px', background: 'rgba(76,175,80,0.1)', border: '1px solid rgba(76,175,80,0.3)', borderRadius: 4, fontSize: 13, marginBottom: 10 }}>
              {queueResult} <Link to="/audits" style={{ color: 'var(--accent)' }}>{t('orgScan.viewAudits')}</Link>
            </div>
          )}

          <div style={{ display: 'flex', flexDirection: 'column', gap: 2, maxHeight: 480, overflowY: 'auto' }}>
            {repos.map(r => (
              <label
                key={r.full_name}
                style={{
                  display: 'flex', alignItems: 'center', gap: 10, padding: '7px 8px',
                  borderRadius: 4, cursor: 'pointer',
                  background: selected.has(r.full_name) ? 'rgba(var(--accent-rgb, 100,149,237),0.06)' : 'transparent',
                }}
              >
                <input
                  type="checkbox"
                  checked={selected.has(r.full_name)}
                  onChange={() => toggle(r.full_name)}
                  style={{ flexShrink: 0 }}
                />
                <span style={{ fontFamily: 'monospace', fontSize: 13, minWidth: 220 }}>
                  <a
                    href={`https://github.com/${r.full_name}`}
                    target="_blank"
                    rel="noreferrer"
                    style={{ color: 'var(--accent)', textDecoration: 'none' }}
                    onClick={e => e.stopPropagation()}
                  >
                    {r.full_name}
                  </a>
                </span>
                {r.language && (
                  <span style={{ fontSize: 11, color: 'var(--muted)', background: 'var(--border)', borderRadius: 3, padding: '1px 5px', flexShrink: 0 }}>
                    {r.language}
                  </span>
                )}
                <span style={{ fontSize: 12, color: 'var(--muted)', flexShrink: 0 }}>
                  ★ {r.stargazers_count.toLocaleString()}
                </span>
                {r.description && (
                  <span style={{ fontSize: 12, color: 'var(--muted)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {r.description}
                  </span>
                )}
              </label>
            ))}
          </div>
        </div>
      )}

      {!loading && repos.length === 0 && org && (
        <p style={{ color: 'var(--muted)', fontSize: 13, marginTop: 8 }}>
          {t('orgScan.noReposFound')}
        </p>
      )}
    </div>
  );
}
