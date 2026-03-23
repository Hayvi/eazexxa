-- BetPro VPS Database Schema
-- Run this on your PostgreSQL instance

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Schema migration ledger (for incremental upgrades on existing DBs)
CREATE TABLE IF NOT EXISTS _migrations (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Role enum
CREATE TYPE app_role AS ENUM ('super_admin', 'admin', 'sub_admin', 'user');

-- Profiles (users)
CREATE TABLE profiles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  plain_pw TEXT,
  role app_role NOT NULL,
  created_by UUID REFERENCES profiles(id),
  balance NUMERIC(12,2) NOT NULL DEFAULT 0,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_profiles_role ON profiles(role);
CREATE INDEX idx_profiles_created_by ON profiles(created_by);
CREATE INDEX idx_profiles_is_active ON profiles(is_active);

-- Refresh tokens (rotating, hashed-at-rest)
CREATE TABLE refresh_tokens (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ,
  replaced_by_token_hash TEXT,
  user_agent TEXT,
  ip_address TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX idx_refresh_tokens_active ON refresh_tokens(user_id, revoked_at, expires_at);

-- Transactions
CREATE TABLE transactions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  sender_id UUID REFERENCES profiles(id),
  receiver_id UUID REFERENCES profiles(id),
  amount NUMERIC(12,2) NOT NULL,
  type TEXT NOT NULL CHECK (type IN ('transfer', 'credit', 'debit')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_transactions_sender ON transactions(sender_id);
CREATE INDEX idx_transactions_receiver ON transactions(receiver_id);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);

-- Withdrawal requests
CREATE TABLE withdrawal_requests (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  requester_id UUID NOT NULL REFERENCES profiles(id),
  target_user_id UUID NOT NULL REFERENCES profiles(id),
  amount NUMERIC(12,2) NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'expired')),
  approved_by UUID REFERENCES profiles(id),
  approved_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_withdrawal_requests_status ON withdrawal_requests(status);
CREATE INDEX idx_withdrawal_requests_requester ON withdrawal_requests(requester_id);
CREATE INDEX idx_withdrawal_requests_target ON withdrawal_requests(target_user_id);

-- Bet slips
CREATE TABLE bet_slips (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES profiles(id),
  model_type TEXT NOT NULL CHECK (model_type IN ('accumulator', 'system')),
  system_k INTEGER,
  total_legs INTEGER NOT NULL DEFAULT 0,
  total_lines INTEGER NOT NULL DEFAULT 0,
  total_stake NUMERIC(12,2) NOT NULL,
  accumulator_odds NUMERIC(70,4),
  potential_win NUMERIC(12,2),
  bonus_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
  bonus_percent NUMERIC(6,2) NOT NULL DEFAULT 0,
  final_payout NUMERIC(12,2),
  settled_win NUMERIC(12,2) NOT NULL DEFAULT 0,
  settled_win_includes_cashout BOOLEAN NOT NULL DEFAULT false,
  total_cashed_out_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
  last_cashout_at TIMESTAMPTZ,
  settled_at TIMESTAMPTZ,
  status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'won', 'lost', 'cancelled', 'half_won', 'half_lost', 'cashed_out')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_bet_slips_user ON bet_slips(user_id);
CREATE INDEX idx_bet_slips_status ON bet_slips(status);
CREATE INDEX idx_bet_slips_user_settled_at ON bet_slips(user_id, settled_at DESC);

-- Bet lines (line = one acca line or one system combination line)
CREATE TABLE bet_lines (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  slip_id UUID NOT NULL REFERENCES bet_slips(id) ON DELETE CASCADE,
  line_index INTEGER NOT NULL,
  line_stake NUMERIC(12,2) NOT NULL,
  line_odds NUMERIC(70,6),
  line_potential_win NUMERIC(12,2),
  settled_win NUMERIC(12,2) NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'won', 'lost', 'cancelled', 'half_won', 'half_lost', 'cashed_out')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (slip_id, line_index)
);

