package inventory

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"what2cook-api/internal/auth"
)

type Handler struct { svc *Service }
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }
type errorBody struct { Error string `json:"error"` }
type listBody struct { Inventories []Inventory `json:"inventories"` }

func (h *Handler) List(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c); if !ok { c.JSON(http.StatusUnauthorized, errorBody{Error:"unauthorized"}); return }
	list, err := h.svc.List(userID); if err != nil { log.Printf("inventory list failed: %v", err); c.JSON(http.StatusInternalServerError, errorBody{Error:"internal error"}); return }
	if list == nil { list = []Inventory{} }; c.JSON(http.StatusOK, listBody{Inventories:list})
}
func (h *Handler) Create(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c); if !ok { c.JSON(http.StatusUnauthorized, errorBody{Error:"unauthorized"}); return }
	var req CreateInventoryRequest; if c.ShouldBindJSON(&req) != nil { c.JSON(http.StatusBadRequest,errorBody{Error:"invalid JSON body"}); return }
	if msg:=ValidateCreateInventory(&req); msg!="" { c.JSON(http.StatusBadRequest,errorBody{Error:msg}); return }
	inv,err:=h.svc.Create(userID,req.Name); if err!=nil { log.Printf("inventory create failed: %v",err); c.JSON(http.StatusInternalServerError,errorBody{Error:"internal error"}); return }; c.JSON(http.StatusCreated,inv)
}
func (h *Handler) Get(c *gin.Context) {
	userID,ok:=auth.UserIDFromContext(c); if !ok { c.JSON(http.StatusUnauthorized,errorBody{Error:"unauthorized"}); return }; inventoryID,err:=uuid.Parse(c.Param("id")); if err!=nil { c.JSON(http.StatusBadRequest,errorBody{Error:"invalid inventory id"}); return }; inv,err:=h.svc.Get(userID,inventoryID); if err!=nil { h.writeServiceError(c,err); return }; c.JSON(http.StatusOK,inv)
}
func (h *Handler) Update(c *gin.Context) {
	userID,ok:=auth.UserIDFromContext(c); if !ok { c.JSON(http.StatusUnauthorized,errorBody{Error:"unauthorized"}); return }; inventoryID,err:=uuid.Parse(c.Param("id")); if err!=nil { c.JSON(http.StatusBadRequest,errorBody{Error:"invalid inventory id"}); return }; var req UpdateInventoryRequest; if c.ShouldBindJSON(&req)!=nil { c.JSON(http.StatusBadRequest,errorBody{Error:"invalid JSON body"}); return }; if msg:=ValidateUpdateInventory(&req); msg!="" { c.JSON(http.StatusBadRequest,errorBody{Error:msg}); return }; inv,err:=h.svc.Update(userID,inventoryID,req.Name); if err!=nil { h.writeServiceError(c,err); return }; c.JSON(http.StatusOK,inv)
}
func (h *Handler) Delete(c *gin.Context) {
	userID,ok:=auth.UserIDFromContext(c); if !ok { c.JSON(http.StatusUnauthorized,errorBody{Error:"unauthorized"}); return }; inventoryID,err:=uuid.Parse(c.Param("id")); if err!=nil { c.JSON(http.StatusBadRequest,errorBody{Error:"invalid inventory id"}); return }; if err:=h.svc.Delete(userID,inventoryID); err!=nil { h.writeServiceError(c,err); return }; c.Status(http.StatusNoContent)
}
func (h *Handler) CreateItem(c *gin.Context) {
	userID,ok:=auth.UserIDFromContext(c); if !ok { c.JSON(http.StatusUnauthorized,errorBody{Error:"unauthorized"}); return }; inventoryID,err:=uuid.Parse(c.Param("id")); if err!=nil { c.JSON(http.StatusBadRequest,errorBody{Error:"invalid inventory id"}); return }; var req CreateItemRequest; if c.ShouldBindJSON(&req)!=nil { c.JSON(http.StatusBadRequest,errorBody{Error:"invalid JSON body"}); return }; if msg:=ValidateCreateItem(&req); msg!="" { c.JSON(http.StatusBadRequest,errorBody{Error:msg}); return }; item,err:=h.svc.AddItem(userID,inventoryID,req.Name,req.Quantity,req.Category); if err!=nil { h.writeServiceError(c,err); return }; c.JSON(http.StatusCreated,item)
}
func (h *Handler) UpdateItem(c *gin.Context) {
	userID,ok:=auth.UserIDFromContext(c); if !ok { c.JSON(http.StatusUnauthorized,errorBody{Error:"unauthorized"}); return }; inventoryID,err:=uuid.Parse(c.Param("id")); if err!=nil { c.JSON(http.StatusBadRequest,errorBody{Error:"invalid inventory id"}); return }; itemID,err:=uuid.Parse(c.Param("itemId")); if err!=nil { c.JSON(http.StatusBadRequest,errorBody{Error:"invalid item id"}); return }; var req UpdateItemRequest; if c.ShouldBindJSON(&req)!=nil { c.JSON(http.StatusBadRequest,errorBody{Error:"invalid JSON body"}); return }; if msg:=ValidateUpdateItem(&req); msg!="" { c.JSON(http.StatusBadRequest,errorBody{Error:msg}); return }; item,err:=h.svc.UpdateItem(userID,inventoryID,itemID,req.Name,req.Quantity!=nil,req.Quantity,req.Category!=nil,req.Category); if err!=nil { h.writeServiceError(c,err); return }; c.JSON(http.StatusOK,item)
}
func (h *Handler) DeleteItem(c *gin.Context) {
	userID,ok:=auth.UserIDFromContext(c); if !ok { c.JSON(http.StatusUnauthorized,errorBody{Error:"unauthorized"}); return }; inventoryID,err:=uuid.Parse(c.Param("id")); if err!=nil { c.JSON(http.StatusBadRequest,errorBody{Error:"invalid inventory id"}); return }; itemID,err:=uuid.Parse(c.Param("itemId")); if err!=nil { c.JSON(http.StatusBadRequest,errorBody{Error:"invalid item id"}); return }; if err:=h.svc.DeleteItem(userID,inventoryID,itemID); err!=nil { h.writeServiceError(c,err); return }; c.Status(http.StatusNoContent)
}
func (h *Handler) writeServiceError(c *gin.Context, err error) {
	switch { case errors.Is(err,ErrNotFound): c.JSON(http.StatusNotFound,errorBody{Error:"not found"}); case errors.Is(err,ErrCannotDeleteDefault): c.JSON(http.StatusBadRequest,errorBody{Error:"cannot delete default inventory"}); default: log.Printf("inventory error: %v",err); c.JSON(http.StatusInternalServerError,errorBody{Error:"internal error"}) }
}
