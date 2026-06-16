import { BrowserRouter, Routes, Route, NavLink, Outlet } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { useLang, type Lang } from './i18n'
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
  const { t, lang, setLang } = useLang()
  return (
    <div className="app-header">
      <div className="container">
        <div className="app-header-inner">
          <div>
            <h1>{t('appTitle')}</h1>
            <p>{t('appTagline')}</p>
          </div>
          <nav className="tab-nav">
            <NavLink to="/" end>{t('navScans')}</NavLink>
            <NavLink to="/trending">{t('navTrending')}</NavLink>
            <NavLink to="/audits">{t('navAudit')}</NavLink>
            <NavLink to="/schedules">{t('navSchedules')}</NavLink>
            <NavLink to="/portfolio">{t('navPortfolio')}</NavLink>
            <NavLink to="/remediation">{t('navRemediation')}</NavLink>
            <NavLink to="/orgscan">{t('navOrgScan')}</NavLink>
          </nav>
          <select
            className="input"
            aria-label={t('langSwitcherLabel')}
            style={{ width: 'auto' }}
            value={lang}
            onChange={(e) => setLang(e.target.value as Lang)}
          >
            <option value="en">EN</option>
            <option value="hu">HU</option>
          </select>
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
