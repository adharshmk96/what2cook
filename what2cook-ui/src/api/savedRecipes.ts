import { apiRequest } from './client'

export type RecipeIngredient = {
  id: string
  recipe_id: string
  name: string
  quantity: string
  position: number
}

export type StepIngredientRef = {
  name: string
  quantity: string
}

export type RecipeStep = {
  id: string
  recipe_id: string
  position: number
  instruction: string
  ingredients_used: StepIngredientRef[]
}

export type Nutrition = {
  calories?: number | null
  protein_g?: number | null
  carbs_g?: number | null
  fat_g?: number | null
  fiber_g?: number | null
  sugar_g?: number | null
  sodium_mg?: number | null
}

export type SavedRecipe = {
  id: string
  user_id: string
  title: string
  summary: string
  minutes: number
  servings: number
  ingredients: RecipeIngredient[]
  steps: RecipeStep[]
  nutrition: Nutrition | null
  created_at: string
  updated_at: string
}

export type IngredientInput = {
  name: string
  quantity: string
}

export type StepInput = {
  instruction: string
  ingredients_used: StepIngredientRef[]
}

export type SaveRecipeInput = {
  title: string
  summary: string
  minutes: number
  servings: number
  ingredients: IngredientInput[]
  steps: StepInput[]
  nutrition: Nutrition | null
}

type ListRecipesResponse = { recipes: SavedRecipe[] }

export async function listSavedRecipes(): Promise<SavedRecipe[]> {
  const data = await apiRequest<ListRecipesResponse>('/recipes')
  if (!data || !Array.isArray(data.recipes)) {
    throw new Error('Unexpected recipes response from server')
  }
  return data.recipes
}

export async function getSavedRecipe(id: string): Promise<SavedRecipe> {
  return apiRequest<SavedRecipe>(`/recipes/${id}`)
}

export async function createSavedRecipe(input: SaveRecipeInput): Promise<SavedRecipe> {
  return apiRequest<SavedRecipe>('/recipes', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export async function updateSavedRecipe(
  id: string,
  input: SaveRecipeInput,
): Promise<SavedRecipe> {
  return apiRequest<SavedRecipe>(`/recipes/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  })
}

export async function deleteSavedRecipe(id: string): Promise<void> {
  await apiRequest<null>(`/recipes/${id}`, { method: 'DELETE' })
}
