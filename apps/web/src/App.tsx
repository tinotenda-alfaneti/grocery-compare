import { Route, Routes } from 'react-router-dom'
import BottomNav from './components/BottomNav'
import ListsPage from './pages/ListsPage'
import ListDetailPage from './pages/ListDetailPage'
import ItemsPage from './pages/ItemsPage'
import ItemDetailPage from './pages/ItemDetailPage'
import LookupPage from './pages/LookupPage'
import GiftCardsPage from './pages/GiftCardsPage'
import SettingsPage from './pages/SettingsPage'

export default function App() {
  return (
    <div className="flex h-full flex-col bg-gray-50 text-gray-900">
      <main className="flex-1 overflow-y-auto pb-16">
        <Routes>
          <Route path="/" element={<ListsPage />} />
          <Route path="/lists/:id" element={<ListDetailPage />} />
          <Route path="/items" element={<ItemsPage />} />
          <Route path="/items/:id" element={<ItemDetailPage />} />
          <Route path="/lookup" element={<LookupPage />} />
          <Route path="/settings" element={<SettingsPage />} />
          <Route path="/settings/gift-cards" element={<GiftCardsPage />} />
        </Routes>
      </main>
      <BottomNav />
    </div>
  )
}
