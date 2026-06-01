import { useState } from 'react'
import { ScanResult } from '../api'

type SortKey = 'repo' | 'stars' | 'score'

const DEFAULT_WIDTHS = [200, 80, 80, 220, 320, 100]
const MIN_COL_WIDTH = 60

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
  const [colWidths, setColWidths] = useState(DEFAULT_WIDTHS)

  function toggleSort(key: SortKey) {
    if (sortKey === key) setAsc(a => !a)
    else { setSortKey(key); setAsc(true) }
  }

  function startResize(colIdx: number, e: React.MouseEvent) {
    e.preventDefault()
    e.stopPropagation()
    const startX = e.clientX
    const startWidth = colWidths[colIdx]

    function onMove(ev: MouseEvent) {
      const newWidth = Math.max(MIN_COL_WIDTH, startWidth + ev.clientX - startX)
      setColWidths(prev => prev.map((w, i) => i === colIdx ? newWidth : w))
    }
    function onUp() {
      document.removeEventListener('mousemove', onMove)
      document.removeEventListener('mouseup', onUp)
    }
    document.addEventListener('mousemove', onMove)
    document.addEventListener('mouseup', onUp)
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

  function th(colIdx: number, key: SortKey | null, label: string) {
    const active = key !== null && sortKey === key
    return (
      <th
        style={{ position: 'relative' }}
        onClick={key ? () => toggleSort(key) : undefined}
      >
        {label}
        {active && <span className="sort-indicator">{asc ? ' ↑' : ' ↓'}</span>}
        <div className="resize-handle" onMouseDown={e => startResize(colIdx, e)} />
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
          <table style={{ tableLayout: 'fixed', width: '100%' }}>
            <colgroup>
              {colWidths.map((w, i) => <col key={i} style={{ width: w }} />)}
            </colgroup>
            <thead>
              <tr>
                {th(0, 'repo', 'Repository')}
                {th(1, 'stars', 'Stars')}
                {th(2, 'score', 'Score')}
                {th(3, null, 'Weak Checks')}
                {th(4, null, 'Description')}
                {th(5, null, 'Links')}
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
                    {r.scorecard_url
                      ? <a href={r.scorecard_url} target="_blank" rel="noopener noreferrer">Scorecard</a>
                      : <span style={{ color: 'var(--muted)', fontSize: 12 }}>CLI scan</span>
                    }
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
