<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import {
  BookOpen,
  ChevronDown,
  ChevronUp,
  Flame,
  Pencil,
  Plus,
  Trash2,
  Users,
  X,
} from 'lucide-vue-next'
import FormError from '../../components/FormError.vue'
import Modal from '../../components/Modal.vue'
import { useSavedRecipesStore } from '../../stores/savedRecipes'
import type {
  Nutrition,
  SavedRecipe,
  SaveRecipeInput,
} from '../../api/savedRecipes'

type FormIngredient = { name: string; quantity: string }
type FormStepIngredient = { name: string; quantity: string }
type FormStep = { instruction: string; ingredientsUsed: FormStepIngredient[] }

const NUTRITION_FIELDS: { key: keyof Nutrition; label: string; unit: string }[] = [
  { key: 'calories', label: 'Calories', unit: 'kcal' },
  { key: 'protein_g', label: 'Protein', unit: 'g' },
  { key: 'carbs_g', label: 'Carbs', unit: 'g' },
  { key: 'fat_g', label: 'Fat', unit: 'g' },
  { key: 'fiber_g', label: 'Fiber', unit: 'g' },
  { key: 'sugar_g', label: 'Sugar', unit: 'g' },
  { key: 'sodium_mg', label: 'Sodium', unit: 'mg' },
]

const store = useSavedRecipesStore()

const modalOpen = ref(false)
const editingId = ref<string | null>(null)
const expandedIds = ref<Set<string>>(new Set())

const formTitle = ref('')
const formSummary = ref('')
const formMinutes = ref('')
const formServings = ref('')
const formIngredients = ref<FormIngredient[]>([{ name: '', quantity: '' }])
const formSteps = ref<FormStep[]>([{ instruction: '', ingredientsUsed: [] }])
const showNutrition = ref(false)
const formNutrition = reactive<Record<string, string>>({
  calories: '',
  protein_g: '',
  carbs_g: '',
  fat_g: '',
  fiber_g: '',
  sugar_g: '',
  sodium_mg: '',
})
const formError = ref('')

const recipes = computed(() => store.recipes)
const isEditing = computed(() => editingId.value !== null)

function toggleExpanded(id: string) {
  const next = new Set(expandedIds.value)
  next.has(id) ? next.delete(id) : next.add(id)
  expandedIds.value = next
}

function resetForm() {
  formTitle.value = ''
  formSummary.value = ''
  formMinutes.value = ''
  formServings.value = ''
  formIngredients.value = [{ name: '', quantity: '' }]
  formSteps.value = [{ instruction: '', ingredientsUsed: [] }]
  showNutrition.value = false
  for (const field of NUTRITION_FIELDS) formNutrition[field.key] = ''
  formError.value = ''
}

function openCreateModal() {
  editingId.value = null
  resetForm()
  modalOpen.value = true
}

function openEditModal(recipe: SavedRecipe) {
  editingId.value = recipe.id
  formTitle.value = recipe.title
  formSummary.value = recipe.summary
  formMinutes.value = recipe.minutes ? String(recipe.minutes) : ''
  formServings.value = recipe.servings ? String(recipe.servings) : ''
  formIngredients.value = recipe.ingredients.length
    ? recipe.ingredients.map((i) => ({ name: i.name, quantity: i.quantity }))
    : [{ name: '', quantity: '' }]
  formSteps.value = recipe.steps.length
    ? recipe.steps.map((s) => ({
        instruction: s.instruction,
        ingredientsUsed: s.ingredients_used.map((i) => ({ name: i.name, quantity: i.quantity })),
      }))
    : [{ instruction: '', ingredientsUsed: [] }]
  showNutrition.value = !!recipe.nutrition
  for (const field of NUTRITION_FIELDS) {
    const value = recipe.nutrition?.[field.key]
    formNutrition[field.key] = value != null ? String(value) : ''
  }
  formError.value = ''
  modalOpen.value = true
}

function closeModal() {
  modalOpen.value = false
  editingId.value = null
  resetForm()
}

function addIngredientRow() {
  formIngredients.value = [...formIngredients.value, { name: '', quantity: '' }]
}

