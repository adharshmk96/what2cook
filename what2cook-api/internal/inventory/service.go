package inventory

import (
	"errors"
	"log"

	"github.com/google/uuid"
)

var ErrCannotDeleteDefault = errors.New("cannot delete default inventory")

type Service struct { repo *Repository }
func NewService(repo *Repository) *Service { return &Service{repo: repo} }

func (s *Service) EnsureDefault(userID uuid.UUID) (*Inventory, error) {
	existing, err := s.repo.FindDefaultByUser(userID)
	if err == nil { return existing, nil }
	if !errors.Is(err, ErrNotFound) { return nil, err }
	inv := &Inventory{UserID: userID, Name: DefaultInventoryName, IsDefault: true}
	if err := s.repo.CreateInventory(inv); err != nil {
		again, findErr := s.repo.FindDefaultByUser(userID)
		if findErr == nil { return again, nil }
		return nil, err
	}
	log.Printf("created default inventory %s for user %s", inv.ID, userID)
	return inv, nil
}

func (s *Service) List(userID uuid.UUID) ([]Inventory, error) {
	if _, err := s.EnsureDefault(userID); err != nil { return nil, err }
	return s.repo.ListByUser(userID)
}
func (s *Service) Create(userID uuid.UUID, name string) (*Inventory, error) {
	if _, err := s.EnsureDefault(userID); err != nil { return nil, err }
	inv := &Inventory{UserID: userID, Name: name, IsDefault: false}
	if err := s.repo.CreateInventory(inv); err != nil { return nil, err }
	return inv, nil
}
func (s *Service) Get(userID, inventoryID uuid.UUID) (*Inventory, error) {
	if _, err := s.EnsureDefault(userID); err != nil { return nil, err }
	return s.repo.FindByIDForUserWithItems(inventoryID, userID)
}
func (s *Service) Update(userID, inventoryID uuid.UUID, name string) (*Inventory, error) {
	if err := s.repo.UpdateInventoryName(inventoryID, userID, name); err != nil { return nil, err }
	return s.repo.FindByIDForUserWithItems(inventoryID, userID)
}
func (s *Service) Delete(userID, inventoryID uuid.UUID) error {
	inv, err := s.repo.FindByIDForUser(inventoryID, userID)
	if err != nil { return err }
	if inv.IsDefault { return ErrCannotDeleteDefault }
	return s.repo.DeleteInventory(inventoryID, userID)
}

func (s *Service) AddItem(userID, inventoryID uuid.UUID, name string, quantity, category *string) (*InventoryItem, error) {
	if _, err := s.repo.FindByIDForUser(inventoryID, userID); err != nil { return nil, err }
	item := &InventoryItem{InventoryID: inventoryID, Name: name, Quantity: quantity, Category: category}
	if err := s.repo.CreateItem(item); err != nil { return nil, err }
	return item, nil
}

func (s *Service) UpdateItem(userID, inventoryID, itemID uuid.UUID, name *string, quantitySet bool, quantity *string, categorySet bool, category *string) (*InventoryItem, error) {
	item, err := s.repo.FindItemForUser(inventoryID, itemID, userID)
	if err != nil { return nil, err }
	if name != nil { item.Name = *name }
	if quantitySet { item.Quantity = normalizeQuantity(quantity) }
	if categorySet { item.Category = normalizeCategory(category) }
	if err := s.repo.UpdateItem(item); err != nil { return nil, err }
	return item, nil
}
func (s *Service) DeleteItem(userID, inventoryID, itemID uuid.UUID) error { return s.repo.DeleteItem(inventoryID, itemID, userID) }
