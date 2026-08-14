<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  ChefHat,
  ClipboardCopy,
  ListPlus,
  Package,
  Pencil,
  Plus,
  ShoppingBasket,
  Trash2,
} from 'lucide-vue-next'
import AiPromptLinks from '../../components/AiPromptLinks.vue'
import FormError from '../../components/FormError.vue'
import Modal from '../../components/Modal.vue'
import { buildRecipePrompt } from '../../lib/recipePrompt'
import { parseInventoryCsv } from '../../lib/parseInventoryCsv'
import { useInventoryStore } from '../../stores/inventory'
import type { InventoryItem } from '../../api/inventory'

const store = useInventoryStore()

type AddMode = 'add' | 'quick-add'

const addModalOpen = ref(false)
const addMode = ref<AddMode>('add')
const quickAddText = ref('')
const newName = ref('')
const newQuantity = ref('')
const dish = ref('')
const selectedIds = ref<Set<string>>(new Set())
const editingId = ref<string | null>(null)
const editName = ref('')
const editQuantity = ref('')
const copyStatus = ref('')
let copyStatusTimer: ReturnType<typeof setTimeout> | null = null

const items = computed(() => store.items)
const allSelected = computed(
  () => items.value.length > 0 && selectedIds.value.size === items.value.length,
)
const canCopy = computed(() => selectedIds.value.size > 0)

const selectedIngredients = computed(() =>
  items.value.filter((item) => selectedIds.value.has(item.id)),
)

const recipePrompt = computed(() =>
  buildRecipePrompt(
    selectedIngredients.value.map((item) => ({
      name: item.name,
      quantity: item.quantity,
    })),
    dish.value,
    'inventory',
  ),
)

function toggleSelect(id: string) {
  const next = new Set(selectedIds.value)
  if (next.has(id)) {
    next.delete(id)
  } else {
    next.add(id)
  }
  selectedIds.value = next
}

function selectAll() {
  selectedIds.value = new Set(items.value.map((item) => item.id))
}

function clearSelection() {
  selectedIds.value = new Set()
}

function toggleSelectAll() {
  if (allSelected.value) {
    clearSelection()
  } else {
    selectAll()
  }
}

function resetAddFields() {
  newName.value = ''
  newQuantity.value = ''
  quickAddText.value = ''
}

function openAddModal(mode: AddMode = 'add') {
  addMode.value = mode
  addModalOpen.value = true
}

function closeAddModal() {
  addModalOpen.value = false
  resetAddFields()
}

async function onAdd() {
  const name = newName.value.trim()
  if (!name) {
    return
  }
  const item = await store.addItem(name, newQuantity.value)
  if (item) {
    closeAddModal()
  }
}

async function onQuickAdd() {
  const entries = parseInventoryCsv(quickAddText.value)
  if (entries.length === 0) {
    return
  }

  for (const entry of entries) {
    const item = await store.addItem(entry.name, entry.quantity)
    if (!item) {
      return
    }
  }

  closeAddModal()
}

function startEdit(item: InventoryItem) {
  editingId.value = item.id
  editName.value = item.name
  editQuantity.value = item.quantity ?? ''
}

function cancelEdit() {
  editingId.value = null
  editName.value = ''
  editQuantity.value = ''
}

async function saveEdit() {
  const id = editingId.value
  if (!id) {
    return
  }
  const updated = await store.updateItem(id, editName.value, editQuantity.value)
  if (updated) {
    cancelEdit()
  }
}

async function onDelete(item: InventoryItem) {
  const ok = window.confirm(`Remove "${item.name}" from inventory?`)
  if (!ok) {
    return
  }
  const removed = await store.removeItem(item.id)
  if (removed) {
    const next = new Set(selectedIds.value)
    next.delete(item.id)
    selectedIds.value = next
    if (editingId.value === item.id) {
      cancelEdit()
    }
  }
}