function removeIngredientRow(index: number) {
  const removed = formIngredients.value[index]
  formIngredients.value = formIngredients.value.filter((_, i) => i !== index)
  if (removed?.name.trim()) {
    for (const step of formSteps.value) {
      step.ingredientsUsed = step.ingredientsUsed.filter(
        (used) => used.name !== removed.name.trim(),
      )
    }
  }
}

function renameIngredient(ingredient: FormIngredient, nextName: string) {
  const oldName = ingredient.name.trim()
  ingredient.name = nextName
  const newName = nextName.trim()
  if (!oldName || oldName === newName) return
  for (const step of formSteps.value) {
    step.ingredientsUsed = step.ingredientsUsed.map((used) =>
      used.name === oldName ? { ...used, name: newName } : used,
    )
  }
}

function addStepRow() {
  formSteps.value = [...formSteps.value, { instruction: '', ingredientsUsed: [] }]
}

function removeStepRow(index: number) {
  formSteps.value = formSteps.value.filter((_, i) => i !== index)
}

const availableIngredientNames = computed(() =>
  formIngredients.value
    .map((i) => i.name.trim())
    .filter((name, index, all) => name && all.indexOf(name) === index),
)

function isStepIngredientUsed(step: FormStep, name: string) {
  return step.ingredientsUsed.some((used) => used.name === name)
}

function toggleStepIngredient(step: FormStep, name: string) {
  if (isStepIngredientUsed(step, name)) {
    step.ingredientsUsed = step.ingredientsUsed.filter((used) => used.name !== name)
    return
  }
  const source = formIngredients.value.find((i) => i.name.trim() === name)
  step.ingredientsUsed = [
    ...step.ingredientsUsed,
    { name, quantity: source?.quantity ?? '' },
  ]
}

function stepIngredientQuantity(step: FormStep, name: string) {
  return step.ingredientsUsed.find((used) => used.name === name)?.quantity ?? ''
}

function setStepIngredientQuantity(step: FormStep, name: string, quantity: string) {
  step.ingredientsUsed = step.ingredientsUsed.map((used) =>
    used.name === name ? { ...used, quantity } : used,
  )
}

function buildNutritionPayload(): Nutrition | null {
  if (!showNutrition.value) return null
  const nutrition: Nutrition = {}
  let hasValue = false
  for (const field of NUTRITION_FIELDS) {
    const raw = formNutrition[field.key].trim()
    if (raw === '') continue
    const num = Number(raw)
    if (Number.isNaN(num)) continue
    nutrition[field.key] = num
    hasValue = true
  }
  return hasValue ? nutrition : null
}

function validateForm(): string {
  if (!formTitle.value.trim()) return 'Title is required'
  const cleanIngredients = formIngredients.value.filter((i) => i.name.trim())
  if (cleanIngredients.length === 0) return 'Add at least one ingredient'
  const cleanSteps = formSteps.value.filter((s) => s.instruction.trim())
  if (cleanSteps.length === 0) return 'Add at least one step'
  return ''
}

async function onSubmit() {
  const message = validateForm()
  if (message) {
    formError.value = message
    return
  }
  formError.value = ''

  const input: SaveRecipeInput = {
    title: formTitle.value.trim(),
    summary: formSummary.value.trim(),
    minutes: Number(formMinutes.value) || 0,
    servings: Number(formServings.value) || 0,
    ingredients: formIngredients.value
      .filter((i) => i.name.trim())
      .map((i) => ({ name: i.name.trim(), quantity: i.quantity.trim() })),
    steps: formSteps.value
      .filter((s) => s.instruction.trim())
      .map((s) => ({
        instruction: s.instruction.trim(),
        ingredients_used: s.ingredientsUsed
          .filter((u) => u.name.trim())
          .map((u) => ({ name: u.name.trim(), quantity: u.quantity.trim() })),
      })),
    nutrition: buildNutritionPayload(),
  }

  const result = editingId.value
    ? await store.update(editingId.value, input)
    : await store.create(input)

  if (result) closeModal()
}

async function onDelete(recipe: SavedRecipe) {
  if (!window.confirm(`Delete "${recipe.title}"? This cannot be undone.`)) return
  await store.remove(recipe.id)
}

