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

export const api = {
  listScans: () => request<Scan[]>('GET', '/api/scans'),
  getScan: (id: number) => request<Scan>('GET', `/api/scans/${id}`),
  createScan: (params: CreateScanParams) => request<Scan>('POST', '/api/scans', params),
  deleteScan: (id: number) => request<void>('DELETE', `/api/scans/${id}`),
  getResults: (id: number) => request<ScanResult[]>('GET', `/api/scans/${id}/results`),
  getTrending: (params: GetTrendingParams) => {
    const q = new URLSearchParams()
    if (params.language) q.set('language', params.language)
    if (params.since) q.set('since', params.since)
    if (params.token) q.set('token', params.token)
    const qs = q.toString()
    return request<ScanResult[]>('GET', `/api/trending${qs ? '?' + qs : ''}`)
  },
};
