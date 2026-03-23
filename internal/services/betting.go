package services

import (
	"context"
	"fmt"

	"github.com/betpro/server/internal/db"
	"github.com/betpro/server/internal/models"
	"github.com/betpro/server/pkg/money"
	"github.com/jackc/pgx/v5"
)

type BetService struct {
	db  *db.DB
	hub BroadcastHub
}

type BroadcastHub interface {
	BroadcastToUser(userID, msgType string, payload interface{})
}

func NewBetService(database *db.DB, hub BroadcastHub) *BetService {
	return &BetService{
		db:  database,
		hub: hub,
	}
}

type PlaceBetRequest struct {
	Stake     money.Money `json:"stake"`
	Bets      []BetInput  `json:"bets"`
	ModelType string      `json:"model_type"`
}

type BetInput struct {
	GameID    string  `json:"game_id"`
	MarketID  string  `json:"market_id"`
	OutcomeID string  `json:"outcome_id"`
	Odds      float64 `json:"odds"`
}

type PlaceBetResponse struct {
	TicketID string      `json:"ticket_id"`
	Stake    money.Money `json:"stake"`
	TotalOdds float64    `json:"total_odds"`
	PotentialWin money.Money `json:"potential_win"`
}

func (s *BetService) PlaceBet(ctx context.Context, userID string, req PlaceBetRequest) (*PlaceBetResponse, error) {
	if req.Stake <= 0 {
		return nil, fmt.Errorf("invalid stake")
	}

	if len(req.Bets) == 0 {
		return nil, fmt.Errorf("no bets provided")
	}

	if req.ModelType != "accumulator" && req.ModelType != "single" {
		return nil, fmt.Errorf("invalid model type")
	}

	totalOdds := 1.0
	for _, bet := range req.Bets {
		if bet.Odds <= 1.0 {
			return nil, fmt.Errorf("invalid odds")
		}
		totalOdds *= bet.Odds
	}

	potentialWin := req.Stake.Mul(totalOdds)

	var ticketID string
	err := s.db.WithTransaction(ctx, func(tx pgx.Tx) error {
		var balance money.Money
		err := tx.QueryRow(ctx, `
			SELECT balance FROM users WHERE id = $1 FOR UPDATE
		`, userID).Scan(&balance)
		if err != nil {
			return fmt.Errorf("get balance: %w", err)
		}

		if balance < req.Stake {
			return models.ErrInsufficientFunds
		}

		_, err = tx.Exec(ctx, `
			UPDATE users SET balance = balance - $1 WHERE id = $2
		`, req.Stake, userID)
		if err != nil {
			return fmt.Errorf("deduct balance: %w", err)
		}

		err = tx.QueryRow(ctx, `
			INSERT INTO tickets (user_id, stake, total_odds, model_type, status, payout)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`, userID, req.Stake, totalOdds, req.ModelType, models.TicketStatusPending, 0).Scan(&ticketID)
		if err != nil {
			return fmt.Errorf("create ticket: %w", err)
		}

		for _, bet := range req.Bets {
			_, err = tx.Exec(ctx, `
				INSERT INTO bets (ticket_id, user_id, game_id, market_id, outcome_id, odds, stake, status)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			`, ticketID, userID, bet.GameID, bet.MarketID, bet.OutcomeID, bet.Odds, req.Stake, models.BetStatusPending)
			if err != nil {
				return fmt.Errorf("create bet: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	s.hub.BroadcastToUser(userID, "bet_placed", map[string]interface{}{
		"ticket_id": ticketID,
		"stake":     req.Stake,
	})

	return &PlaceBetResponse{
		TicketID:     ticketID,
		Stake:        req.Stake,
		TotalOdds:    totalOdds,
		PotentialWin: potentialWin,
	}, nil
}

func (s *BetService) GetTickets(ctx context.Context, userID string, limit, offset int) ([]*models.Ticket, error) {
	query := `
		SELECT id, user_id, stake, total_odds, model_type, status, payout, created_at, settled_at
		FROM tickets
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*models.Ticket
	for rows.Next() {
		ticket := &models.Ticket{}
		err := rows.Scan(
			&ticket.ID, &ticket.UserID, &ticket.Stake, &ticket.TotalOdds,
			&ticket.ModelType, &ticket.Status, &ticket.Payout,
			&ticket.CreatedAt, &ticket.SettledAt,
		)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, ticket)
	}

	return tickets, nil
}