store.load()
</script>

<template>
  <section class="dash-panel saved-recipes" aria-labelledby="saved-recipes-title">
    <div class="saved-recipes__header">
      <div>
        <h1 id="saved-recipes-title" class="dash-panel__title">
          <BookOpen class="icon icon--lg" aria-hidden="true" />
          Saved Recipes
        </h1>
        <p class="dash-panel__desc">
          Keep your favorite recipes with ingredients, step-by-step
          instructions, and optional nutrition info.
        </p>
      </div>
      <button class="btn-primary" type="button" @click="openCreateModal">
        <Plus class="icon" aria-hidden="true" />
        New recipe
      </button>
    </div>

    <FormError :error="store.error" />
    <p v-if="store.loading" class="saved-recipes__status" role="status">
      Loading recipes…
    </p>

    <template v-else>
      <ul v-if="recipes.length > 0" class="saved-recipes__list">
        <li v-for="recipe in recipes" :key="recipe.id" class="recipe-item">
          <div class="recipe-item__header">
            <button
              class="recipe-item__toggle"
              type="button"
              @click="toggleExpanded(recipe.id)"
            >
              <component
                :is="expandedIds.has(recipe.id) ? ChevronUp : ChevronDown"
                class="icon icon--sm"
                aria-hidden="true"
              />
              <div class="recipe-item__identity">
                <h2 class="recipe-item__title">{{ recipe.title }}</h2>
                <p v-if="recipe.summary" class="recipe-item__summary">
                  {{ recipe.summary }}
                </p>
              </div>
            </button>

            <div class="recipe-item__meta">
              <span v-if="recipe.minutes" class="recipe-item__meta-chip">
                {{ recipe.minutes }} min
              </span>
              <span v-if="recipe.servings" class="recipe-item__meta-chip">
                <Users class="icon icon--sm" aria-hidden="true" />
                {{ recipe.servings }}
              </span>
              <span v-if="recipe.nutrition" class="recipe-item__meta-chip">
                <Flame class="icon icon--sm" aria-hidden="true" />
                {{ recipe.nutrition.calories ?? '—' }} kcal
              </span>
            </div>

            <div class="recipe-item__actions">
              <button
                class="btn-ghost"
                type="button"
                :disabled="store.saving"
                @click="openEditModal(recipe)"
              >
                <Pencil class="icon icon--sm" aria-hidden="true" />
                Edit
              </button>
              <button
                class="btn-ghost recipe-item__delete"
                type="button"
                :disabled="store.saving"
                @click="onDelete(recipe)"
              >
                <Trash2 class="icon icon--sm" aria-hidden="true" />
                Delete
              </button>
            </div>
          </div>

          <div v-if="expandedIds.has(recipe.id)" class="recipe-item__detail">
            <div class="recipe-item__section">
              <h3>Ingredients</h3>
              <ul class="recipe-item__ingredients">
                <li v-for="ingredient in recipe.ingredients" :key="ingredient.id">
                  <span class="recipe-item__ingredient-name">{{ ingredient.name }}</span>
                  <span v-if="ingredient.quantity" class="recipe-item__ingredient-qty">
                    {{ ingredient.quantity }}
                  </span>
                </li>
              </ul>
            </div>

            <div class="recipe-item__section">
              <h3>Steps</h3>
              <ol class="recipe-item__steps">
                <li v-for="step in recipe.steps" :key="step.id">
                  <p>{{ step.instruction }}</p>
                  <p v-if="step.ingredients_used.length" class="recipe-item__step-uses">
                    Uses:
                    <span
                      v-for="(used, index) in step.ingredients_used"
                      :key="used.name"
                    >
                      {{ used.name }}<template v-if="used.quantity"> ({{ used.quantity }})</template
                      ><template v-if="index < step.ingredients_used.length - 1">, </template>
                    </span>
                  </p>
                </li>
              </ol>
            </div>

            <div v-if="recipe.nutrition" class="recipe-item__section">
              <h3>Nutrition</h3>
              <dl class="recipe-item__nutrition">
                <template v-for="field in NUTRITION_FIELDS" :key="field.key">
                  <template v-if="recipe.nutrition[field.key] != null">
                    <dt>{{ field.label }}</dt>
                    <dd>{{ recipe.nutrition[field.key] }} {{ field.unit }}</dd>
                  </template>
                </template>
              </dl>
            </div>
          </div>
        </li>
      </ul>

      <p v-else class="saved-recipes__empty">
        <BookOpen class="icon icon--lg" aria-hidden="true" />
        <span>No saved recipes yet — create your first one.</span>
        <button class="btn-primary" type="button" @click="openCreateModal">
          <Plus class="icon" aria-hidden="true" />
          New recipe
        </button>
      </p>
    </template>

    <Modal
      :open="modalOpen"
      :title="isEditing ? 'Edit recipe' : 'New recipe'"
      title-id="saved-recipe-modal-title"
      @close="closeModal"
    >
      <form class="recipe-form" @submit.prevent="onSubmit">
        <div class="recipe-form__row">
          <label class="field recipe-form__title">
            <span>Title</span>
            <input
              v-model="formTitle"
              type="text"
              maxlength="120"
              placeholder="e.g. Chicken & Garlic Skillet"
              :disabled="store.saving"
            />
          </label>
        </div>

        <label class="field">
          <span>Summary (optional)</span>
          <input
            v-model="formSummary"
            type="text"
            maxlength="500"
            placeholder="A quick, one-pan weeknight dinner"
            :disabled="store.saving"
          />
        </label>

        <div class="recipe-form__row">
          <label class="field">
            <span>Minutes</span>
            <input
              v-model="formMinutes"
              type="number"
              min="0"
              placeholder="25"
              :disabled="store.saving"
            />
          </label>
          <label class="field">
            <span>Servings</span>
            <input
              v-model="formServings"
              type="number"
              min="0"
              placeholder="2"
              :disabled="store.saving"
            />
          </label>
        </div>

        <fieldset class="recipe-form__section">
          <legend>Ingredients</legend>
          <div
            v-for="(ingredient, index) in formIngredients"
            :key="index"
            class="recipe-form__ingredient-row"
          >
            <label class="field">
              <span class="sr-only">Ingredient name</span>
              <input
                :value="ingredient.name"
                type="text"
                maxlength="80"
                placeholder="Ingredient"
                :disabled="store.saving"
                @input="renameIngredient(ingredient, ($event.target as HTMLInputElement).value)"
              />
            </label>
            <label class="field">
              <span class="sr-only">Quantity</span>
              <input
                v-model="ingredient.quantity"
                type="text"
                maxlength="40"
                placeholder="Quantity (e.g. 500g)"
                :disabled="store.saving"
              />
            </label>
            <button
              class="btn-ghost recipe-form__remove"
              type="button"
              :disabled="store.saving || formIngredients.length === 1"
              @click="removeIngredientRow(index)"
              aria-label="Remove ingredient"
            >
              <X class="icon icon--sm" aria-hidden="true" />
            </button>
          </div>
          <button
            class="link-button"
            type="button"
            :disabled="store.saving"
            @click="addIngredientRow"
          >
            <Plus class="icon icon--sm" aria-hidden="true" />
            Add ingredient
          </button>
        </fieldset>

        <fieldset class="recipe-form__section">
          <legend>Steps</legend>
          <div
            v-for="(step, index) in formSteps"
            :key="index"
            class="recipe-form__step"
          >
            <div class="recipe-form__step-header">
              <span class="recipe-form__step-number">{{ index + 1 }}</span>
              <label class="field recipe-form__step-instruction">
                <span class="sr-only">Step instruction</span>
                <textarea
                  v-model="step.instruction"
                  rows="2"
                  maxlength="1000"
                  placeholder="Describe this step"
                  :disabled="store.saving"
                ></textarea>
              </label>
              <button
                class="btn-ghost recipe-form__remove"
                type="button"
                :disabled="store.saving || formSteps.length === 1"
                @click="removeStepRow(index)"
                aria-label="Remove step"
              >
                <X class="icon icon--sm" aria-hidden="true" />
              </button>
            </div>

            <div v-if="availableIngredientNames.length" class="recipe-form__step-ingredients">
              <p class="recipe-form__step-ingredients-label">Ingredients used in this step:</p>
              <div class="recipe-form__step-chips">
                <div
                  v-for="name in availableIngredientNames"
                  :key="name"
                  class="recipe-form__step-chip"
                  :class="{ 'is-active': isStepIngredientUsed(step, name) }"
                >
                  <label>
                    <input
                      type="checkbox"
                      :checked="isStepIngredientUsed(step, name)"
                      :disabled="store.saving"
                      @change="toggleStepIngredient(step, name)"
                    />
                    <span>{{ name }}</span>
                  </label>
                  <input
                    v-if="isStepIngredientUsed(step, name)"
                    class="recipe-form__step-chip-qty"
                    type="text"
                    maxlength="40"
                    placeholder="qty"
                    :value="stepIngredientQuantity(step, name)"
                    :disabled="store.saving"
                    @input="
                      setStepIngredientQuantity(
                        step,
                        name,
                        ($event.target as HTMLInputElement).value,
                      )
                    "
                  />
                </div>
              </div>
            </div>
          </div>
          <button
            class="link-button"
            type="button"
            :disabled="store.saving"
            @click="addStepRow"
          >
            <Plus class="icon icon--sm" aria-hidden="true" />
            Add step
          </button>
        </fieldset>

        <fieldset class="recipe-form__section">
          <legend>
            <label class="recipe-form__nutrition-toggle">
              <input v-model="showNutrition" type="checkbox" :disabled="store.saving" />
              <span>Nutritional value (optional)</span>
            </label>
          </legend>
          <div v-if="showNutrition" class="recipe-form__nutrition-grid">
            <label v-for="field in NUTRITION_FIELDS" :key="field.key" class="field">
              <span>{{ field.label }} ({{ field.unit }})</span>
              <input
                v-model="formNutrition[field.key]"
                type="number"
                min="0"
                step="any"
                :disabled="store.saving"
              />
            </label>
          </div>
        </fieldset>

        <p v-if="formError" class="form-error" role="alert">{{ formError }}</p>

        <div class="recipe-form__actions">
          <button class="btn-primary" type="submit" :disabled="store.saving">
            {{ isEditing ? 'Save changes' : 'Save recipe' }}
          </button>
          <button
            class="btn-ghost"
            type="button"
            :disabled="store.saving"
            @click="closeModal"
          >
            Cancel
          </button>
        </div>
      </form>
    </Modal>
  </section>
