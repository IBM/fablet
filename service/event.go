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
type BlockEventReq struct {
	BaseRequest
	ChannelID string `json:"channelID"`
}

// ChaincodeEventReq chaincode event reqeust.
type ChaincodeEventReq struct {
	BaseRequest
	ChannelID   string `json:"channelID"`
	ChaincodeID string `json:"chaincodeID"`
	EventFilter string `json:"eventFilter"`
}

// BlockEventResult block event result
type BlockEventResult struct {
	Number     uint64 `json:"number"`
	TXNumber   int    `json:"TXNumber"`
	UpdateTime int64  `json:"updateTime"`
	SourceURL  string `json:"sourceURL"`
}

// ChaincodeEventResult chaincode event result
type ChaincodeEventResult struct {
	TXID        string `json:"TXID"`
	ChaincodeID string `json:"chaincodeID"`
	EventName   string `json:"eventName"`
	Payload     string `json:"payload"`
	BlockNumber uint64 `json:"blockNumber"`
	SourceURL   string `json:"sourceURL"`
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
	reqBody := &BlockEventReq{}
	if err := wsConn.ReadJSON(reqBody); err != nil {
		return err
	}
	conn, err := getConnOfReq(reqBody.GetReqConn(), true)
	if err != nil {
		return err
	}

	eventChan := make(chan *fab.FilteredBlockEvent, 1)
	closeChan := make(chan int, 1)
	eventCloseChan := make(chan int, 1)
	go api.MonitorBlockEvent(conn, reqBody.ChannelID, eventChan, closeChan, eventCloseChan)

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
			logger.Debugf("Get a block event from %s.", event.SourceURL)
			// wsConn.SetWriteDeadline(time.Now().Add(WSWriteDeadline))
			if wsConn == nil {
				return errors.Errorf("Websocket connection is nil.")
			}
			if event == nil {
				// Allow error
				// return errors.Errorf("Event is nil")
			} else {
				// TODO In fact, we don't know the block time now.
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
		case <-eventCloseChan:
			return nil
		}
	}
}

// HandleChaincodeEvent handle event
func HandleChaincodeEvent(wsConn *websocket.Conn) error {
	logger.Info("Service HandleChaincodeEvent")
	reqBody := &ChaincodeEventReq{}
	if err := wsConn.ReadJSON(reqBody); err != nil {
		return err
	}
	conn, err := getConnOfReq(reqBody.GetReqConn(), true)
	if err != nil {
		return err
	}

	eventChan := make(chan *fab.CCEvent, 1)
	closeChan := make(chan int, 1)
	eventCloseChan := make(chan int, 1)
	go api.MonitorChaincodeEvent(conn, reqBody.ChannelID, reqBody.ChaincodeID, reqBody.EventFilter, eventChan, closeChan, eventCloseChan)

	pingTicker := time.NewTicker(WSPingInterval)

	defer func() {
		logger.Info("Service HandleChaincodeEvent end")
		// TODO This should happen in context likes web service, then the event handler has chance to process closeChan.
		closeChan <- 0
		pingTicker.Stop()
	}()

	for {
		select {
		case event := <-eventChan:
			logger.Debugf("Get a chaincode event from %s.", event.SourceURL)
			// wsConn.SetWriteDeadline(time.Now().Add(WSWriteDeadline))
			if wsConn == nil {
				return errors.Errorf("Websocket connection is nil.")
			}
			if event == nil {
				// Allow error
				// return errors.Errorf("Event is nil")
			} else {
				// TODO In fact, we don't know the block time now.
				ccEventRes := ChaincodeEventResult{
					TXID:        event.TxID,
					ChaincodeID: event.ChaincodeID,
					EventName:   event.EventName,
					Payload:     string(event.Payload),
					BlockNumber: event.BlockNumber,
					SourceURL:   event.SourceURL,
				}

				resultJSON, _ := json.Marshal(ccEventRes)
				if err := wsConn.WriteMessage(websocket.TextMessage, resultJSON); err != nil {
					return errors.WithMessage(err, "Error of write event data.")
				}
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
		case <-eventCloseChan:
			return nil
		}
	}
}
