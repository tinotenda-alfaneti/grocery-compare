import { useQuery } from '@tanstack/react-query'
import { useNavigate, useParams } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { api } from '../api/client'
import StoreMappingCard from '../components/StoreMappingCard'

export default function ItemDetailPage() {
  const { id } = useParams()
  const itemId = Number(id)
  const navigate = useNavigate()

  const { data: item } = useQuery({ queryKey: ['items', itemId], queryFn: () => api.items.get(itemId) })
  const { data: stores } = useQuery({ queryKey: ['stores'], queryFn: api.stores.list })
  const { data: mappings } = useQuery({ queryKey: ['mappings', itemId], queryFn: () => api.mappings.listForItem(itemId) })

  if (!item || !stores) return <div className="p-4 text-gray-500">Loading…</div>

  const mappingByStore = Object.fromEntries((mappings ?? []).map((m) => [m.storeId, m]))

  return (
    <div className="p-4">
      <button className="mb-3 flex items-center gap-1 text-sm text-gray-600" onClick={() => navigate(-1)}>
        <ArrowLeft size={16} /> Back
      </button>
      <h1 className="mb-4 text-xl font-semibold">{item.name}</h1>

      <div className="space-y-3">
        {stores.map((store) => (
          <StoreMappingCard key={store.id} store={store} mapping={mappingByStore[store.id]} itemId={itemId} />
        ))}
      </div>
    </div>
  )
}
