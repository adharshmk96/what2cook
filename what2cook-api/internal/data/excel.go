package data

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"what2cook-api/internal/inventory"
)

const (
	sheetUser        = "User"
	sheetInventories = "Inventories"
	sheetItems       = "Items"
)

func encodeExcel(snapshot *ExportSnapshot) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	defaultSheet := f.GetSheetName(0)
	if err := f.SetSheetName(defaultSheet, sheetUser); err != nil {
		return nil, fmt.Errorf("rename user sheet: %w", err)
	}

	if _, err := f.NewSheet(sheetInventories); err != nil {
		return nil, fmt.Errorf("create inventories sheet: %w", err)
	}
	if _, err := f.NewSheet(sheetItems); err != nil {
		return nil, fmt.Errorf("create items sheet: %w", err)
	}

	if err := writeUserSheet(f, snapshot.User); err != nil {
		return nil, err
	}
	if err := writeInventoriesSheet(f, snapshot.Inventories); err != nil {
		return nil, err
	}
	if err := writeItemsSheet(f, snapshot.Inventories); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("write excel: %w", err)
	}
	return buf.Bytes(), nil
}

func writeUserSheet(f *excelize.File, user UserExport) error {
	headers := []string{"email", "email_verified_at", "created_at"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheetUser, cell, header); err != nil {
			return err
		}
	}

	values := []any{user.Email, formatTime(user.EmailVerifiedAt), user.CreatedAt.UTC().Format(time.RFC3339)}
	for i, value := range values {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		if err := f.SetCellValue(sheetUser, cell, value); err != nil {
			return err
		}
	}
	return nil
}

func writeInventoriesSheet(f *excelize.File, inventories []InventoryExport) error {
	headers := []string{"name", "is_default"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheetInventories, cell, header); err != nil {
			return err
		}
	}

	row := 2
	for _, inv := range inventories {
		defaultValue := "false"
		if inv.IsDefault {
			defaultValue = "true"
		}
		if err := f.SetCellValue(sheetInventories, coordAt(1, row), inv.Name); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetInventories, coordAt(2, row), defaultValue); err != nil {
			return err
		}
		row++
	}
	return nil
}

func writeItemsSheet(f *excelize.File, inventories []InventoryExport) error {
	headers := []string{"inventory_name", "name", "quantity", "category"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheetItems, cell, header); err != nil {
			return err
		}
	}

	row := 2
	for _, inv := range inventories {
		for _, item := range inv.Items {
			if err := f.SetCellValue(sheetItems, coordAt(1, row), inv.Name); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetItems, coordAt(2, row), item.Name); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetItems, coordAt(3, row), inventory.DerefString(item.Quantity)); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetItems, coordAt(4, row), inventory.DerefString(item.Category)); err != nil {
				return err
			}
			row++
		}
	}
	return nil
}

func decodeExcel(raw []byte) (ExportSnapshot, error) {
	f, err := excelize.OpenReader(bytes.NewReader(raw))
	if err != nil {
		return ExportSnapshot{}, fmt.Errorf("parse excel: %w", err)
	}
	defer f.Close()

	snapshot := ExportSnapshot{}
	userRows, err := f.GetRows(sheetUser)
	if err == nil && len(userRows) >= 2 {
		snapshot.User = parseUserRows(userRows)
	}

	inventoryRows, err := f.GetRows(sheetInventories)
	if err != nil || len(inventoryRows) < 2 {
		return ExportSnapshot{}, fmt.Errorf("%w: missing Inventories sheet", ErrEmptyImport)
	}

	itemRows, err := f.GetRows(sheetItems)
	if err != nil {
		itemRows = nil
	}

	inventoryMap := make(map[string]*InventoryExport)
	for _, row := range inventoryRows[1:] {
		if len(strings.TrimSpace(strings.Join(row, ""))) == 0 {
			continue
		}
		name := inventory.NormalizeInventoryName(cellAt(row, 0))
		if name == "" {
			continue
		}
		inventoryMap[name] = &InventoryExport{
			Name:      name,
			IsDefault: inventory.ParseBoolish(cellAt(row, 1)),
			Items:     []ItemExport{},
		}
	}

	if len(inventoryMap) == 0 {
		return ExportSnapshot{}, ErrEmptyImport
	}

	if len(itemRows) >= 2 {
		for _, row := range itemRows[1:] {
			if len(strings.TrimSpace(strings.Join(row, ""))) == 0 {
				continue
			}
			invName := inventory.NormalizeInventoryName(cellAt(row, 0))
			if invName == "" {
				invName = inventory.DefaultInventoryName
			}
			inv, ok := inventoryMap[invName]
			if !ok {
				inv = &InventoryExport{Name: invName, Items: []ItemExport{}}
				inventoryMap[invName] = inv
			}
			itemName := strings.TrimSpace(cellAt(row, 1))
			if itemName == "" {
				continue
			}
			inv.Items = append(inv.Items, ItemExport{
				Name:     itemName,
				Quantity: inventory.StringValue(cellAt(row, 2)),
				Category: inventory.StringValue(cellAt(row, 3)),
			})
		}
	}

	for _, inv := range inventoryMap {
		snapshot.Inventories = append(snapshot.Inventories, *inv)
	}

	return snapshot, nil
}

func parseUserRows(rows [][]string) UserExport {
	user := UserExport{Email: cellAt(rows[1], 0)}
	if verified := cellAt(rows[1], 1); verified != "" {
		if parsed, err := time.Parse(time.RFC3339, verified); err == nil {
			user.EmailVerifiedAt = &parsed
		}
	}
	if created := cellAt(rows[1], 2); created != "" {
		if parsed, err := time.Parse(time.RFC3339, created); err == nil {
			user.CreatedAt = parsed
		}
	}
	return user
}

func cellAt(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
}

func coordAt(col, row int) string {
	name, _ := excelize.CoordinatesToCellName(col, row)
	return name
}
