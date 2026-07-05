import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { Plus } from 'lucide-react'
import { api } from '../api/client'

export default function ListsPage() {
  const [name, setName] = useState('')
  const queryClient = useQueryClient()
  const navigate = useNavigate()

  const { data: lists, isLoading } = useQuery({ queryKey: ['lists'], queryFn: api.lists.list })

  const createList = useMutation({
    mutationFn: (n: string) => api.lists.create(n),
    onSuccess: (created) => {
      queryClient.invalidateQueries({ queryKey: ['lists'] })
      setName('')
      navigate(`/lists/${created.id}`)
    }
  })

  return (
    <div className="p-4">
      <h1 className="mb-4 text-xl font-semibold">Shopping lists</h1>

      <form
        className="mb-4 flex gap-2"
        onSubmit={(e) => {
          e.preventDefault()
          if (name.trim()) createList.mutate(name.trim())
        }}
      >
        <input
          className="flex-1 rounded-lg border border-gray-300 px-3 py-2"
          placeholder="e.g. Weekly shop"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
        <button
          type="submit"
          className="flex items-center gap-1 rounded-lg bg-green-600 px-3 py-2 text-white disabled:opacity-50"
          disabled={createList.isPending}
        >
          <Plus size={18} /> New
        </button>
      </form>

      {isLoading && <p className="text-gray-500">Loading…</p>}

      <ul className="space-y-2">
        {lists?.map((list) => (
          <li key={list.id}>
            <button
              className="w-full rounded-lg border border-gray-200 bg-white p-3 text-left shadow-sm"
              onClick={() => navigate(`/lists/${list.id}`)}
            >
              <div className="font-medium">{list.name}</div>
              <div className="text-xs text-gray-500">Updated {new Date(list.updatedAt).toLocaleDateString()}</div>
            </button>
          </li>
        ))}
        {lists && lists.length === 0 && <p className="text-gray-500">No lists yet — create one above.</p>}
      </ul>
    </div>
  )
}
