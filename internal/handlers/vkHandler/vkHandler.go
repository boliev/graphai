package vkHandler

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type Handler struct {
	vkSecureKey string
}

func NewHandler(vkSecureKey string) *Handler {
	return &Handler{
		vkSecureKey: vkSecureKey,
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

	userID, err := strconv.ParseInt(form["user_id"], 10, 64)
	if err != nil {
		h.writeVKError(w, http.StatusBadRequest, 100, "invalid user_id", true)
		return
	}

	itemID, err := strconv.ParseInt(form["item_id"], 10, 64)
	if err != nil {
		h.writeVKError(w, http.StatusBadRequest, 100, "invalid item_id", true)
		return
	}

	// КРИТИЧНО:
	// VK может ретраить order_status_change для того же order_id.
	// Значит обработка должна быть ИДЕМПОТЕНТНОЙ.
	//
	// Ниже псевдологика:
	//
	// 1. Начать транзакцию
	// 2. Найти платеж по vk_order_id = orderID
	// 3. Если уже есть:
	//      вернуть тот же app_order_id, что возвращали раньше
	// 4. Если нет:
	//      - проверить товар
	//      - начислить пользователю баланс/кредиты
	//      - сохранить платеж и app_order_id
	// 5. Commit
	//
	// Здесь просто заглушка:
	//appOrderID, err := processPurchaseIdempotently(orderID, userID, itemID)
	//if err != nil {
	//	h.writeVKError(w, http.StatusOK, 100, "failed to process order", true)
	//	return
	//}
	_ = userID
	_ = itemID

	appOrderID := int64(rand.Intn(1000))
	var resp vkSuccessChargeable
	resp.Response.OrderID = orderID
	resp.Response.AppOrderID = appOrderID

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
