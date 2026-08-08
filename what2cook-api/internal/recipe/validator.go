package recipe

import "strings"

const (
	minIngredients = 1
	maxIngredients = 30
	maxNameLen    = 80
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
