export type ParsedInventoryEntry = {
  name: string
  quantity: string | null
}

/**
 * Parse comma-separated inventory entries.
 * Format: "chicken: 100g, tomato, ginger: 100g"
 * First ":" in a token separates name from optional quantity.
 */
export function parseInventoryCsv(raw: string): ParsedInventoryEntry[] {
  const entries: ParsedInventoryEntry[] = []

  for (const part of raw.split(',')) {
    const token = part.trim()
    if (!token) {
      continue
    }

    const colonIndex = token.indexOf(':')
    if (colonIndex === -1) {
      entries.push({ name: token, quantity: null })
      continue
    }

    const name = token.slice(0, colonIndex).trim()
    const quantity = token.slice(colonIndex + 1).trim()
    if (!name) {
      continue
    }

    entries.push({
      name,
      quantity: quantity || null,
    })
  }

  return entries
}