</template>

<style scoped>
.saved-recipes__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
}

.saved-recipes__status {
  margin-top: 1rem;
  color: var(--text-muted, #6b7280);
}

.saved-recipes__list {
  display: grid;
  gap: 1rem;
  margin-top: 1.25rem;
  padding: 0;
  list-style: none;
}

.recipe-item {
  overflow: hidden;
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 1rem;
  background: var(--surface, #fff);
}

.recipe-item__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
  padding: 1rem 1.1rem;
}

.recipe-item__toggle {
  display: flex;
  align-items: flex-start;
  gap: 0.6rem;
  flex: 1 1 220px;
  min-width: 0;
  border: 0;
  background: none;
  padding: 0;
  cursor: pointer;
  text-align: left;
  color: inherit;
  font: inherit;
}

.recipe-item__title {
  margin: 0;
  font-size: 1rem;
  font-weight: 750;
}

.recipe-item__summary {
  margin: 0.25rem 0 0;
  color: var(--text-muted, #6b7280);
  font-size: 0.85rem;
}

.recipe-item__meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.recipe-item__meta-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.25rem 0.6rem;
  border-radius: 999px;
  background: var(--surface-subtle, rgba(127, 127, 127, 0.08));
  font-size: 0.78rem;
  color: var(--text-muted, #6b7280);
}

.recipe-item__actions {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.recipe-item__delete {
  color: var(--danger, #dc2626);
}

.recipe-item__detail {
  display: grid;
  gap: 1rem;
  padding: 0 1.1rem 1.1rem;
  border-top: 1px solid var(--border-color, #e5e7eb);
}

.recipe-item__section h3 {
  margin: 1rem 0 0.5rem;
  font-size: 0.85rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-muted, #6b7280);
}

.recipe-item__ingredients {
  margin: 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: 0.3rem;
}

.recipe-item__ingredients li {
  display: flex;
  justify-content: space-between;
  gap: 0.6rem;
  padding: 0.35rem 0.6rem;
  border-radius: 0.5rem;
  background: var(--surface-subtle, rgba(127, 127, 127, 0.04));
}

.recipe-item__ingredient-qty {
  color: var(--text-muted, #6b7280);
  font-size: 0.85rem;
}

.recipe-item__steps {
  margin: 0;
  padding-left: 1.2rem;
  display: grid;
  gap: 0.6rem;
}

.recipe-item__steps p {
  margin: 0;
}

.recipe-item__step-uses {
  margin-top: 0.2rem !important;
  font-size: 0.82rem;
  color: var(--text-muted, #6b7280);
}

.recipe-item__nutrition {
  margin: 0;
  display: grid;
  grid-template-columns: max-content 1fr;
  gap: 0.3rem 0.8rem;
}

.recipe-item__nutrition dt {
  color: var(--text-muted, #6b7280);
  font-size: 0.85rem;
}

.recipe-item__nutrition dd {
  margin: 0;
  font-weight: 600;
}

.saved-recipes__empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  padding: 2.5rem 1rem;
  color: var(--text-muted, #6b7280);
  text-align: center;
}

.recipe-form {
  display: grid;
  gap: 1rem;
}

.recipe-form__row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
}

.recipe-form__row:has(.recipe-form__title) {
  grid-template-columns: 1fr;
}

.recipe-form__section {
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 0.85rem;
  padding: 0.85rem 0.9rem 1rem;
  display: grid;
  gap: 0.6rem;
}

.recipe-form__section legend {
  padding: 0 0.4rem;
  font-weight: 650;
  font-size: 0.85rem;
}

.recipe-form__nutrition-toggle {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  font-weight: 650;
  cursor: pointer;
}

.recipe-form__ingredient-row {
  display: grid;
  grid-template-columns: 1.4fr 1fr auto;
  gap: 0.5rem;
  align-items: center;
}

.recipe-form__step {
  display: grid;
  gap: 0.5rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px dashed var(--border-color, #e5e7eb);
}

.recipe-form__step:last-of-type {
  border-bottom: 0;
  padding-bottom: 0;
}

.recipe-form__step-header {
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: 0.5rem;
  align-items: start;
}

.recipe-form__step-number {
  display: grid;
  place-items: center;
  width: 1.6rem;
  height: 1.6rem;
  margin-top: 0.35rem;
  border-radius: 50%;
  background: var(--surface-subtle, rgba(127, 127, 127, 0.1));
  font-size: 0.75rem;
  font-weight: 700;
}

.recipe-form__step-instruction textarea {
  width: 100%;
  padding: 0.55rem 0.7rem;
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 0.6rem;
  font: inherit;
  resize: vertical;
}

.recipe-form__remove {
  padding: 0.4rem;
}

.recipe-form__step-ingredients {
  padding-left: 2.1rem;
}

.recipe-form__step-ingredients-label {
  margin: 0 0 0.35rem;
  font-size: 0.78rem;
  color: var(--text-muted, #6b7280);
}

.recipe-form__step-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.recipe-form__step-chip {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.3rem 0.55rem;
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 999px;
  background: var(--surface, #fff);
}

.recipe-form__step-chip.is-active {
  background: var(--surface-subtle, rgba(127, 127, 127, 0.08));
}

.recipe-form__step-chip label {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.82rem;
  cursor: pointer;
}

.recipe-form__step-chip-qty {
  width: 4.5rem;
  padding: 0.15rem 0.4rem;
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 0.4rem;
  font-size: 0.78rem;
}

.recipe-form__nutrition-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 0.6rem;
}

.recipe-form__actions {
  display: flex;
  gap: 0.6rem;
}

@media (max-width: 640px) {
  .recipe-form__row {
    grid-template-columns: 1fr;
  }

  .recipe-item__header {
    flex-direction: column;
  }

  .recipe-item__actions {
    align-self: flex-end;
  }
}
</style>