CREATE INDEX idx_bet_lines_slip_id ON bet_lines(slip_id);
CREATE INDEX idx_bet_lines_status ON bet_lines(status);

-- Individual bets
CREATE TABLE bets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES profiles(id),
  slip_id UUID REFERENCES bet_slips(id) ON DELETE CASCADE,
  line_id UUID REFERENCES bet_lines(id) ON DELETE CASCADE,
  selection_index INTEGER,
  match_id TEXT,
  bet_type TEXT,
  odds NUMERIC(10,4) NOT NULL,
  stake NUMERIC(12,2) NOT NULL,
  potential_win NUMERIC(12,2),
  status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'won', 'lost', 'cancelled', 'half_won', 'half_lost', 'cashed_out')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  -- Rich game data
  home_team TEXT,
  away_team TEXT,
  league_name TEXT,
  sport_name TEXT,
  match_type TEXT,
  market_name TEXT,
  selection_name TEXT,
  match_date TIMESTAMPTZ,
  region_name TEXT,
  -- Settlement matching columns
  market_key TEXT,
  event_key TEXT,
  event_base NUMERIC(6,2)
);

CREATE INDEX idx_bets_user ON bets(user_id);
CREATE INDEX idx_bets_slip_id ON bets(slip_id);
CREATE INDEX idx_bets_line_id ON bets(line_id);
CREATE INDEX idx_bets_created_at ON bets(created_at DESC);
CREATE INDEX idx_bets_match_market ON bets(match_id, market_key);
CREATE INDEX idx_bets_pending ON bets(status) WHERE status = 'pending';

-- Cashout audit trail
CREATE TABLE bet_cashout_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  slip_id UUID NOT NULL REFERENCES bet_slips(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
  cashout_type TEXT NOT NULL CHECK (cashout_type IN ('full', 'partial')),
  requested_amount NUMERIC(12,2) NOT NULL,
  offered_amount NUMERIC(12,2) NOT NULL,
  accepted_amount NUMERIC(12,2) NOT NULL,
  applied_fraction NUMERIC(10,6) NOT NULL,
  stake_before NUMERIC(12,2) NOT NULL,
  stake_after NUMERIC(12,2) NOT NULL,
  quote_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_bet_cashout_events_slip_created
  ON bet_cashout_events(slip_id, created_at DESC);
CREATE INDEX idx_bet_cashout_events_user_created
  ON bet_cashout_events(user_id, created_at DESC);

-- Settlement audit trail (for dispute/debug workflows)
CREATE TABLE settlement_audit_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  run_id TEXT NOT NULL,
  game_id TEXT NOT NULL,
  bet_id UUID REFERENCES bets(id) ON DELETE SET NULL,
  user_id UUID REFERENCES profiles(id) ON DELETE SET NULL,
  status_before TEXT NOT NULL,
  status_after TEXT NOT NULL,
  reason TEXT,
  match_strategy TEXT,
  bet_market_family TEXT,
  settlement_market_family TEXT,
  market_name TEXT,
  selection_name TEXT,
  canonical_bet_market TEXT,
  canonical_settlement_market TEXT,
  settlement_market TEXT,
  settlement_winners JSONB NOT NULL DEFAULT '[]'::jsonb,
  unmatched_reasons JSONB NOT NULL DEFAULT '{}'::jsonb,
  sample_context JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_settlement_audit_game_created
  ON settlement_audit_logs(game_id, created_at DESC);
CREATE INDEX idx_settlement_audit_bet_created
  ON settlement_audit_logs(bet_id, created_at DESC);
CREATE INDEX idx_settlement_audit_user_created
  ON settlement_audit_logs(user_id, created_at DESC);
CREATE INDEX idx_settlement_audit_run
  ON settlement_audit_logs(run_id);

-- Presence sessions
CREATE TABLE presence_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
  session_id TEXT NOT NULL,
  device_id TEXT NOT NULL,
  ip_address TEXT,
  country TEXT,
  city TEXT,
  lat DOUBLE PRECISION,
  lng DOUBLE PRECISION,
  started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  ended_at TIMESTAMPTZ,
  end_reason TEXT,
  UNIQUE(user_id, session_id)
);

