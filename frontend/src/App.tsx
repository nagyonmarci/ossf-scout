import { BrowserRouter, Routes, Route } from 'react-router-dom'
import ScanList from './pages/ScanList'
import ScanDetail from './pages/ScanDetail'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<ScanList />} />
        <Route path="/scans/:id" element={<ScanDetail />} />
      </Routes>
    </BrowserRouter>
  )
}
