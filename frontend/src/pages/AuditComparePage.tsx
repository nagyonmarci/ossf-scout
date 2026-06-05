import { useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import Markdown from 'react-markdown';
import { api } from '../api';

const PROVIDERS = ['anthropic', 'openai', 'gemini', 'ollama', ''] as const;
type Provider = typeof PROVIDERS[number];

const MODELS: Record<string, string[]> = {
  anthropic: ['claude-opus-4-8', 'claude-sonnet-4-6', 'claude-haiku-4-5-20251001'],
  openai: ['gpt-4o', 'gpt-4o-mini', 'o3-mini'],
  gemini: ['gemini-2.0-flash', 'gemini-1.5-pro', 'gemini-1.5-flash'],
};

interface ProviderConfig {
  provider: Provider;
  model: string;
  anthropicKey: string;
  openaiKey: string;
  geminiKey: string;
  ollamaUrl: string;
  ollamaModel: string;
}

const defaultConfig = (): ProviderConfig => ({
  provider: 'anthropic',
  model: 'claude-sonnet-4-6',
  anthropicKey: localStorage.getItem('audit.anthropicKey') || '',
  openaiKey: localStorage.getItem('audit.openaiKey') || '',
  geminiKey: localStorage.getItem('audit.geminiKey') || '',
  ollamaUrl: '',
  ollamaModel: '',
});

function ProviderForm({ label, cfg, onChange }: {
  label: string;
  cfg: ProviderConfig;
  onChange: (c: ProviderConfig) => void;
}) {
  const set = (patch: Partial<ProviderConfig>) => onChange({ ...cfg, ...patch });
  return (
    <div className="card" style={{ flex: 1, minWidth: 0 }}>
      <h3 style={{ margin: '0 0 10px', fontSize: 14 }}>{label}</h3>
      <div style={{ display: 'flex', gap: 4, flexWrap: 'wrap', marginBottom: 8 }}>
        {PROVIDERS.map(p => (
          <button
            key={p}
            className={`btn${cfg.provider === p ? ' btn-primary' : ''}`}
            style={{ padding: '3px 10px', fontSize: 12, background: cfg.provider === p ? undefined : 'transparent', border: '1px solid var(--border)' }}
            onClick={() => set({ provider: p, model: MODELS[p]?.[0] || '' })}
          >
            {p || 'Snapshot'}
          </button>
        ))}
      </div>
      {cfg.provider === 'anthropic' && (
        <>
          <input className="input" style={{ marginBottom: 6, fontSize: 12 }} type="password"
            placeholder="Anthropic API key" value={cfg.anthropicKey}
            onChange={e => { set({ anthropicKey: e.target.value }); localStorage.setItem('audit.anthropicKey', e.target.value); }} />
          <select className="input" style={{ fontSize: 12 }} value={cfg.model}
            onChange={e => set({ model: e.target.value })}>
            {MODELS.anthropic.map(m => <option key={m} value={m}>{m}</option>)}
          </select>
        </>
      )}
      {cfg.provider === 'openai' && (
        <>
          <input className="input" style={{ marginBottom: 6, fontSize: 12 }} type="password"
            placeholder="OpenAI API key" value={cfg.openaiKey}
            onChange={e => { set({ openaiKey: e.target.value }); localStorage.setItem('audit.openaiKey', e.target.value); }} />
          <select className="input" style={{ fontSize: 12 }} value={cfg.model}
            onChange={e => set({ model: e.target.value })}>
            {MODELS.openai.map(m => <option key={m} value={m}>{m}</option>)}
          </select>
        </>
      )}
      {cfg.provider === 'gemini' && (
        <>
          <input className="input" style={{ marginBottom: 6, fontSize: 12 }} type="password"
            placeholder="Gemini API key" value={cfg.geminiKey}
            onChange={e => { set({ geminiKey: e.target.value }); localStorage.setItem('audit.geminiKey', e.target.value); }} />
          <select className="input" style={{ fontSize: 12 }} value={cfg.model}
            onChange={e => set({ model: e.target.value })}>
            {MODELS.gemini.map(m => <option key={m} value={m}>{m}</option>)}
          </select>
        </>
      )}
      {cfg.provider === 'ollama' && (
        <>
          <input className="input" style={{ marginBottom: 6, fontSize: 12 }} type="text"
            placeholder="Ollama URL (blank = server default)" value={cfg.ollamaUrl}
            onChange={e => set({ ollamaUrl: e.target.value })} />
          <input className="input" style={{ fontSize: 12 }} type="text"
            placeholder="Model (e.g. llama3)" value={cfg.ollamaModel}
            onChange={e => set({ ollamaModel: e.target.value })} />
        </>
      )}
    </div>
  );
}

function cfgToParams(cfg: ProviderConfig) {
  return {
    repo: '',
    provider: cfg.provider || undefined,
    model: cfg.provider === 'ollama' ? cfg.ollamaModel || undefined : cfg.model || undefined,
    anthropic_key: cfg.provider === 'anthropic' ? cfg.anthropicKey || undefined : undefined,
    openai_key: cfg.provider === 'openai' ? cfg.openaiKey || undefined : undefined,
    gemini_key: cfg.provider === 'gemini' ? cfg.geminiKey || undefined : undefined,
    ollama_url: cfg.provider === 'ollama' ? cfg.ollamaUrl || undefined : undefined,
  };
}

type CompareResult = {
  report: string; input_tokens: number; output_tokens: number; error?: string; provider: string; model: string;
};

function ResultPanel({ result, label }: { result: CompareResult; label: string }) {
  const label2 = result.model ? `${result.provider || 'snapshot'} / ${result.model}` : (result.provider || 'snapshot');
  return (
    <div style={{ flex: 1, minWidth: 0, display: 'flex', flexDirection: 'column', gap: 6 }}>
      <div style={{ fontSize: 12, color: 'var(--muted)', fontWeight: 600 }}>
        {label}: <span style={{ color: 'var(--fg)' }}>{label2}</span>
        {result.input_tokens > 0 && (
          <span style={{ marginLeft: 8 }}>
            {result.input_tokens.toLocaleString()} in / {result.output_tokens.toLocaleString()} out
          </span>
        )}
      </div>
      {result.error ? (
        <p className="error-msg">{result.error}</p>
      ) : (
        <div className="card audit-report" style={{ flex: 1, overflow: 'auto', maxHeight: '80vh' }}>
          <Markdown>{result.report}</Markdown>
        </div>
      )}
    </div>
  );
}

export default function AuditComparePage() {
  const { id } = useParams<{ id: string }>();
  const [cfgA, setCfgA] = useState<ProviderConfig>(defaultConfig);
  const [cfgB, setCfgB] = useState<ProviderConfig>(() => ({ ...defaultConfig(), provider: 'openai', model: 'gpt-4o' }));
  const [running, setRunning] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [results, setResults] = useState<{ a: CompareResult; b: CompareResult } | null>(null);

  async function run() {
    if (!id) return;
    setRunning(true);
    setError(null);
    setResults(null);
    try {
      const res = await api.compareAudit(id, cfgToParams(cfgA) as any, cfgToParams(cfgB) as any);
      setResults(res);
    } catch (e) {
      setError(String(e));
    } finally {
      setRunning(false);
    }
  }

  return (
    <div className="container">
      <Link to={`/audits/${id}`} className="back-link">← Back to audit</Link>
      <h2>Compare models — audit {id?.slice(0, 8)}…</h2>
      <p style={{ color: 'var(--muted)', fontSize: 13, marginTop: -8 }}>
        Run two providers in parallel against the same saved context and compare reports side by side.
        Results are NOT saved — this is a preview only.
      </p>

      {error && <p className="error-msg">{error}</p>}

      <div style={{ display: 'flex', gap: 12, marginBottom: 12, flexWrap: 'wrap' }}>
        <ProviderForm label="Model A" cfg={cfgA} onChange={setCfgA} />
        <ProviderForm label="Model B" cfg={cfgB} onChange={setCfgB} />
      </div>

      <button className="btn btn-primary" onClick={run} disabled={running}>
        {running ? 'Running both providers… (this takes 1–3 min each)' : 'Run comparison'}
      </button>

      {running && (
        <p style={{ color: 'var(--muted)', fontSize: 13, marginTop: 8 }}>
          Both providers are running in parallel. Results will appear when both complete.
        </p>
      )}

      {results && (
        <div style={{ display: 'flex', gap: 12, marginTop: 16, flexWrap: 'wrap' }}>
          <ResultPanel result={results.a} label="A" />
          <ResultPanel result={results.b} label="B" />
        </div>
      )}
    </div>
  );
}
