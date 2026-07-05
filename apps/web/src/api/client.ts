import type {
  CanonicalItem,
  GiftCardDiscount,
  ItemComparison,
  PriceObservation,
  ProductMapping,
  Recommendation,
  Settings,
  ShoppingList,
  ShoppingListItem,
  Store
} from './types'

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`/api${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...init
  })
  const text = await res.text()
  const body = text ? JSON.parse(text) : undefined

  if (!res.ok) {
    throw new Error(body?.error || `Request failed: ${res.status}`)
  }
  return body as T
}

const post = <T>(path: string, body?: unknown) =>
  request<T>(path, { method: 'POST', body: body !== undefined ? JSON.stringify(body) : undefined })
const patch = <T>(path: string, body?: unknown) =>
  request<T>(path, { method: 'PATCH', body: body !== undefined ? JSON.stringify(body) : undefined })
const del = <T>(path: string) => request<T>(path, { method: 'DELETE' })

export const api = {
  stores: {
    list: () => request<Store[]>('/stores'),
    setIncluded: (id: number, includedInComparisons: boolean) =>
      patch<Store>(`/stores/${id}`, { includedInComparisons })
  },

  items: {
    search: (query: string) => request<CanonicalItem[]>(`/items?query=${encodeURIComponent(query)}`),
    get: (id: number) => request<CanonicalItem>(`/items/${id}`),
    create: (name: string, category?: string) => post<CanonicalItem>('/items', { name, category }),
    archive: (id: number) => del<void>(`/items/${id}`),
    compare: (id: number) => request<ItemComparison[]>(`/items/${id}/compare`)
  },

  mappings: {
    listForItem: (itemId: number) => request<ProductMapping[]>(`/items/${itemId}/mappings`),
    create: (itemId: number, input: { storeId: number; productName: string; productUrl?: string; packSize?: string }) =>
      post<ProductMapping>(`/items/${itemId}/mappings`, input),
    remove: (mappingId: number) => del<void>(`/mappings/${mappingId}`),
    addPrice: (mappingId: number, pricePence: number) => post<void>(`/mappings/${mappingId}/price`, { pricePence }),
    priceHistory: (mappingId: number) => request<PriceObservation[]>(`/mappings/${mappingId}/price-history`),
    addPromo: (
      mappingId: number,
      input: { promoPricePence: number; promoLabel?: string; effectiveFrom: string; effectiveTo: string }
    ) => post<{ id: number }>(`/mappings/${mappingId}/promo`, input),
    deletePromo: (promoId: number) => del<void>(`/promos/${promoId}`),
    addMemberPrice: (mappingId: number, input: { memberPricePence: number; effectiveFrom: string; effectiveTo?: string }) =>
      post<{ id: number }>(`/mappings/${mappingId}/member-price`, input),
    deleteMemberPrice: (memberPriceId: number) => del<void>(`/member-prices/${memberPriceId}`)
  },

  lists: {
    list: () => request<ShoppingList[]>('/lists'),
    get: (id: number) => request<{ list: ShoppingList; items: ShoppingListItem[] }>(`/lists/${id}`),
    create: (name: string) => post<ShoppingList>('/lists', { name }),
    remove: (id: number) => del<void>(`/lists/${id}`),
    addItem: (listId: number, input: { canonicalItemId: number; quantity: number; notes?: string }) =>
      post<{ id: number }>(`/lists/${listId}/items`, input),
    updateItem: (listId: number, itemId: number, input: { quantity?: number; notes?: string }) =>
      patch<void>(`/lists/${listId}/items/${itemId}`, input),
    removeItem: (listId: number, itemId: number) => del<void>(`/lists/${listId}/items/${itemId}`),
    compare: (id: number) => request<Recommendation>(`/lists/${id}/compare`)
  },

  giftCards: {
    listAll: () => request<GiftCardDiscount[]>('/gift-card-discounts'),
    current: () => request<Record<number, number>>('/gift-card-discounts/current'),
    create: (input: {
      storeId: number
      discountPercent: number
      effectiveFrom: string
      effectiveTo?: string
      notes?: string
    }) => post<{ id: number }>('/gift-card-discounts', input),
    remove: (id: number) => del<void>(`/gift-card-discounts/${id}`)
  },

  settings: {
    get: () => request<Settings>('/settings'),
    update: (input: { secondStopMinSavingPence?: number; secondStopMinSavingPercent?: number }) =>
      patch<Settings>('/settings', input)
  }
}
