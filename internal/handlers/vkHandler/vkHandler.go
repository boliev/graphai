package vkHandler

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/boliev/graphai/internal/domain/order"
	"github.com/boliev/graphai/internal/domain/user"
	"github.com/jackc/pgx/v5"
)

type txManager interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}
type Handler struct {
	vkSecureKey  string
	txManager    txManager
	orderService *order.Service
	userService  *user.Service
}

func NewHandler(vkSecureKey string, txManager txManager, orderService *order.Service, userService *user.Service) *Handler {
	return &Handler{
		vkSecureKey:  vkSecureKey,
		txManager:    txManager,
		orderService: orderService,
		userService:  userService,
	}
}

func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.writeVKError(w, http.StatusBadRequest, 100, "invalid form body", true)
		return
	}

	// VK шлет POST form fields.
	form := map[string]string{}
	for k := range r.PostForm {
		form[k] = r.PostForm.Get(k)
	}

	if !h.validateVKSignature(form, h.vkSecureKey) {
		h.writeVKError(w, http.StatusForbidden, 10, "bad signature", true)
		return
	}

	notificationType := form["notification_type"]

	switch notificationType {
	case "get_item", "get_item_test":
		h.handleGetItem(w, form)

	case "order_status_change", "order_status_change_test":
		h.handleOrderStatusChange(w, form)

	default:
		h.writeVKError(w, http.StatusBadRequest, 100, "unknown notification_type", true)
	}
}

func (h *Handler) handleGetItem(w http.ResponseWriter, form map[string]string) {
	itemName := form["item"]

	item := getProduct(itemName)
	if item == nil {
		h.writeVKError(w, http.StatusOK, 20, "item not found", true)
		return
	}

	var resp vkSuccessGetItem
	resp.Response.ItemID = item.ID
	resp.Response.Title = item.Title
	resp.Response.PhotoURL = item.PhotoURL
	resp.Response.Price = item.Price

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleOrderStatusChange(w http.ResponseWriter, form map[string]string) {
	ctx := context.Background()
	// Обычно нас интересует только chargeable.
	// Именно здесь нужно атомарно выдать товар/кредиты пользователю.
	status := form["status"]
	if status != "chargeable" {
		h.writeVKError(w, http.StatusOK, 100, "unsupported status", true)
		return
	}

	orderID, err := strconv.ParseInt(form["order_id"], 10, 64)
	if err != nil {
		h.writeVKError(w, http.StatusBadRequest, 100, "invalid order_id", true)
		return
	}

	vkUserID, err := strconv.ParseInt(form["user_id"], 10, 64)
	if err != nil {
		h.writeVKError(w, http.StatusBadRequest, 100, "invalid user_id", true)
		return
	}

	usr, err := h.userService.FindByVKID(ctx, vkUserID)
	if err != nil {
		h.writeVKError(w, http.StatusBadRequest, 100, "user doesn't exists", true)
		log.Printf("user %v doesn't exists", vkUserID)
		return
	}

	itemID, err := strconv.ParseInt(form["item_id"], 10, 64)
	if err != nil {
		h.writeVKError(w, http.StatusBadRequest, 100, "invalid item_id", true)
		return
	}

	product := getProductById(itemID)
	if product == nil {
		h.writeVKError(w, http.StatusOK, 20, "item not found", true)
		log.Printf("item %d not found", itemID)
		return
	}

	tx, err := h.txManager.Begin(ctx)
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			log.Printf("failed to rollback transaction: %s", err)
		}
	}()

	if err != nil {
		h.writeVKError(w, http.StatusInternalServerError, 100, "error begin transaction", true)
		log.Printf("error begin transaction: %s", err.Error())
		return
	}

	ord, err := h.orderService.GetOrderByVkOrderIdTx(ctx, tx, orderID)
	if err != nil {
		h.writeVKError(w, http.StatusInternalServerError, 100, "error get order", true)
		log.Printf("error get order %d: %s", orderID, err.Error())
		return
	}

	if ord != nil {
		vkOrderId, err := strconv.ParseInt(ord.VkOrderID, 10, 64)
		if err != nil {
			h.writeVKError(w, http.StatusBadRequest, 100, "invalid order_id", true)
			log.Printf("invalid order_id: %d", ord.VkOrderID)
			return
		}

		var resp vkSuccessChargeable
		resp.Response.OrderID = vkOrderId
		resp.Response.AppOrderID = ord.ID

		h.writeJSON(w, http.StatusOK, resp)
	}

	err = h.userService.IncreaseCreditsTx(ctx, tx, usr.ID, product.Credits)
	if err != nil {
		h.writeVKError(w, http.StatusInternalServerError, 100, "error increase credits", true)
		log.Printf("error increase credits for user %d: %s", usr.ID, err.Error())
		return
	}

	newOrder, err := h.orderService.UpsertTx(ctx, tx, &order.Order{
		VkOrderID: string(orderID),
		UserID:    usr.ID,
		Product:   product.Name,
	})

	if err != nil {
		h.writeVKError(w, http.StatusInternalServerError, 100, "error upsert order", true)
		log.Printf("error upsert order: %s", err.Error())
		return
	}
	err = tx.Commit(ctx)
	if err != nil {
		h.writeVKError(w, http.StatusInternalServerError, 100, "error commit transaction", true)
		log.Printf("error commit transaction: %s", err.Error())
		return
	}

	var resp vkSuccessChargeable
	resp.Response.OrderID = orderID
	resp.Response.AppOrderID = newOrder.ID

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) validateVKSignature(form map[string]string, secureKey string) bool {
	sig := form["sig"]
	if sig == "" {
		return false
	}

	// sig исключаем из подписи.
	keys := make([]string, 0, len(form))
	for k := range form {
		if k == "sig" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(form[k])
	}

	sum := md5.Sum([]byte(b.String() + secureKey))
	expected := hex.EncodeToString(sum[:])

	return expected == sig
}

func (h *Handler) writeJSON(w http.ResponseWriter, statusCode int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handler) writeVKError(w http.ResponseWriter, statusCode, code int, msg string, critical bool) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	var resp vkError
	resp.Error.Code = code
	resp.Error.Msg = msg
	resp.Error.Critical = critical

	_ = json.NewEncoder(w).Encode(resp)
}
