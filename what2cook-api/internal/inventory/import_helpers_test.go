package inventory

import "testing"

func strPtr(v string) *string { return &v }

func TestItemIdentityIgnoresCase(t *testing.T) {
	cat := "Meat"
	if ItemIdentity("Chicken", &cat) != ItemIdentity("chicken", strPtr("meat")) {
		t.Fatal("name+category identity should be case-insensitive")
	}
}

func TestFilterNewItemsSkipsExistingAndInsertsNew(t *testing.T) {
	existing := []InventoryItem{
		{Name: "chicken", Category: strPtr("Meat")},
		{Name: "tomato", Category: strPtr("Produce")},
	}
	incoming := []ItemImport{
		{Name: "Chicken", Quantity: strPtr("1kg"), Category: strPtr("meat")},
		{Name: "ginger", Category: strPtr("Produce")},
		{Name: "tomato", Category: strPtr("Produce")},
		{Name: "Ginger", Category: strPtr("produce")},
	}

	toInsert, skipped := FilterNewItems(existing, incoming)
	if skipped != 3 {
		t.Fatalf("skipped = %d, want 3 (existing chicken+tomato and duplicate ginger)", skipped)
	}
	if len(toInsert) != 1 {
		t.Fatalf("toInsert = %d, want 1", len(toInsert))
	}
	if toInsert[0].Name != "ginger" {
		t.Fatalf("inserted name = %q, want ginger", toInsert[0].Name)
	}
}

func TestFilterNewItemsTreatsCategoryAsPartOfIdentity(t *testing.T) {
	existing := []InventoryItem{
		{Name: "salt", Category: strPtr("Spices")},
	}
	incoming := []ItemImport{
		{Name: "salt", Category: strPtr("Baking")},
		{Name: "salt", Category: nil},
	}

	toInsert, skipped := FilterNewItems(existing, incoming)
	if skipped != 0 {
		t.Fatalf("skipped = %d, want 0", skipped)
	}
	if len(toInsert) != 2 {
		t.Fatalf("toInsert = %d, want 2", len(toInsert))
	}
}

func TestFilterNewItemsEmptyExistingInsertsAllUnique(t *testing.T) {
	incoming := []ItemImport{
		{Name: "rice"},
		{Name: "Rice"},
		{Name: "beans"},
	}
	toInsert, skipped := FilterNewItems(nil, incoming)
	if skipped != 1 {
		t.Fatalf("skipped = %d, want 1", skipped)
	}
	if len(toInsert) != 2 {
		t.Fatalf("toInsert = %d, want 2", len(toInsert))
	}
}
