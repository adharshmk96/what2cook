<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ChefHat, ClipboardCopy, Package, Pencil, Plus, ShoppingBasket, Trash2 } from 'lucide-vue-next'
import AiPromptLinks from '../../components/AiPromptLinks.vue'
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
const categories = computed(() => [...new Set(items.value.map(i => i.category?.trim()).filter((v): v is string => !!v))].sort((a,b) => a.localeCompare(b)))
const groupedItems = computed(() => {
  const groups = new Map<string, InventoryItem[]>()
  for (const item of items.value) {
    const key = item.category?.trim() || 'Uncategorized'
    groups.set(key, [...(groups.get(key) ?? []), item])
  }
  return [...groups.entries()].sort(([a], [b]) => a === 'Uncategorized' ? 1 : b === 'Uncategorized' ? -1 : a.localeCompare(b))
})
const allSelected = computed(() => items.value.length > 0 && selectedIds.value.size === items.value.length)
const selectedIngredients = computed(() => items.value.filter(i => selectedIds.value.has(i.id)))
const canCopy = computed(() => selectedIds.value.size > 0)
const recipePrompt = computed(() => buildRecipePrompt(selectedIngredients.value.map(i => ({ name: i.name, quantity: i.quantity })), dish.value, 'inventory'))

function toggleSelect(id: string) { const next = new Set(selectedIds.value); next.has(id) ? next.delete(id) : next.add(id); selectedIds.value = next }
function toggleSelectAll() { selectedIds.value = allSelected.value ? new Set() : new Set(items.value.map(i => i.id)) }
function resetAddFields() { newName.value=''; newQuantity.value=''; newCategory.value='' }
function closeAddModal() { addModalOpen.value=false; resetAddFields() }
async function onAdd() { const item = await store.addItem(newName.value, newQuantity.value, newCategory.value); if (item) closeAddModal() }
function startEdit(item: InventoryItem) { editingId.value=item.id; editName.value=item.name; editQuantity.value=item.quantity ?? ''; editCategory.value=item.category ?? '' }
function cancelEdit() { editingId.value=null; editName.value=''; editQuantity.value=''; editCategory.value='' }
async function saveEdit() { if (!editingId.value) return; const updated=await store.updateItem(editingId.value,editName.value,editQuantity.value,editCategory.value); if(updated) cancelEdit() }
async function onDelete(item: InventoryItem) { if(!window.confirm(`Remove "${item.name}" from inventory?`)) return; if(await store.removeItem(item.id)){ const next=new Set(selectedIds.value); next.delete(item.id); selectedIds.value=next; if(editingId.value===item.id) cancelEdit() } }
async function onCopyPrompt() { if(!canCopy.value)return; try { await navigator.clipboard.writeText(recipePrompt.value); copyStatus.value='Prompt copied' } catch { copyStatus.value='Could not copy — try again' }; setTimeout(()=>copyStatus.value='',2500) }
onMounted(()=>void store.loadDefault())
</script>

