<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useRecipesStore } from '../../stores/recipes'

const router = useRouter()
const recipes = useRecipesStore()

onMounted(() => {
  if (!recipes.hasResults()) {
    void router.replace({ name: 'dashboard-quick-recipe' })
  }
})
</script>

<template>
  <section
    v-if="recipes.hasResults()"
    class="dash-panel recipe-results"
    aria-labelledby="recipe-results-title"
  >
    <h1 id="recipe-results-title" class="dash-panel__title">Your recipes</h1>
    <p class="dash-panel__desc">
      Placeholder dishes for
      <strong>{{ recipes.ingredients.join(', ') }}</strong>.
    </p>

    <ul class="recipe-results__list">
      <li v-for="recipe in recipes.recipes" :key="recipe.id" class="recipe-card">
        <div class="recipe-card__meta">
          <span class="recipe-card__minutes">{{ recipe.minutes }} min</span>
        </div>
        <h2 class="recipe-card__title">{{ recipe.title }}</h2>
        <p class="recipe-card__summary">{{ recipe.summary }}</p>
        <p class="recipe-card__ingredients">
          Uses: {{ recipe.ingredients_used.join(', ') }}
        </p>
      </li>
    </ul>

    <div class="recipe-results__actions">
      <RouterLink class="btn-primary" :to="{ name: 'dashboard-quick-recipe' }">
        Try again
      </RouterLink>
    </div>
  </section>
</template>
