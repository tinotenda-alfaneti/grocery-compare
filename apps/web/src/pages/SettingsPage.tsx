import { useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { ChevronRight } from 'lucide-react'
import { api } from '../api/client'

export default function SettingsPage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const { data: settings } = useQuery({ queryKey: ['settings'], queryFn: api.settings.get })
  const { data: stores } = useQuery({ queryKey: ['stores'], queryFn: api.stores.list })

  const [thresholdPounds, setThresholdPounds] = useState('')

  useEffect(() => {
    if (settings) setThresholdPounds((settings.secondStopMinSavingPence / 100).toFixed(2))
  }, [settings])

  const updateSettings = useMutation({
    mutationFn: (pricePence: number) => api.settings.update({ secondStopMinSavingPence: pricePence }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['settings'] })
  })
  const toggleStore = useMutation({
    mutationFn: ({ id, included }: { id: number; included: boolean }) => api.stores.setIncluded(id, included),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['stores'] })
  })

  return (
    <div className="p-4">
      <h1 className="mb-4 text-xl font-semibold">Settings</h1>

      <section className="mb-6">
        <h2 className="mb-2 text-sm font-medium text-gray-500">Second-stop threshold</h2>
        <p className="mb-2 text-xs text-gray-500">
          Only recommend splitting your shop across two stores if it saves at least this much.
        </p>
        <div className="flex gap-2">
          <div className="flex items-center rounded border border-gray-300 px-2">
            <span className="text-gray-400">£</span>
            <input
              type="number"
              step="0.5"
              className="w-20 px-1 py-2 outline-none"
              value={thresholdPounds}
              onChange={(e) => setThresholdPounds(e.target.value)}
            />
          </div>
          <button
            className="rounded bg-green-600 px-3 py-2 text-sm text-white"
            onClick={() => updateSettings.mutate(Math.round(Number(thresholdPounds) * 100))}
          >
            Save
          </button>
        </div>
      </section>

      <section className="mb-6">
        <h2 className="mb-2 text-sm font-medium text-gray-500">Stores to compare</h2>
        <div className="space-y-2">
          {stores?.map((store) => (
            <label key={store.id} className="flex items-center justify-between rounded-lg border border-gray-200 bg-white p-3">
              <span>{store.name}</span>
              <input
                type="checkbox"
                checked={store.includedInComparisons}
                onChange={(e) => toggleStore.mutate({ id: store.id, included: e.target.checked })}
              />
            </label>
          ))}
        </div>
      </section>

      <section>
        <button
          className="flex w-full items-center justify-between rounded-lg border border-gray-200 bg-white p-3"
          onClick={() => navigate('/settings/gift-cards')}
        >
          <span>Gift-card discounts</span>
          <ChevronRight size={18} />
        </button>
      </section>
    </div>
  )
}
