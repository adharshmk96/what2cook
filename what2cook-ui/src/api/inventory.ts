import { apiRequest } from './client'

export type InventoryItem = {
  id: string
  inventory_id: string
  name: string
  quantity: string | null
  created_at: string
  updated_at: string
}

export type Inventory = {
  id: string
  user_id: string
  name: string
  is_default: boolean
  items?: InventoryItem[]
  created_at: string
  updated_at: string
}

type ListInventoriesResponse = {
  inventories: Inventory[]
}

export async function listInventories(): Promise<Inventory[]> {
  const data = await apiRequest<ListInventoriesResponse>('/inventories')
  if (!data || !Array.isArray(data.inventories)) {
    throw new Error('Unexpected inventories response from server')
  }
  return data.inventories
}

export async function getInventory(id: string): Promise<Inventory> {
  const data = await apiRequest<Inventory>(`/inventories/${id}`)
  if (!data || typeof data.id !== 'string') {
    throw new Error('Unexpected inventory response from server')
  }
  return {
    ...data,
    items: Array.isArray(data.items) ? data.items : [],
  }
}

export async function createInventory(name: string): Promise<Inventory> {
  return apiRequest<Inventory>('/inventories', {
    method: 'POST',
    body: JSON.stringify({ name }),
  })
}

export async function updateInventory(
  id: string,
  name: string,
): Promise<Inventory> {
  return apiRequest<Inventory>(`/inventories/${id}`, {
    method: 'PATCH',
    body: JSON.stringify({ name }),
  })
}

export async function deleteInventory(id: string): Promise<void> {
  await apiRequest<null>(`/inventories/${id}`, { method: 'DELETE' })
}

export async function createItem(
  inventoryId: string,
  name: string,
  quantity?: string | null,
): Promise<InventoryItem> {
  const body: { name: string; quantity?: string } = { name }
  if (quantity != null && quantity.trim()) {
    body.quantity = quantity.trim()
  }
  return apiRequest<InventoryItem>(`/inventories/${inventoryId}/items`, {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export async function updateItem(
  inventoryId: string,
  itemId: string,
  name: string,
  quantity?: string | null,
): Promise<InventoryItem> {
  return apiRequest<InventoryItem>(
    `/inventories/${inventoryId}/items/${itemId}`,
    {
      method: 'PATCH',
      body: JSON.stringify({
        name,
        quantity: quantity?.trim() ? quantity.trim() : '',
      }),
    },
  )
}

export async function deleteItem(
  inventoryId: string,
  itemId: string,
): Promise<void> {
  await apiRequest<null>(`/inventories/${inventoryId}/items/${itemId}`, {
    method: 'DELETE',
  })
}
