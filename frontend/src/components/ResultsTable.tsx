import { useState, useRef, useEffect } from 'react'
import { ScanResult } from '../api'

type SortKey = 'repo' | 'stars' | 'issues' | 'score'

const DEFAULT_WIDTHS = [220, 70, 70, 80, 240, 0]  // 0 = Description takes remaining space
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

function parseCheckTag(tag: string): { name: string; checkScore: string } {
  const m = tag.match(/^(.+)\((-?\d+)\)$/)
  if (!m) return { name: tag, checkScore: '' }
  return { name: m[1], checkScore: m[2] }
}

function checkDocsUrl(name: string): string {
  return `https://github.com/ossf/scorecard/blob/main/docs/checks.md#${name.toLowerCase()}`
}

function CheckTag({ tag }: { tag: string }) {
  const { name, checkScore } = parseCheckTag(tag)
  const scoreNum = parseInt(checkScore, 10)
  const scoreColor = scoreNum === -1 ? 'var(--muted)' : scoreNum < 3 ? 'var(--red)' : 'var(--orange)'
  return (
    <a
      href={checkDocsUrl(name)}
      target="_blank"
      rel="noopener noreferrer"
      className="tag tag-link"
      title={`View ${name} check documentation`}
    >
      {name}
      {checkScore !== '' && (
        <span className="tag-score" style={{ color: scoreColor }}>
          {scoreNum === -1 ? 'N/A' : scoreNum}
        </span>
      )}
    </a>
  )
}

const COLS: { label: string; sortKey: SortKey | null }[] = [
  { label: 'Repository', sortKey: 'repo' },
  { label: 'Stars',      sortKey: 'stars' },
  { label: 'Issues',     sortKey: 'issues' },
  { label: 'Score',      sortKey: 'score' },
  { label: 'Weak Checks', sortKey: null },
  { label: 'Description', sortKey: null },
]

export default function ResultsTable({ results }: { results: ScanResult[] }) {
  const [sortKey, setSortKey] = useState<SortKey>('score')
  const [asc, setAsc] = useState(true)
  const [filter, setFilter] = useState('')
  const [hideNA, setHideNA] = useState(false)
  const [colWidths, setColWidths] = useState(DEFAULT_WIDTHS.slice(0, COLS.length))
  const [stickyVisible, setStickyVisible] = useState(false)
  const [tableLeft, setTableLeft] = useState(0)
  const [tableWidth, setTableWidth] = useState(0)
  const theadRef = useRef<HTMLTableSectionElement>(null)
  const tableRef = useRef<HTMLTableElement>(null)

  useEffect(() => {
    const el = theadRef.current
    if (!el) return
    const updatePos = () => {
      if (tableRef.current) {
        const r = tableRef.current.getBoundingClientRect()
        setTableLeft(r.left)
        setTableWidth(r.width)
      }
    }
    const observer = new IntersectionObserver(
      ([entry]) => {
        setStickyVisible(entry.boundingClientRect.top < 0)
        updatePos()
      },
      { threshold: 0 }
    )
    observer.observe(el)
    window.addEventListener('scroll', updatePos, { passive: true })
    window.addEventListener('resize', updatePos, { passive: true })
    return () => {
      observer.disconnect()
      window.removeEventListener('scroll', updatePos)
      window.removeEventListener('resize', updatePos)
    }
  }, [])

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
    else if (sortKey === 'issues') diff = a.open_issues - b.open_issues
    else {
      const sa = a.score === -1 ? 999 : a.score
      const sb = b.score === -1 ? 999 : b.score
      diff = sa - sb
    }
    return asc ? diff : -diff
  })

  const filtered = sorted.filter(r => {
    if (hideNA && r.score === -1) return false
    if (!filter) return true
    const q = filter.toLowerCase()
    return r.repo.toLowerCase().includes(q)
      || r.description.toLowerCase().includes(q)
      || r.weak_checks.some(c => c.toLowerCase().includes(q))
  })

  function renderTh(colIdx: number, col: typeof COLS[number]) {
    const active = col.sortKey !== null && sortKey === col.sortKey
    return (
      <th
        key={col.label}
        style={{ position: 'relative', width: colWidths[colIdx] || undefined }}
        onClick={col.sortKey ? () => toggleSort(col.sortKey!) : undefined}
      >
        {col.label}
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
      <div className="filter-bar">
        <input
          className="filter-input"
          placeholder="Filter by repo, description, check…"
          value={filter}
          onChange={e => setFilter(e.target.value)}
        />
        <label className="filter-toggle">
          <input
            type="checkbox"
            checked={hideNA}
            onChange={e => setHideNA(e.target.checked)}
            style={{ width: 'auto' }}
          />
          Hide N/A
        </label>
      </div>

      {filtered.length === 0 && <p className="empty">No results match the filter.</p>}

      {filtered.length > 0 && (
        <>
          {stickyVisible && (
            <div className="sticky-header" style={{ left: tableLeft, width: tableWidth }}>
              <table style={{ tableLayout: 'fixed', width: tableWidth }}>
                <colgroup>
                  {colWidths.map((w, i) => <col key={i} style={{ width: w || undefined }} />)}
                </colgroup>
                <thead>
                  <tr>{COLS.map((col, i) => renderTh(i, col))}</tr>
                </thead>
              </table>
            </div>
          )}

          <div className="table-wrap">
            <table ref={tableRef} style={{ tableLayout: 'fixed', width: '100%' }}>
              <colgroup>
                {colWidths.map((w, i) => <col key={i} style={{ width: w || undefined }} />)}
              </colgroup>
              <thead ref={theadRef}>
                <tr>{COLS.map((col, i) => renderTh(i, col))}</tr>
              </thead>
              <tbody>
                {filtered.map(r => (
                  <tr key={r.repo}>
                    <td className="repo-name">
                      <a href={r.repo_url} target="_blank" rel="noopener noreferrer">{r.repo}</a>
                    </td>
                    <td>
                      {r.stars.toLocaleString()}
                      {(r.stars_today ?? 0) > 0 && (
                        <span className="stars-today"> +{r.stars_today!.toLocaleString()}</span>
                      )}
                    </td>
                    <td title="Open issues + pull requests">{r.open_issues.toLocaleString()}</td>
                    <td>
                      {r.scorecard_url
                        ? <a href={r.scorecard_url} target="_blank" rel="noopener noreferrer" className={scoreClass(r.score)}>{scoreLabel(r.score)}</a>
                        : <span className={scoreClass(r.score)} title="CLI scan — not indexed online">{scoreLabel(r.score)}</span>
                      }
                    </td>
                    <td>
                      <div className="tags">
                        {r.weak_checks.map(c => <CheckTag key={c} tag={c} />)}
                      </div>
                    </td>
                    <td className="description">{r.description}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      )}
    </>
  )
}
