package service

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

// WSHandler To handle all incoming http request
type WSHandler func(wsConn *websocket.Conn) error

// WS to return websocket handler
func WS(wsh WSHandler) HTTPHandler {
	return func(res http.ResponseWriter, req *http.Request) {
		logger.Infof("Websocket starts.")

		wsConn, err := upgrader.Upgrade(res, req, nil)
		if err != nil {
			logger.Errorf("Error in websocket: %s", err.Error())
			return
		}
		defer func() {
			logger.Infof("Websocket quits.")
			wsConn.Close()
		}()

		if err := wsh(wsConn); err != nil {
			logger.Errorf("Error in websocket: %s", err.Error())
		}
	}
}
