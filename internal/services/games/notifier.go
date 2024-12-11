package games

import (
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog"
	"investment-game-backend/internal/models"
	"sync"
)

type TeamsNotifier struct {
	conns map[*websocket.Conn]struct{}
	mx    sync.Mutex
	log   *zerolog.Logger
}

func NewTeamsNotifier(log *zerolog.Logger) *TeamsNotifier {
	return &TeamsNotifier{
		log:   log,
		conns: make(map[*websocket.Conn]struct{}),
		mx:    sync.Mutex{},
	}
}

type tradePeriodChangedMessage struct {
	IsTradeStage bool `json:"isTradeStage"`
}

func (n *TeamsNotifier) NotifyTradePeriodChanged(isTrade bool) {
	msg, _ := jsoniter.Marshal(tradePeriodChangedMessage{IsTradeStage: isTrade})

	n.mx.Lock()
	defer n.mx.Unlock()

	n.log.Trace().
		Bool("is_trade_period", isTrade).
		Int("conns_count", len(n.conns)).
		Msg("notify: trade period changed")

	for conn := range n.conns {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			n.RemoveConnection(conn)
			_ = conn.Close()
		}
	}
}

type roundPeriodChangedMessage struct {
	IsRoundStage bool `json:"isRoundStage"`
}

func (n *TeamsNotifier) NotifyRoundPeriodChanged(isRound bool) {
	msg, _ := jsoniter.Marshal(roundPeriodChangedMessage{IsRoundStage: isRound})

	n.mx.Lock()
	defer n.mx.Unlock()

	n.log.Trace().
		Bool("is_round_period", isRound).
		Int("conns_count", len(n.conns)).
		Msg("notify: round period changed")

	for conn := range n.conns {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			n.RemoveConnection(conn)
			_ = conn.Close()
		}
	}
}

type gameStateChangedMessage struct {
	GameState models.GameState `json:"gameState"`
}

func (n *TeamsNotifier) NotifyGameStateChanged(state models.GameState) {
	msg, _ := jsoniter.Marshal(gameStateChangedMessage{GameState: state})

	n.mx.Lock()
	defer n.mx.Unlock()

	n.log.Trace().
		Int("game_state", int(state)).
		Int("conns_count", len(n.conns)).
		Msg("notify: game state changed")

	for conn := range n.conns {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			n.RemoveConnection(conn)
			_ = conn.Close()
		}
	}
}

func (n *TeamsNotifier) RegisterConnection(conn *websocket.Conn) {
	n.mx.Lock()
	defer n.mx.Unlock()
	n.conns[conn] = struct{}{}
}

func (n *TeamsNotifier) RemoveConnection(conn *websocket.Conn) {
	n.mx.Lock()
	defer n.mx.Unlock()
	delete(n.conns, conn)
}