<template>
  <section class="dash-panel inventory" aria-labelledby="inventory-title">
    <div class="inventory-header">
      <div><h1 id="inventory-title" class="dash-panel__title inventory-title"><Package class="icon icon--lg" /> Inventory</h1><p class="dash-panel__desc">Track pantry ingredients by category and quantity, then copy a prompt for an AI cook.</p></div>
      <button class="btn-primary inventory-add-btn" type="button" :disabled="store.loading || store.saving" @click="addModalOpen=true"><Plus class="icon" /> Add</button>
    </div>

    <Modal :open="addModalOpen" title="Add ingredient" title-id="inventory-add-title" @close="closeAddModal">
      <form class="inventory-add" @submit.prevent="onAdd">
        <label class="field"><span>Ingredient</span><input v-model="newName" type="text" autocomplete="off" placeholder="e.g. chicken" maxlength="80" :disabled="store.saving" /></label>
        <label class="field"><span>Quantity (optional)</span><input v-model="newQuantity" type="text" autocomplete="off" placeholder="e.g. 500g" maxlength="40" :disabled="store.saving" /></label>
        <label class="field"><span>Category (optional)</span><input v-model="newCategory" type="text" list="inventory-categories" autocomplete="off" placeholder="e.g. Meat, Vegetables, Spices" maxlength="60" :disabled="store.saving" /></label>
        <datalist id="inventory-categories"><option v-for="category in categories" :key="category" :value="category" /></datalist>
        <button class="btn-primary inventory-add__submit" type="submit" :disabled="store.saving || !newName.trim()"><Plus class="icon" /> Add</button>
      </form>
    </Modal>

    <FormError :error="store.error" />
    <p v-if="store.loading" class="inventory-status" role="status">Loading pantry…</p>
    <template v-else>
      <div class="inventory-toolbar">
        <label class="inventory-toolbar__select-all"><input type="checkbox" :checked="allSelected" :disabled="items.length===0" :indeterminate="selectedIds.size>0 && selectedIds.size<items.length" @change="toggleSelectAll" /><span>Select all</span></label>
        <button v-if="selectedIds.size>0" class="link-button" type="button" @click="selectedIds=new Set()">Clear selection</button>
      </div>

      <div v-if="items.length>0">
        <section v-for="[category, categoryItems] in groupedItems" :key="category" class="inventory-category">
          <h2 class="inventory-category__title">{{ category }} <span>({{ categoryItems.length }})</span></h2>
          <ul class="inventory-list" :aria-label="`${category} ingredients`">
            <li v-for="item in categoryItems" :key="item.id" class="inventory-card">
              <template v-if="editingId===item.id">
                <div class="inventory-card__edit">
                  <label class="field"><span class="sr-only">Name</span><input v-model="editName" type="text" maxlength="80" :disabled="store.saving" /></label>
                  <label class="field"><span class="sr-only">Quantity</span><input v-model="editQuantity" type="text" placeholder="Quantity" maxlength="40" :disabled="store.saving" /></label>
                  <label class="field"><span class="sr-only">Category</span><input v-model="editCategory" type="text" list="inventory-categories" placeholder="Category" maxlength="60" :disabled="store.saving" /></label>
                  <div class="inventory-card__edit-actions"><button class="btn-primary" type="button" :disabled="store.saving || !editName.trim()" @click="saveEdit">Save</button><button class="btn-ghost" type="button" :disabled="store.saving" @click="cancelEdit">Cancel</button></div>
                </div>
              </template>
              <template v-else>
                <label class="inventory-card__select"><input type="checkbox" :checked="selectedIds.has(item.id)" @change="toggleSelect(item.id)" /><span class="sr-only">Select {{ item.name }}</span></label>
                <div class="inventory-card__body"><p class="inventory-card__name">{{ item.name }}</p><p class="inventory-card__qty">{{ item.quantity?.trim() ? item.quantity : 'No quantity' }}</p></div>
                <div class="inventory-card__actions"><button class="btn-ghost" type="button" :disabled="store.saving" @click="startEdit(item)"><Pencil class="icon icon--sm" /> Edit</button><button class="btn-ghost inventory-card__delete" type="button" :disabled="store.saving" @click="onDelete(item)"><Trash2 class="icon icon--sm" /> Delete</button></div>
              </template>
            </li>
          </ul>
        </section>
      </div>
      <p v-else class="inventory-empty"><ShoppingBasket class="icon icon--lg" /><span>No ingredients yet — tap Add to get started.</span><button class="btn-primary inventory-empty__add" type="button" @click="addModalOpen=true"><Plus class="icon" /> Add</button></p>
    </template>

    <label class="field inventory-dish"><span>Dish (optional)</span><input v-model="dish" type="text" autocomplete="off" placeholder="e.g. chicken curry" maxlength="120" /></label>
    <div class="inventory-actions"><button class="btn-primary" type="button" disabled title="Coming soon"><ChefHat class="icon" /> Generate (soon)</button><div class="prompt-actions"><button class="btn-ghost" type="button" :disabled="!canCopy" @click="onCopyPrompt"><ClipboardCopy class="icon" /> Copy prompt</button><AiPromptLinks :prompt="recipePrompt" :disabled="!canCopy" /></div></div>
    <p v-if="copyStatus" class="inventory-copy-status" role="status">{{ copyStatus }}</p>
  </section>
</template>

<style scoped>
.inventory-category { margin-top: 1.25rem; }
.inventory-category__title { margin: 0 0 .6rem; font-size: .9rem; font-weight: 700; letter-spacing: .02em; }
.inventory-category__title span { opacity: .6; font-weight: 500; }
</style>
