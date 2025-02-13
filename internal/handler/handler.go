package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/vaim25ye/avito/internal/repository"
)

type Handler struct {
	repo *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{repo: r}
}

// --------------------
// POST /users
// JSON: { "name":"Вася", "password":"secret", "balance":1000 }
// --------------------
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	type createUserReq struct {
		Name     string `json:"name"`
		Password string `json:"password"`
		Balance  int    `json:"balance"`
	}
	var req createUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	user, err := h.repo.CreateUser(r.Context(), req.Name, req.Password, req.Balance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(user)
}

// --------------------
// GET /get_user?id=1
// --------------------
func (h *Handler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(user)
}

// --------------------
// POST /transfer
// body: { "from_user":1, "to_user":2, "amount":300 }
// --------------------
func (h *Handler) Transfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type transferReq struct {
		FromUser int `json:"from_user"`
		ToUser   int `json:"to_user"`
		Amount   int `json:"amount"`
	}
	var req transferReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	err := h.repo.Transfer(r.Context(), req.FromUser, req.ToUser, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// --------------------
// POST /purchase
// body: { "user_id":1, "merch_id":2, "amount":3 }
// --------------------
func (h *Handler) PurchaseMerch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type purchaseReq struct {
		UserID  int `json:"user_id"`
		MerchID int `json:"merch_id"`
		Amount  int `json:"amount"`
	}
	var req purchaseReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	err := h.repo.PurchaseMerch(r.Context(), req.UserID, req.MerchID, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "purchase success",
	})
}