async function onCopyPrompt() {
  if (!canCopy.value) {
    return
  }
  try {
    await navigator.clipboard.writeText(recipePrompt.value)
    copyStatus.value = 'Prompt copied'
  } catch (err) {
    console.warn('Clipboard write failed', err)
    copyStatus.value = 'Could not copy — try again'
  }
  if (copyStatusTimer) {
    clearTimeout(copyStatusTimer)
  }
  copyStatusTimer = setTimeout(() => {
    copyStatus.value = ''
  }, 2500)
}

onMounted(() => {
  void store.loadDefault()
})
</script>

<template>
  <section class="dash-panel inventory" aria-labelledby="inventory-title">
    <div class="inventory-header">
      <div>
        <h1 id="inventory-title" class="dash-panel__title inventory-title">
          <Package class="icon icon--lg" aria-hidden="true" />
          Inventory
        </h1>
        <p class="dash-panel__desc">
          Track pantry ingredients and quantities, then copy a prompt for an AI cook.
        </p>
      </div>
      <button
        class="btn-primary inventory-add-btn"
        type="button"
        :disabled="store.loading || store.saving"
        @click="openAddModal()"
      >
        <Plus class="icon" aria-hidden="true" />
        Add
      </button>
    </div>

    <Modal
      :open="addModalOpen"
      title="Add ingredients"
      title-id="inventory-add-title"
      @close="closeAddModal"
    >
      <div class="inventory-add-tabs" role="tablist" aria-label="Add mode">
        <button
          class="inventory-add-tabs__tab"
          :class="{ 'is-active': addMode === 'add' }"
          type="button"
          role="tab"
          :aria-selected="addMode === 'add'"
          @click="addMode = 'add'"
        >
          <Plus class="icon icon--sm" aria-hidden="true" />
          Add
        </button>
        <button
          class="inventory-add-tabs__tab"
          :class="{ 'is-active': addMode === 'quick-add' }"
          type="button"
          role="tab"
          :aria-selected="addMode === 'quick-add'"
          @click="addMode = 'quick-add'"
        >
          <ListPlus class="icon icon--sm" aria-hidden="true" />
          Quick add
        </button>
      </div>

      <form
        v-if="addMode === 'add'"
        class="inventory-add"
        @submit.prevent="onAdd"
      >
        <label class="field">
          <span>Ingredient</span>
          <input
            v-model="newName"
            type="text"
            autocomplete="off"
            placeholder="e.g. chicken"
            :disabled="store.loading || store.saving"
            maxlength="80"
          />
        </label>
        <label class="field">
          <span>Quantity (optional)</span>
          <input
            v-model="newQuantity"
            type="text"
            autocomplete="off"
            placeholder="e.g. 500g"
            :disabled="store.loading || store.saving"
            maxlength="40"
          />
        </label>
        <button
          class="btn-primary inventory-add__submit"
          type="submit"
          :disabled="store.loading || store.saving || !newName.trim()"
        >
          <Plus class="icon" aria-hidden="true" />
          Add
        </button>
      </form>

      <form
        v-else
        class="inventory-quick-add"
        @submit.prevent="onQuickAdd"
      >
        <label class="field">
          <span>Ingredients (CSV)</span>
          <input
            v-model="quickAddText"
            type="text"
            autocomplete="off"
            placeholder="chicken: 100g, tomato, ginger: 100g"
            :disabled="store.loading || store.saving"
          />
        </label>
        <p class="inventory-quick-add__hint">
          Separate items with commas. Use <code>name: quantity</code> for amounts.
        </p>
        <button
          class="btn-primary inventory-add__submit"
          type="submit"
          :disabled="store.loading || store.saving || !quickAddText.trim()"
        >
          <Plus class="icon" aria-hidden="true" />
          Add all
        </button>
      </form>
    </Modal>

    <FormError :error="store.error" />

    <p v-if="store.loading" class="inventory-status" role="status">Loading pantry…</p>

    <template v-else>
      <div class="inventory-toolbar">
        <label class="inventory-toolbar__select-all">
          <input
            type="checkbox"
            :checked="allSelected"
            :disabled="items.length === 0"
            :indeterminate="
              selectedIds.size > 0 && selectedIds.size < items.length
            "
            @change="toggleSelectAll"
          />
          <span>Select all</span>
        </label>
        <button
          v-if="selectedIds.size > 0"
          class="link-button"
          type="button"
          @click="clearSelection"
        >
          Clear selection
        </button>
      </div>

      <ul v-if="items.length > 0" class="inventory-list" aria-label="Ingredients">
        <li v-for="item in items" :key="item.id" class="inventory-card">
          <template v-if="editingId === item.id">
            <div class="inventory-card__edit">
              <label class="field">
                <span class="sr-only">Name</span>
                <input
                  v-model="editName"
                  type="text"
                  maxlength="80"
                  :disabled="store.saving"
                />
              </label>
              <label class="field">
                <span class="sr-only">Quantity</span>
                <input
                  v-model="editQuantity"
                  type="text"
                  placeholder="Quantity"
                  maxlength="40"
                  :disabled="store.saving"
                />
              </label>
              <div class="inventory-card__edit-actions">
                <button
                  class="btn-primary"
                  type="button"
                  :disabled="store.saving || !editName.trim()"
                  @click="saveEdit"
                >
                  Save
                </button>
                <button
                  class="btn-ghost"
                  type="button"
                  :disabled="store.saving"
                  @click="cancelEdit"
                >
                  Cancel
                </button>
              </div>
            </div>
          </template>
          <template v-else>
            <label class="inventory-card__select">
              <input
                type="checkbox"
                :checked="selectedIds.has(item.id)"
                @change="toggleSelect(item.id)"
              />
              <span class="sr-only">Select {{ item.name }}</span>
            </label>
            <div class="inventory-card__body">
              <p class="inventory-card__name">{{ item.name }}</p>
              <p class="inventory-card__qty">
                {{ item.quantity?.trim() ? item.quantity : 'No quantity' }}
              </p>
            </div>
            <div class="inventory-card__actions">
              <button
                class="btn-ghost"
                type="button"
                :disabled="store.saving"
                @click="startEdit(item)"
              >
                <Pencil class="icon icon--sm" aria-hidden="true" />
                Edit
              </button>
              <button
                class="btn-ghost inventory-card__delete"
                type="button"
                :disabled="store.saving"
                @click="onDelete(item)"
              >
                <Trash2 class="icon icon--sm" aria-hidden="true" />
                Delete
              </button>
            </div>
          </template>
        </li>
      </ul>
      <p v-else class="inventory-empty">
        <ShoppingBasket class="icon icon--lg" aria-hidden="true" />
        <span>No ingredients yet — tap Add to get started.</span>
        <button
          class="btn-primary inventory-empty__add"
          type="button"
          :disabled="store.saving"
          @click="openAddModal()"
        >
          <Plus class="icon" aria-hidden="true" />
          Add
        </button>
      </p>
    </template>

    <label class="field inventory-dish">
      <span>Dish (optional)</span>
      <input
        v-model="dish"
        type="text"
        autocomplete="off"
        placeholder="e.g. chicken curry"
        maxlength="120"
      />
    </label>

    <div class="inventory-actions">
      <button class="btn-primary" type="button" disabled title="Coming soon">
        <ChefHat class="icon" aria-hidden="true" />
        Generate (soon)
      </button>
      <div class="prompt-actions">
        <button
          class="btn-ghost"
          type="button"
          :disabled="!canCopy"
          @click="onCopyPrompt"
        >
          <ClipboardCopy class="icon" aria-hidden="true" />
          Copy prompt
        </button>
        <AiPromptLinks :prompt="recipePrompt" :disabled="!canCopy" />
      </div>
    </div>
    <p v-if="copyStatus" class="inventory-copy-status" role="status">
      {{ copyStatus }}
    </p>
  </section>
</template>
