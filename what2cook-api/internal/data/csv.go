package data

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"what2cook-api/internal/inventory"
)

const csvHeader = "email,email_verified_at,user_created_at,inventory_name,inventory_is_default,item_name,item_quantity,item_category"

func encodeCSV(snapshot *ExportSnapshot) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	if err := w.Write(strings.Split(csvHeader, ",")); err != nil {
		return nil, fmt.Errorf("write csv header: %w", err)
	}

	email := snapshot.User.Email
	verified := formatTime(snapshot.User.EmailVerifiedAt)
	created := snapshot.User.CreatedAt.UTC().Format(time.RFC3339)

	if len(snapshot.Inventories) == 0 {
		if err := w.Write([]string{email, verified, created, inventory.DefaultInventoryName, "true", "", "", ""}); err != nil {
			return nil, fmt.Errorf("write csv row: %w", err)
		}
		w.Flush()
		return buf.Bytes(), w.Error()
	}

	for _, inv := range snapshot.Inventories {
		defaultFlag := "false"
		if inv.IsDefault {
			defaultFlag = "true"
		}

		if len(inv.Items) == 0 {
			if err := w.Write([]string{
				email,
				verified,
				created,
				inv.Name,
				defaultFlag,
				"",
				"",
				"",
			}); err != nil {
				return nil, fmt.Errorf("write csv row: %w", err)
			}
			continue
		}

		for _, item := range inv.Items {
			if err := w.Write([]string{
				email,
				verified,
				created,
				inv.Name,
				defaultFlag,
				item.Name,
				inventory.DerefString(item.Quantity),
				inventory.DerefString(item.Category),
			}); err != nil {
				return nil, fmt.Errorf("write csv row: %w", err)
			}
		}
	}

	w.Flush()
	return buf.Bytes(), w.Error()
}

func decodeCSV(raw []byte) (ExportSnapshot, error) {
	reader := csv.NewReader(bytes.NewReader(raw))
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return ExportSnapshot{}, fmt.Errorf("parse csv: %w", err)
	}
	if len(records) < 2 {
		return ExportSnapshot{}, fmt.Errorf("%w: csv must include a header and at least one data row", ErrEmptyImport)
	}

	header := normalizeHeader(records[0])
	col := mapHeaderColumns(header)
	if col["inventory_name"] < 0 || col["item_name"] < 0 {
		return ExportSnapshot{}, fmt.Errorf("invalid csv header: expected %s", csvHeader)
	}

	snapshot := ExportSnapshot{}
	inventoryMap := make(map[string]*InventoryExport)

	for _, row := range records[1:] {
		if len(strings.TrimSpace(strings.Join(row, ""))) == 0 {
			continue
		}

		if snapshot.User.Email == "" {
			snapshot.User.Email = cell(row, col["email"])
			if verified := cell(row, col["email_verified_at"]); verified != "" {
				if parsed, parseErr := time.Parse(time.RFC3339, verified); parseErr == nil {
					snapshot.User.EmailVerifiedAt = &parsed
				}
			}
			if created := cell(row, col["user_created_at"]); created != "" {
				if parsed, parseErr := time.Parse(time.RFC3339, created); parseErr == nil {
					snapshot.User.CreatedAt = parsed
				}
			}
		}

		invName := inventory.NormalizeInventoryName(cell(row, col["inventory_name"]))
		if invName == "" {
			invName = inventory.DefaultInventoryName
		}

		inv, ok := inventoryMap[invName]
		if !ok {
			inv = &InventoryExport{
				Name:      invName,
				IsDefault: inventory.ParseBoolish(cell(row, col["inventory_is_default"])),
				Items:     []ItemExport{},
			}
			inventoryMap[invName] = inv
		}

		itemName := strings.TrimSpace(cell(row, col["item_name"]))
		if itemName == "" {
			continue
		}

		inv.Items = append(inv.Items, ItemExport{
			Name:     itemName,
			Quantity: inventory.StringValue(cell(row, col["item_quantity"])),
			Category: inventory.StringValue(cell(row, col["item_category"])),
		})
	}

	for _, inv := range inventoryMap {
		snapshot.Inventories = append(snapshot.Inventories, *inv)
	}

	if len(snapshot.Inventories) == 0 {
		return ExportSnapshot{}, ErrEmptyImport
	}

	return snapshot, nil
}

func normalizeHeader(header []string) []string {
	out := make([]string, len(header))
	for i, col := range header {
		out[i] = strings.ToLower(strings.TrimSpace(col))
	}
	return out
}

func mapHeaderColumns(header []string) map[string]int {
	columns := map[string]int{
		"email":                  -1,
		"email_verified_at":      -1,
		"user_created_at":        -1,
		"inventory_name":         -1,
		"inventory_is_default":   -1,
		"item_name":              -1,
		"item_quantity":          -1,
		"item_category":          -1,
	}
	for i, col := range header {
		if _, ok := columns[col]; ok {
			columns[col] = i
		}
	}
	return columns
}

func cell(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return row[index]
}

func formatTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
