import { useEffect, useState, useCallback } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { api, Audit, AuditStatus, CostStats } from '../api';
import StatusBadge from '../components/StatusBadge';
import { useLang } from '../i18n';

type Provider = '' | 'anthropic' | 'openai' | 'gemini' | 'ollama';

const PROVIDER_MODELS: Record<string, { id: string; label: string; hint: string; inputRate: number; outputRate: number }[]> = {
  anthropic: [
    { id: 'claude-opus-4-8',            label: 'Claude Opus 4.8',    hint: '~$0.17–$0.31/run',  inputRate: 5.0,  outputRate: 25.0 },
    { id: 'claude-sonnet-4-6',          label: 'Claude Sonnet 4.6',  hint: '~$0.10–$0.19/run',  inputRate: 3.0,  outputRate: 15.0 },
    { id: 'claude-haiku-4-5-20251001',  label: 'Claude Haiku 4.5',   hint: '~$0.06–$0.12/run',  inputRate: 1.0,  outputRate: 5.0  },
    { id: 'claude-opus-4-7',            label: 'Claude Opus 4.7',    hint: '~$0.17–$0.31/run',  inputRate: 5.0,  outputRate: 25.0 },
    { id: 'claude-opus-4-6',            label: 'Claude Opus 4.6',    hint: '~$0.17–$0.31/run',  inputRate: 5.0,  outputRate: 25.0 },
    { id: 'claude-sonnet-4-5-20250929', label: 'Claude Sonnet 4.5',  hint: '~$0.10–$0.19/run',  inputRate: 3.0,  outputRate: 15.0 },
    { id: 'claude-opus-4-5-20251101',   label: 'Claude Opus 4.5',    hint: '~$0.17–$0.31/run',  inputRate: 5.0,  outputRate: 25.0 },
  ],
  openai: [
    { id: 'gpt-5.5',      label: 'GPT-5.5 (1M ctx)',        hint: '~$0.35–$1.40/run',           inputRate: 5.0,  outputRate: 30.0 },
    { id: 'gpt-5.4',      label: 'GPT-5.4 (1M ctx)',        hint: '~$0.18–$0.70/run',           inputRate: 2.5,  outputRate: 15.0 },
    { id: 'gpt-5.4-mini', label: 'GPT-5.4 mini (400K ctx)', hint: '~$0.05–$0.20/run',           inputRate: 0.75, outputRate: 4.50 },
    { id: 'gpt-4.1',      label: 'GPT-4.1 (1M ctx)',        hint: '~$0.09–$0.35/run · 30K TPM (T1)',   inputRate: 2.0,  outputRate: 8.0  },
    { id: 'gpt-4.1-mini', label: 'GPT-4.1 mini (1M ctx)',   hint: '~$0.02–$0.07/run · 200K TPM (T1)',  inputRate: 0.40, outputRate: 1.60 },
    { id: 'gpt-4.1-nano', label: 'GPT-4.1 nano (1M ctx)',   hint: '~$0.005–$0.02/run · 200K TPM (T1)', inputRate: 0.10, outputRate: 0.40 },
    { id: 'gpt-4o',       label: 'GPT-4o (128K ctx)',        hint: '~$0.10–$0.50/run · 30K TPM (T1)',   inputRate: 2.5,  outputRate: 10.0 },
    { id: 'gpt-4o-mini',  label: 'GPT-4o mini (128K ctx)',  hint: '~$0.005–$0.03/run · 200K TPM (T1)', inputRate: 0.15, outputRate: 0.60 },
    { id: 'o1',           label: 'o1 (200K ctx)',            hint: '~$0.70–$2.80/run · 30K TPM (T1)',   inputRate: 15.0, outputRate: 60.0 },
    { id: 'o4-mini',      label: 'o4-mini (200K ctx)',       hint: '~$0.05–$0.20/run · 50K TPM (T1)',   inputRate: 1.10, outputRate: 4.40 },
    { id: 'o3',           label: 'o3 (200K ctx)',            hint: '~$0.45–$1.80/run · 30K TPM (T1)',   inputRate: 10.0, outputRate: 40.0 },
    { id: 'o3-mini',      label: 'o3-mini (200K ctx)',       hint: '~$0.05–$0.20/run · 50K TPM (T1)',   inputRate: 1.10, outputRate: 4.40 },
  ],
  gemini: [
    { id: 'gemini-3.5-flash',      label: 'Gemini 3.5 Flash',      hint: '~$0.10–$0.40/run',   inputRate: 1.50,  outputRate: 9.0  },
    { id: 'gemini-3.1-flash-lite', label: 'Gemini 3.1 Flash Lite', hint: '~$0.015–$0.07/run',  inputRate: 0.25,  outputRate: 1.50 },
    { id: 'gemini-2.5-pro',        label: 'Gemini 2.5 Pro',        hint: '~$0.09–$0.36/run',   inputRate: 1.25,  outputRate: 10.0 },
    { id: 'gemini-2.5-flash',      label: 'Gemini 2.5 Flash',      hint: '~$0.027–$0.09/run',  inputRate: 0.30,  outputRate: 2.50 },
    { id: 'gemini-2.5-flash-lite', label: 'Gemini 2.5 Flash Lite', hint: '~$0.005–$0.02/run',  inputRate: 0.10,  outputRate: 0.40 },
    { id: 'gemini-2.0-flash',      label: 'Gemini 2.0 Flash',      hint: '~$0.005–$0.02/run',  inputRate: 0.10,  outputRate: 0.40 },
    { id: 'gemini-1.5-pro',        label: 'Gemini 1.5 Pro',        hint: '~$0.09–$0.37/run',   inputRate: 1.25,  outputRate: 5.0  },
    { id: 'gemini-1.5-flash',      label: 'Gemini 1.5 Flash',      hint: '~$0.003–$0.015/run', inputRate: 0.075, outputRate: 0.30 },
  ],
};

