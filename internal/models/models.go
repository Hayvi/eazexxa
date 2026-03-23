package models

import (
	"time"

	"github.com/betpro/server/pkg/money"
)

type User struct {
	ID        string    `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"`
	Role      string    `json:"role" db:"role"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	Balance   money.Money `json:"balance" db:"balance"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Profile struct {
	ID       string      `json:"id" db:"id"`
	Balance  money.Money `json:"balance" db:"balance"`
	IsActive bool        `json:"is_active" db:"is_active"`
	Role     string      `json:"role" db:"role"`
}

type Bet struct {
	ID         string      `json:"id" db:"id"`
	UserID     string      `json:"user_id" db:"user_id"`
	GameID     string      `json:"game_id" db:"game_id"`
	MarketID   string      `json:"market_id" db:"market_id"`
	OutcomeID  string      `json:"outcome_id" db:"outcome_id"`
	Odds       float64     `json:"odds" db:"odds"`
	Stake      money.Money `json:"stake" db:"stake"`
	Status     string      `json:"status" db:"status"`
	CreatedAt  time.Time   `json:"created_at" db:"created_at"`
}

type Ticket struct {
	ID         string      `json:"id" db:"id"`
	UserID     string      `json:"user_id" db:"user_id"`
	Stake      money.Money `json:"stake" db:"stake"`
	TotalOdds  float64     `json:"total_odds" db:"total_odds"`
	ModelType  string      `json:"model_type" db:"model_type"`
	SystemK    *int        `json:"system_k,omitempty" db:"system_k"`
	Status     string      `json:"status" db:"status"`
	Payout     money.Money `json:"payout" db:"payout"`
	CreatedAt  time.Time   `json:"created_at" db:"created_at"`
	SettledAt  *time.Time  `json:"settled_at,omitempty" db:"settled_at"`
}

type Game struct {
	ID            string    `json:"id" db:"id"`
	SportID       string    `json:"sport_id" db:"sport_id"`
	CompetitionID string    `json:"competition_id" db:"competition_id"`
	HomeTeam      string    `json:"home_team" db:"home_team"`
	AwayTeam      string    `json:"away_team" db:"away_team"`
	StartTime     time.Time `json:"start_time" db:"start_time"`
	Status        string    `json:"status" db:"status"`
}

type Result struct {
	GameID     string     `json:"game_id" db:"game_id"`
	HomeScore  int        `json:"home_score" db:"home_score"`
	AwayScore  int        `json:"away_score" db:"away_score"`
	Status     string     `json:"status" db:"status"`
	SettledAt  *time.Time `json:"settled_at,omitempty" db:"settled_at"`
}

type Withdrawal struct {
	ID        string      `json:"id" db:"id"`
	UserID    string      `json:"user_id" db:"user_id"`
	Amount    money.Money `json:"amount" db:"amount"`
	Status    string      `json:"status" db:"status"`
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
	ExpiresAt *time.Time  `json:"expires_at,omitempty" db:"expires_at"`
}

const (
	TicketStatusPending  = "pending"
	TicketStatusWon      = "won"
	TicketStatusLost     = "lost"
	TicketStatusCashout  = "cashout"
	TicketStatusCanceled = "canceled"

	BetStatusPending  = "pending"
	BetStatusWon      = "won"
	BetStatusLost     = "lost"
	BetStatusVoid     = "void"

	WithdrawalStatusPending  = "pending"
	WithdrawalStatusApproved = "approved"
	WithdrawalStatusRejected = "rejected"
	WithdrawalStatusExpired  = "expired"

	RoleUser       = "user"
	RoleAdmin      = "admin"
	RoleSuperAdmin = "super_admin"
)
