package inventory

import "strings"

const (
	maxNameLen     = 80
	maxQuantityLen = 40
	maxCategoryLen = 60
)

func normalizeName(raw string) string {
	name := strings.TrimSpace(raw)
	if len(name) > maxNameLen {
		name = name[:maxNameLen]
	}
	return name
}

func normalizeQuantity(raw *string) *string {
	if raw == nil {
		return nil
	}
	q := strings.TrimSpace(*raw)
	if q == "" {
		return nil
	}
	if len(q) > maxQuantityLen {
		q = q[:maxQuantityLen]
	}
	return &q
}

func normalizeCategory(raw *string) *string {
	if raw == nil {
		return nil
	}
	category := strings.TrimSpace(*raw)
	if category == "" {
		return nil
	}
	if len(category) > maxCategoryLen {
		category = category[:maxCategoryLen]
	}
	return &category
}

// ValidateCreateInventory returns an error message, or empty if valid.
func ValidateCreateInventory(req *CreateInventoryRequest) string {
	if req == nil {
		return "name is required"
	}
	req.Name = normalizeName(req.Name)
	if req.Name == "" {
		return "name is required"
	}
	return ""
}

// ValidateUpdateInventory returns an error message, or empty if valid.
func ValidateUpdateInventory(req *UpdateInventoryRequest) string {
	if req == nil {
		return "name is required"
	}
	req.Name = normalizeName(req.Name)
	if req.Name == "" {
		return "name is required"
	}
	return ""
}

// ValidateCreateItem returns an error message, or empty if valid.
func ValidateCreateItem(req *CreateItemRequest) string {
	if req == nil {
		return "name is required"
	}
	req.Name = normalizeName(req.Name)
	if req.Name == "" {
		return "name is required"
	}
	req.Quantity = normalizeQuantity(req.Quantity)
	req.Category = normalizeCategory(req.Category)
	return ""
}

// ValidateUpdateItem returns an error message, or empty if valid.
func ValidateUpdateItem(req *UpdateItemRequest) string {
	if req == nil {
		return "name is required"
	}
	if req.Name == nil && req.Quantity == nil && req.Category == nil {
		return "at least one field is required"
	}
	if req.Name != nil {
		normalized := normalizeName(*req.Name)
		if normalized == "" {
			return "name is required"
		}
		req.Name = &normalized
	}
	if req.Quantity != nil {
		trimmed := strings.TrimSpace(*req.Quantity)
		if len(trimmed) > maxQuantityLen {
			trimmed = trimmed[:maxQuantityLen]
		}
		req.Quantity = &trimmed
	}
	if req.Category != nil {
		trimmed := strings.TrimSpace(*req.Category)
		if len(trimmed) > maxCategoryLen {
			trimmed = trimmed[:maxCategoryLen]
		}
		req.Category = &trimmed
	}
	return ""
}