const ALL_MODELS = Object.values(PROVIDER_MODELS).flat();

function formatDate(s: string) {
  return new Date(s).toLocaleString();
}

function costEstimate(provider: string, model: string, inputTokens: number | null, outputTokens: number | null): string {
  if (provider === 'ollama') return 'free (local)';
  if (!provider) return 'free';
  if (!inputTokens && !outputTokens) return '—';
  const m = ALL_MODELS.find((x) => x.id === model);
  const inputRate  = m?.inputRate  ?? 15;
  const outputRate = m?.outputRate ?? 75;
  const cost = ((inputTokens ?? 0) / 1_000_000) * inputRate
             + ((outputTokens ?? 0) / 1_000_000) * outputRate;
  return `~$${cost.toFixed(3)}`;
}

function modeLabel(provider: string, model: string): string {
  if (provider === 'ollama') return model ? `Ollama · ${model}` : 'Ollama';
  if (provider === 'anthropic') {
    const m = PROVIDER_MODELS.anthropic.find((x) => x.id === model);
    return m?.label ?? model ?? 'Anthropic';
  }
  if (provider === 'openai') {
    const m = PROVIDER_MODELS.openai.find((x) => x.id === model);
    return m?.label ?? model ?? 'OpenAI';
  }
  if (provider === 'gemini') {
    const m = PROVIDER_MODELS.gemini.find((x) => x.id === model);
    return m?.label ?? model ?? 'Gemini';
  }
  return 'Snapshot';
}

const LS = {
  get: (k: string, fallback: string) => localStorage.getItem(k) ?? fallback,
  set: (k: string, v: string) => localStorage.setItem(k, v),
};

