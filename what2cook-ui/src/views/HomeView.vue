<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const ingredient = ref('')
const ingredients = ref(['Chicken', 'Tomatoes', 'Rice'])

const ingredientSummary = computed(() => ingredients.value.join(', '))

function addIngredient() {
  const value = ingredient.value.trim()
  if (!value || ingredients.value.some((item) => item.toLowerCase() === value.toLowerCase())) {
    ingredient.value = ''
    return
  }

  ingredients.value.push(value)
  ingredient.value = ''
}

function removeIngredient(value: string) {
  ingredients.value = ingredients.value.filter((item) => item !== value)
}

async function findRecipes() {
  await router.push({
    name: auth.isAuthenticated ? 'dashboard-quick-recipe' : 'register',
    query: { ingredients: ingredientSummary.value },
  })
}

async function onLogout() {
  await auth.logout()
  await router.replace({ name: 'login' })
}
</script>

<template>
  <div class="home-shell">
    <div class="home-atmosphere" aria-hidden="true" />
    <header class="home-header landing-header">
      <RouterLink class="brand-lockup" :to="{ name: 'home' }" aria-label="what2cook home">
        <img src="/icon.png" alt="" width="44" height="44" />
        <span>what2cook</span>
      </RouterLink>
      <nav class="home-nav" aria-label="Main navigation">
        <template v-if="auth.isAuthenticated">
          <RouterLink :to="{ name: 'dashboard' }">Dashboard</RouterLink>
          <RouterLink :to="{ name: 'dashboard-account' }">Account</RouterLink>
          <button type="button" class="btn-ghost" @click="onLogout">Sign out</button>
        </template>
        <template v-else>
          <RouterLink :to="{ name: 'login' }">Log in</RouterLink>
          <RouterLink class="nav-cta" :to="{ name: 'register' }">Start cooking</RouterLink>
        </template>
      </nav>
    </header>
    <main class="landing-main">
      <section class="hero-copy" aria-labelledby="hero-title">
        <p class="eyebrow"><span aria-hidden="true">●</span> Your pantry, reimagined</p>
        <h1 id="hero-title">Turn what you have into <em>what you’ll love.</em></h1>
        <p class="hero-lead">
          Add the ingredients already in your kitchen and discover delicious recipes you can make right now.
        </p>
        <div class="trust-row" aria-label="what2cook benefits">
          <span>Less waste</span><span>Quick ideas</span><span>No guesswork</span>
        </div>
      </section>

      <section class="pantry-card" aria-labelledby="pantry-title">
        <div class="pantry-card__heading">
          <div>
            <p class="step-label">Step 1</p>
            <h2 id="pantry-title">What’s in your kitchen?</h2>
          </div>
          <span class="ingredient-count">{{ ingredients.length }} items</span>
        </div>
        <form class="ingredient-form" @submit.prevent="addIngredient">
          <label class="sr-only" for="ingredient">Add an ingredient</label>
          <input
            id="ingredient"
            v-model="ingredient"
            type="text"
            autocomplete="off"
            placeholder="Try eggs, spinach, pasta…"
          />
          <button type="submit" aria-label="Add ingredient">+</button>
        </form>
        <div class="ingredient-list" aria-live="polite">
          <button
            v-for="item in ingredients"
            :key="item"
            type="button"
            :aria-label="`Remove ${item}`"
            @click="removeIngredient(item)"
          >
            {{ item }} <span aria-hidden="true">×</span>
          </button>
        </div>
        <button class="find-button" type="button" :disabled="ingredients.length === 0" @click="findRecipes">
          Find recipes <span aria-hidden="true">→</span>
        </button>
        <p class="card-note">Free to start · Add or remove ingredients anytime</p>
      </section>

      <section class="how-it-works" aria-labelledby="how-title">
        <p class="step-label">Simple by design</p>
        <h2 id="how-title">Dinner in three easy steps</h2>
        <ol>
          <li><strong>1</strong><span><b>List your ingredients</b>Use what’s already at home.</span></li>
          <li><strong>2</strong><span><b>Explore your matches</b>See recipes built around your pantry.</span></li>
          <li><strong>3</strong><span><b>Cook with confidence</b>Follow a clear, simple recipe.</span></li>
        </ol>
      </section>
    </main>
    <footer class="landing-footer">Make more. Waste less. Eat well.</footer>
  </div>
</template>
