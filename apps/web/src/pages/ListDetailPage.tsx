import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate, useParams } from 'react-router-dom'
import { ArrowLeft, Trash2 } from 'lucide-react'
import { api } from '../api/client'
import ItemTypeahead from '../components/ItemTypeahead'
import CompareResults from '../components/CompareResults'

export default function ListDetailPage() {
  const { id } = useParams()
  const listId = Number(id)
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [showCompare, setShowCompare] = useState(false)

  const { data, isLoading } = useQuery({ queryKey: ['lists', listId], queryFn: () => api.lists.get(listId) })
  const { data: rec, refetch: runCompare, isFetching: comparing } = useQuery({
    queryKey: ['lists', listId, 'compare'],
    queryFn: () => api.lists.compare(listId),
    enabled: false
  })

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['lists', listId] })

  const addItem = useMutation({
    mutationFn: (canonicalItemId: number) => api.lists.addItem(listId, { canonicalItemId, quantity: 1 }),
    onSuccess: invalidate
  })
  const updateQuantity = useMutation({
    mutationFn: ({ itemId, quantity }: { itemId: number; quantity: number }) =>
      api.lists.updateItem(listId, itemId, { quantity }),
    onSuccess: invalidate
  })
  const removeItem = useMutation({
    mutationFn: (itemId: number) => api.lists.removeItem(listId, itemId),
    onSuccess: invalidate
  })

  if (isLoading || !data) return <div className="p-4 text-gray-500">Loading…</div>

  const itemNames = Object.fromEntries(data.items.map((it) => [it.canonicalItemId, it.canonicalItemName]))

  return (
    <div className="p-4">
      <button className="mb-3 flex items-center gap-1 text-sm text-gray-600" onClick={() => navigate('/')}>
        <ArrowLeft size={16} /> Lists
      </button>
      <h1 className="mb-4 text-xl font-semibold">{data.list.name}</h1>

      <div className="mb-4">
        <ItemTypeahead onSelect={(item) => addItem.mutate(item.id)} />
      </div>

      <ul className="mb-4 space-y-2">
        {data.items.map((it) => (
          <li key={it.id} className="flex items-center justify-between rounded-lg border border-gray-200 bg-white p-3">
            <div>
              <div className="font-medium">{it.canonicalItemName}</div>
              <button
                className="text-xs text-gray-400 underline"
                onClick={() => navigate(`/items/${it.canonicalItemId}`)}
              >
                edit store mappings
              </button>
            </div>
            <div className="flex items-center gap-2">
              <input
                type="number"
                min={1}
                className="w-14 rounded border border-gray-300 px-2 py-1 text-center"
                value={it.quantity}
                onChange={(e) => updateQuantity.mutate({ itemId: it.id, quantity: Number(e.target.value) || 1 })}
              />
              <button className="text-red-500" onClick={() => removeItem.mutate(it.id)}>
                <Trash2 size={18} />
              </button>
            </div>
          </li>
        ))}
        {data.items.length === 0 && <p className="text-gray-500">Add items above to get started.</p>}
      </ul>

      {data.items.length > 0 && (
        <button
          className="mb-4 w-full rounded-lg bg-green-600 py-3 font-medium text-white disabled:opacity-50"
          disabled={comparing}
          onClick={async () => {
            setShowCompare(true)
            await runCompare()
          }}
        >
          {comparing ? 'Comparing…' : 'Compare stores'}
        </button>
      )}

      {showCompare && rec && <CompareResults rec={rec} itemNames={itemNames} />}
    </div>
  )
}
