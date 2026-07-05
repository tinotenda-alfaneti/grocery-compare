import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { api } from '../api/client'
import type { CanonicalItem } from '../api/types'

export default function ItemTypeahead({ onSelect }: { onSelect: (item: CanonicalItem) => void }) {
  const [query, setQuery] = useState('')
  const { data: results } = useQuery({
    queryKey: ['items', 'search', query],
    queryFn: () => api.items.search(query),
    enabled: query.trim().length > 0
  })

  const exactMatch = results?.some((r) => r.name.toLowerCase() === query.trim().toLowerCase())

  return (
    <div>
      <input
        className="w-full rounded-lg border border-gray-300 px-3 py-2"
        placeholder="Search or add an item…"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
      />
      {query.trim() && (
        <ul className="mt-1 divide-y divide-gray-100 rounded-lg border border-gray-200 bg-white">
          {results?.map((it) => (
            <li key={it.id}>
              <button
                className="w-full px-3 py-2 text-left hover:bg-gray-50"
                onClick={() => {
                  onSelect(it)
                  setQuery('')
                }}
              >
                {it.name}
              </button>
            </li>
          ))}
          {!exactMatch && (
            <li>
              <button
                className="w-full px-3 py-2 text-left text-green-700 hover:bg-gray-50"
                onClick={async () => {
                  const created = await api.items.create(query.trim())
                  onSelect(created)
                  setQuery('')
                }}
              >
                + Add "{query.trim()}" as a new item
              </button>
            </li>
          )}
        </ul>
      )}
    </div>
  )
}
