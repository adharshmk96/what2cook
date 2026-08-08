package recipe

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

const simulateAIDelay = 1500 * time.Millisecond

// Service generates placeholder recipes (simulated AI).
type Service struct{}

// NewService creates a recipe service.
func NewService() *Service {
	return &Service{}
}

// Generate simulates an AI call and returns placeholder recipes.
func (s *Service) Generate(ingredients []string) (*GenerateResponse, error) {
	log.Printf("recipe: simulating AI generate for %d ingredients", len(ingredients))
	time.Sleep(simulateAIDelay)

	joined := strings.Join(ingredients, ", ")
	primary := ingredients[0]
	secondary := primary
	if len(ingredients) > 1 {
		secondary = ingredients[1]
	}

	recipes := []Recipe{
		{
			ID:              uuid.NewString(),
			Title:           fmt.Sprintf("%s & %s Skillet", titleCase(primary), titleCase(secondary)),
			Summary:         fmt.Sprintf("A quick one-pan placeholder dish using %s.", joined),
			Minutes:         25,
			IngredientsUsed: ingredients,
		},
		{
			ID:              uuid.NewString(),
			Title:           fmt.Sprintf("Roasted %s Bowl", titleCase(primary)),
			Summary:         fmt.Sprintf("Oven-roasted placeholder recipe starring %s with %s.", primary, joined),
			Minutes:         35,
			IngredientsUsed: ingredients,
		},
		{
			ID:              uuid.NewString(),
			Title:           fmt.Sprintf("%s Soup of the Day", titleCase(primary)),
			Summary:         fmt.Sprintf("A cozy simmered soup built from %s. Placeholder only — no real AI yet.", joined),
			Minutes:         40,
			IngredientsUsed: ingredients,
		},
	}

	return &GenerateResponse{
		Ingredients: ingredients,
		Recipes:     recipes,
	}, nil
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(strings.ToLower(s))
	runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
	return string(runes)
}
