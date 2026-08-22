package data

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"

	"what2cook-api/internal/inventory"
)

const csvHeader = "inventory_name,inventory_is_default,item_name,item_quantity,item_category"

func encodeCSV(snapshot *ExportSnapshot) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	if err := w.Write(strings.Split(csvHeader, ",")); err != nil {
		return nil, fmt.Errorf("write csv header: %w", err)
	}

	if len(snapshot.Inventories) == 0 {
		if err := w.Write([]string{inventory.DefaultInventoryName, "true", "", "", ""}); err != nil {
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

		invName := inventory.NormalizeInventoryName(cell(row, col["inventory_name"]))
		if invName == "" {
			invName = inventory.DefaultInventoryName
		}

		key := inventory.InventoryIdentity(invName)
		inv, ok := inventoryMap[key]
		if !ok {
			inv = &InventoryExport{
				Name:      invName,
				IsDefault: inventory.ParseBoolish(cell(row, col["inventory_is_default"])),
				Items:     []ItemExport{},
			}
			inventoryMap[key] = inv
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
		"inventory_name":       -1,
		"inventory_is_default": -1,
		"item_name":            -1,
		"item_quantity":        -1,
		"item_category":        -1,
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
