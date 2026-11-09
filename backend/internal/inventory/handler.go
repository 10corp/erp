package inventory

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type HTTPReceiveStockRequest struct {
	CompanyID       uuid.UUID `json:"company_id"`
	ItemCode        string    `json:"item_code"`
	ItemName        string    `json:"item_name"`
	FromWarehouseID uuid.UUID `json:"from_warehouse_id"` // Supplier Virtual Location
	ToWarehouseID   uuid.UUID `json:"to_warehouse_id"`   // Internal Warehouse
	Qty             string    `json:"qty"`
	Rate            string    `json:"rate"`
	VoucherNo       string    `json:"voucher_no"`
	Remarks         string    `json:"remarks"`
}

func (h *Handler) ReceiveStock(w http.ResponseWriter, r *http.Request) {
	var req HTTPReceiveStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json payload"})
		return
	}

	qty, err := decimal.NewFromString(req.Qty)
	if err != nil || qty.IsZero() || qty.IsNegative() {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "qty must be greater than zero"})
		return
	}

	rate, err := decimal.NewFromString(req.Rate)
	if err != nil || rate.IsNegative() {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "rate must be non-negative"})
		return
	}

	entry, err := h.service.ReceiveStock(r.Context(), ReceiveStockInput{
		CompanyID:       req.CompanyID,
		ItemCode:        req.ItemCode,
		ItemName:        req.ItemName,
		FromWarehouseID: req.FromWarehouseID,
		ToWarehouseID:   req.ToWarehouseID,
		Qty:             qty,
		Rate:            rate,
		VoucherType:     "Purchase Receipt",
		VoucherID:       uuid.New(),
		VoucherNo:       req.VoucherNo,
		Remarks:         req.Remarks,
	})
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusCreated, entry)
}

type HTTPDeliverStockRequest struct {
	CompanyID           uuid.UUID `json:"company_id"`
	CustomerName        string    `json:"customer_name"`
	PostingDate         string    `json:"posting_date"` // YYYY-MM-DD
	FromWarehouseID     uuid.UUID `json:"from_warehouse_id"`
	ToWarehouseID       uuid.UUID `json:"to_warehouse_id"`
	COGSAccountID       uuid.UUID `json:"cogs_account_id"`
	StockAssetAccountID uuid.UUID `json:"stock_asset_account_id"`
	COGSAccountCode     string    `json:"cogs_account_code"`
	StockAssetCode      string    `json:"stock_asset_code"`
	Items               []struct {
		ItemCode string `json:"item_code"`
		ItemName string `json:"item_name"`
		Qty      string `json:"qty"`
	} `json:"items"`
	Remarks string `json:"remarks"`
}

func (h *Handler) DeliverStock(w http.ResponseWriter, r *http.Request) {
	var req HTTPDeliverStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json payload"})
		return
	}

	postingDate, err := time.Parse("2006-01-02", req.PostingDate)
	if err != nil {
		postingDate = time.Now()
	}

	var items []DeliverItemInput
	for _, it := range req.Items {
		q, err := decimal.NewFromString(it.Qty)
		if err != nil || q.IsZero() || q.IsNegative() {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid item quantity"})
			return
		}
		items = append(items, DeliverItemInput{
			ItemCode: it.ItemCode,
			ItemName: it.ItemName,
			Qty:      q,
		})
	}

	dn, glEntries, err := h.service.DeliverStock(r.Context(), DeliverStockRequest{
		CompanyID:           req.CompanyID,
		CustomerName:        req.CustomerName,
		PostingDate:         postingDate,
		FromWarehouseID:     req.FromWarehouseID,
		ToWarehouseID:       req.ToWarehouseID,
		COGSAccountID:       req.COGSAccountID,
		StockAssetAccountID: req.StockAssetAccountID,
		COGSAccountCode:     req.COGSAccountCode,
		StockAssetCode:      req.StockAssetCode,
		Items:               items,
		Remarks:             req.Remarks,
	})
	if err != nil {
		if err == ErrInsufficientStock {
			respondJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":          "posted",
		"delivery_note":   dn,
		"cogs_gl_entries": len(glEntries),
		"message":         "Delivery Note submitted, stock deducted, and Perpetual COGS automatically posted to General Ledger.",
	})
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
