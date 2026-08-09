package inventory

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("not found")

// Repository persists inventory entities.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates an inventory repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CreateInventory inserts an inventory.
func (r *Repository) CreateInventory(inv *Inventory) error {
	if err := r.db.Create(inv).Error; err != nil {
		return fmt.Errorf("create inventory: %w", err)
	}
	return nil
}

// ListByUser returns inventories for a user (without items).
func (r *Repository) ListByUser(userID uuid.UUID) ([]Inventory, error) {
	var list []Inventory
	err := r.db.Where("user_id = ?", userID).Order("is_default DESC, created_at ASC").Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("list inventories: %w", err)
	}
	return list, nil
}

// FindDefaultByUser returns the default inventory for a user, if any.
func (r *Repository) FindDefaultByUser(userID uuid.UUID) (*Inventory, error) {
	var inv Inventory
	err := r.db.Where("user_id = ? AND is_default = ?", userID, true).First(&inv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find default inventory: %w", err)
	}
	return &inv, nil
}

// FindByIDForUser loads an inventory owned by userID.
func (r *Repository) FindByIDForUser(id, userID uuid.UUID) (*Inventory, error) {
	var inv Inventory
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&inv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find inventory: %w", err)
	}
	return &inv, nil
}

// FindByIDForUserWithItems loads an inventory with its items.
func (r *Repository) FindByIDForUserWithItems(id, userID uuid.UUID) (*Inventory, error) {
	var inv Inventory
	err := r.db.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).Where("id = ? AND user_id = ?", id, userID).First(&inv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find inventory with items: %w", err)
	}
	return &inv, nil
}

// UpdateInventoryName updates the name of an inventory owned by userID.
func (r *Repository) UpdateInventoryName(id, userID uuid.UUID, name string) error {
	res := r.db.Model(&Inventory{}).Where("id = ? AND user_id = ?", id, userID).Update("name", name)
	if res.Error != nil {
		return fmt.Errorf("update inventory: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteInventory deletes a non-default inventory owned by userID (and its items).
func (r *Repository) DeleteInventory(id, userID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var inv Inventory
		err := tx.Where("id = ? AND user_id = ? AND is_default = ?", id, userID, false).First(&inv).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("find inventory for delete: %w", err)
		}
		if err := tx.Where("inventory_id = ?", id).Delete(&InventoryItem{}).Error; err != nil {
			return fmt.Errorf("delete inventory items: %w", err)
		}
		res := tx.Delete(&inv)
		if res.Error != nil {
			return fmt.Errorf("delete inventory: %w", res.Error)
		}
		return nil
	})
}

// CreateItem inserts an inventory item.
func (r *Repository) CreateItem(item *InventoryItem) error {
	if err := r.db.Create(item).Error; err != nil {
		return fmt.Errorf("create item: %w", err)
	}
	return nil
}

// FindItemForUser loads an item that belongs to an inventory owned by userID.
func (r *Repository) FindItemForUser(inventoryID, itemID, userID uuid.UUID) (*InventoryItem, error) {
	var item InventoryItem
	err := r.db.
		Joins("JOIN inventories ON inventories.id = inventory_items.inventory_id").
		Where("inventory_items.id = ? AND inventory_items.inventory_id = ? AND inventories.user_id = ?", itemID, inventoryID, userID).
		First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find item: %w", err)
	}
	return &item, nil
}

// UpdateItem persists name and quantity for an item.
func (r *Repository) UpdateItem(item *InventoryItem) error {
	updates := map[string]interface{}{
		"name":     item.Name,
		"quantity": item.Quantity,
	}
	if err := r.db.Model(item).Updates(updates).Error; err != nil {
		return fmt.Errorf("update item: %w", err)
	}
	return nil
}

// DeleteItem deletes an item belonging to an inventory owned by userID.
func (r *Repository) DeleteItem(inventoryID, itemID, userID uuid.UUID) error {
	sub := r.db.Model(&Inventory{}).Select("id").Where("id = ? AND user_id = ?", inventoryID, userID)
	res := r.db.Where("id = ? AND inventory_id IN (?)", itemID, sub).Delete(&InventoryItem{})
	if res.Error != nil {
		return fmt.Errorf("delete item: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
