package data

import (
	"strings"
	"testing"
	"time"
)

func TestCSVRoundTrip(t *testing.T) {
	verified := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	created := time.Date(2025, 12, 1, 8, 0, 0, 0, time.UTC)
	qty := "500g"
	cat := "Meat"

	original := &ExportSnapshot{
		User: UserExport{
			Email:           "chef@example.com",
			EmailVerifiedAt: &verified,
			CreatedAt:       created,
		},
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

	parsed, err := decodeCSV(raw)
	if err != nil {
		t.Fatalf("decodeCSV: %v", err)
	}

	if parsed.User.Email != original.User.Email {
		t.Fatalf("email = %q, want %q", parsed.User.Email, original.User.Email)
	}
	if len(parsed.Inventories) != 1 {
		t.Fatalf("inventories = %d, want 1", len(parsed.Inventories))
	}
	if len(parsed.Inventories[0].Items) != 1 {
		t.Fatalf("items = %d, want 1", len(parsed.Inventories[0].Items))
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
		User: UserExport{Email: "a@b.com", CreatedAt: time.Now()},
		Inventories: []InventoryExport{
			{Name: "Pantry", IsDefault: true, Items: []ItemExport{}},
		},
	}
	raw, err := encodeCSV(snapshot)
	if err != nil {
		t.Fatalf("encodeCSV: %v", err)
	}
	if !strings.Contains(string(raw), "inventory_name") {
		t.Fatalf("csv missing header: %s", string(raw))
	}
}
