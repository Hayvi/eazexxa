package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/betpro/server/internal/middleware"
	"github.com/betpro/server/internal/services"
	"github.com/betpro/server/pkg/logger"
)

type BetHandler struct {
	betService *services.BetService
}

func NewBetHandler(betService *services.BetService) *BetHandler {
	return &BetHandler{betService: betService}
}

func (h *BetHandler) PlaceBet(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req services.PlaceBetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result, err := h.betService.PlaceBet(r.Context(), claims.UserID, req)
	if err != nil {
		logger.Error("failed to place bet", "error", err, "user_id", claims.UserID)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, result)
}

func (h *BetHandler) GetTickets(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	tickets, err := h.betService.GetTickets(r.Context(), claims.UserID, limit, offset)
	if err != nil {
		logger.Error("failed to get tickets", "error", err, "user_id", claims.UserID)
		respondError(w, http.StatusInternalServerError, "Server error")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"tickets": tickets,
		"limit":   limit,
		"offset":  offset,
	})
}
