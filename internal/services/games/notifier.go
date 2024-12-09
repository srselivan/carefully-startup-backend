package games

import (
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"sync"
)

type TeamsNotifier struct {
	conns map[*websocket.Conn]struct{}
	mx    sync.Mutex
}

func NewTeamsNotifier() *TeamsNotifier {
	return &TeamsNotifier{
		conns: make(map[*websocket.Conn]struct{}),
		mx:    sync.Mutex{},
	}
}

type message struct {
	IsTradeStage bool `json:"isTradeStage"`
}

func (n *TeamsNotifier) NotifyTradePeriodChanged(isTrade bool) {
	msg, _ := jsoniter.Marshal(message{IsTradeStage: isTrade})

	n.mx.Lock()
	defer n.mx.Unlock()

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
