package models

type GameState int8

const (
	GameStateClosed GameState = iota - 1
	GameStatePaused
	GameStateOpened
)

type RoundState int8

const (
	RoundStateNotStarted RoundState = iota - 1
	RoundStatePaused
	RoundStateStarted
)

type Game struct {
	State        GameState
	CurrentRound int
	RoundState   RoundState
}
