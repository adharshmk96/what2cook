package inventory

import (
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func testRepo(t *testing.T) *Repository {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&Inventory{}, &InventoryItem{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return NewRepository(db)
}

func TestMergeForUserInsertsNewAndSkipsExisting(t *testing.T) {
	repo := testRepo(t)
	userID := uuid.New()
	inv := &Inventory{UserID: userID, Name: DefaultInventoryName, IsDefault: true}
	if err := repo.CreateInventory(inv); err != nil {
		t.Fatalf("create inventory: %v", err)
	}
	if err := repo.CreateItem(&InventoryItem{
		InventoryID: inv.ID,
		Name:        "chicken",
		Category:    strPtr("Meat"),
	}); err != nil {
		t.Fatalf("create item: %v", err)
	}

	result, err := repo.MergeForUser(userID, []InventoryImport{
		{
			Name:      DefaultInventoryName,
			IsDefault: true,
			Items: []ItemImport{
				{Name: "Chicken", Quantity: strPtr("500g"), Category: strPtr("meat")},
				{Name: "ginger", Category: strPtr("Produce")},
			},
		},
	})
	if err != nil {
		t.Fatalf("MergeForUser: %v", err)
	}
	if result.ItemsInserted != 1 {
		t.Fatalf("inserted = %d, want 1", result.ItemsInserted)
	}
	if result.ItemsSkipped != 1 {
		t.Fatalf("skipped = %d, want 1", result.ItemsSkipped)
	}
	if result.InventoriesCreated != 0 {
		t.Fatalf("inventories created = %d, want 0", result.InventoriesCreated)
	}

	loaded, err := repo.FindByIDForUserWithItems(inv.ID, userID)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if len(loaded.Items) != 2 {
		t.Fatalf("items = %d, want 2", len(loaded.Items))
	}
}

func TestMergeForUserCreatesDataForAnotherUser(t *testing.T) {
	repo := testRepo(t)
	sourceUser := uuid.New()
	targetUser := uuid.New()

	source := &Inventory{UserID: sourceUser, Name: DefaultInventoryName, IsDefault: true}
	if err := repo.CreateInventory(source); err != nil {
		t.Fatalf("create source inventory: %v", err)
	}
	if err := repo.CreateItem(&InventoryItem{
		InventoryID: source.ID,
		Name:        "tomato",
		Category:    strPtr("Produce"),
	}); err != nil {
		t.Fatalf("create source item: %v", err)
	}

	target := &Inventory{UserID: targetUser, Name: DefaultInventoryName, IsDefault: true}
	if err := repo.CreateInventory(target); err != nil {
		t.Fatalf("create target inventory: %v", err)
	}

	result, err := repo.MergeForUser(targetUser, []InventoryImport{
		{
			Name: DefaultInventoryName,
			Items: []ItemImport{
				{Name: "tomato", Category: strPtr("Produce")},
			},
		},
	})
	if err != nil {
		t.Fatalf("MergeForUser: %v", err)
	}
	if result.ItemsInserted != 1 {
		t.Fatalf("inserted = %d, want 1", result.ItemsInserted)
	}

	sourceLoaded, err := repo.FindByIDForUserWithItems(source.ID, sourceUser)
	if err != nil {
		t.Fatalf("reload source: %v", err)
	}
	if len(sourceLoaded.Items) != 1 {
		t.Fatalf("source items changed: %d", len(sourceLoaded.Items))
	}

	targetLoaded, err := repo.FindByIDForUserWithItems(target.ID, targetUser)
	if err != nil {
		t.Fatalf("reload target: %v", err)
	}
	if len(targetLoaded.Items) != 1 {
		t.Fatalf("target items = %d, want 1", len(targetLoaded.Items))
	}
	if targetLoaded.Items[0].ID == sourceLoaded.Items[0].ID {
		t.Fatal("imported item reused source id")
	}
}
