export interface Store {
  id: number
  name: string
  slug: string
  supportsMemberPricing: boolean
  memberPricingLabel?: string
  includedInComparisons: boolean
}

export interface CanonicalItem {
  id: number
  name: string
  category?: string
  notes?: string
  archived: boolean
}

export interface ProductMapping {
  id: number
  canonicalItemId: number
  storeId: number
  productName: string
  productUrl?: string
  packSize?: string
  active: boolean
  isManual: boolean
  currentPricePence?: number
}

export interface PriceObservation {
  id: number
  mappingId: number
  pricePence: number
  observedAt: string
  source: 'manual' | 'scraped'
}

export interface ShoppingList {
  id: number
  name: string
  archived: boolean
  createdAt: string
  updatedAt: string
}

export interface ShoppingListItem {
  id: number
  shoppingListId: number
  canonicalItemId: number
  canonicalItemName: string
  quantity: number
  notes?: string
  sortOrder: number
}

export interface GiftCardDiscount {
  id: number
  storeId: number
  discountPercent: number
  effectiveFrom: string
  effectiveTo?: string
  notes?: string
}

export interface Settings {
  secondStopMinSavingPence: number
  secondStopMinSavingPercent?: number
  pinEnabled: boolean
}

export interface CompareStore {
  id: number
  name: string
  slug: string
  supportsMemberPricing: boolean
  giftCardDiscountPercent: number
}

export interface StoreResult {
  store: CompareStore
  subtotalPence: number
  discountPercent: number
  totalPence: number
  coveredCount: number
  totalCount: number
  coverageRatio: number
}

export interface SplitAssignment {
  canonicalItemId: number
  storeId: number
  pricePence: number
}

export interface SplitResult {
  storeA: CompareStore
  storeB: CompareStore
  totalAPence: number
  totalBPence: number
  totalPence: number
  coverageRatio: number
  assignments: SplitAssignment[]
}

export interface Recommendation {
  perStore: StoreResult[]
  bestSingle: StoreResult
  bestSplit?: SplitResult
  recommendation: 'single' | 'split'
  savingPence: number
  savingPercent: number
  unmappedAnywhere: number[] | null
}

export interface ItemComparison {
  storeId: number
  storeName: string
  available: boolean
  pricePence: number
}
