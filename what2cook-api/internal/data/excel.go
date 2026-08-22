package data

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"

	"what2cook-api/internal/inventory"
)

const (
	sheetInventories = "Inventories"
	sheetItems       = "Items"
)

func encodeExcel(snapshot *ExportSnapshot) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	defaultSheet := f.GetSheetName(0)
	if err := f.SetSheetName(defaultSheet, sheetInventories); err != nil {
		return nil, fmt.Errorf("rename inventories sheet: %w", err)
	}
	if _, err := f.NewSheet(sheetItems); err != nil {
		return nil, fmt.Errorf("create items sheet: %w", err)
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

	inventoryMap := make(map[string]*InventoryExport)

	if inventoryRows, rowsErr := f.GetRows(sheetInventories); rowsErr == nil && len(inventoryRows) >= 2 {
		for _, row := range inventoryRows[1:] {
			if len(strings.TrimSpace(strings.Join(row, ""))) == 0 {
				continue
			}
			name := inventory.NormalizeInventoryName(cellAt(row, 0))
			if name == "" {
				continue
			}
			inventoryMap[inventory.InventoryIdentity(name)] = &InventoryExport{
				Name:      name,
				IsDefault: inventory.ParseBoolish(cellAt(row, 1)),
				Items:     []ItemExport{},
			}
		}
	}

	itemRows, err := f.GetRows(sheetItems)
	if err != nil {
		itemRows = nil
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
			key := inventory.InventoryIdentity(invName)
			inv, ok := inventoryMap[key]
			if !ok {
				inv = &InventoryExport{Name: invName, Items: []ItemExport{}}
				inventoryMap[key] = inv
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

	if len(inventoryMap) == 0 {
		return ExportSnapshot{}, ErrEmptyImport
	}

	snapshot := ExportSnapshot{}
	for _, inv := range inventoryMap {
		snapshot.Inventories = append(snapshot.Inventories, *inv)
	}

	return snapshot, nil
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
