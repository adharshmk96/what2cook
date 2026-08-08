import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as recipesApi from '../api/recipes'
import type { Recipe } from '../api/recipes'
import { ApiError } from '../api/client'

export const useRecipesStore = defineStore('recipes', () => {
  const ingredients = ref<string[]>([])
  const recipes = ref<Recipe[]>([])
  const loading = ref(false)
  const error = ref<unknown>(null)

  function clearResults() {
    recipes.value = []
    error.value = null
  }

  function hasResults() {
    return recipes.value.length > 0
  }

  async function generate(nextIngredients: string[]) {
    const cleaned = nextIngredients
      .map((item) => item.trim())
      .filter(Boolean)

    if (cleaned.length === 0) {
      error.value = new Error('Add at least one ingredient')
      return null
    }

    loading.value = true
    error.value = null
    ingredients.value = cleaned

    try {
      const result = await recipesApi.generateRecipes(cleaned)
      ingredients.value = result.ingredients
      recipes.value = result.recipes
      return result
    } catch (err) {
      console.warn('Recipe generate failed', err)
      recipes.value = []
      error.value = err instanceof ApiError || err instanceof Error
        ? err
        : new Error('Could not generate recipes')
      return null
    } finally {
      loading.value = false
    }
  }

  return {
    ingredients,
    recipes,
    loading,
    error,
    clearResults,
    hasResults,
    generate,
  }
})
