import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../api/client'
import { formatPence, todayISODate } from '../lib/money'
import type { ProductMapping, Store } from '../api/types'

type Panel = null | 'create' | 'price' | 'promo' | 'member'

export default function StoreMappingCard({
  store,
  mapping,
  itemId
}: {
  store: Store
  mapping?: ProductMapping
  itemId: number
}) {
  const [panel, setPanel] = useState<Panel>(null)
  const queryClient = useQueryClient()
  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['mappings', itemId] })

  const createMapping = useMutation({
    mutationFn: (productName: string) => api.mappings.create(itemId, { storeId: store.id, productName }),
    onSuccess: () => {
      invalidate()
      setPanel(null)
    }
  })
  const removeMapping = useMutation({
    mutationFn: (mappingId: number) => api.mappings.remove(mappingId),
    onSuccess: invalidate
  })
  const addPrice = useMutation({
    mutationFn: (pricePence: number) => api.mappings.addPrice(mapping!.id, pricePence),
    onSuccess: () => {
      invalidate()
      setPanel(null)
    }
  })
  const addPromo = useMutation({
    mutationFn: (input: { promoPricePence: number; promoLabel?: string; effectiveFrom: string; effectiveTo: string }) =>
      api.mappings.addPromo(mapping!.id, input),
    onSuccess: () => {
      invalidate()
      setPanel(null)
    }
  })
  const addMemberPrice = useMutation({
    mutationFn: (input: { memberPricePence: number; effectiveFrom: string }) =>
      api.mappings.addMemberPrice(mapping!.id, input),
    onSuccess: () => {
      invalidate()
      setPanel(null)
    }
  })

  if (!mapping) {
    return (
      <div className="rounded-lg border border-dashed border-gray-300 p-3">
        <div className="mb-2 flex items-center justify-between">
          <span className="font-medium">{store.name}</span>
          <span className="text-xs text-gray-400">not mapped</span>
        </div>
        {panel === 'create' ? (
          <CreateForm onCancel={() => setPanel(null)} onSubmit={(name) => createMapping.mutate(name)} />
        ) : (
          <button className="text-sm text-green-700 underline" onClick={() => setPanel('create')}>
            + Map a product
          </button>
        )}
      </div>
    )
  }

  return (
    <div className="rounded-lg border border-gray-200 bg-white p-3">
      <div className="mb-1 flex items-center justify-between">
        <span className="font-medium">{store.name}</span>
        <button className="text-xs text-red-500" onClick={() => removeMapping.mutate(mapping.id)}>
          remove
        </button>
      </div>
      <div className="text-sm text-gray-700">{mapping.productName}</div>
      {mapping.packSize && <div className="text-xs text-gray-400">{mapping.packSize}</div>}
      <div className="mt-1 text-lg font-semibold">
        {mapping.currentPricePence != null ? formatPence(mapping.currentPricePence) : '—'}
      </div>

      <div className="mt-2 flex flex-wrap gap-2 text-xs">
        <button className="rounded bg-gray-100 px-2 py-1" onClick={() => setPanel(panel === 'price' ? null : 'price')}>
          Update price
        </button>
        <button className="rounded bg-gray-100 px-2 py-1" onClick={() => setPanel(panel === 'promo' ? null : 'promo')}>
          Add promo
        </button>
        {store.supportsMemberPricing && (
          <button className="rounded bg-gray-100 px-2 py-1" onClick={() => setPanel(panel === 'member' ? null : 'member')}>
            Add {store.memberPricingLabel ?? 'member'} price
          </button>
        )}
      </div>

      {panel === 'price' && (
        <PriceForm onCancel={() => setPanel(null)} onSubmit={(p) => addPrice.mutate(p)} label="Price (£)" />
      )}
      {panel === 'promo' && (
        <PromoForm onCancel={() => setPanel(null)} onSubmit={(input) => addPromo.mutate(input)} />
      )}
      {panel === 'member' && (
        <MemberForm onCancel={() => setPanel(null)} onSubmit={(input) => addMemberPrice.mutate(input)} />
      )}
    </div>
  )
}

