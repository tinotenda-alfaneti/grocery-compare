import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { api } from '../api/client'
import { todayISODate } from '../lib/money'

export default function GiftCardsPage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const { data: stores } = useQuery({ queryKey: ['stores'], queryFn: api.stores.list })
  const { data: current } = useQuery({ queryKey: ['gift-cards', 'current'], queryFn: api.giftCards.current })
  const { data: history } = useQuery({ queryKey: ['gift-cards', 'all'], queryFn: api.giftCards.listAll })

  const [drafts, setDrafts] = useState<Record<number, string>>({})

  const create = useMutation({
    mutationFn: (storeId: number) =>
      api.giftCards.create({
        storeId,
        discountPercent: Number(drafts[storeId] || 0),
        effectiveFrom: todayISODate()
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['gift-cards'] })
    }
  })

  if (!stores) return <div className="p-4 text-gray-500">Loading…</div>

  return (
    <div className="p-4">
      <button className="mb-3 flex items-center gap-1 text-sm text-gray-600" onClick={() => navigate('/settings')}>
        <ArrowLeft size={16} /> Settings
      </button>
      <h1 className="mb-1 text-xl font-semibold">Gift-card discounts</h1>
      <p className="mb-4 text-sm text-gray-500">
        Enter the current discount % you can get on each store's gift cards via your employee perks site. This is applied
        as a discount to that store's basket total — never fetched automatically.
      </p>

      <div className="space-y-3">
        {stores.map((store) => {
          const currentPct = current?.[store.id] ?? 0
          return (
            <div key={store.id} className="rounded-lg border border-gray-200 bg-white p-3">
              <div className="mb-2 flex items-center justify-between">
                <span className="font-medium">{store.name}</span>
                <span className="text-lg font-semibold text-green-700">{currentPct}% off</span>
              </div>
              <div className="flex gap-2">
                <input
                  type="number"
                  step="0.1"
                  className="w-24 rounded border border-gray-300 px-2 py-1 text-sm"
                  placeholder="New %"
                  value={drafts[store.id] ?? ''}
                  onChange={(e) => setDrafts((d) => ({ ...d, [store.id]: e.target.value }))}
                />
                <button
                  className="rounded bg-green-600 px-3 py-1 text-sm text-white"
                  onClick={() => create.mutate(store.id)}
                >
                  Update
                </button>
              </div>
              {history && history.filter((h) => h.storeId === store.id).length > 0 && (
                <details className="mt-2 text-xs text-gray-500">
                  <summary>History</summary>
                  <ul className="mt-1 space-y-0.5">
                    {history
                      .filter((h) => h.storeId === store.id)
                      .map((h) => (
                        <li key={h.id}>
                          {h.discountPercent}% from {h.effectiveFrom}
                          {h.effectiveTo ? ` to ${h.effectiveTo}` : ''}
                        </li>
                      ))}
                  </ul>
                </details>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