CREATE INDEX idx_presence_sessions_last_seen ON presence_sessions(last_seen_at);
CREATE INDEX idx_presence_sessions_ended ON presence_sessions(ended_at);

-- Presence history (archived sessions)
CREATE TABLE presence_history (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  session_id TEXT NOT NULL,
  device_id TEXT NOT NULL,
  ip_address TEXT,
  country TEXT,
  city TEXT,
  lat DOUBLE PRECISION,
  lng DOUBLE PRECISION,
  started_at TIMESTAMPTZ NOT NULL,
  last_seen_at TIMESTAMPTZ NOT NULL,
  ended_at TIMESTAMPTZ,
  end_reason TEXT,
  archived_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_presence_history_user ON presence_history(user_id, started_at DESC);

-- Performance indexes for 5K+ users
CREATE INDEX idx_profiles_active ON profiles(is_active) WHERE is_active = true;
CREATE INDEX idx_profiles_role_created ON profiles(role, created_at DESC) WHERE is_active = true;
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_type_created ON transactions(type, created_at DESC);
CREATE INDEX idx_presence_sessions_user ON presence_sessions(user_id);
CREATE INDEX idx_presence_sessions_session ON presence_sessions(session_id);
CREATE INDEX idx_presence_history_ended ON presence_history(ended_at DESC);

-- Updated_at trigger
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_profiles_updated_at
BEFORE UPDATE ON profiles
FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Prevent transactions from being modified (immutable ledger)
CREATE OR REPLACE FUNCTION prevent_transactions_update()
RETURNS TRIGGER AS $$
BEGIN
  RAISE EXCEPTION 'transactions are immutable';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_transactions_no_update
BEFORE UPDATE ON transactions
FOR EACH ROW EXECUTE FUNCTION prevent_transactions_update();

CREATE TRIGGER trg_transactions_no_delete
BEFORE DELETE ON transactions
FOR EACH ROW EXECUTE FUNCTION prevent_transactions_update();

-- Create initial super admin (password: changeme123)
-- Hash generated with: SELECT crypt('changeme123', gen_salt('bf'));
INSERT INTO profiles (username, password_hash, plain_pw, role, balance)
VALUES (
  'root_admin',
  '$2a$06$rKh8XQ8Y8Y8Y8Y8Y8Y8Y8.XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX',
  'changeme123',
  'super_admin',
  0
);

-- NOTE: Generate a real bcrypt hash for production!
-- You can use: node -e "console.log(require('bcrypt').hashSync('yourpassword', 10))"

-- BetSlip Codes (shareable betslips)
CREATE TABLE IF NOT EXISTS betslip_codes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code TEXT UNIQUE NOT NULL,
  created_by UUID REFERENCES profiles(id),
  bets JSONB NOT NULL,
  accumulator_odds NUMERIC(70,4),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at TIMESTAMPTZ NOT NULL,
  usage_count INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_betslip_codes_code ON betslip_codes(code);
CREATE INDEX IF NOT EXISTS idx_betslip_codes_expires ON betslip_codes(expires_at);
CREATE INDEX IF NOT EXISTS idx_betslip_codes_created_by ON betslip_codes(created_by);

-- Results finality tracking (first observed transition to results_live=0)
CREATE TABLE IF NOT EXISTS public.results_game_finality (
  game_id TEXT PRIMARY KEY,
  sport_id INTEGER,
  sport_name TEXT,
  team1_name TEXT,
  team2_name TEXT,
  game_date_ts BIGINT,
  last_results_live INTEGER,
  first_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  first_final_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_results_game_finality_first_final_at
  ON public.results_game_finality(first_final_at DESC);

CREATE INDEX IF NOT EXISTS idx_results_game_finality_last_seen_at
  ON public.results_game_finality(last_seen_at DESC);
