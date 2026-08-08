import { apiRequest } from './client'

export type Recipe = {
  id: string
  title: string
  summary: string
  minutes: number
  ingredients_used: string[]
}

export type GenerateRecipesResponse = {
  ingredients: string[]
  recipes: Recipe[]
}

export async function generateRecipes(
  ingredients: string[],
): Promise<GenerateRecipesResponse> {
  const data = await apiRequest<GenerateRecipesResponse>('/recipes/generate', {
    method: 'POST',
    body: JSON.stringify({ ingredients }),
  })

  if (!data || !Array.isArray(data.recipes)) {
    throw new Error('Unexpected recipe response from server')
  }

  return {
    ingredients: Array.isArray(data.ingredients) ? data.ingredients : ingredients,
    recipes: data.recipes,
  }
}
