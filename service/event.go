package service

import (
	"encoding/json"
	"time"

	"github.com/IBM/fablet/api"
	"github.com/gorilla/websocket"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
)

// EventReq use fields instead of anonlymous fields, to have a more clear structure.
type EventReq struct {
	BaseRequest
	ChannelID string `json:"channelID"`
}

// BlockEventResult block event result
type BlockEventResult struct {
	Number     uint64 `json:"number"`
	TXNumber   int    `json:"TXNumber"`
	UpdateTime int64  `json:"updateTime"`
	SourceURL  string `json:"sourceURL"`
}

const (
	// WSPingInterval interval of ping
	WSPingInterval = time.Second * 10
	// WSWriteDeadline deadline time of write
	WSWriteDeadline = time.Second * 10
)

// HandleBlockEvent handle event
func HandleBlockEvent(wsConn *websocket.Conn) error {
	logger.Info("Service HandleBlockEvent")
	reqBody := &EventReq{}
	if err := wsConn.ReadJSON(reqBody); err != nil {
		return err
	}
	conn, err := getConnOfReq(reqBody.GetReqConn(), true)
	if err != nil {
		return err
	}

	eventChan := make(chan *fab.FilteredBlockEvent, 1)
	closeChan := make(chan int, 1)
	go api.MonitorBlockEvent(conn, reqBody.ChannelID, eventChan, closeChan)

	pingTicker := time.NewTicker(WSPingInterval)

	defer func() {
		logger.Info("Service HandleBlockEvent end")
		// TODO closeChan might be a little later, so then MonitorBlockEvent will ends a little later.
		closeChan <- 0
		pingTicker.Stop()
	}()

	for {
		select {
		case event := <-eventChan:
			logger.Debugf("Get an event from %s.", event.SourceURL)
			// wsConn.SetWriteDeadline(time.Now().Add(WSWriteDeadline))
			if wsConn == nil {
				return errors.Errorf("Websocket connection is nil.")
			}
			if event == nil {
				return errors.Errorf("Event is nil")
			}
			blockEventRes := BlockEventResult{
				Number:     event.FilteredBlock.GetNumber(),
				TXNumber:   len(event.FilteredBlock.GetFilteredTransactions()),
				UpdateTime: time.Now().UnixNano() / 1000000,
				SourceURL:  event.SourceURL,
			}

			resultJSON, _ := json.Marshal(blockEventRes)
			if err := wsConn.WriteMessage(websocket.TextMessage, resultJSON); err != nil {
				return errors.WithMessage(err, "Error of write event data.")
			}
		case <-pingTicker.C:
			// logger.Debug("Ping................")
			// wsConn.SetWriteDeadline(time.Now().Add(time.Second * 3))
			// TODO ping is not enough, there always be 60 seconds later after connection close.
			if wsConn == nil {
				return errors.Errorf("Websocket connection is nil.")
			}
			if err := wsConn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return errors.WithMessage(err, "Error of write websocket ping.")
			}
		}
	}
}
