package data

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"what2cook-api/internal/inventory"
)

var (
	ErrInvalidFormat = errors.New("invalid format")
	ErrEmptyImport   = errors.New("empty import file")
)

// Service handles portable inventory export and import.
type Service struct {
	invRepo *inventory.Repository
}

// NewService creates a data service.
func NewService(invRepo *inventory.Repository) *Service {
	return &Service{invRepo: invRepo}
}

// Export returns file bytes, content type, and download filename.
func (s *Service) Export(userID uuid.UUID, format string) ([]byte, string, string, error) {
	snapshot, err := s.loadSnapshot(userID)
	if err != nil {
		return nil, "", "", err
	}

	switch strings.ToLower(strings.TrimSpace(format)) {
	case "csv":
		data, err := encodeCSV(snapshot)
		if err != nil {
			return nil, "", "", err
		}
		return data, "text/csv; charset=utf-8", "what2cook-export.csv", nil
	case "xlsx", "excel":
		data, err := encodeExcel(snapshot)
		if err != nil {
			return nil, "", "", err
		}
		return data, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "what2cook-export.xlsx", nil
	default:
		return nil, "", "", fmt.Errorf("%w: use csv or xlsx", ErrInvalidFormat)
	}
}

// Import merges inventories and items from a file into the current user's data.
// Existing items (same name + category) are skipped; new items are inserted.
func (s *Service) Import(userID uuid.UUID, raw []byte, format string) (*ImportResult, error) {
	if len(raw) == 0 {
		return nil, ErrEmptyImport
	}

	var snapshot ExportSnapshot
	var err error

	switch strings.ToLower(strings.TrimSpace(format)) {
	case "csv":
		snapshot, err = decodeCSV(raw)
	case "xlsx", "excel":
		snapshot, err = decodeExcel(raw)
	default:
		return nil, fmt.Errorf("%w: use csv or xlsx", ErrInvalidFormat)
	}
	if err != nil {
		return nil, err
	}

	imports, _, err := snapshotToImports(snapshot)
	if err != nil {
		return nil, err
	}

	merged, err := s.invRepo.MergeForUser(userID, imports)
	if err != nil {
		return nil, err
	}

	return &ImportResult{
		Inventories: merged.InventoriesCreated,
		Items:       merged.ItemsInserted,
		Skipped:     merged.ItemsSkipped,
	}, nil
}

func (s *Service) loadSnapshot(userID uuid.UUID) (*ExportSnapshot, error) {
	inventories, err := s.invRepo.ListByUserWithItems(userID)
	if err != nil {
		return nil, err
	}
	if len(inventories) == 0 {
		inv, ensureErr := s.invRepo.FindDefaultByUser(userID)
		if ensureErr == nil {
			inventories = []inventory.Inventory{*inv}
		}
	}

	snapshot := &ExportSnapshot{
		Inventories: make([]InventoryExport, 0, len(inventories)),
	}

	for _, inv := range inventories {
		export := InventoryExport{
			Name:      inv.Name,
			IsDefault: inv.IsDefault,
			Items:     make([]ItemExport, 0, len(inv.Items)),
		}
		for _, item := range inv.Items {
			export.Items = append(export.Items, ItemExport{
				Name:     item.Name,
				Quantity: item.Quantity,
				Category: item.Category,
			})
		}
		snapshot.Inventories = append(snapshot.Inventories, export)
	}

	return snapshot, nil
}

func snapshotToImports(snapshot ExportSnapshot) ([]inventory.InventoryImport, int, error) {
	if len(snapshot.Inventories) == 0 {
		return nil, 0, nil
	}

	imports := make([]inventory.InventoryImport, 0, len(snapshot.Inventories))
	itemCount := 0

	for _, inv := range snapshot.Inventories {
		name := inventory.NormalizeInventoryName(inv.Name)
		if name == "" {
			name = inventory.DefaultInventoryName
		}

		imp := inventory.InventoryImport{
			Name:      name,
			IsDefault: inv.IsDefault,
			Items:     make([]inventory.ItemImport, 0, len(inv.Items)),
		}

		for _, item := range inv.Items {
			normalized, err := inventory.NormalizeItemImport(item.Name, item.Quantity, item.Category)
			if err != nil {
				return nil, 0, err
			}
			if normalized == nil {
				continue
			}
			imp.Items = append(imp.Items, *normalized)
			itemCount++
		}

		imports = append(imports, imp)
	}

	return imports, itemCount, nil
}
