<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { buildRecipePrompt } from '../../lib/recipePrompt'

const route = useRoute()

const draft = ref('')
const ingredients = ref<string[]>([])
const dish = ref('')
const copyStatus = ref('')
let copyStatusTimer: ReturnType<typeof setTimeout> | null = null

const canCopy = computed(() => ingredients.value.length > 0)

function parseTokens(raw: string): string[] {
  return raw
    .split(',')
    .map((part) => part.trim())
    .filter(Boolean)
}

function addTokens(raw: string) {
  for (const token of parseTokens(raw)) {
    const exists = ingredients.value.some(
      (item) => item.toLowerCase() === token.toLowerCase(),
    )
    if (!exists) {
      ingredients.value.push(token)
    }
  }
}

function commitDraft() {
  const value = draft.value
  if (!value.trim()) {
    draft.value = ''
    return
  }
  addTokens(value)
  draft.value = ''
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Enter' || event.key === ',') {
    event.preventDefault()
    commitDraft()
  }
}

function onInput(event: Event) {
  const target = event.target as HTMLInputElement
  const value = target.value
  if (!value.includes(',')) {
    return
  }
  const parts = value.split(',')
  const complete = parts.slice(0, -1).join(',')
  const remainder = parts[parts.length - 1] ?? ''
  if (complete.trim()) {
    addTokens(complete)
  }
  draft.value = remainder
}

function onPaste(event: ClipboardEvent) {
  const text = event.clipboardData?.getData('text') ?? ''
  if (!text.includes(',')) {
    return
  }
  event.preventDefault()
  addTokens(text)
  draft.value = ''
}

function removeIngredient(value: string) {
  ingredients.value = ingredients.value.filter((item) => item !== value)
}

function seedFromQuery() {
  const raw = route.query.ingredients
  if (typeof raw !== 'string' || !raw.trim()) {
    return
  }
  addTokens(raw)
}

async function onCopyPrompt() {
  if (!canCopy.value) {
    return
  }
  const prompt = buildRecipePrompt(
    ingredients.value.map((name) => ({ name })),
    dish.value,
  )
  try {
    await navigator.clipboard.writeText(prompt)
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
  seedFromQuery()
})
</script>

<template>
  <section class="dash-panel quick-recipe" aria-labelledby="quick-recipe-title">
    <h1 id="quick-recipe-title" class="dash-panel__title">Quick Recipe</h1>
    <p class="dash-panel__desc">
      Type ingredients as CSV — press comma or Enter to add each one, then copy a
      prompt for an AI cook.
    </p>

    <label class="field quick-recipe__field">
      <span>Ingredients</span>
      <input
        v-model="draft"
        type="text"
        autocomplete="off"
        placeholder="chicken, rice, tomatoes…"
        @keydown="onKeydown"
        @input="onInput"
        @paste="onPaste"
        @blur="commitDraft"
      />
    </label>

    <div class="ingredient-list quick-recipe__badges" aria-live="polite">
      <button
        v-for="item in ingredients"
        :key="item"
        type="button"
        :aria-label="`Remove ${item}`"
        @click="removeIngredient(item)"
      >
        {{ item }} <span aria-hidden="true">×</span>
      </button>
      <p v-if="ingredients.length === 0" class="quick-recipe__empty">
        No ingredients yet
      </p>
    </div>

    <label class="field quick-recipe__field">
      <span>Dish (optional)</span>
      <input
        v-model="dish"
        type="text"
        autocomplete="off"
        placeholder="e.g. chicken curry"
        maxlength="120"
      />
    </label>

    <div class="quick-recipe__actions">
      <button class="btn-primary" type="button" disabled title="Coming soon">
        Generate recipe — Coming soon
      </button>
      <button
        class="btn-ghost"
        type="button"
        :disabled="!canCopy"
        @click="onCopyPrompt"
      >
        Copy prompt
      </button>
    </div>
    <p v-if="copyStatus" class="quick-recipe__copy-status" role="status">
      {{ copyStatus }}
    </p>
  </section>
</template>
