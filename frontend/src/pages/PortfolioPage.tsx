import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { api, PortfolioRepo } from '../api';

const scoreColor = (score: number) => {
  if (score >= 7) return 'var(--success, #4caf50)';
  if (score >= 4) return 'var(--warning, #ff9800)';
  return 'var(--danger, #f44336)';
};

const scoreBar = (score: number) => {
  const pct = Math.round((score / 10) * 100);
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
      <div style={{ flex: 1, height: 6, background: 'var(--border)', borderRadius: 3, overflow: 'hidden' }}>
        <div style={{ width: `${pct}%`, height: '100%', background: scoreColor(score), borderRadius: 3 }} />
      </div>
      <span style={{ minWidth: 32, fontSize: 12, color: scoreColor(score), fontWeight: 600 }}>
        {score.toFixed(1)}
      </span>
    </div>
  );
};

export default function PortfolioPage() {
  const [repos, setRepos] = useState<PortfolioRepo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filterInput, setFilterInput] = useState('');
  const [sortBy, setSortBy] = useState<'score' | 'stars' | 'audits'>('score');

  useEffect(() => {
    api.getPortfolio()
      .then(setRepos)
      .catch(e => setError(String(e)))
      .finally(() => setLoading(false));
  }, []);

  const filtered = repos
    .filter(r => !filterInput || r.repo.toLowerCase().includes(filterInput.toLowerCase()))
    .sort((a, b) => {
      if (sortBy === 'score') return a.latest_score - b.latest_score;
      if (sortBy === 'stars') return b.stars - a.stars;
      return b.audit_count - a.audit_count;
    });

  const allChecks = Array.from(new Set(repos.flatMap(r => r.weak_checks ?? []))).sort();

  return (
    <div className="container">
      <h2>Portfolio</h2>
      <p style={{ color: 'var(--muted)', fontSize: 13, marginTop: -8 }}>
        Aggregated view of all repos in your scan history. Sorted by OSSF score (lowest first).
      </p>

      {error && <p className="error-msg">{error}</p>}

      <div style={{ display: 'flex', gap: 8, marginBottom: 16, flexWrap: 'wrap' }}>
        <input
          className="input"
          style={{ maxWidth: 280 }}
          placeholder="Filter repos…"
          value={filterInput}
          onChange={e => setFilterInput(e.target.value)}
        />
        <div style={{ display: 'flex', gap: 4 }}>
          {(['score', 'stars', 'audits'] as const).map(s => (
            <button
              key={s}
              className={`btn${sortBy === s ? ' btn-primary' : ''}`}
              style={{ padding: '4px 10px', fontSize: 13, background: sortBy === s ? undefined : 'transparent', border: '1px solid var(--border)' }}
              onClick={() => setSortBy(s)}
            >
              {s === 'score' ? 'Score ↑' : s === 'stars' ? 'Stars ↓' : 'Audits ↓'}
            </button>
          ))}
        </div>
      </div>

      {loading && <p style={{ color: 'var(--muted)' }}>Loading…</p>}

      {!loading && filtered.length === 0 && (
        <p style={{ color: 'var(--muted)' }}>
          No repos found. Run a scan first to populate the portfolio.
        </p>
      )}

      {filtered.length > 0 && (
        <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 13 }}>
            <thead>
              <tr style={{ background: 'var(--surface-2, rgba(255,255,255,0.03))', borderBottom: '1px solid var(--border)' }}>
                <th style={{ padding: '10px 14px', textAlign: 'left', fontWeight: 600 }}>Repo</th>
                <th style={{ padding: '10px 14px', textAlign: 'left', fontWeight: 600, minWidth: 140 }}>OSSF Score</th>
                <th style={{ padding: '10px 14px', textAlign: 'right', fontWeight: 600 }}>Stars</th>
                <th style={{ padding: '10px 14px', textAlign: 'left', fontWeight: 600 }}>Weak checks</th>
                <th style={{ padding: '10px 14px', textAlign: 'right', fontWeight: 600 }}>Audits</th>
                <th style={{ padding: '10px 14px', textAlign: 'left', fontWeight: 600 }}>Last audit</th>
                <th style={{ padding: '10px 14px', textAlign: 'left', fontWeight: 600 }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((r, i) => (
                <tr
                  key={r.repo}
                  style={{ borderBottom: i < filtered.length - 1 ? '1px solid var(--border)' : undefined }}
                >
                  <td style={{ padding: '10px 14px' }}>
                    <a
                      href={`https://github.com/${r.repo}`}
                      target="_blank"
                      rel="noreferrer"
                      style={{ color: 'var(--accent)', textDecoration: 'none', fontFamily: 'monospace' }}
                    >
                      {r.repo}
                    </a>
                    {r.language && (
                      <span style={{ marginLeft: 6, fontSize: 11, color: 'var(--muted)', background: 'var(--border)', borderRadius: 3, padding: '1px 5px' }}>
                        {r.language}
                      </span>
                    )}
                  </td>
                  <td style={{ padding: '10px 14px', minWidth: 140 }}>
                    {r.latest_score > 0 ? scoreBar(r.latest_score) : <span style={{ color: 'var(--muted)' }}>—</span>}
                  </td>
                  <td style={{ padding: '10px 14px', textAlign: 'right', color: 'var(--muted)' }}>
                    {r.stars > 0 ? r.stars.toLocaleString() : '—'}
                  </td>
                  <td style={{ padding: '10px 14px', maxWidth: 260 }}>
                    <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
                      {(r.weak_checks ?? []).slice(0, 5).map(c => (
                        <span
                          key={c}
                          style={{ fontSize: 11, background: 'rgba(244,67,54,0.12)', color: '#ef5350', borderRadius: 3, padding: '1px 5px' }}
                        >
                          {c}
                        </span>
                      ))}
                      {(r.weak_checks ?? []).length > 5 && (
                        <span style={{ fontSize: 11, color: 'var(--muted)' }}>+{r.weak_checks.length - 5}</span>
                      )}
                    </div>
                  </td>
                  <td style={{ padding: '10px 14px', textAlign: 'right', color: 'var(--muted)' }}>
                    {r.audit_count}
                  </td>
                  <td style={{ padding: '10px 14px', color: 'var(--muted)', whiteSpace: 'nowrap' }}>
                    {r.last_audit_at ? new Date(r.last_audit_at).toLocaleDateString() : '—'}
                  </td>
                  <td style={{ padding: '10px 14px', whiteSpace: 'nowrap' }}>
                    <Link
                      to={`/audits?repo=${encodeURIComponent(r.repo)}`}
                      className="btn"
                      style={{ padding: '3px 10px', fontSize: 12, background: 'transparent', border: '1px solid var(--accent)', color: 'var(--accent)' }}
                    >
                      Audit
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {allChecks.length > 0 && (
        <div className="card" style={{ marginTop: 16 }}>
          <h3 style={{ margin: '0 0 10px', fontSize: 14 }}>Weak check heatmap</h3>
          <p style={{ fontSize: 12, color: 'var(--muted)', margin: '0 0 10px' }}>
            How many repos fail each OSSF check (across {repos.length} repos):
          </p>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6 }}>
            {allChecks.map(check => {
              const count = repos.filter(r => (r.weak_checks ?? []).includes(check)).length;
              const opacity = Math.max(0.2, count / repos.length);
              return (
                <div
                  key={check}
                  style={{
                    padding: '4px 8px',
                    borderRadius: 4,
                    fontSize: 12,
                    background: `rgba(244,67,54,${opacity})`,
                    color: opacity > 0.5 ? '#fff' : 'var(--fg)',
                    cursor: 'default',
                  }}
                  title={`${count} of ${repos.length} repos fail this check`}
                >
                  {check} ({count})
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}