export default function AuditPage() {
  const { t } = useLang();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [audits, setAudits] = useState<Audit[]>([]);
  const [costStats, setCostStats] = useState<CostStats | null>(null);
  const [repo, setRepo] = useState(() => searchParams.get('repo') ?? '');
  const [token, setToken] = useState(() => LS.get('audit.ghToken', ''));
  const [provider, setProvider] = useState<Provider>(() => LS.get('audit.provider', '') as Provider);

  // Per-provider state
  const [anthropicKey, setAnthropicKey] = useState(() => LS.get('audit.anthropicKey', ''));
  const [openaiKey, setOpenaiKey]       = useState(() => LS.get('audit.openaiKey', ''));
  const [geminiKey, setGeminiKey]       = useState(() => LS.get('audit.geminiKey', ''));

  const [anthropicModel, setAnthropicModel] = useState(() => LS.get('audit.model', PROVIDER_MODELS.anthropic[0].id));
  const [openaiModel, setOpenaiModel]       = useState(() => LS.get('audit.openaiModel', PROVIDER_MODELS.openai[0].id));
  const [geminiModel, setGeminiModel]       = useState(() => LS.get('audit.geminiModel', PROVIDER_MODELS.gemini[0].id));

  const [ollamaURL, setOllamaURL]   = useState(() => LS.get('audit.ollamaURL', ''));
  const [ollamaModel, setOllamaModel] = useState(() => LS.get('audit.ollamaModel', ''));
  const [splitGeneration, setSplitGeneration] = useState(() => LS.get('audit.splitGeneration', 'true') === 'true');
  const [analysisModel, setAnalysisModel] = useState(() => LS.get('audit.analysisModel', ''));
  const [skipSecrets, setSkipSecrets] = useState(() => LS.get('audit.skipSecrets', 'false') === 'true');

  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    try {
      const [a, c] = await Promise.all([api.listAudits(), api.getCostStats(30)]);
      setAudits(a);
      setCostStats(c);
    } catch (e) {
      setError(String(e));
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  useEffect(() => {
    if (localStorage.getItem('audit.ollamaURL') === 'http://localhost:11434') {
      localStorage.removeItem('audit.ollamaURL');
      setOllamaURL('');
    }
  }, []);

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
      const p = provider;
      await api.createAudit({
        repo: repo.trim(),
        github_token: token || undefined,
        provider: p || undefined,
        anthropic_key: p === 'anthropic' ? (anthropicKey || undefined) : undefined,
        openai_key:    p === 'openai'    ? (openaiKey    || undefined) : undefined,
        gemini_key:    p === 'gemini'    ? (geminiKey    || undefined) : undefined,
        model: p === 'anthropic' ? anthropicModel
             : p === 'openai'    ? openaiModel
             : p === 'gemini'    ? geminiModel
             : p === 'ollama'    ? ollamaModel
             : undefined,
        split_generation: (p === 'anthropic' || p === 'ollama') ? splitGeneration || undefined : undefined,
        analysis_model: splitGeneration ? (analysisModel || undefined) : undefined,
        skip_secrets: skipSecrets || undefined,
        ollama_url: p === 'ollama' ? (ollamaURL || undefined) : undefined,
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

  const providerLabels: Record<Provider, string> = {
    '': t('auditPage.providerSnapshot'),
    anthropic: 'Anthropic',
    openai: 'OpenAI',
    gemini: 'Gemini',
    ollama: 'Ollama',
  };

  const currentModels = PROVIDER_MODELS[provider] ?? [];
  const currentModel = provider === 'anthropic' ? anthropicModel
    : provider === 'openai' ? openaiModel
    : provider === 'gemini' ? geminiModel
    : '';
  const selectedModel = currentModels.find((m) => m.id === currentModel);

  return (
    <div className="container">
      <div className="card">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
          <h2 style={{ margin: 0 }}>{t('auditPage.newAudit')}</h2>
          {costStats && costStats.total_usd > 0 && (
            <span style={{ fontSize: 12, color: 'var(--muted)', background: 'var(--surface)', padding: '3px 10px', borderRadius: 12, border: '1px solid var(--border)' }}>
              {t('auditPage.costStats', { cost: costStats.total_usd.toFixed(3), count: costStats.audit_count })}
            </span>
          )}
        </div>
        <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <label>
            <span style={{ fontSize: 13, color: 'var(--muted)' }}>{t('auditPage.targetRepo')}</span>
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

          {/* Provider selector */}
          <div>
            <span style={{ fontSize: 13, color: 'var(--muted)' }}>{t('auditPage.aiProvider')}</span>
            <div style={{ display: 'flex', gap: 6, marginTop: 4, flexWrap: 'wrap' }}>
              {(['', 'anthropic', 'openai', 'gemini', 'ollama'] as Provider[]).map((p) => (
                <button
                  key={p}
                  type="button"
                  className={`btn${provider === p ? ' btn-primary' : ''}`}
                  style={{ padding: '5px 14px', fontSize: 13, background: provider === p ? undefined : 'transparent', border: '1px solid var(--border)', color: provider === p ? undefined : 'var(--muted)' }}
                  onClick={() => { setProvider(p); LS.set('audit.provider', p); }}
                >
                  {providerLabels[p]}
                </button>
              ))}
            </div>
          </div>

          {/* Anthropic fields */}
          {provider === 'anthropic' && (
            <>
              <label>
                <span style={{ fontSize: 13, color: 'var(--muted)' }}>
                  {t('auditPage.anthropicKeyLabel')}{' '}
                  <span style={{ fontWeight: 400 }}>{t('auditPage.anthropicKeyHint')}</span>
                </span>
                <input type="password" className="input" placeholder="sk-ant-…" value={anthropicKey}
                  onChange={(e) => { setAnthropicKey(e.target.value); LS.set('audit.anthropicKey', e.target.value); }}
                  style={{ marginTop: 4 }} />
              </label>
              <label>
                <span style={{ fontSize: 13, color: 'var(--muted)' }}>
                  {t('auditPage.modelLabel')}{selectedModel && <span style={{ fontWeight: 400 }}> — {splitGeneration ? t('auditPage.haikuSynthesizes', { model: selectedModel.label }) : selectedModel.hint}</span>}
                </span>
                <select className="input" value={anthropicModel}
                  onChange={(e) => { setAnthropicModel(e.target.value); LS.set('audit.model', e.target.value); }}
                  style={{ marginTop: 4 }}>
                  {PROVIDER_MODELS.anthropic.map((m) => <option key={m.id} value={m.id}>{m.label} — {m.hint}</option>)}
                </select>
              </label>
              <label style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <input type="checkbox" checked={splitGeneration}
                  onChange={(e) => { setSplitGeneration(e.target.checked); LS.set('audit.splitGeneration', String(e.target.checked)); }} />
                <span style={{ fontSize: 13, color: 'var(--muted)' }}>{t('auditPage.splitGenerationAnthropicHint')}</span>
              </label>
              {splitGeneration && (
                <label>
                  <span style={{ fontSize: 13, color: 'var(--muted)' }}>{t('auditPage.analysisModelLabel')} <span style={{ fontWeight: 400 }}>{t('auditPage.analysisModelEmptyHaiku')}</span></span>
                  <input type="text" className="input" placeholder="claude-haiku-4-5-20251001" value={analysisModel}
                    onChange={(e) => { setAnalysisModel(e.target.value); LS.set('audit.analysisModel', e.target.value); }}
                    style={{ marginTop: 4 }} />
                </label>
              )}
            </>
          )}

          {/* OpenAI fields */}
          {provider === 'openai' && (
            <>
              <label>
                <span style={{ fontSize: 13, color: 'var(--muted)' }}>
                  {t('auditPage.openaiKeyLabel')}{' '}
                  <span style={{ fontWeight: 400 }}>{t('auditPage.openaiKeyHint')}</span>
                </span>
                <input type="password" className="input" placeholder="sk-…" value={openaiKey}
                  onChange={(e) => { setOpenaiKey(e.target.value); LS.set('audit.openaiKey', e.target.value); }}
                  style={{ marginTop: 4 }} />
              </label>
              <label>
                <span style={{ fontSize: 13, color: 'var(--muted)' }}>
                  {t('auditPage.modelLabel')}{selectedModel && <span style={{ fontWeight: 400 }}> — {selectedModel.hint}</span>}
                </span>
                <select className="input" value={openaiModel}
                  onChange={(e) => { setOpenaiModel(e.target.value); LS.set('audit.openaiModel', e.target.value); }}
                  style={{ marginTop: 4 }}>
                  {PROVIDER_MODELS.openai.map((m) => <option key={m.id} value={m.id}>{m.label} — {m.hint}</option>)}
                </select>
              </label>
            </>
          )}

          {/* Gemini fields */}
          {provider === 'gemini' && (
            <>
              <label>
                <span style={{ fontSize: 13, color: 'var(--muted)' }}>
                  {t('auditPage.geminiKeyLabel')}{' '}
                  <span style={{ fontWeight: 400 }}>{t('auditPage.geminiKeyHint')}</span>
                </span>
                <input type="password" className="input" placeholder="AIza…" value={geminiKey}
                  onChange={(e) => { setGeminiKey(e.target.value); LS.set('audit.geminiKey', e.target.value); }}
                  style={{ marginTop: 4 }} />
              </label>
              <label>
                <span style={{ fontSize: 13, color: 'var(--muted)' }}>
                  {t('auditPage.modelLabel')}{selectedModel && <span style={{ fontWeight: 400 }}> — {selectedModel.hint}</span>}
                </span>
                <select className="input" value={geminiModel}
                  onChange={(e) => { setGeminiModel(e.target.value); LS.set('audit.geminiModel', e.target.value); }}
                  style={{ marginTop: 4 }}>
                  {PROVIDER_MODELS.gemini.map((m) => <option key={m.id} value={m.id}>{m.label} — {m.hint}</option>)}
                </select>
              </label>
            </>
          )}

          {/* Ollama fields */}
          {provider === 'ollama' && (
            <>
              <label>
                <span style={{ fontSize: 13, color: 'var(--muted)' }}>
                  {t('auditPage.ollamaUrlLabel')}{' '}
                  <span style={{ fontWeight: 400 }}>{t('auditPage.ollamaUrlHint')}</span>
                </span>
                <input type="text" className="input" placeholder="http://host.docker.internal:11434" value={ollamaURL}
                  onChange={(e) => { setOllamaURL(e.target.value); LS.set('audit.ollamaURL', e.target.value); }}
                  style={{ marginTop: 4 }} />
              </label>
              <label>
                <span style={{ fontSize: 13, color: 'var(--muted)' }}>{t('auditPage.finalReportModelLabel')} (<code>ollama list</code>)</span>
                <input type="text" className="input" placeholder="llama3.2, qwen2.5, deepseek-r1:8b, …" value={ollamaModel}
                  onChange={(e) => { setOllamaModel(e.target.value); LS.set('audit.ollamaModel', e.target.value); }}
                  style={{ marginTop: 4 }} />
              </label>
              <label style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <input type="checkbox" checked={splitGeneration}
                  onChange={(e) => { setSplitGeneration(e.target.checked); LS.set('audit.splitGeneration', String(e.target.checked)); }} />
                <span style={{ fontSize: 13, color: 'var(--muted)' }}>{t('auditPage.splitGenerationHint')}</span>
              </label>
              {splitGeneration && (
                <label>
                  <span style={{ fontSize: 13, color: 'var(--muted)' }}>{t('auditPage.analysisModelLabel')} <span style={{ fontWeight: 400 }}>{t('auditPage.analysisModelEmptyFinal')}</span></span>
                  <input type="text" className="input" placeholder="llama3.2, qwen2.5:7b, …" value={analysisModel}
                    onChange={(e) => { setAnalysisModel(e.target.value); LS.set('audit.analysisModel', e.target.value); }}
                    style={{ marginTop: 4 }} />
                </label>
              )}
            </>
          )}

          <label>
            <span style={{ fontSize: 13, color: 'var(--muted)' }}>
              {t('auditPage.githubTokenLabel')}{' '}
              <span style={{ fontWeight: 400 }}>{t('auditPage.githubTokenHint')}</span>
            </span>
            <input type="password" className="input" placeholder="ghp_…" value={token}
              onChange={(e) => { setToken(e.target.value); LS.set('audit.ghToken', e.target.value); }}
              style={{ marginTop: 4 }} />
          </label>

          <label style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <input type="checkbox" checked={skipSecrets}
              onChange={(e) => { setSkipSecrets(e.target.checked); LS.set('audit.skipSecrets', String(e.target.checked)); }} />
            <span style={{ fontSize: 13, color: 'var(--muted)' }}>{t('auditPage.skipSecretsLabel')}</span>
          </label>

          {error && <p className="error-msg">{error}</p>}
          <button type="submit" className="btn btn-primary" disabled={submitting || !repo.trim()}>
            {submitting ? t('common.starting') : t('auditPage.runAudit')}
          </button>
        </form>
        <p style={{ fontSize: 12, color: 'var(--muted)', marginTop: 8 }}>
          {provider === 'ollama'    ? t('auditPage.modeDescOllama')
          : provider === 'anthropic' ? t('auditPage.modeDescAnthropic')
          : provider === 'openai'   ? t('auditPage.modeDescOpenai')
          : provider === 'gemini'   ? t('auditPage.modeDescGemini')
          : t('auditPage.modeDescSnapshot')}
        </p>
      </div>

      <div className="card">
        <h2>{t('auditPage.auditHistory')}</h2>
        {audits.length === 0 ? (
          <p className="empty">{t('auditPage.noAudits')}</p>
        ) : (
          <div className="table-wrap">
            <table className="scans-table">
              <thead>
                <tr>
                  <th>{t('auditPage.colRepository')}</th>
                  <th>{t('auditPage.colStarted')}</th>
                  <th>{t('auditPage.colCompleted')}</th>
                  <th>{t('auditPage.colStatus')}</th>
                  <th>{t('auditPage.colMode')}</th>
                  <th>{t('auditPage.colCost')}</th>
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
                      <button className="btn btn-danger" onClick={(e) => deleteAudit(e, a.id)}>{t('common.delete')}</button>
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
