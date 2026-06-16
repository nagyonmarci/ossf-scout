import { useEffect, useState, useCallback, useRef } from 'react';
import { Link, useParams, useNavigate } from 'react-router-dom';
import Markdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { api, Audit } from '../api';
import StatusBadge from '../components/StatusBadge';
import { useLang } from '../i18n';

interface SCNode { action: string; tag: string; sha: string; resolved: boolean; files: string[] }

function SupplyChainGraph({ auditId }: { auditId: string }) {
  const { t } = useLang();
  const [nodes, setNodes] = useState<SCNode[] | null>(null);
  const [repo, setRepo] = useState('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch(`/api/audits/${auditId}/supply-chain`)
      .then(r => r.json())
      .then(d => { setNodes(d.nodes ?? []); setRepo(d.repo ?? ''); })
      .catch(() => setNodes([]))
      .finally(() => setLoading(false));
  }, [auditId]);

  if (loading) return <p style={{ color: 'var(--muted)', fontSize: 13 }}>{t('auditDetail.loadingSupplyChain')}</p>;
  if (!nodes || nodes.length === 0) return <p style={{ color: 'var(--muted)', fontSize: 13 }}>{t('auditDetail.noSupplyChain')}</p>;

  const NODE_W = 220, NODE_H = 44, GAP_X = 40, GAP_Y = 20;
  const ROOT_W = 160, ROOT_H = 36;
  const svgW = ROOT_W + GAP_X + NODE_W + 32;
  const svgH = Math.max(ROOT_H + 20, nodes.length * (NODE_H + GAP_Y));
  const rootX = 8, rootY = svgH / 2 - ROOT_H / 2;
  const nodesX = rootX + ROOT_W + GAP_X;

  return (
    <div style={{ overflowX: 'auto', marginTop: 8 }}>
      <svg width={svgW} height={svgH} style={{ display: 'block', fontFamily: 'monospace' }}>
        {/* Root node */}
        <rect x={rootX} y={rootY} width={ROOT_W} height={ROOT_H} rx={6}
          fill="rgba(100,149,237,0.15)" stroke="rgba(100,149,237,0.6)" strokeWidth={1.5} />
        <text x={rootX + ROOT_W / 2} y={rootY + ROOT_H / 2 + 5} textAnchor="middle"
          fill="var(--fg, #e0e0e0)" fontSize={12} fontWeight={600}>
          {repo.split('/')[1] || repo}
        </text>

        {nodes.map((n, i) => {
          const ny = i * (NODE_H + GAP_Y) + GAP_Y / 2;
          const color = n.resolved ? 'rgba(76,175,80,0.15)' : 'rgba(255,152,0,0.15)';
          const stroke = n.resolved ? 'rgba(76,175,80,0.6)' : 'rgba(255,152,0,0.6)';
          const cy = ny + NODE_H / 2;
          return (
            <g key={n.action + n.tag + i}>
              <line x1={rootX + ROOT_W} y1={rootY + ROOT_H / 2} x2={nodesX} y2={cy}
                stroke="var(--border, #444)" strokeWidth={1} strokeDasharray="4 3" />
              <rect x={nodesX} y={ny} width={NODE_W} height={NODE_H} rx={5}
                fill={color} stroke={stroke} strokeWidth={1.5} />
              <text x={nodesX + 8} y={ny + 16} fill="var(--fg, #e0e0e0)" fontSize={11} fontWeight={600}>
                {n.action.length > 28 ? '…' + n.action.slice(-27) : n.action}
              </text>
              <text x={nodesX + 8} y={ny + 30} fill="var(--muted, #888)" fontSize={10}>
                {n.tag ? `@${n.tag.slice(0, 20)}` : '(no tag)'}{' '}
                {n.resolved ? `· ${n.sha.slice(0, 8)}` : '· unresolved'}
              </text>
            </g>
          );
        })}
      </svg>
      <p style={{ fontSize: 11, color: 'var(--muted)', marginTop: 4 }}>
        {t('auditDetail.supplyChainLegend', { count: nodes.length })}
      </p>
    </div>
  );
}

const ALL_MODELS: Record<string, { id: string; label: string }[]> = {
  anthropic: [
    { id: 'claude-opus-4-8',           label: 'Opus 4'   },
    { id: 'claude-sonnet-4-6',         label: 'Sonnet 4' },
    { id: 'claude-haiku-4-5-20251001', label: 'Haiku 4'  },
  ],
  openai: [
    { id: 'gpt-4o',      label: 'GPT-4o'      },
    { id: 'gpt-4o-mini', label: 'GPT-4o mini' },
    { id: 'o3-mini',     label: 'o3-mini'     },
  ],
  gemini: [
    { id: 'gemini-2.0-flash', label: 'Gemini 2.0 Flash' },
    { id: 'gemini-1.5-pro',   label: 'Gemini 1.5 Pro'   },
    { id: 'gemini-1.5-flash', label: 'Gemini 1.5 Flash'  },
  ],
};

