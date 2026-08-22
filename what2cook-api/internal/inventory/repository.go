package inventory

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("not found")

type Repository struct { db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) CreateInventory(inv *Inventory) error {
	if err := r.db.Create(inv).Error; err != nil { return fmt.Errorf("create inventory: %w", err) }
	return nil
}

func (r *Repository) ListByUser(userID uuid.UUID) ([]Inventory, error) {
	var list []Inventory
	if err := r.db.Where("user_id = ?", userID).Order("is_default DESC, created_at ASC").Find(&list).Error; err != nil { return nil, fmt.Errorf("list inventories: %w", err) }
	return list, nil
}

func (r *Repository) FindDefaultByUser(userID uuid.UUID) (*Inventory, error) {
	var inv Inventory
	err := r.db.Where("user_id = ? AND is_default = ?", userID, true).First(&inv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, ErrNotFound }
	if err != nil { return nil, fmt.Errorf("find default inventory: %w", err) }
	return &inv, nil
}

func (r *Repository) FindByIDForUser(id, userID uuid.UUID) (*Inventory, error) {
	var inv Inventory
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&inv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, ErrNotFound }
	if err != nil { return nil, fmt.Errorf("find inventory: %w", err) }
	return &inv, nil
}

func (r *Repository) FindByIDForUserWithItems(id, userID uuid.UUID) (*Inventory, error) {
	var inv Inventory
	err := r.db.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("category ASC, created_at ASC")
	}).Where("id = ? AND user_id = ?", id, userID).First(&inv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, ErrNotFound }
	if err != nil { return nil, fmt.Errorf("find inventory with items: %w", err) }
	return &inv, nil
}

func (r *Repository) UpdateInventoryName(id, userID uuid.UUID, name string) error {
	res := r.db.Model(&Inventory{}).Where("id = ? AND user_id = ?", id, userID).Update("name", name)
	if res.Error != nil { return fmt.Errorf("update inventory: %w", res.Error) }
	if res.RowsAffected == 0 { return ErrNotFound }
	return nil
}

func (r *Repository) DeleteInventory(id, userID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var inv Inventory
		err := tx.Where("id = ? AND user_id = ? AND is_default = ?", id, userID, false).First(&inv).Error
		if errors.Is(err, gorm.ErrRecordNotFound) { return ErrNotFound }
		if err != nil { return fmt.Errorf("find inventory for delete: %w", err) }
		if err := tx.Where("inventory_id = ?", id).Delete(&InventoryItem{}).Error; err != nil { return fmt.Errorf("delete inventory items: %w", err) }
		if err := tx.Delete(&inv).Error; err != nil { return fmt.Errorf("delete inventory: %w", err) }
		return nil
	})
}

func (r *Repository) CreateItem(item *InventoryItem) error {
	if err := r.db.Create(item).Error; err != nil { return fmt.Errorf("create item: %w", err) }
	return nil
}

func (r *Repository) FindItemForUser(inventoryID, itemID, userID uuid.UUID) (*InventoryItem, error) {
	var item InventoryItem
	err := r.db.Joins("JOIN inventories ON inventories.id = inventory_items.inventory_id").Where("inventory_items.id = ? AND inventory_items.inventory_id = ? AND inventories.user_id = ?", itemID, inventoryID, userID).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, ErrNotFound }
	if err != nil { return nil, fmt.Errorf("find item: %w", err) }
	return &item, nil
}

func (r *Repository) UpdateItem(item *InventoryItem) error {
	updates := map[string]interface{}{"name": item.Name, "quantity": item.Quantity, "category": item.Category}
	if err := r.db.Model(item).Updates(updates).Error; err != nil { return fmt.Errorf("update item: %w", err) }
	return nil
}

func (r *Repository) DeleteItem(inventoryID, itemID, userID uuid.UUID) error {
	sub := r.db.Model(&Inventory{}).Select("id").Where("id = ? AND user_id = ?", inventoryID, userID)
	res := r.db.Where("id = ? AND inventory_id IN (?)", itemID, sub).Delete(&InventoryItem{})
	if res.Error != nil { return fmt.Errorf("delete item: %w", res.Error) }
	if res.RowsAffected == 0 { return ErrNotFound }
	return nil
}

// ListByUserWithItems returns all inventories for a user with nested items.
func (r *Repository) ListByUserWithItems(userID uuid.UUID) ([]Inventory, error) {
	var list []Inventory
	err := r.db.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("category ASC, name ASC")
	}).Where("user_id = ?", userID).Order("is_default DESC, created_at ASC").Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("list inventories with items: %w", err)
	}
	return list, nil
}

// InventoryImport is a normalized inventory payload for bulk import.
type InventoryImport struct {
	Name      string
	IsDefault bool
	Items     []ItemImport
}

// ItemImport is a normalized inventory item payload for bulk import.
type ItemImport struct {
	Name     string
	Quantity *string
	Category *string
}

// ReplaceAllForUser deletes existing inventories and recreates them from imports.
func (r *Repository) ReplaceAllForUser(userID uuid.UUID, imports []InventoryImport) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var existing []Inventory
		if err := tx.Where("user_id = ?", userID).Find(&existing).Error; err != nil {
			return fmt.Errorf("list inventories for replace: %w", err)
		}

		for _, inv := range existing {
			if err := tx.Where("inventory_id = ?", inv.ID).Delete(&InventoryItem{}).Error; err != nil {
				return fmt.Errorf("delete inventory items: %w", err)
			}
		}
		if err := tx.Where("user_id = ?", userID).Delete(&Inventory{}).Error; err != nil {
			return fmt.Errorf("delete inventories: %w", err)
		}

		if len(imports) == 0 {
			inv := &Inventory{UserID: userID, Name: DefaultInventoryName, IsDefault: true}
			if err := tx.Create(inv).Error; err != nil {
				return fmt.Errorf("create default inventory: %w", err)
			}
			return nil
		}

		hasDefault := false
		for _, imp := range imports {
			if imp.IsDefault {
				hasDefault = true
				break
			}
		}
		if !hasDefault {
			imports[0].IsDefault = true
		}

		for _, imp := range imports {
			inv := &Inventory{
				UserID:    userID,
				Name:      imp.Name,
				IsDefault: imp.IsDefault,
			}
			if err := tx.Create(inv).Error; err != nil {
				return fmt.Errorf("create inventory: %w", err)
			}
			for _, item := range imp.Items {
				row := &InventoryItem{
					InventoryID: inv.ID,
					Name:        item.Name,
					Quantity:    item.Quantity,
					Category:    item.Category,
				}
				if err := tx.Create(row).Error; err != nil {
					return fmt.Errorf("create inventory item: %w", err)
				}
			}
		}
		return nil
	})
}
