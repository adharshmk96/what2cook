package recipe

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SavedRecipe is a user-authored recipe with structured ingredients, steps,
// and optional nutritional info.
type SavedRecipe struct {
	ID          uuid.UUID          `gorm:"type:text;primaryKey" json:"id"`
	UserID      uuid.UUID          `gorm:"type:text;index;not null" json:"user_id"`
	Title       string             `gorm:"size:120;not null" json:"title"`
	Summary     string             `gorm:"size:500" json:"summary"`
	Minutes     int                `gorm:"not null;default:0" json:"minutes"`
	Servings    int                `gorm:"not null;default:0" json:"servings"`
	Ingredients []RecipeIngredient `gorm:"foreignKey:RecipeID;constraint:OnDelete:CASCADE" json:"ingredients"`
	Steps       []RecipeStep       `gorm:"foreignKey:RecipeID;constraint:OnDelete:CASCADE" json:"steps"`
	Nutrition   *Nutrition         `gorm:"serializer:json" json:"nutrition"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// BeforeCreate assigns a UUID if missing.
func (r *SavedRecipe) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// RecipeIngredient is one ingredient + quantity in a recipe's ingredient list.
type RecipeIngredient struct {
	ID       uuid.UUID `gorm:"type:text;primaryKey" json:"id"`
	RecipeID uuid.UUID `gorm:"type:text;index;not null" json:"recipe_id"`
	Name     string    `gorm:"size:80;not null" json:"name"`
	Quantity string    `gorm:"size:40" json:"quantity"`
	Position int       `gorm:"not null;default:0" json:"position"`
}

// BeforeCreate assigns a UUID if missing.
func (i *RecipeIngredient) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

// RecipeStep is one instruction step, naming the ingredients (with
// quantities) it uses.
type RecipeStep struct {
	ID              uuid.UUID        `gorm:"type:text;primaryKey" json:"id"`
	RecipeID        uuid.UUID        `gorm:"type:text;index;not null" json:"recipe_id"`
	Position        int              `gorm:"not null;default:0" json:"position"`
	Instruction     string           `gorm:"size:1000;not null" json:"instruction"`
	IngredientsUsed []StepIngredient `gorm:"serializer:json" json:"ingredients_used"`
}

// BeforeCreate assigns a UUID if missing.
func (s *RecipeStep) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// StepIngredient names an ingredient + quantity used within a single step.
type StepIngredient struct {
	Name     string `json:"name"`
	Quantity string `json:"quantity"`
}

// Nutrition is optional per-serving nutritional info.
type Nutrition struct {
	Calories *float64 `json:"calories,omitempty"`
	ProteinG *float64 `json:"protein_g,omitempty"`
	CarbsG   *float64 `json:"carbs_g,omitempty"`
	FatG     *float64 `json:"fat_g,omitempty"`
	FiberG   *float64 `json:"fiber_g,omitempty"`
	SugarG   *float64 `json:"sugar_g,omitempty"`
	SodiumMg *float64 `json:"sodium_mg,omitempty"`
}

// IngredientInput is the request payload for one recipe ingredient.
type IngredientInput struct {
	Name     string `json:"name"`
	Quantity string `json:"quantity"`
}

// StepInput is the request payload for one recipe step.
type StepInput struct {
	Instruction     string           `json:"instruction"`
	IngredientsUsed []StepIngredient `json:"ingredients_used"`
}

// SaveRecipeRequest is the JSON body for creating or updating a saved recipe.
type SaveRecipeRequest struct {
	Title       string            `json:"title"`
	Summary     string            `json:"summary"`
	Minutes     int               `json:"minutes"`
	Servings    int               `json:"servings"`
	Ingredients []IngredientInput `json:"ingredients"`
	Steps       []StepInput       `json:"steps"`
	Nutrition   *Nutrition        `json:"nutrition"`
}
