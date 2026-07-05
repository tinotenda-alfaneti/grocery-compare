import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { api } from '../api/client'
import { formatPence } from '../lib/money'
import ItemTypeahead from '../components/ItemTypeahead'
import type { CanonicalItem } from '../api/types'

export default function LookupPage() {
  const [selected, setSelected] = useState<CanonicalItem | null>(null)

  const { data: results, isFetching } = useQuery({
    queryKey: ['items', selected?.id, 'compare'],
    queryFn: () => api.items.compare(selected!.id),
    enabled: !!selected
  })

  const cheapest = results?.filter((r) => r.available).sort((a, b) => a.pricePence - b.pricePence)[0]

  return (
    <div className="p-4">
      <h1 className="mb-4 text-xl font-semibold">Single-item lookup</h1>
      <ItemTypeahead onSelect={setSelected} />

      {selected && (
        <div className="mt-4">
          <h2 className="mb-2 font-medium">{selected.name}</h2>
          {isFetching && <p className="text-gray-500">Loading…</p>}
          <div className="space-y-2">
            {results?.map((r) => (
              <div
                key={r.storeId}
                className={`flex items-center justify-between rounded-lg border p-3 ${
                  cheapest && r.storeId === cheapest.storeId ? 'border-green-500 bg-green-50' : 'border-gray-200 bg-white'
                }`}
              >
                <span className="font-medium">{r.storeName}</span>
                <span>{r.available ? formatPence(r.pricePence) : <span className="text-gray-400">not mapped</span>}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
