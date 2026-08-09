package inventory

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const DefaultInventoryName = "My Pantry"

// Inventory is a named pantry owned by a user.
type Inventory struct {
	ID        uuid.UUID       `gorm:"type:text;primaryKey" json:"id"`
	UserID    uuid.UUID       `gorm:"type:text;index;not null" json:"user_id"`
	Name      string          `gorm:"size:80;not null" json:"name"`
	IsDefault bool            `gorm:"not null;default:false;index" json:"is_default"`
	Items     []InventoryItem `gorm:"foreignKey:InventoryID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// BeforeCreate assigns a UUID if missing.
func (inv *Inventory) BeforeCreate(tx *gorm.DB) error {
	if inv.ID == uuid.Nil {
		inv.ID = uuid.New()
	}
	return nil
}

// InventoryItem is an ingredient inside an inventory.
type InventoryItem struct {
	ID          uuid.UUID `gorm:"type:text;primaryKey" json:"id"`
	InventoryID uuid.UUID `gorm:"type:text;index;not null" json:"inventory_id"`
	Name        string    `gorm:"size:80;not null" json:"name"`
	Quantity    *string   `gorm:"size:40" json:"quantity"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BeforeCreate assigns a UUID if missing.
func (item *InventoryItem) BeforeCreate(tx *gorm.DB) error {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	return nil
}

// CreateInventoryRequest is the body for POST /inventories.
type CreateInventoryRequest struct {
	Name string `json:"name"`
}

// UpdateInventoryRequest is the body for PATCH /inventories/:id.
type UpdateInventoryRequest struct {
	Name string `json:"name"`
}

// CreateItemRequest is the body for POST /inventories/:id/items.
type CreateItemRequest struct {
	Name     string  `json:"name"`
	Quantity *string `json:"quantity"`
}

// UpdateItemRequest is the body for PATCH /inventories/:id/items/:itemId.
type UpdateItemRequest struct {
	Name     *string `json:"name"`
	Quantity *string `json:"quantity"`
}
