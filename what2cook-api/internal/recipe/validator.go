package recipe

import (
	"fmt"
	"strings"
)

const (
	minIngredients = 1
	maxIngredients = 30
	maxNameLen     = 80
)

const (
	maxTitleLen          = 120
	maxSummaryLen        = 500
	maxRecipeIngredients = 50
	maxRecipeSteps       = 50
	maxQuantityLen       = 40
	maxInstructionLen    = 1000
)

// NormalizeIngredients trims, drops empties, and dedupes case-insensitively.
func NormalizeIngredients(raw []string) []string {
	seen := make(map[string]struct{}, len(raw))
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		name := strings.TrimSpace(item)
		if name == "" {
			continue
		}
		if len(name) > maxNameLen {
			name = name[:maxNameLen]
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, name)
	}
	return out
}

// ValidateGenerate returns an error message, or empty string if valid.
func ValidateGenerate(req *GenerateRequest) string {
	if req == nil {
		return "ingredients are required"
	}
	normalized := NormalizeIngredients(req.Ingredients)
	if len(normalized) < minIngredients {
		return "at least one ingredient is required"
	}
	if len(normalized) > maxIngredients {
		return "too many ingredients (max 30)"
	}
	req.Ingredients = normalized
	return ""
}

// ValidateSaveRecipe trims/normalizes a save-recipe request in place and
// returns an error message, or empty string if valid.
func ValidateSaveRecipe(req *SaveRecipeRequest) string {
	if req == nil {
		return "title is required"
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		return "title is required"
	}
	req.Title = truncate(req.Title, maxTitleLen)

	req.Summary = strings.TrimSpace(req.Summary)
	req.Summary = truncate(req.Summary, maxSummaryLen)

	if req.Minutes < 0 {
		return "minutes cannot be negative"
	}
	if req.Servings < 0 {
		return "servings cannot be negative"
	}

	if len(req.Ingredients) == 0 {
		return "at least one ingredient is required"
	}
	if len(req.Ingredients) > maxRecipeIngredients {
		return fmt.Sprintf("too many ingredients (max %d)", maxRecipeIngredients)
	}
	for i := range req.Ingredients {
		req.Ingredients[i].Name = strings.TrimSpace(req.Ingredients[i].Name)
		if req.Ingredients[i].Name == "" {
			return "ingredient name is required"
		}
		req.Ingredients[i].Name = truncate(req.Ingredients[i].Name, maxNameLen)
		req.Ingredients[i].Quantity = strings.TrimSpace(req.Ingredients[i].Quantity)
		req.Ingredients[i].Quantity = truncate(req.Ingredients[i].Quantity, maxQuantityLen)
	}

	if len(req.Steps) == 0 {
		return "at least one step is required"
	}
	if len(req.Steps) > maxRecipeSteps {
		return fmt.Sprintf("too many steps (max %d)", maxRecipeSteps)
	}
	for i := range req.Steps {
		req.Steps[i].Instruction = strings.TrimSpace(req.Steps[i].Instruction)
		if req.Steps[i].Instruction == "" {
			return "step instruction is required"
		}
		req.Steps[i].Instruction = truncate(req.Steps[i].Instruction, maxInstructionLen)
		for j := range req.Steps[i].IngredientsUsed {
			req.Steps[i].IngredientsUsed[j].Name = strings.TrimSpace(req.Steps[i].IngredientsUsed[j].Name)
			req.Steps[i].IngredientsUsed[j].Name = truncate(req.Steps[i].IngredientsUsed[j].Name, maxNameLen)
			req.Steps[i].IngredientsUsed[j].Quantity = strings.TrimSpace(req.Steps[i].IngredientsUsed[j].Quantity)
			req.Steps[i].IngredientsUsed[j].Quantity = truncate(req.Steps[i].IngredientsUsed[j].Quantity, maxQuantityLen)
		}
	}

	if req.Nutrition != nil {
		if msg := validateNutrition(req.Nutrition); msg != "" {
			return msg
		}
		if isEmptyNutrition(req.Nutrition) {
			req.Nutrition = nil
		}
	}

	return ""
}

// truncate cuts s to at most max runes, never splitting a multi-byte rune.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max])
}

func isEmptyNutrition(n *Nutrition) bool {
	return n.Calories == nil && n.ProteinG == nil && n.CarbsG == nil &&
		n.FatG == nil && n.FiberG == nil && n.SugarG == nil && n.SodiumMg == nil
}

func validateNutrition(n *Nutrition) string {
	fields := map[string]*float64{
		"calories": n.Calories,
		"protein":  n.ProteinG,
		"carbs":    n.CarbsG,
		"fat":      n.FatG,
		"fiber":    n.FiberG,
		"sugar":    n.SugarG,
		"sodium":   n.SodiumMg,
	}
	for name, value := range fields {
		if value != nil && *value < 0 {
			return fmt.Sprintf("nutrition %s cannot be negative", name)
		}
	}
	return ""
}
