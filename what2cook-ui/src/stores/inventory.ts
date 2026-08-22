import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import * as inventoryApi from '../api/inventory'
import type { Inventory, InventoryItem } from '../api/inventory'
import { ApiError } from '../api/client'

export const useInventoryStore = defineStore('inventory', () => {
  const inventory = ref<Inventory | null>(null)
  const loading = ref(false)
  const saving = ref(false)
  const error = ref<unknown>(null)
  const items = computed(() => inventory.value?.items ?? [])
  const inventoryId = computed(() => inventory.value?.id ?? null)

  function setError(err: unknown, fallback: string) {
    error.value = err instanceof ApiError || err instanceof Error ? err : new Error(fallback)
  }

  async function loadDefault() {
    loading.value = true; error.value = null
    try {
      const list = await inventoryApi.listInventories()
      const defaultInv = list.find((inv) => inv.is_default) ?? list[0] ?? null
      if (!defaultInv) throw new Error('No inventory available')
      const full = await inventoryApi.getInventory(defaultInv.id)
      inventory.value = full
      return full
    } catch (err) {
      console.warn('Load inventory failed', err); inventory.value = null; setError(err, 'Could not load inventory'); return null
    } finally { loading.value = false }
  }

  async function addItem(name: string, quantity?: string | null, category?: string | null) {
    const id = inventoryId.value
    if (!id) { error.value = new Error('Inventory not loaded'); return null }
    const trimmed = name.trim()
    if (!trimmed) { error.value = new Error('Ingredient name is required'); return null }
    saving.value = true; error.value = null
    try {
      const item = await inventoryApi.createItem(id, trimmed, quantity, category)
      if (inventory.value) inventory.value = { ...inventory.value, items: [...(inventory.value.items ?? []), item] }
      return item
    } catch (err) { console.warn('Add inventory item failed', err); setError(err, 'Could not add ingredient'); return null }
    finally { saving.value = false }
  }

  async function updateItem(itemId: string, name: string, quantity?: string | null, category?: string | null) {
    const id = inventoryId.value
    if (!id) { error.value = new Error('Inventory not loaded'); return null }
    const trimmed = name.trim()
    if (!trimmed) { error.value = new Error('Ingredient name is required'); return null }
    saving.value = true; error.value = null
    try {
      const updated = await inventoryApi.updateItem(id, itemId, trimmed, quantity, category)
      if (inventory.value) inventory.value = { ...inventory.value, items: (inventory.value.items ?? []).map((item) => item.id === itemId ? updated : item) }
      return updated
    } catch (err) { console.warn('Update inventory item failed', err); setError(err, 'Could not update ingredient'); return null }
    finally { saving.value = false }
  }

  async function removeItem(itemId: string) {
    const id = inventoryId.value
    if (!id) { error.value = new Error('Inventory not loaded'); return false }
    saving.value = true; error.value = null
    try {
      await inventoryApi.deleteItem(id, itemId)
      if (inventory.value) inventory.value = { ...inventory.value, items: (inventory.value.items ?? []).filter((item) => item.id !== itemId) }
      return true
    } catch (err) { console.warn('Delete inventory item failed', err); setError(err, 'Could not delete ingredient'); return false }
    finally { saving.value = false }
  }

  function clearError() { error.value = null }
  return { inventory, items, inventoryId, loading, saving, error, loadDefault, addItem, updateItem, removeItem, clearError }
})

export type { Inventory, InventoryItem }
