package recipe

// GenerateRequest is the JSON body for POST /recipes/generate.
type GenerateRequest struct {
	Ingredients []string `json:"ingredients"`
}

// Recipe is a placeholder recipe returned by the generate endpoint.
type Recipe struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Summary          string   `json:"summary"`
	Minutes          int      `json:"minutes"`
	IngredientsUsed  []string `json:"ingredients_used"`
}

// GenerateResponse is returned after simulated recipe generation.
type GenerateResponse struct {
	Ingredients []string `json:"ingredients"`
	Recipes     []Recipe `json:"recipes"`
}
