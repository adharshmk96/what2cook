package inventory

import "strings"

const (
	maxNameLen     = 80
	maxQuantityLen = 40
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
	return ""
}

// ValidateUpdateItem returns an error message, or empty if valid.
func ValidateUpdateItem(req *UpdateItemRequest) string {
	if req == nil {
		return "name is required"
	}
	if req.Name == nil && req.Quantity == nil {
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
		// Empty string clears quantity; keep a sentinel empty pointer so the
		// handler can tell the field was provided, then normalize in service.
		trimmed := strings.TrimSpace(*req.Quantity)
		if len(trimmed) > maxQuantityLen {
			trimmed = trimmed[:maxQuantityLen]
		}
		req.Quantity = &trimmed
	}
	return ""
}
