<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  ChefHat,
  ClipboardCopy,
  Layers3,
  Package,
  Pencil,
  Plus,
  ShoppingBasket,
  Trash2,
} from 'lucide-vue-next'
import AiPromptLinks from '../../components/AiPromptLinks.vue'
import CategoryCombobox from '../../components/CategoryCombobox.vue'
import FormError from '../../components/FormError.vue'
import Modal from '../../components/Modal.vue'
import { buildRecipePrompt } from '../../lib/recipePrompt'
import { useInventoryStore } from '../../stores/inventory'
import type { InventoryItem } from '../../api/inventory'

const store = useInventoryStore()
const addModalOpen = ref(false)
const newName = ref('')
const newQuantity = ref('')
const newCategory = ref('')
const dish = ref('')
const selectedIds = ref<Set<string>>(new Set())
const editingId = ref<string | null>(null)
const editName = ref('')
const editQuantity = ref('')
const editCategory = ref('')
const copyStatus = ref('')

const items = computed(() => store.items)
const categories = computed(() =>
  [
    ...new Set(
      items.value
        .map((item) => item.category?.trim())
        .filter((value): value is string => !!value),
    ),
  ].sort((a, b) => a.localeCompare(b)),
)

const groupedItems = computed(() => {
  const groups = new Map<string, InventoryItem[]>()

  for (const item of items.value) {
    const key = item.category?.trim() || 'Uncategorized'
    groups.set(key, [...(groups.get(key) ?? []), item])
  }

  return [...groups.entries()]
    .map(([category, categoryItems]) => ({
      category,
      items: [...categoryItems].sort((a, b) => a.name.localeCompare(b.name)),
    }))
    .sort((a, b) => {
      if (a.category === 'Uncategorized') return 1
      if (b.category === 'Uncategorized') return -1
      return a.category.localeCompare(b.category)
    })
})

const allSelected = computed(
  () => items.value.length > 0 && selectedIds.value.size === items.value.length,
)
const selectedIngredients = computed(() =>
  items.value.filter((item) => selectedIds.value.has(item.id)),
)
const canCopy = computed(() => selectedIds.value.size > 0)
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
  next.has(id) ? next.delete(id) : next.add(id)
  selectedIds.value = next
}

function toggleSelectAll() {
  selectedIds.value = allSelected.value
    ? new Set()
    : new Set(items.value.map((item) => item.id))
}

function isCategorySelected(categoryItems: InventoryItem[]) {
  return (
    categoryItems.length > 0 &&
    categoryItems.every((item) => selectedIds.value.has(item.id))
  )
}

function isCategoryPartiallySelected(categoryItems: InventoryItem[]) {
  const selectedCount = categoryItems.filter((item) =>
    selectedIds.value.has(item.id),
  ).length
  return selectedCount > 0 && selectedCount < categoryItems.length
}

function toggleCategorySelection(categoryItems: InventoryItem[]) {
  const next = new Set(selectedIds.value)
  const shouldClear = isCategorySelected(categoryItems)

  for (const item of categoryItems) {
    if (shouldClear) {
      next.delete(item.id)
    } else {
      next.add(item.id)
    }
  }

  selectedIds.value = next
}

function resetAddFields() {
  newName.value = ''
  newQuantity.value = ''
  newCategory.value = ''
}

function closeAddModal() {
  addModalOpen.value = false
  resetAddFields()
}

async function onAdd() {
  const item = await store.addItem(
    newName.value,
    newQuantity.value,
    newCategory.value,
  )
  if (item) closeAddModal()
}

function startEdit(item: InventoryItem) {
  editingId.value = item.id
  editName.value = item.name
  editQuantity.value = item.quantity ?? ''
  editCategory.value = item.category ?? ''
}

function cancelEdit() {
  editingId.value = null
  editName.value = ''
  editQuantity.value = ''
  editCategory.value = ''
}

async function saveEdit() {
  if (!editingId.value) return

  const updated = await store.updateItem(
    editingId.value,
    editName.value,
    editQuantity.value,
    editCategory.value,
  )
  if (updated) cancelEdit()
}

async function onDelete(item: InventoryItem) {
  if (!window.confirm(`Remove "${item.name}" from inventory?`)) return

  if (await store.removeItem(item.id)) {
    const next = new Set(selectedIds.value)
    next.delete(item.id)
    selectedIds.value = next
    if (editingId.value === item.id) cancelEdit()
  }
}

