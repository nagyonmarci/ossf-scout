import { useState } from 'react'
import { ScanResult } from '../api'

type SortKey = 'repo' | 'stars' | 'score'

function scoreClass(score: number) {
  if (score === -1) return 'score score-na'
  if (score < 3) return 'score score-red'
  if (score < 5) return 'score score-orange'
  return 'score score-yellow'
}

function scoreLabel(score: number) {
  return score === -1 ? 'N/A' : score.toFixed(1)
}

export default function ResultsTable({ results }: { results: ScanResult[] }) {
  const [sortKey, setSortKey] = useState<SortKey>('score')
  const [asc, setAsc] = useState(true)
  const [filter, setFilter] = useState('')

  function toggleSort(key: SortKey) {
    if (sortKey === key) setAsc(a => !a)
    else { setSortKey(key); setAsc(true) }
  }

  const sorted = [...results].sort((a, b) => {
    let diff = 0
    if (sortKey === 'repo') diff = a.repo.localeCompare(b.repo)
    else if (sortKey === 'stars') diff = a.stars - b.stars
    else {
      const sa = a.score === -1 ? 999 : a.score
      const sb = b.score === -1 ? 999 : b.score
      diff = sa - sb
    }
    return asc ? diff : -diff
  })

  const filtered = sorted.filter(r => {
    if (!filter) return true
    const q = filter.toLowerCase()
    return r.repo.toLowerCase().includes(q)
      || r.description.toLowerCase().includes(q)
      || r.weak_checks.some(c => c.toLowerCase().includes(q))
  })

  function th(key: SortKey, label: string) {
    const active = sortKey === key
    return (
      <th onClick={() => toggleSort(key)}>
        {label}
        {active && <span className="sort-indicator">{asc ? ' ↑' : ' ↓'}</span>}
      </th>
    )
  }

  if (results.length === 0) {
    return <p className="empty">No results found for this scan.</p>
  }

  return (
    <>
      <input
        className="filter-input"
        placeholder="Filter by repo, description, check…"
        value={filter}
        onChange={e => setFilter(e.target.value)}
      />
      {filtered.length === 0 && (
        <p className="empty">No results match the filter.</p>
      )}
      {filtered.length > 0 && (
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                {th('repo', 'Repository')}
                {th('stars', 'Stars')}
                {th('score', 'Score')}
                <th>Weak Checks</th>
                <th>Description</th>
                <th>Links</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map(r => (
                <tr key={r.id}>
                  <td className="repo-name">{r.repo}</td>
                  <td>{r.stars.toLocaleString()}</td>
                  <td><span className={scoreClass(r.score)}>{scoreLabel(r.score)}</span></td>
                  <td>
                    <div className="tags">
                      {r.weak_checks.map(c => <span key={c} className="tag">{c}</span>)}
                    </div>
                  </td>
                  <td className="description">{r.description}</td>
                  <td className="links-cell">
                    <a href={r.repo_url} target="_blank" rel="noopener noreferrer">GitHub</a>
                    <a href={r.scorecard_url} target="_blank" rel="noopener noreferrer">Scorecard</a>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </>
  )
}
