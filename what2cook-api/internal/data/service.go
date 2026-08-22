package data

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"what2cook-api/internal/auth"
	"what2cook-api/internal/inventory"
)

var (
	ErrInvalidFormat = errors.New("invalid format")
	ErrEmptyImport   = errors.New("empty import file")
)

// Service handles user data export and import.
type Service struct {
	authRepo *auth.Repository
	invRepo  *inventory.Repository
}

// NewService creates a data service.
func NewService(authRepo *auth.Repository, invRepo *inventory.Repository) *Service {
	return &Service{authRepo: authRepo, invRepo: invRepo}
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

// Import replaces the user's inventory data from a file.
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

	imports, itemCount, err := snapshotToImports(snapshot)
	if err != nil {
		return nil, err
	}

	if err := s.invRepo.ReplaceAllForUser(userID, imports); err != nil {
		return nil, err
	}

	return &ImportResult{
		Inventories: len(imports),
		Items:       itemCount,
	}, nil
}

func (s *Service) loadSnapshot(userID uuid.UUID) (*ExportSnapshot, error) {
	user, err := s.authRepo.FindUserByID(userID)
	if err != nil {
		return nil, err
	}

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
		User: UserExport{
			Email:           user.Email,
			EmailVerifiedAt: user.EmailVerifiedAt,
			CreatedAt:       user.CreatedAt,
		},
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
