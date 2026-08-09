export type PromptIngredient = {
  name: string
  quantity?: string | null
}

function formatIngredientLine(item: PromptIngredient): string {
  const name = item.name.trim()
  const quantity = item.quantity?.trim() ?? ''
  if (quantity) {
    return `- ${name}: ${quantity}`
  }
  return `- ${name}: `
}

/**
 * Builds the shared AI prompt for Inventory and Quick Recipe copy actions.
 */
export function buildRecipePrompt(
  ingredients: PromptIngredient[],
  dish?: string | null,
): string {
  const lines = ingredients
    .map((item) => ({
      name: item.name.trim(),
      quantity: item.quantity?.trim() || '',
    }))
    .filter((item) => item.name)

  const ingredientBlock = lines.map(formatIngredientLine).join('\n')
  const dishName = dish?.trim() ?? ''

  if (dishName) {
    return [
      `Generate a recipe for "${dishName}" using these ingredients:`,
      ingredientBlock,
      '',
      'Include:',
      '- A full ingredients list with quantities',
      '- Step-by-step instructions',
      '- For each step: estimated minutes and which ingredients are used',
    ].join('\n')
  }

  return [
    'Generate a list of recipes I can cook with these ingredients:',
    ingredientBlock,
    '',
    'For each recipe include:',
    '- Title',
    '- Ingredients list with quantities',
    '- Step-by-step instructions',
    '- For each step: estimated minutes and which ingredients are used',
  ].join('\n')
}
