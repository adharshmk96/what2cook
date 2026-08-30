package recipe

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("not found")

// SavedRepository persists SavedRecipe rows.
type SavedRepository struct{ db *gorm.DB }

// NewSavedRepository creates a saved recipe repository.
func NewSavedRepository(db *gorm.DB) *SavedRepository { return &SavedRepository{db: db} }

func (r *SavedRepository) preload(tx *gorm.DB) *gorm.DB {
	return tx.Preload("Ingredients", func(db *gorm.DB) *gorm.DB {
		return db.Order("position ASC")
	}).Preload("Steps", func(db *gorm.DB) *gorm.DB {
		return db.Order("position ASC")
	})
}

// Create inserts a new recipe with its ingredients and steps.
func (r *SavedRepository) Create(recipe *SavedRecipe) error {
	if err := r.db.Create(recipe).Error; err != nil {
		return fmt.Errorf("create recipe: %w", err)
	}
	return nil
}

// ListByUser returns all recipes owned by a user, newest first.
func (r *SavedRepository) ListByUser(userID uuid.UUID) ([]SavedRecipe, error) {
	var list []SavedRecipe
	err := r.preload(r.db).Where("user_id = ?", userID).Order("created_at DESC").Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("list recipes: %w", err)
	}
	return list, nil
}

// FindByIDForUser returns one recipe owned by a user.
func (r *SavedRepository) FindByIDForUser(id, userID uuid.UUID) (*SavedRecipe, error) {
	var rec SavedRecipe
	err := r.preload(r.db).Where("id = ? AND user_id = ?", id, userID).First(&rec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find recipe: %w", err)
	}
	return &rec, nil
}

// Replace overwrites a recipe's fields, ingredients, and steps in a transaction.
func (r *SavedRepository) Replace(recipe *SavedRecipe) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var existing SavedRecipe
		err := tx.Where("id = ? AND user_id = ?", recipe.ID, recipe.UserID).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("find recipe for update: %w", err)
		}

		if err := tx.Where("recipe_id = ?", recipe.ID).Delete(&RecipeIngredient{}).Error; err != nil {
			return fmt.Errorf("delete ingredients: %w", err)
		}
		if err := tx.Where("recipe_id = ?", recipe.ID).Delete(&RecipeStep{}).Error; err != nil {
			return fmt.Errorf("delete steps: %w", err)
		}

		fields := []string{"Title", "Summary", "Minutes", "Servings", "Nutrition"}
		if err := tx.Model(&SavedRecipe{}).Where("id = ?", recipe.ID).Select(fields).Updates(recipe).Error; err != nil {
			return fmt.Errorf("update recipe: %w", err)
		}

		for i := range recipe.Ingredients {
			recipe.Ingredients[i].RecipeID = recipe.ID
			if err := tx.Create(&recipe.Ingredients[i]).Error; err != nil {
				return fmt.Errorf("create ingredient: %w", err)
			}
		}
		for i := range recipe.Steps {
			recipe.Steps[i].RecipeID = recipe.ID
			if err := tx.Create(&recipe.Steps[i]).Error; err != nil {
				return fmt.Errorf("create step: %w", err)
			}
		}
		return nil
	})
}

// Delete removes a recipe owned by a user (ingredients/steps cascade).
func (r *SavedRepository) Delete(id, userID uuid.UUID) error {
	res := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&SavedRecipe{})
	if res.Error != nil {
		return fmt.Errorf("delete recipe: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
