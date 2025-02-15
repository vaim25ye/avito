package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/vaim25ye/avito/internal/cache"
	"github.com/vaim25ye/avito/internal/repository"
)

type Handler struct {
	repo  repository.Repo // <-- интерфейс!
	cache *cache.Cache
}

func NewHandler(r repository.Repo, c *cache.Cache) *Handler {
	return &Handler{repo: r, cache: c}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	type request struct {
		Name     string `json:"name"`
		Password string `json:"password"`
		Balance  int    `json:"balance"`
	}
	var req request
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

	info, ok := h.cache.GetUserInfoByID(userID)
	if !ok {
		http.Error(w, "user not found in cache", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(info)
}

func (h *Handler) Transfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	type reqBody struct {
		FromUser int `json:"from_user"`
		ToUser   int `json:"to_user"`
		Amount   int `json:"amount"`
	}
	var body reqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	err := h.repo.Transfer(r.Context(), body.FromUser, body.ToUser, body.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

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
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "purchase success"})
}