function CreateForm({ onSubmit, onCancel }: { onSubmit: (name: string) => void; onCancel: () => void }) {
  const [name, setName] = useState('')
  return (
    <div className="flex gap-2">
      <input
        autoFocus
        className="flex-1 rounded border border-gray-300 px-2 py-1 text-sm"
        placeholder="Product name at this store"
        value={name}
        onChange={(e) => setName(e.target.value)}
      />
      <button className="rounded bg-green-600 px-2 py-1 text-sm text-white" onClick={() => name.trim() && onSubmit(name.trim())}>
        Save
      </button>
      <button className="text-sm text-gray-400" onClick={onCancel}>
        Cancel
      </button>
    </div>
  )
}

function PriceForm({
  onSubmit,
  onCancel,
  label
}: {
  onSubmit: (pricePence: number) => void
  onCancel: () => void
  label: string
}) {
  const [pounds, setPounds] = useState('')
  return (
    <div className="mt-2 flex gap-2">
      <input
        autoFocus
        type="number"
        step="0.01"
        className="w-24 rounded border border-gray-300 px-2 py-1 text-sm"
        placeholder={label}
        value={pounds}
        onChange={(e) => setPounds(e.target.value)}
      />
      <button
        className="rounded bg-green-600 px-2 py-1 text-sm text-white"
        onClick={() => {
          const p = Math.round(Number(pounds) * 100)
          if (p > 0) onSubmit(p)
        }}
      >
        Save
      </button>
      <button className="text-sm text-gray-400" onClick={onCancel}>
        Cancel
      </button>
    </div>
  )
}

function PromoForm({
  onSubmit,
  onCancel
}: {
  onSubmit: (input: { promoPricePence: number; promoLabel?: string; effectiveFrom: string; effectiveTo: string }) => void
  onCancel: () => void
}) {
  const [pounds, setPounds] = useState('')
  const [label, setLabel] = useState('')
  const [from, setFrom] = useState(todayISODate())
  const [to, setTo] = useState(todayISODate())
  return (
    <div className="mt-2 space-y-2 text-sm">
      <input
        type="number"
        step="0.01"
        className="w-full rounded border border-gray-300 px-2 py-1"
        placeholder="Promo price (£)"
        value={pounds}
        onChange={(e) => setPounds(e.target.value)}
      />
      <input
        className="w-full rounded border border-gray-300 px-2 py-1"
        placeholder="Label, e.g. 3 for £5 (optional)"
        value={label}
        onChange={(e) => setLabel(e.target.value)}
      />
      <div className="flex gap-2">
        <input type="date" className="flex-1 rounded border border-gray-300 px-2 py-1" value={from} onChange={(e) => setFrom(e.target.value)} />
        <input type="date" className="flex-1 rounded border border-gray-300 px-2 py-1" value={to} onChange={(e) => setTo(e.target.value)} />
      </div>
      <div className="flex gap-2">
        <button
          className="rounded bg-green-600 px-2 py-1 text-white"
          onClick={() => {
            const p = Math.round(Number(pounds) * 100)
            if (p > 0) onSubmit({ promoPricePence: p, promoLabel: label || undefined, effectiveFrom: from, effectiveTo: to })
          }}
        >
          Save
        </button>
        <button className="text-gray-400" onClick={onCancel}>
          Cancel
        </button>
      </div>
    </div>
  )
}

function MemberForm({
  onSubmit,
  onCancel
}: {
  onSubmit: (input: { memberPricePence: number; effectiveFrom: string }) => void
  onCancel: () => void
}) {
  const [pounds, setPounds] = useState('')
  const [from, setFrom] = useState(todayISODate())
  return (
    <div className="mt-2 space-y-2 text-sm">
      <input
        type="number"
        step="0.01"
        className="w-full rounded border border-gray-300 px-2 py-1"
        placeholder="Member/Clubcard price (£)"
        value={pounds}
        onChange={(e) => setPounds(e.target.value)}
      />
      <input type="date" className="w-full rounded border border-gray-300 px-2 py-1" value={from} onChange={(e) => setFrom(e.target.value)} />
      <div className="flex gap-2">
        <button
          className="rounded bg-green-600 px-2 py-1 text-white"
          onClick={() => {
            const p = Math.round(Number(pounds) * 100)
            if (p > 0) onSubmit({ memberPricePence: p, effectiveFrom: from })
          }}
        >
          Save
        </button>
        <button className="text-gray-400" onClick={onCancel}>
          Cancel
        </button>
      </div>
    </div>
  )
}
