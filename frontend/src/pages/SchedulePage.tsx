import { useEffect, useState, useCallback } from 'react';
import { api, Schedule, UpdateScheduleParams } from '../api';
import { useLang } from '../i18n';

function formatDate(s: string | null) {
  if (!s) return '—';
  return new Date(s).toLocaleString();
}

function intervalLabel(h: number, t: (key: any, params?: Record<string, string | number>) => string) {
  if (h < 24) return `${h}h`;
  const d = Math.round(h / 24);
  return d === 1 ? t('schedulePage.daily') : d === 7 ? t('schedulePage.weekly') : d === 30 ? t('schedulePage.monthly') : `${d}d`;
}

function windowLabel(start: number, end: number, t: (key: any, params?: Record<string, string | number>) => string) {
  if (start < 0 || end <= start) return t('schedulePage.anyTime');
  return `${String(start).padStart(2, '0')}:00–${String(end).padStart(2, '0')}:00 UTC`;
}

interface EditState extends UpdateScheduleParams {
  id: string;
}

export default function SchedulePage() {
  const { t } = useLang();
  const INTERVAL_OPTIONS = [
    { value: 24,  label: t('schedulePage.intervalDaily') },
    { value: 72,  label: t('schedulePage.interval3Days') },
    { value: 168, label: t('schedulePage.intervalWeekly') },
    { value: 336, label: t('schedulePage.interval2Weeks') },
    { value: 720, label: t('schedulePage.intervalMonthly') },
  ];
  const [schedules, setSchedules] = useState<Schedule[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // New schedule form
  const [newRepo, setNewRepo] = useState('');
  const [newInterval, setNewInterval] = useState(168);
  const [newProvider, setNewProvider] = useState('');
  const [newModel, setNewModel] = useState('');
  const [newWindowStart, setNewWindowStart] = useState(-1);
  const [newWindowEnd, setNewWindowEnd] = useState(-1);
  const [submitting, setSubmitting] = useState(false);

  // Inline editing
  const [editing, setEditing] = useState<EditState | null>(null);

  const load = useCallback(async () => {
    try {
      setSchedules(await api.listSchedules());
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    if (!newRepo.trim()) return;
    setSubmitting(true);
    setError(null);
    try {
      await api.createSchedule({
        repo: newRepo.trim(),
        interval_h: newInterval,
        provider: newProvider || undefined,
        model: newModel || undefined,
        time_window_start: newWindowStart,
        time_window_end: newWindowEnd,
      });
      setNewRepo('');
      load();
    } catch (e) {
      setError(String(e));
    } finally {
      setSubmitting(false);
    }
  }

  async function handleToggle(s: Schedule) {
    await api.updateSchedule(s.id, {
      interval_h: s.interval_h,
      provider: s.provider,
      model: s.model,
      enabled: !s.enabled,
      time_window_start: s.time_window_start,
      time_window_end: s.time_window_end,
    });
    load();
  }

  async function handleDelete(id: string) {
    if (!confirm(t('schedulePage.deleteConfirm'))) return;
    await api.deleteSchedule(id);
    setSchedules((prev) => prev.filter((s) => s.id !== id));
  }

  async function handleTrigger(id: string) {
    await api.triggerSchedule(id);
    load();
  }

  async function handlePauseAll() {
    if (!confirm(t('schedulePage.pauseAllConfirm'))) return;
    await api.pauseAllSchedules();
    load();
  }

  async function handleSaveEdit() {
    if (!editing) return;
    await api.updateSchedule(editing.id, {
      interval_h: editing.interval_h,
      provider: editing.provider,
      model: editing.model,
      enabled: editing.enabled,
      time_window_start: editing.time_window_start,
      time_window_end: editing.time_window_end,
    });
    setEditing(null);
    load();
  }

  async function handleEnableSuggested(s: Schedule) {
    await api.updateSchedule(s.id, {
      interval_h: s.interval_h,
      provider: s.provider,
      model: s.model,
      enabled: true,
      time_window_start: s.time_window_start,
      time_window_end: s.time_window_end,
    });
    load();
  }

  const active = schedules.filter((s) => s.enabled);
  const disabled = schedules.filter((s) => !s.enabled && s.auto_detected);

  if (loading) return <div className="container"><p style={{ color: 'var(--muted)' }}>{t('common.loading')}</p></div>;

  return (
    <div className="container">
      {/* New schedule form */}
      <div className="card">
        <h2>{t('schedulePage.newSchedule')}</h2>
        <form onSubmit={handleCreate} style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <label>
            <span style={{ fontSize: 13, color: 'var(--muted)' }}>{t('schedulePage.repositoryLabel')}</span>
            <input type="text" className="input" placeholder="e.g. golang/go" value={newRepo}
              onChange={(e) => setNewRepo(e.target.value)} required style={{ marginTop: 4 }} />
          </label>
          <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap' }}>
            <label style={{ flex: 1, minWidth: 150 }}>
              <span style={{ fontSize: 13, color: 'var(--muted)' }}>{t('schedulePage.intervalLabel')}</span>
              <select className="input" value={newInterval} onChange={(e) => setNewInterval(Number(e.target.value))} style={{ marginTop: 4 }}>
                {INTERVAL_OPTIONS.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
              </select>
            </label>
            <label style={{ flex: 1, minWidth: 150 }}>
              <span style={{ fontSize: 13, color: 'var(--muted)' }}>{t('schedulePage.providerLabel')}</span>
              <select className="input" value={newProvider} onChange={(e) => setNewProvider(e.target.value)} style={{ marginTop: 4 }}>
                <option value="">{t('schedulePage.staticSnapshotFree')}</option>
                <option value="anthropic">Anthropic</option>
                <option value="openai">OpenAI</option>
                <option value="gemini">Gemini</option>
                <option value="ollama">Ollama</option>
              </select>
            </label>
            <label style={{ flex: 1, minWidth: 150 }}>
              <span style={{ fontSize: 13, color: 'var(--muted)' }}>{t('schedulePage.modelLabel')}</span>
              <input type="text" className="input" placeholder="default" value={newModel}
                onChange={(e) => setNewModel(e.target.value)} style={{ marginTop: 4 }} />
            </label>
          </div>
          <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap' }}>
            <label style={{ flex: 1, minWidth: 150 }}>
              <span style={{ fontSize: 13, color: 'var(--muted)' }}>{t('schedulePage.windowStartLabel')}</span>
              <input type="number" className="input" min={-1} max={23} value={newWindowStart}
                onChange={(e) => setNewWindowStart(Number(e.target.value))} style={{ marginTop: 4 }} />
            </label>
            <label style={{ flex: 1, minWidth: 150 }}>
              <span style={{ fontSize: 13, color: 'var(--muted)' }}>{t('schedulePage.windowEndLabel')}</span>
              <input type="number" className="input" min={-1} max={24} value={newWindowEnd}
                onChange={(e) => setNewWindowEnd(Number(e.target.value))} style={{ marginTop: 4 }} />
            </label>
          </div>
          {error && <p className="error-msg">{error}</p>}
          <button type="submit" className="btn btn-primary" disabled={submitting || !newRepo.trim()}>
            {submitting ? t('common.creating') : t('schedulePage.createSchedule')}
          </button>
        </form>
      </div>

      {/* Active schedules */}
      <div className="card">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h2>{t('schedulePage.activeSchedules')}</h2>
          {active.length > 0 && (
            <button className="btn btn-danger" style={{ fontSize: 12, padding: '3px 10px' }} onClick={handlePauseAll}>
              {t('schedulePage.pauseAll')}
            </button>
          )}
        </div>
        {active.length === 0 ? (
          <p className="empty">{t('schedulePage.noActiveSchedules')}</p>
        ) : (
          <div className="table-wrap">
            <table className="scans-table">
              <thead>
                <tr>
                  <th>{t('schedulePage.colRepository')}</th>
                  <th>{t('schedulePage.colInterval')}</th>
                  <th>{t('schedulePage.colWindow')}</th>
                  <th>{t('schedulePage.colProviderModel')}</th>
                  <th>{t('schedulePage.colLastRun')}</th>
                  <th>{t('schedulePage.colNextRun')}</th>
                  <th>{t('schedulePage.colSource')}</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {active.map((s) => (
                  editing?.id === s.id ? (
                    <tr key={s.id}>
                      <td><code>{s.repo}</code></td>
                      <td>
                        <select className="input" style={{ fontSize: 12, padding: '2px 4px' }}
                          value={editing.interval_h} onChange={(e) => setEditing({ ...editing, interval_h: Number(e.target.value) })}>
                          {INTERVAL_OPTIONS.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
                        </select>
                      </td>
                      <td>
                        <div style={{ display: 'flex', gap: 4, alignItems: 'center' }}>
                          <input type="number" min={-1} max={23} style={{ width: 50, fontSize: 12, padding: '2px 4px', background: 'var(--surface)', border: '1px solid var(--border)', color: 'var(--text)', borderRadius: 4 }}
                            value={editing.time_window_start} onChange={(e) => setEditing({ ...editing, time_window_start: Number(e.target.value) })} />
                          <span style={{ color: 'var(--muted)' }}>–</span>
                          <input type="number" min={-1} max={24} style={{ width: 50, fontSize: 12, padding: '2px 4px', background: 'var(--surface)', border: '1px solid var(--border)', color: 'var(--text)', borderRadius: 4 }}
                            value={editing.time_window_end} onChange={(e) => setEditing({ ...editing, time_window_end: Number(e.target.value) })} />
                        </div>
                      </td>
                      <td>
                        <div style={{ display: 'flex', gap: 4 }}>
                          <select className="input" style={{ fontSize: 12, padding: '2px 4px' }}
                            value={editing.provider} onChange={(e) => setEditing({ ...editing, provider: e.target.value })}>
                            <option value="">{t('schedulePage.snapshotOption')}</option>
                            <option value="anthropic">Anthropic</option>
                            <option value="openai">OpenAI</option>
                            <option value="gemini">Gemini</option>
                            <option value="ollama">Ollama</option>
                          </select>
                          <input type="text" style={{ width: 100, fontSize: 12, padding: '2px 4px', background: 'var(--surface)', border: '1px solid var(--border)', color: 'var(--text)', borderRadius: 4 }}
                            placeholder="model" value={editing.model} onChange={(e) => setEditing({ ...editing, model: e.target.value })} />
                        </div>
                      </td>
                      <td>{formatDate(s.last_run_at)}</td>
                      <td>{formatDate(s.next_run_at)}</td>
                      <td style={{ fontSize: 11, color: 'var(--muted)' }}>{s.auto_detected ? t('schedulePage.auto') : t('schedulePage.manual')}</td>
                      <td style={{ display: 'flex', gap: 4 }}>
                        <button className="btn btn-primary" style={{ fontSize: 12, padding: '3px 8px' }} onClick={handleSaveEdit}>{t('common.save')}</button>
                        <button className="btn" style={{ fontSize: 12, padding: '3px 8px' }} onClick={() => setEditing(null)}>{t('common.cancel')}</button>
                      </td>
                    </tr>
                  ) : (
                    <tr key={s.id}>
                      <td><code>{s.repo}</code></td>
                      <td style={{ color: 'var(--muted)', fontSize: 12 }}>{intervalLabel(s.interval_h, t)}</td>
                      <td style={{ color: 'var(--muted)', fontSize: 12 }}>{windowLabel(s.time_window_start, s.time_window_end, t)}</td>
                      <td style={{ color: 'var(--muted)', fontSize: 12 }}>{s.provider || t('schedulePage.snapshot')}{s.model ? ` · ${s.model}` : ''}</td>
                      <td style={{ fontSize: 12 }}>{formatDate(s.last_run_at)}</td>
                      <td style={{ fontSize: 12 }}>{formatDate(s.next_run_at)}</td>
                      <td style={{ fontSize: 11, color: 'var(--muted)' }}>
                        {s.auto_detected ? <span title={s.detect_reason}>{t('schedulePage.auto')}</span> : t('schedulePage.manual')}
                      </td>
                      <td style={{ display: 'flex', gap: 4 }}>
                        <button className="btn" style={{ fontSize: 12, padding: '3px 8px' }}
                          onClick={() => setEditing({ id: s.id, interval_h: s.interval_h, provider: s.provider, model: s.model, enabled: s.enabled, time_window_start: s.time_window_start, time_window_end: s.time_window_end })}>
                          {t('common.edit')}
                        </button>
                        <button className="btn" style={{ fontSize: 12, padding: '3px 8px' }} onClick={() => handleTrigger(s.id)}>{t('schedulePage.runNow')}</button>
                        <button className="btn btn-danger" style={{ fontSize: 12, padding: '3px 8px' }} onClick={() => handleToggle(s)}>{t('schedulePage.disable')}</button>
                        <button className="btn btn-danger" style={{ fontSize: 12, padding: '3px 8px' }} onClick={() => handleDelete(s.id)}>{t('common.delete')}</button>
                      </td>
                    </tr>
                  )
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Auto-detected suggestions */}
      {disabled.length > 0 && (
        <div className="card">
          <h2>{t('schedulePage.suggestedSchedules')}</h2>
          <p style={{ fontSize: 13, color: 'var(--muted)', marginBottom: 12 }}>
            {t('schedulePage.suggestedSchedulesDesc')}
          </p>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {disabled.map((s) => (
              <div key={s.id} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '8px 12px', background: 'var(--surface)', borderRadius: 6, border: '1px solid var(--border)' }}>
                <div>
                  <code style={{ fontSize: 13 }}>{s.repo}</code>
                  {s.detect_reason && <span style={{ fontSize: 11, color: 'var(--muted)', marginLeft: 10 }}>{s.detect_reason}</span>}
                </div>
                <div style={{ display: 'flex', gap: 6 }}>
                  <button className="btn btn-primary" style={{ fontSize: 12, padding: '3px 10px' }} onClick={() => handleEnableSuggested(s)}>{t('schedulePage.enable')}</button>
                  <button className="btn btn-danger" style={{ fontSize: 12, padding: '3px 10px' }} onClick={() => handleDelete(s.id)}>{t('schedulePage.dismiss')}</button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
