package inventory

import (
	"fmt"
	"strings"
)

// NormalizeInventoryName trims and caps an inventory name.
func NormalizeInventoryName(raw string) string {
	name := strings.TrimSpace(raw)
	if len(name) > 80 {
		name = name[:80]
	}
	return name
}

// InventoryIdentity is a case-insensitive key for matching inventories by name.
func InventoryIdentity(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// ItemIdentity is a case-insensitive key for matching items by name + category.
func ItemIdentity(name string, category *string) string {
	return strings.ToLower(strings.TrimSpace(name)) + "\x00" + strings.ToLower(strings.TrimSpace(DerefString(category)))
}

// FilterNewItems returns items that do not already exist in the inventory.
// Existing items are matched by name + category (case-insensitive). Duplicates
// inside incoming are also skipped. Quantity is not part of identity.
func FilterNewItems(existing []InventoryItem, incoming []ItemImport) (toInsert []ItemImport, skipped int) {
	seen := make(map[string]struct{}, len(existing)+len(incoming))
	for _, item := range existing {
		seen[ItemIdentity(item.Name, item.Category)] = struct{}{}
	}

	for _, item := range incoming {
		key := ItemIdentity(item.Name, item.Category)
		if _, exists := seen[key]; exists {
			skipped++
			continue
		}
		seen[key] = struct{}{}
		toInsert = append(toInsert, item)
	}
	return toInsert, skipped
}

// NormalizeItemImport validates and normalizes an imported item row.
func NormalizeItemImport(name string, quantity, category *string) (*ItemImport, error) {
	normalizedName := normalizeName(name)
	if normalizedName == "" {
		return nil, nil
	}

	return &ItemImport{
		Name:     normalizedName,
		Quantity: normalizeQuantity(quantity),
		Category: normalizeCategory(category),
	}, nil
}

// ParseBoolish parses common spreadsheet boolean values.
func ParseBoolish(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "y":
		return true
	default:
		return false
	}
}

// StringValue returns a trimmed string pointer or nil when empty.
func StringValue(raw string) *string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// DerefString returns the string value or empty when nil.
func DerefString(raw *string) string {
	if raw == nil {
		return ""
	}
	return *raw
}

// ValidateImportFileSize rejects files that are too large.
func ValidateImportFileSize(size int64) error {
	const maxBytes = 5 << 20 // 5 MiB
	if size > maxBytes {
		return fmt.Errorf("file too large (max %d MB)", maxBytes>>20)
	}
	return nil
}
