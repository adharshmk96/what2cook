package data

import (
	"bytes"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestCSVRoundTrip(t *testing.T) {
	qty := "500g"
	cat := "Meat"

	original := &ExportSnapshot{
		Inventories: []InventoryExport{
			{
				Name:      "My Pantry",
				IsDefault: true,
				Items: []ItemExport{
					{Name: "chicken", Quantity: &qty, Category: &cat},
				},
			},
		},
	}

	raw, err := encodeCSV(original)
	if err != nil {
		t.Fatalf("encodeCSV: %v", err)
	}
	if strings.Contains(string(raw), "email") {
		t.Fatalf("csv should not include user fields: %s", string(raw))
	}

	parsed, err := decodeCSV(raw)
	if err != nil {
		t.Fatalf("decodeCSV: %v", err)
	}

	if len(parsed.Inventories) != 1 {
		t.Fatalf("inventories = %d, want 1", len(parsed.Inventories))
	}
	if parsed.Inventories[0].Name != "My Pantry" {
		t.Fatalf("inventory name = %q", parsed.Inventories[0].Name)
	}
	if len(parsed.Inventories[0].Items) != 1 {
		t.Fatalf("items = %d, want 1", len(parsed.Inventories[0].Items))
	}
	if parsed.Inventories[0].Items[0].Name != "chicken" {
		t.Fatalf("item name = %q", parsed.Inventories[0].Items[0].Name)
	}
}

func TestCSVDecodeIgnoresLegacyUserColumns(t *testing.T) {
	raw := []byte(strings.Join([]string{
		"email,email_verified_at,user_created_at,inventory_name,inventory_is_default,item_name,item_quantity,item_category",
		"chef@example.com,2026-01-02T03:04:05Z,2025-12-01T08:00:00Z,My Pantry,true,chicken,500g,Meat",
	}, "\n"))

	parsed, err := decodeCSV(raw)
	if err != nil {
		t.Fatalf("decodeCSV: %v", err)
	}
	if len(parsed.Inventories) != 1 || len(parsed.Inventories[0].Items) != 1 {
		t.Fatalf("parsed inventories=%d items=%d", len(parsed.Inventories), len(parsed.Inventories[0].Items))
	}
	if parsed.Inventories[0].Items[0].Name != "chicken" {
		t.Fatalf("item name = %q", parsed.Inventories[0].Items[0].Name)
	}
}

func TestDetectFormatFromFilename(t *testing.T) {
	if got := detectFormat("backup.XLSX"); got != "xlsx" {
		t.Fatalf("detectFormat xlsx = %q", got)
	}
	if got := detectFormat("backup.csv"); got != "csv" {
		t.Fatalf("detectFormat csv = %q", got)
	}
}

func TestCSVHeaderPresent(t *testing.T) {
	snapshot := &ExportSnapshot{
		Inventories: []InventoryExport{
			{Name: "Pantry", IsDefault: true, Items: []ItemExport{}},
		},
	}
	raw, err := encodeCSV(snapshot)
	if err != nil {
		t.Fatalf("encodeCSV: %v", err)
	}
	header := strings.Split(strings.Split(string(raw), "\n")[0], ",")
	for _, col := range header {
		if strings.Contains(col, "email") || strings.Contains(col, "user_") {
			t.Fatalf("csv header includes user field %q: %s", col, string(raw))
		}
	}
	if !strings.Contains(string(raw), "inventory_name") {
		t.Fatalf("csv missing header: %s", string(raw))
	}
}

func TestExcelRoundTripOmitsUserSheet(t *testing.T) {
	qty := "1kg"
	cat := "Produce"
	original := &ExportSnapshot{
		Inventories: []InventoryExport{
			{
				Name:      "My Pantry",
				IsDefault: true,
				Items: []ItemExport{
					{Name: "tomato", Quantity: &qty, Category: &cat},
				},
			},
		},
	}

	raw, err := encodeExcel(original)
	if err != nil {
		t.Fatalf("encodeExcel: %v", err)
	}

	book, err := excelize.OpenReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("open excel: %v", err)
	}
	defer book.Close()
	for _, name := range book.GetSheetList() {
		if strings.EqualFold(name, "User") {
			t.Fatal("excel export should not include a User sheet")
		}
	}

	parsed, err := decodeExcel(raw)
	if err != nil {
		t.Fatalf("decodeExcel: %v", err)
	}
	if len(parsed.Inventories) != 1 || len(parsed.Inventories[0].Items) != 1 {
		t.Fatalf("parsed inventories=%d", len(parsed.Inventories))
	}
	if parsed.Inventories[0].Items[0].Name != "tomato" {
		t.Fatalf("item name = %q", parsed.Inventories[0].Items[0].Name)
	}
}
