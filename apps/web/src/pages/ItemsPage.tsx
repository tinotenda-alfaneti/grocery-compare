import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { api } from '../api/client'

export default function ItemsPage() {
  const [query, setQuery] = useState('')
  const navigate = useNavigate()
  const { data: items, isLoading } = useQuery({ queryKey: ['items', 'all', query], queryFn: () => api.items.search(query) })

  return (
    <div className="p-4">
      <h1 className="mb-4 text-xl font-semibold">Items</h1>
      <input
        className="mb-4 w-full rounded-lg border border-gray-300 px-3 py-2"
        placeholder="Search your items…"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
      />

      {isLoading && <p className="text-gray-500">Loading…</p>}

      <ul className="space-y-2">
        {items?.map((it) => (
          <li key={it.id}>
            <button
              className="w-full rounded-lg border border-gray-200 bg-white p-3 text-left shadow-sm"
              onClick={() => navigate(`/items/${it.id}`)}
            >
              <div className="font-medium">{it.name}</div>
              {it.category && <div className="text-xs text-gray-500">{it.category}</div>}
            </button>
          </li>
        ))}
        {items && items.length === 0 && (
          <p className="text-gray-500">No items yet. Add items from a shopping list or here.</p>
        )}
      </ul>
    </div>
  )
}
