package data

// ExportSnapshot is a portable inventory dump. It is not tied to a user.
type ExportSnapshot struct {
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

// ImportResult summarizes an import merge.
type ImportResult struct {
	Inventories int `json:"inventories"`
	Items       int `json:"items"`
	Skipped     int `json:"skipped"`
}
