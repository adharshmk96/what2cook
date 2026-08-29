package recipe

import "github.com/google/uuid"

// SavedService applies saved-recipe business rules over the repository.
type SavedService struct{ repo *SavedRepository }

// NewSavedService creates a saved recipe service.
func NewSavedService(repo *SavedRepository) *SavedService { return &SavedService{repo: repo} }

// List returns all recipes owned by a user.
func (s *SavedService) List(userID uuid.UUID) ([]SavedRecipe, error) {
	return s.repo.ListByUser(userID)
}

// Get returns one recipe owned by a user.
func (s *SavedService) Get(userID, id uuid.UUID) (*SavedRecipe, error) {
	return s.repo.FindByIDForUser(id, userID)
}

// Create saves a new recipe for a user.
func (s *SavedService) Create(userID uuid.UUID, req *SaveRecipeRequest) (*SavedRecipe, error) {
	rec := buildRecipe(userID, uuid.New(), req)
	if err := s.repo.Create(rec); err != nil {
		return nil, err
	}
	return rec, nil
}

// Update overwrites an existing recipe owned by a user.
func (s *SavedService) Update(userID, id uuid.UUID, req *SaveRecipeRequest) (*SavedRecipe, error) {
	rec := buildRecipe(userID, id, req)
	if err := s.repo.Replace(rec); err != nil {
		return nil, err
	}
	return rec, nil
}

// Delete removes a recipe owned by a user.
func (s *SavedService) Delete(userID, id uuid.UUID) error {
	return s.repo.Delete(id, userID)
}

func buildRecipe(userID, id uuid.UUID, req *SaveRecipeRequest) *SavedRecipe {
	ingredients := make([]RecipeIngredient, len(req.Ingredients))
	for i, in := range req.Ingredients {
		ingredients[i] = RecipeIngredient{Name: in.Name, Quantity: in.Quantity, Position: i}
	}
	steps := make([]RecipeStep, len(req.Steps))
	for i, in := range req.Steps {
		steps[i] = RecipeStep{Position: i, Instruction: in.Instruction, IngredientsUsed: in.IngredientsUsed}
	}
	return &SavedRecipe{
		ID:          id,
		UserID:      userID,
		Title:       req.Title,
		Summary:     req.Summary,
		Minutes:     req.Minutes,
		Servings:    req.Servings,
		Ingredients: ingredients,
		Steps:       steps,
		Nutrition:   req.Nutrition,
	}
}
