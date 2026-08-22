package data

import "time"

// UserExport is the user profile section of an export.
type UserExport struct {
	Email           string     `json:"email"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

// ExportSnapshot is the full export payload for a user.
type ExportSnapshot struct {
	User        UserExport        `json:"user"`
	Inventories []InventoryExport `json:"inventories"`
}

// InventoryExport is an inventory with nested items.
type InventoryExport struct {
	Name      string       `json:"name"`
	IsDefault bool         `json:"is_default"`
	Items     []ItemExport `json:"items"`
}

// ItemExport is a single inventory item.
type ItemExport struct {
	Name     string  `json:"name"`
	Quantity *string `json:"quantity"`
	Category *string `json:"category"`
}

// ImportResult summarizes an import operation.
type ImportResult struct {
	Inventories int `json:"inventories"`
	Items       int `json:"items"`
}
