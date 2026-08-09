export type PromptIngredient = {
  name: string
  quantity?: string | null
}

export type PromptMode = 'inventory' | 'quick'

function formatIngredientLine(item: PromptIngredient): string {
  const name = item.name.trim()
  const quantity = item.quantity?.trim() ?? ''
  if (quantity) {
    return `- ${name}: ${quantity}`
  }
  return `- ${name}: `
}

const RESPONSE_FORMAT = [
  'Response format for each recipe:',
  '',
  'Ingredients:',
  '- List every ingredient with quantity in g, tbsp, or ml',
  '',
  'Steps:',
  'For each step include:',
  '- Title',
  '- Time (minutes)',
  '- Ingredients used (with quantity in g, tbsp, or ml)',
  '- What to do',
].join('\n')

/**
 * Builds the shared AI prompt for Inventory and Quick Recipe copy actions.
 */
export function buildRecipePrompt(
  ingredients: PromptIngredient[],
  dish?: string | null,
  mode: PromptMode = 'quick',
): string {
  const lines = ingredients
    .map((item) => ({
      name: item.name.trim(),
      quantity: item.quantity?.trim() || '',
    }))
    .filter((item) => item.name)

  const ingredientBlock = lines.map(formatIngredientLine).join('\n')
  const dishName = dish?.trim() ?? ''
  const inventoryLead =
    mode === 'inventory'
      ? ['This is my full inventory. Select necessary items only.', '']
      : []

  if (dishName) {
    return [
      ...inventoryLead,
      `Generate a recipe for "${dishName}" using these ingredients:`,
      ingredientBlock,
      '',
      RESPONSE_FORMAT,
    ].join('\n')
  }

  return [
    ...inventoryLead,
    'Generate a list of recipes I can cook with these ingredients:',
    ingredientBlock,
    '',
    'For each recipe include a Title, then follow this format:',
    '',
    RESPONSE_FORMAT,
  ].join('\n')
}
