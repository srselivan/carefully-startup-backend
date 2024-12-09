package models

type GameState int8

const (
	GameStateClosed GameState = iota - 1
	GameStatePaused
	GameStateOpened
	GameStateStarted
	GameStateStopGenerally
)

type TradeState int8

const (
	TradeStateNotStarted TradeState = iota
	TradeStateStarted
)

type Game struct {
	State        GameState
	CurrentRound int
	TradeState   TradeState
	CurrentGame  int64
}
