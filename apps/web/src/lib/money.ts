export function formatPence(pence: number): string {
  return `£${(pence / 100).toFixed(2)}`
}

export function todayISODate(): string {
  return new Date().toISOString().slice(0, 10)
}
