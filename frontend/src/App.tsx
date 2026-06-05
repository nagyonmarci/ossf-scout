import { BrowserRouter, Routes, Route, NavLink, Outlet } from 'react-router-dom'
import { useEffect, useState } from 'react'
import ScanList from './pages/ScanList'
import ScanDetail from './pages/ScanDetail'
import TrendingPage from './pages/TrendingPage'
import AuditPage from './pages/AuditPage'
import AuditDetail from './pages/AuditDetail'
import SchedulePage from './pages/SchedulePage'
import PortfolioPage from './pages/PortfolioPage'
import RemediationPage from './pages/RemediationPage'
import OrgScanPage from './pages/OrgScanPage'
import AuditComparePage from './pages/AuditComparePage'

function AppHeader() {
  return (
    <div className="app-header">
      <div className="container">
        <div className="app-header-inner">
          <div>
            <h1>ossf-scout</h1>
            <p>Find GitHub repos with weak OpenSSF Scorecard scores</p>
          </div>
          <nav className="tab-nav">
            <NavLink to="/" end>Scans</NavLink>
            <NavLink to="/trending">Trending</NavLink>
            <NavLink to="/audits">Audit</NavLink>
            <NavLink to="/schedules">Schedules</NavLink>
            <NavLink to="/portfolio">Portfolio</NavLink>
            <NavLink to="/remediation">Remediation</NavLink>
            <NavLink to="/orgscan">Org scan</NavLink>
          </nav>
        </div>
      </div>
    </div>
  )
}

function Layout() {
  return (
    <>
      <AppHeader />
      <Outlet />
    </>
  )
}

function BackToTop() {
  const [visible, setVisible] = useState(false)

  useEffect(() => {
    const onScroll = () => setVisible(window.scrollY > 300)
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [])

  if (!visible) return null
  return (
    <button className="back-to-top" onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })} title="Back to top">
      ↑
    </button>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<ScanList />} />
          <Route path="/trending" element={<TrendingPage />} />
          <Route path="/scans/:id" element={<ScanDetail />} />
          <Route path="/audits" element={<AuditPage />} />
          <Route path="/audits/:id" element={<AuditDetail />} />
          <Route path="/audits/:id/compare" element={<AuditComparePage />} />
          <Route path="/schedules" element={<SchedulePage />} />
          <Route path="/portfolio" element={<PortfolioPage />} />
          <Route path="/remediation" element={<RemediationPage />} />
          <Route path="/orgscan" element={<OrgScanPage />} />
        </Route>
      </Routes>
      <BackToTop />
    </BrowserRouter>
  )
}
