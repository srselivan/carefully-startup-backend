package v1

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (r *Router) initWebsocketRouter(router chi.Router) {
	router.Route("/websocket", func(subRouter chi.Router) {
		subRouter.Get("/trade-updates", r.tradeUpdates)
	})
}

func (r *Router) tradeUpdates(resp http.ResponseWriter, req *http.Request) {
	conn, err := r.upgrader.Upgrade(resp, req, nil)
	if err != nil {
		r.log.Error().Err(err).Msg("failed to upgrade websocket connection")
		return
	}
	r.teamsNotifier.RegisterConnection(conn)
	r.log.Trace().Msg("websocket connection upgraded")

	defer func() {
		r.teamsNotifier.RemoveConnection(conn)
		err = conn.Close()
		r.log.Trace().Err(err).Msg("websocket connection closed")
	}()

	for {
		if _, _, err = conn.ReadMessage(); err != nil {
			break
		}
	}
}
