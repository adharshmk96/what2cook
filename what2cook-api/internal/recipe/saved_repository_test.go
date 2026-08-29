package recipe

import (
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&SavedRecipe{}, &RecipeIngredient{}, &RecipeStep{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return db
}

func TestSavedRecipeCRUD(t *testing.T) {
	db := newTestDB(t)
	repo := NewSavedRepository(db)
	userID := uuid.New()

	calories := 420.5
	req := &SaveRecipeRequest{
		Title:    "Chicken Skillet",
		Summary:  "A quick weeknight dinner.",
		Minutes:  25,
		Servings: 2,
		Ingredients: []IngredientInput{
			{Name: "Chicken breast", Quantity: "500g"},
			{Name: "Garlic", Quantity: "2 cloves"},
		},
		Steps: []StepInput{
			{
				Instruction: "Sear the chicken until golden.",
				IngredientsUsed: []StepIngredient{
					{Name: "Chicken breast", Quantity: "500g"},
				},
			},
			{
				Instruction: "Add garlic and simmer.",
				IngredientsUsed: []StepIngredient{
					{Name: "Garlic", Quantity: "2 cloves"},
				},
			},
		},
		Nutrition: &Nutrition{Calories: &calories},
	}
	if msg := ValidateSaveRecipe(req); msg != "" {
		t.Fatalf("validate: %s", msg)
	}

	rec := buildRecipe(userID, uuid.New(), req)
	if err := repo.Create(rec); err != nil {
		t.Fatalf("create: %v", err)
	}

	fetched, err := repo.FindByIDForUser(rec.ID, userID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if len(fetched.Ingredients) != 2 || len(fetched.Steps) != 2 {
		t.Fatalf("expected 2 ingredients and 2 steps, got %d/%d", len(fetched.Ingredients), len(fetched.Steps))
	}
	if fetched.Nutrition == nil || fetched.Nutrition.Calories == nil || *fetched.Nutrition.Calories != calories {
		t.Fatalf("nutrition not round-tripped: %+v", fetched.Nutrition)
	}
	if len(fetched.Steps[0].IngredientsUsed) != 1 || fetched.Steps[0].IngredientsUsed[0].Name != "Chicken breast" {
		t.Fatalf("step ingredients not round-tripped: %+v", fetched.Steps[0].IngredientsUsed)
	}

	// Replace with fewer ingredients/steps and no nutrition.
	updateReq := &SaveRecipeRequest{
		Title:    "Chicken Skillet (updated)",
		Summary:  "Still quick.",
		Minutes:  20,
		Servings: 2,
		Ingredients: []IngredientInput{
			{Name: "Chicken breast", Quantity: "500g"},
		},
		Steps: []StepInput{
			{Instruction: "Sear and serve.", IngredientsUsed: nil},
		},
	}
	if msg := ValidateSaveRecipe(updateReq); msg != "" {
		t.Fatalf("validate update: %s", msg)
	}
	updated := buildRecipe(userID, rec.ID, updateReq)
	if err := repo.Replace(updated); err != nil {
		t.Fatalf("replace: %v", err)
	}

	fetched, err = repo.FindByIDForUser(rec.ID, userID)
	if err != nil {
		t.Fatalf("find after update: %v", err)
	}
	if fetched.Title != "Chicken Skillet (updated)" {
		t.Fatalf("title not updated: %s", fetched.Title)
	}
	if len(fetched.Ingredients) != 1 || len(fetched.Steps) != 1 {
		t.Fatalf("expected 1 ingredient and 1 step after update, got %d/%d", len(fetched.Ingredients), len(fetched.Steps))
	}
	if fetched.Nutrition != nil {
		t.Fatalf("expected nutrition cleared, got %+v", fetched.Nutrition)
	}

	list, err := repo.ListByUser(userID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 recipe, got %d", len(list))
	}

	if err := repo.Delete(rec.ID, userID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := repo.FindByIDForUser(rec.ID, userID); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}
