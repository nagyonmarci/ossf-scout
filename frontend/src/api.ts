export type ScanStatus = 'running' | 'done' | 'error';

export interface Scan {
  id: number;
  created_at: string;
  finished_at: string | null;
  status: ScanStatus;
  error_msg?: string | null;
  language: string;
  min_stars: number;
  max_score: number;
  limit: number;
  workers: number;
  check_filter: string;
  topic: string;
  keyword: string;
  total_repos: number | null;
  result_count: number | null;
}

export interface ScanResult {
  id?: number;
  scan_id?: number;
  repo: string;
  stars: number;
  stars_today?: number;
  open_issues: number;
  score: number;
  language: string;
  description: string;
  weak_checks: string[];
  scorecard_url: string;
  repo_url: string;
}

export interface CreateScanParams {
  language: string;
  min_stars: number;
  max_score: number;
  limit: number;
  workers: number;
  check_filter: string;
  github_token: string;
  use_cli_fallback: boolean;
  pushed_after: string;
  min_maintained: number;
  topic: string;
  keyword: string;
  single_repo: string;
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch(path, {
    method,
    headers: body ? { 'Content-Type': 'application/json' } : {},
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(`${method} ${path} → ${res.status}: ${text}`);
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}

export interface GetTrendingParams {
  language?: string;
  since?: string;
  token?: string;
}

// ── Audit ─────────────────────────────────────────────────────────────────────

export type AuditStatus = 'pending' | 'running' | 'done' | 'error';

export interface Audit {
  id: string;
  repo: string;
  status: AuditStatus;
  model: string;
  provider: string;
  created_at: string;
  completed_at: string | null;
  report: string | null;
  error: string | null;
  input_tokens: number | null;
  output_tokens: number | null;
  has_context: boolean;
}

export interface CreateAuditParams {
  repo: string;
  github_token?: string;
  anthropic_key?: string;
  openai_key?: string;
  gemini_key?: string;
  model?: string;
  analysis_model?: string;
  split_generation?: boolean;
  provider?: string;
  ollama_url?: string;
}

export interface GenerateAuditParams {
  provider?: string;
  anthropic_key?: string;
  openai_key?: string;
  gemini_key?: string;
  model?: string;
  analysis_model?: string;
  split_generation?: boolean;
  ollama_url?: string;
}

// ── Schedules ─────────────────────────────────────────────────────────────────

export interface Schedule {
  id: string;
  repo: string;
  interval_h: number;
  provider: string;
  model: string;
  enabled: boolean;
  time_window_start: number;
  time_window_end: number;
  cli_fallback: boolean;
  auto_detected: boolean;
  detect_reason?: string;
  last_run_at: string | null;
  next_run_at: string;
  created_at: string;
}

export interface CreateScheduleParams {
  repo: string;
  interval_h?: number;
  provider?: string;
  model?: string;
  time_window_start?: number;
  time_window_end?: number;
  cli_fallback?: boolean;
}

export interface UpdateScheduleParams {
  interval_h: number;
  provider: string;
  model: string;
  enabled: boolean;
  time_window_start: number;
  time_window_end: number;
}

// ── Cost stats ────────────────────────────────────────────────────────────────

export interface CostStats {
  total_usd: number;
  by_model: Record<string, number>;
  total_input_tokens: number;
  total_output_tokens: number;
  audit_count: number;
}

export const api = {
  listScans: () => request<Scan[]>('GET', '/api/scans'),
  getScan: (id: number) => request<Scan>('GET', `/api/scans/${id}`),
  createScan: (params: CreateScanParams) => request<Scan>('POST', '/api/scans', params),
  deleteScan: (id: number) => request<void>('DELETE', `/api/scans/${id}`),
  getResults: (id: number) => request<ScanResult[]>('GET', `/api/scans/${id}/results`),
  listAudits: () => request<Audit[]>('GET', '/api/audits'),
  getAudit: (id: string) => request<Audit>('GET', `/api/audits/${id}`),
  createAudit: (params: CreateAuditParams) => request<Audit>('POST', '/api/audits', params),
  generateAudit: (id: string, params: GenerateAuditParams) => request<Audit>('POST', `/api/audits/${id}/generate`, params),
  deleteAudit: (id: string) => request<void>('DELETE', `/api/audits/${id}`),
  listSchedules: () => request<Schedule[]>('GET', '/api/schedules'),
  createSchedule: (params: CreateScheduleParams) => request<Schedule>('POST', '/api/schedules', params),
  updateSchedule: (id: string, params: UpdateScheduleParams) => request<void>('PUT', `/api/schedules/${id}`, params),
  deleteSchedule: (id: string) => request<void>('DELETE', `/api/schedules/${id}`),
  triggerSchedule: (id: string) => request<void>('POST', `/api/schedules/${id}/run`, {}),
  getCostStats: (days?: number) => request<CostStats>('GET', `/api/stats/costs${days ? `?days=${days}` : ''}`),
  getIssuesPRs: (owner: string, repo: string, refresh?: boolean) =>
    request<{ repo: string; summary: string; cached: string }>('GET', `/api/issues-prs/${owner}/${repo}${refresh ? '?refresh=true' : ''}`),
  getTrending: (params: GetTrendingParams) => {
    const q = new URLSearchParams()
    if (params.language) q.set('language', params.language)
    if (params.since) q.set('since', params.since)
    if (params.token) q.set('token', params.token)
    const qs = q.toString()
    return request<ScanResult[]>('GET', `/api/trending${qs ? '?' + qs : ''}`)
  },
};
