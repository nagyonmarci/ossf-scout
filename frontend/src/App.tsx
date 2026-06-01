import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { useEffect, useState } from 'react'
import ScanList from './pages/ScanList'
import ScanDetail from './pages/ScanDetail'

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
        <Route path="/" element={<ScanList />} />
        <Route path="/scans/:id" element={<ScanDetail />} />
      </Routes>
      <BackToTop />
    </BrowserRouter>
  )
}
