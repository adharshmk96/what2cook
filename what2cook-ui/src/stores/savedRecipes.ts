import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as savedRecipesApi from '../api/savedRecipes'
import type { SavedRecipe, SaveRecipeInput } from '../api/savedRecipes'
import { ApiError } from '../api/client'

export const useSavedRecipesStore = defineStore('savedRecipes', () => {
  const recipes = ref<SavedRecipe[]>([])
  const loading = ref(false)
  const saving = ref(false)
  const error = ref<unknown>(null)

  function setError(err: unknown, fallback: string) {
    error.value = err instanceof ApiError || err instanceof Error ? err : new Error(fallback)
  }

  async function load() {
    loading.value = true
    error.value = null
    try {
      recipes.value = await savedRecipesApi.listSavedRecipes()
      return recipes.value
    } catch (err) {
      console.warn('Load saved recipes failed', err)
      recipes.value = []
      setError(err, 'Could not load saved recipes')
      return null
    } finally {
      loading.value = false
    }
  }

  async function create(input: SaveRecipeInput) {
    saving.value = true
    error.value = null
    try {
      const recipe = await savedRecipesApi.createSavedRecipe(input)
      recipes.value = [recipe, ...recipes.value]
      return recipe
    } catch (err) {
      console.warn('Create saved recipe failed', err)
      setError(err, 'Could not save recipe')
      return null
    } finally {
      saving.value = false
    }
  }

  async function update(id: string, input: SaveRecipeInput) {
    saving.value = true
    error.value = null
    try {
      const recipe = await savedRecipesApi.updateSavedRecipe(id, input)
      recipes.value = recipes.value.map((r) => (r.id === id ? recipe : r))
      return recipe
    } catch (err) {
      console.warn('Update saved recipe failed', err)
      setError(err, 'Could not update recipe')
      return null
    } finally {
      saving.value = false
    }
  }

  async function remove(id: string) {
    saving.value = true
    error.value = null
    try {
      await savedRecipesApi.deleteSavedRecipe(id)
      recipes.value = recipes.value.filter((r) => r.id !== id)
      return true
    } catch (err) {
      console.warn('Delete saved recipe failed', err)
      setError(err, 'Could not delete recipe')
      return false
    } finally {
      saving.value = false
    }
  }

  function clearError() {
    error.value = null
  }

  return { recipes, loading, saving, error, load, create, update, remove, clearError }
})

export type { SavedRecipe }