const LS = {
  get: (k: string, fallback: string) => localStorage.getItem(k) ?? fallback,
  set: (k: string, v: string) => localStorage.setItem(k, v),
};

export default function AuditDetail() {
  const { t } = useLang();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [audit, setAudit] = useState<Audit | null>(null);
  const [error, setError] = useState<string | null>(null);
  const prevStatus = useRef<string | null>(null);

  // generate-form state
  const [showGenForm, setShowGenForm] = useState(false);
  const [genProvider, setGenProvider] = useState(() => LS.get('audit.provider', 'ollama'));
  const [genApiKey, setGenApiKey] = useState(() => LS.get('audit.anthropicKey', ''));
  const [genOpenAIKey, setGenOpenAIKey] = useState(() => LS.get('audit.openaiKey', ''));
  const [genGeminiKey, setGenGeminiKey] = useState(() => LS.get('audit.geminiKey', ''));
  const [genModel, setGenModel] = useState(() => LS.get('audit.model', ALL_MODELS.anthropic[0].id));
  const [genOllamaURL, setGenOllamaURL] = useState(() => LS.get('audit.ollamaURL', ''));
  const [genOllamaModel, setGenOllamaModel] = useState(() => LS.get('audit.ollamaModel', ''));
  const [genSplitGeneration, setGenSplitGeneration] = useState(() => LS.get('audit.splitGeneration', 'true') === 'true');
  const [genAnalysisModel, setGenAnalysisModel] = useState(() => LS.get('audit.analysisModel', ''));
  const [genSubmitting, setGenSubmitting] = useState(false);
  const [genError, setGenError] = useState<string | null>(null);
  const [extracting, setExtracting] = useState(false);
  const [extractMsg, setExtractMsg] = useState<string | null>(null);

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
        openai_key:    genProvider === 'openai'    ? genOpenAIKey || undefined : undefined,
        gemini_key:    genProvider === 'gemini'    ? genGeminiKey || undefined : undefined,
        model: genProvider === 'anthropic' ? genModel || undefined
             : genProvider === 'openai'    ? genModel || undefined
             : genProvider === 'gemini'    ? genModel || undefined
             : genProvider === 'ollama'    ? genOllamaModel || undefined
             : undefined,
        split_generation: (genProvider === 'anthropic' || genProvider === 'ollama') ? genSplitGeneration : undefined,
        analysis_model: genSplitGeneration ? genAnalysisModel || undefined : undefined,
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
        {t('auditDetail.backToAudits')}
      </Link>

      {error && <p className="error-msg">{error}</p>}

      {audit && (
        <>
          <div className="card">
            <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 16 }}>
              <h2 style={{ margin: 0 }}>
                {t('auditDetail.auditHeading')} <code>{audit.repo}</code>
              </h2>
              <StatusBadge status={audit.status} />
            </div>

            <div className="detail-meta">
              <span className="meta-item">
                <span className="meta-label">{t('auditDetail.started')}</span>
                {new Date(audit.created_at).toLocaleString()}
              </span>
              {audit.completed_at && (
                <span className="meta-item">
                  <span className="meta-label">{t('auditDetail.completed')}</span>
                  {new Date(audit.completed_at).toLocaleString()}
                </span>
              )}
              <span className="meta-item">
                <span className="meta-label">{t('auditDetail.mode')}</span>
                {audit.provider === 'ollama'
                  ? `Ollama · ${audit.model || '?'}`
                  : audit.provider === 'anthropic'
                  ? audit.model || 'Anthropic'
                  : audit.provider === 'openai'
                  ? audit.model || 'OpenAI'
                  : audit.provider === 'gemini'
                  ? audit.model || 'Gemini'
                  : t('auditDetail.staticSnapshot')}
              </span>
              {(audit.input_tokens ?? 0) > 0 && (
                <span className="meta-item">
                  <span className="meta-label">{t('auditDetail.tokensInOut')}</span>
                  {audit.input_tokens!.toLocaleString()} / {(audit.output_tokens ?? 0).toLocaleString()}
                </span>
              )}
            </div>

            {(audit.status === 'pending' || audit.status === 'running') && (
              <p style={{ color: 'var(--muted)', fontSize: 13 }}>
                {audit.status === 'pending'
                  ? t('auditDetail.waitingToStart')
                  : t('auditDetail.cloningAndAnalyzing')}
              </p>
            )}
            {audit.status === 'error' && <p className="error-msg">{t('auditDetail.error', { msg: audit.error ?? '' })}</p>}
            <div style={{ display: 'flex', gap: 8, marginTop: 8, flexWrap: 'wrap' }}>
              {(audit.status === 'done' || (audit.status === 'error' && audit.report)) && (
                <>
                  <button className="btn btn-primary" onClick={downloadReport}>
                    {t('auditDetail.downloadMd')}
                  </button>
                  <a className="btn no-print" href={`/api/audits/${audit.id}/export.pdf`} download>
                    {t('auditDetail.downloadPdf')}
                  </a>
                </>
              )}
              {audit.has_context && (
                <>
                  <a className="btn" href={`/api/audits/${audit.id}/context.md`} download>
                    {t('auditDetail.downloadAiContext')}
                  </a>
                  <Link
                    to={`/audits/${audit.id}/compare`}
                    className="btn"
                    style={{ background: 'transparent', border: '1px solid var(--border)' }}
                  >
                    {t('auditDetail.compareModels')}
                  </Link>
                </>
              )}
              {(audit.status === 'done' || (audit.status === 'error' && audit.report)) && (
                <>
                  <a className="btn" href={`/api/audits/${audit.id}/export.sarif`} download>
                    {t('auditDetail.exportSarif')}
                  </a>
                  <a className="btn" href={`/api/audits/${audit.id}/export.json`} download>
                    {t('auditDetail.exportJson')}
                  </a>
                </>
              )}
              {(audit.status === 'done' || (audit.status === 'error' && audit.report)) && (
                <>
                  <Link
                    to={`/remediation?audit_id=${audit.id}`}
                    className="btn"
                    style={{ background: 'transparent', border: '1px solid var(--border)' }}
                  >
                    {t('auditDetail.trackFindings')}
                  </Link>
                  <button
                    className="btn"
                    style={{ background: 'transparent', border: '1px solid var(--border)' }}
                    disabled={extracting}
                    onClick={async () => {
                      setExtracting(true);
                      setExtractMsg(null);
                      try {
                        const r = await api.extractFindings(audit!.id);
                        setExtractMsg(t('auditDetail.findingsAdded', { count: r.created }));
                      } catch (e) {
                        setExtractMsg(t('auditDetail.extractFailed', { error: String(e) }));
                      } finally {
                        setExtracting(false);
                      }
                    }}
                  >
                    {extracting ? t('auditDetail.extracting') : t('auditDetail.extractToTracker')}
                  </button>
                  {extractMsg && <span style={{ fontSize: 12, color: 'var(--muted)', alignSelf: 'center' }}>{extractMsg}</span>}
                </>
              )}
              {canRunWithAI && (
                audit.has_context ? (
                  <button
                    className="btn"
                    style={{ background: 'transparent', border: '1px solid var(--accent)', color: 'var(--accent)' }}
                    onClick={() => setShowGenForm(v => !v)}
                  >
                    {t('auditDetail.runWithAi')}
                  </button>
                ) : (
                  <button
                    className="btn"
                    style={{ background: 'transparent', border: '1px solid var(--accent)', color: 'var(--accent)' }}
                    onClick={() => navigate(`/audits?repo=${encodeURIComponent(audit!.repo)}`)}
                  >
                    {t('auditDetail.runWithAi')}
                  </button>
                )
              )}
            </div>

            {showGenForm && audit.has_context && (
              <div style={{ marginTop: 12, padding: 12, border: '1px solid var(--border)', borderRadius: 6 }}>
                <p style={{ margin: '0 0 8px', fontSize: 13, color: 'var(--muted)' }}>
                  {t('auditDetail.regenerateHint')}
                </p>
                <div style={{ display: 'flex', gap: 6, marginBottom: 10, flexWrap: 'wrap' }}>
                  {(['ollama', 'anthropic', 'openai', 'gemini', ''] as const).map(p => (
                    <button
                      key={p}
                      className={`btn${genProvider === p ? ' btn-primary' : ''}`}
                      style={{ padding: '4px 12px', fontSize: 13, background: genProvider === p ? undefined : 'transparent', border: '1px solid var(--border)', color: genProvider === p ? undefined : 'var(--muted)' }}
                      onClick={() => { setGenProvider(p); LS.set('audit.provider', p); }}
                    >
                      {p === '' ? t('auditDetail.snapshot') : p === 'anthropic' ? 'Anthropic' : p === 'openai' ? 'OpenAI' : p === 'gemini' ? 'Gemini' : 'Ollama'}
                    </button>
                  ))}
                </div>
                {genProvider === 'anthropic' && (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: 8 }}>
                    <input type="password" className="input" placeholder={t('auditDetail.anthropicApiKeyPlaceholder')} value={genApiKey}
                      onChange={e => { setGenApiKey(e.target.value); LS.set('audit.anthropicKey', e.target.value); }} />
                    <select className="input" value={genModel}
                      onChange={e => { setGenModel(e.target.value); LS.set('audit.model', e.target.value); }}>
                      {ALL_MODELS.anthropic.map(m => <option key={m.id} value={m.id}>{m.label}</option>)}
                    </select>
                    <label style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                      <input
                        type="checkbox"
                        checked={genSplitGeneration}
                        onChange={e => { setGenSplitGeneration(e.target.checked); LS.set('audit.splitGeneration', String(e.target.checked)); }}
                      />
                      <span style={{ fontSize: 13, color: 'var(--muted)' }}>{t('auditPage.splitGenerationHint')}</span>
                    </label>
                    {genSplitGeneration && (
                      <input type="text" className="input" placeholder={t('auditDetail.analysisModelOptionalPlaceholder')}
                        value={genAnalysisModel} onChange={e => { setGenAnalysisModel(e.target.value); LS.set('audit.analysisModel', e.target.value); }} />
                    )}
                  </div>
                )}
                {genProvider === 'openai' && (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: 8 }}>
                    <input type="password" className="input" placeholder={t('auditDetail.openaiApiKeyPlaceholder')} value={genOpenAIKey}
                      onChange={e => { setGenOpenAIKey(e.target.value); LS.set('audit.openaiKey', e.target.value); }} />
                    <select className="input" value={genModel}
                      onChange={e => { setGenModel(e.target.value); LS.set('audit.model', e.target.value); }}>
                      {ALL_MODELS.openai.map(m => <option key={m.id} value={m.id}>{m.label}</option>)}
                    </select>
                  </div>
                )}
                {genProvider === 'gemini' && (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: 8 }}>
                    <input type="password" className="input" placeholder={t('auditDetail.geminiApiKeyPlaceholder')} value={genGeminiKey}
                      onChange={e => { setGenGeminiKey(e.target.value); LS.set('audit.geminiKey', e.target.value); }} />
                    <select className="input" value={genModel}
                      onChange={e => { setGenModel(e.target.value); LS.set('audit.model', e.target.value); }}>
                      {ALL_MODELS.gemini.map(m => <option key={m.id} value={m.id}>{m.label}</option>)}
                    </select>
                  </div>
                )}
                {genProvider === 'ollama' && (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: 8 }}>
                    <input
                      type="text"
                      className="input"
                      placeholder={t('auditDetail.ollamaUrlPlaceholder')}
                      value={genOllamaURL}
                      onChange={e => { setGenOllamaURL(e.target.value); LS.set('audit.ollamaURL', e.target.value); }}
                    />
                    <input
                      type="text"
                      className="input"
                      placeholder={t('auditDetail.finalReportModelPlaceholder')}
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
                        {t('auditPage.splitGenerationHint')}
                      </span>
                    </label>
                    {genSplitGeneration && (
                      <input
                        type="text"
                        className="input"
                        placeholder={t('auditDetail.analysisModelOptionalPlaceholder')}
                        value={genAnalysisModel}
                        onChange={e => { setGenAnalysisModel(e.target.value); LS.set('audit.analysisModel', e.target.value); }}
                      />
                    )}
                  </div>
                )}
                <div style={{ display: 'flex', gap: 8 }}>
                  <button className="btn btn-primary" onClick={submitGenerate} disabled={genSubmitting}>
                    {genSubmitting ? t('common.starting') : t('auditDetail.generateReport')}
                  </button>
                  <button
                    className="btn"
                    style={{ background: 'transparent', border: '1px solid var(--border)' }}
                    onClick={() => setShowGenForm(false)}
                  >
                    {t('common.cancel')}
                  </button>
                </div>
                {genError && <p className="error-msg" style={{ marginTop: 6 }}>{genError}</p>}
              </div>
            )}
          </div>

          {audit.report && (audit.status === 'done' || audit.status === 'error') && (
            <div className="card audit-report">
              <Markdown remarkPlugins={[remarkGfm]}>{audit.report}</Markdown>
            </div>
          )}

          {audit.has_context && (audit.status === 'done' || audit.status === 'error') && (
            <div className="card">
              <h3 style={{ margin: '0 0 8px', fontSize: 14 }}>{t('auditDetail.supplyChainGraph')}</h3>
              <p style={{ fontSize: 12, color: 'var(--muted)', margin: '0 0 8px' }}>
                {t('auditDetail.supplyChainDesc')}
              </p>
              <SupplyChainGraph auditId={audit.id} />
            </div>
          )}
        </>
      )}
    </div>
  );
}