async function onCopyPrompt() {
  if (!canCopy.value) return

  try {
    await navigator.clipboard.writeText(recipePrompt.value)
    copyStatus.value = 'Prompt copied'
  } catch {
    copyStatus.value = 'Could not copy — try again'
  }

  setTimeout(() => (copyStatus.value = ''), 2500)
}

onMounted(() => void store.loadDefault())
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
          Track pantry ingredients by category and quantity, then copy a prompt
          for an AI cook.
        </p>
      </div>
      <button
        class="btn-primary inventory-add-btn"
        type="button"
        :disabled="store.loading || store.saving"
        @click="addModalOpen = true"
      >
        <Plus class="icon" aria-hidden="true" />
        Add
      </button>
    </div>

    <Modal
      :open="addModalOpen"
      title="Add ingredient"
      title-id="inventory-add-title"
      @close="closeAddModal"
    >
      <form class="inventory-add" @submit.prevent="onAdd">
        <label class="field">
          <span>Ingredient</span>
          <input
            v-model="newName"
            type="text"
            autocomplete="off"
            placeholder="e.g. chicken"
            maxlength="80"
            :disabled="store.saving"
          />
        </label>
        <label class="field">
          <span>Quantity (optional)</span>
          <input
            v-model="newQuantity"
            type="text"
            autocomplete="off"
            placeholder="e.g. 500g"
            maxlength="40"
            :disabled="store.saving"
          />
        </label>
        <label class="field">
          <span>Category (optional)</span>
          <CategoryCombobox
            v-model="newCategory"
            :options="categories"
            input-id="inventory-add-category"
            list-id="inventory-add-category-list"
            :disabled="store.saving"
          />
        </label>
        <button
          class="btn-primary inventory-add__submit"
          type="submit"
          :disabled="store.saving || !newName.trim()"
        >
          <Plus class="icon" aria-hidden="true" />
          Add
        </button>
      </form>
    </Modal>

    <FormError :error="store.error" />
    <p v-if="store.loading" class="inventory-status" role="status">
      Loading pantry…
    </p>

    <template v-else>
      <div class="inventory-summary">
        <div class="inventory-summary__stat">
          <strong>{{ items.length }}</strong>
          <span>{{ items.length === 1 ? 'ingredient' : 'ingredients' }}</span>
        </div>
        <div class="inventory-summary__divider" aria-hidden="true"></div>
        <div class="inventory-summary__stat">
          <strong>{{ groupedItems.length }}</strong>
          <span>{{ groupedItems.length === 1 ? 'category' : 'categories' }}</span>
        </div>
      </div>

      <div class="inventory-toolbar">
        <label class="inventory-toolbar__select-all">
          <input
            type="checkbox"
            :checked="allSelected"
            :disabled="items.length === 0"
            :indeterminate="selectedIds.size > 0 && selectedIds.size < items.length"
            @change="toggleSelectAll"
          />
          <span>Select all</span>
        </label>
        <span v-if="selectedIds.size > 0" class="inventory-toolbar__selection">
          {{ selectedIds.size }} selected
        </span>
        <button
          v-if="selectedIds.size > 0"
          class="link-button"
          type="button"
          @click="selectedIds = new Set()"
        >
          Clear selection
        </button>
      </div>

      <div v-if="items.length > 0" class="inventory-categories">
        <section
          v-for="group in groupedItems"
          :key="group.category"
          class="inventory-category"
        >
          <div class="inventory-category__header">
            <div class="inventory-category__identity">
              <span class="inventory-category__icon" aria-hidden="true">
                <Layers3 class="icon" />
              </span>
              <div>
                <h2 class="inventory-category__title">{{ group.category }}</h2>
                <p class="inventory-category__meta">
                  {{ group.items.length }}
                  {{ group.items.length === 1 ? 'ingredient' : 'ingredients' }}
                </p>
              </div>
            </div>

            <label class="inventory-category__select-all">
              <input
                type="checkbox"
                :checked="isCategorySelected(group.items)"
                :indeterminate="isCategoryPartiallySelected(group.items)"
                @change="toggleCategorySelection(group.items)"
              />
              <span>Select category</span>
            </label>
          </div>

          <ul
            class="inventory-list inventory-category__list"
            :aria-label="`${group.category} ingredients`"
          >
            <li
              v-for="item in group.items"
              :key="item.id"
              class="inventory-card inventory-category__card"
              :class="{ 'is-selected': selectedIds.has(item.id) }"
            >
              <template v-if="editingId === item.id">
                <div class="inventory-card__edit inventory-category__edit">
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
                  <label class="field">
                    <span class="sr-only">Category</span>
                    <CategoryCombobox
                      v-model="editCategory"
                      :options="categories"
                      list-id="inventory-edit-category-list"
                      placeholder="Category"
                      aria-label="Category"
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
                    class="btn-ghost inventory-category__action"
                    type="button"
                    :disabled="store.saving"
                    @click="startEdit(item)"
                  >
                    <Pencil class="icon icon--sm" aria-hidden="true" />
                    Edit
                  </button>
                  <button
                    class="btn-ghost inventory-card__delete inventory-category__action"
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
        </section>
      </div>

      <p v-else class="inventory-empty">
        <ShoppingBasket class="icon icon--lg" aria-hidden="true" />
        <span>No ingredients yet — tap Add to get started.</span>
        <button
          class="btn-primary inventory-empty__add"
          type="button"
          @click="addModalOpen = true"
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

