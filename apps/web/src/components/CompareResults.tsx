import { formatPence } from '../lib/money'
import type { Recommendation } from '../api/types'

export default function CompareResults({
  rec,
  itemNames = {}
}: {
  rec: Recommendation
  itemNames?: Record<number, string>
}) {
  const headline =
    rec.recommendation === 'split' && rec.bestSplit
      ? `Split: ${rec.bestSplit.storeA.name} ${formatPence(rec.bestSplit.totalAPence)} + ${
          rec.bestSplit.storeB.name
        } ${formatPence(rec.bestSplit.totalBPence)} — save ${formatPence(rec.savingPence)} (${rec.savingPercent.toFixed(
          0
        )}%)`
      : `Shop at ${rec.bestSingle.store.name} — ${formatPence(rec.bestSingle.totalPence)}`

  const subtext =
    rec.recommendation === 'single' && rec.bestSplit && rec.savingPence > 0
      ? `Splitting across two stores only saves ${formatPence(rec.savingPence)} — not worth the extra trip.`
      : rec.recommendation === 'single'
      ? 'Cheapest option is one store.'
      : undefined

  return (
    <div className="space-y-4">
      <div className="rounded-xl bg-green-600 p-4 text-white shadow">
        <div className="text-lg font-semibold">{headline}</div>
        {subtext && <div className="mt-1 text-sm text-green-100">{subtext}</div>}
      </div>

      {rec.unmappedAnywhere && rec.unmappedAnywhere.length > 0 && (
        <div className="rounded-lg border border-amber-300 bg-amber-50 p-3 text-sm text-amber-800">
          {rec.unmappedAnywhere.length} item(s) on this list aren't mapped to any store yet, so they're excluded from
          the totals below.
        </div>
      )}

      <div className="space-y-2">
        {rec.perStore
          .slice()
          .sort((a, b) => a.totalPence - b.totalPence)
          .map((r) => (
            <div
              key={r.store.id}
              className={`rounded-lg border p-3 ${
                r.store.id === rec.bestSingle.store.id ? 'border-green-500 bg-green-50' : 'border-gray-200 bg-white'
              }`}
            >
              <div className="flex items-center justify-between">
                <span className="font-medium">{r.store.name}</span>
                <span className="font-semibold">{formatPence(r.totalPence)}</span>
              </div>
              <div className="mt-0.5 text-xs text-gray-500">
                {r.coveredCount}/{r.totalCount} items covered
                {r.discountPercent > 0 && ` · ${r.discountPercent}% gift-card discount applied`}
                {r.coverageRatio < 1 && ' · partial basket'}
              </div>
            </div>
          ))}
      </div>

      {rec.bestSplit && rec.recommendation === 'split' && (
        <div className="rounded-lg border border-gray-200 bg-white p-3">
          <div className="mb-2 text-sm font-medium">Split assignment</div>
          <ul className="space-y-1 text-sm text-gray-700">
            {rec.bestSplit.assignments.map((a, i) => (
              <li key={i} className="flex justify-between">
                <span>
                  {itemNames[a.canonicalItemId] ?? `Item #${a.canonicalItemId}`} →{' '}
                  {a.storeId === rec.bestSplit!.storeA.id ? rec.bestSplit!.storeA.name : rec.bestSplit!.storeB.name}
                </span>
                <span>{formatPence(a.pricePence)}</span>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  )
}