<style scoped>
.inventory-summary {
  display: inline-flex;
  align-items: center;
  gap: 0.9rem;
  margin: 1rem 0 0.35rem;
  padding: 0.7rem 0.9rem;
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 0.85rem;
  background: var(--surface-subtle, rgba(127, 127, 127, 0.04));
}

.inventory-summary__stat {
  display: flex;
  align-items: baseline;
  gap: 0.35rem;
}

.inventory-summary__stat strong {
  font-size: 1rem;
}

.inventory-summary__stat span,
.inventory-category__meta,
.inventory-toolbar__selection {
  color: var(--text-muted, #6b7280);
  font-size: 0.8rem;
}

.inventory-summary__divider {
  width: 1px;
  height: 1.15rem;
  background: var(--border-color, #e5e7eb);
}

.inventory-toolbar__selection {
  margin-left: auto;
}

.inventory-categories {
  display: grid;
  gap: 1rem;
  margin-top: 0.9rem;
}

.inventory-category {
  overflow: hidden;
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 1rem;
  background: var(--surface, #fff);
}

.inventory-category__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.9rem 1rem;
  border-bottom: 1px solid var(--border-color, #e5e7eb);
  background: var(--surface-subtle, rgba(127, 127, 127, 0.04));
}

.inventory-category__identity {
  display: flex;
  align-items: center;
  gap: 0.7rem;
  min-width: 0;
}

.inventory-category__icon {
  display: grid;
  width: 2.15rem;
  height: 2.15rem;
  flex: 0 0 auto;
  place-items: center;
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 0.65rem;
  background: var(--surface, #fff);
}

.inventory-category__title {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 750;
}

.inventory-category__meta {
  margin: 0.15rem 0 0;
}

.inventory-category__select-all {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  flex: 0 0 auto;
  cursor: pointer;
  font-size: 0.8rem;
  color: var(--text-muted, #6b7280);
}

.inventory-category__list {
  margin: 0;
  padding: 0.5rem;
}

.inventory-category__card {
  border: 0;
  border-radius: 0.75rem;
  transition: background-color 120ms ease, box-shadow 120ms ease;
}

.inventory-category__card + .inventory-category__card {
  border-top: 1px solid var(--border-color, #e5e7eb);
  border-top-left-radius: 0;
  border-top-right-radius: 0;
}

.inventory-category__card.is-selected {
  background: var(--surface-subtle, rgba(127, 127, 127, 0.06));
  box-shadow: inset 3px 0 0 currentColor;
}

.inventory-category__edit {
  width: 100%;
}

.inventory-category__action {
  padding-inline: 0.55rem;
}

@media (max-width: 640px) {
  .inventory-summary {
    display: flex;
    width: 100%;
    justify-content: center;
  }

  .inventory-category__header {
    align-items: flex-start;
  }

  .inventory-category__select-all span {
    display: none;
  }

  .inventory-category__card {
    align-items: flex-start;
  }

  .inventory-category__action {
    padding-inline: 0.4rem;
  }
}
</style>
