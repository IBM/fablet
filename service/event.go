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

type ErrorResult struct {
	Error string `json:error`
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

	errChan := make(chan error, 1)
	// Monitor the client connection.
	go func() {
		for {
			logger.Debug("Begin to waiting for reading block event request")
			if _, _, err := wsConn.ReadMessage(); err != nil {
				// TODO To see if it err when client disconnected.
				logger.Errorf("Reading block event reqeust with error: %s.", err.Error())
				errChan <- err
				return
			}
		}
	}()

	eventChan := make(chan *fab.FilteredBlockEvent, 1)
	closeChan := make(chan int, 1)
	eventCloseChan := make(chan int, 1)
	go api.MonitorBlockEvent(conn, reqBody.ChannelID, eventChan, closeChan, eventCloseChan)

	pingTicker := time.NewTicker(WSPingInterval)

	// TODO defer in sequence
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
		case err := <-errChan:
			return err
		case <-pingTicker.C:
			// TODO ping is not enough, there always be tens of seconds later after connection close.
			// Places a reader might be better. See the chaincode event.
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

	listeners := 0

	reqChan := make(chan *ChaincodeEventReq, 1)
	errChan := make(chan error, 1)

	eventChan := make(chan *fab.CCEvent, 1)
	closeChan := make(chan int, 1)
	eventCloseChan := make(chan error, 1)

	pingTicker := time.NewTicker(WSPingInterval)

	// TODO defer in sequence
	defer func() {
		logger.Info("Service HandleChaincodeEvent end")
		// TODO This should happen in context likes web service, then the event handler has chance to process closeChan.
		closeChan <- 0
		pingTicker.Stop()
	}()

	go readCCEventRequest(wsConn, reqChan, errChan)

	for {
		select {
		case reqBody := <-reqChan:
			logger.Debug("Received a reqBody and then begin a new event connection.")
			conn, err := getConnOfReq(reqBody.GetReqConn(), true)
			if err != nil {
				return err
			}

			listeners++

			// Close the existing event registration
			if listeners > 1 {
				closeChan <- 0
			}

			go api.MonitorChaincodeEvent(conn, reqBody.ChannelID, reqBody.ChaincodeID, reqBody.EventFilter, eventChan, closeChan, eventCloseChan)
		case err := <-errChan:
			return err
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
			// wsConn.SetWriteDeadline(time.Now().Add(WSWriteDeadline))
			if err := wsConn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return errors.WithMessage(err, "Error of write websocket ping.")
			}
		case <-eventCloseChan:
			// TODO It doesn't work yet - to use eventCloseChan to pass the error in API.

			// The error need to be handled by frontend, since the chaincode event filter might be changed.
			// Althought the block event don't need to to do this.
			// Anyway, the eventCloseChan might with error or nil.
			// if err != nil {
			// 	logger.Error("Chaincode event listener return error: %s.", err.Error())
			// 	errRes := ErrorResult{
			// 		Error: err.Error(),
			// 	}
			// 	resultJSON, _ := json.Marshal(errRes)
			// 	if msgErr := wsConn.WriteMessage(websocket.TextMessage, resultJSON); msgErr != nil {
			// 		return errors.WithMessage(err, "Error of write event data.")
			// 	}
			// }

			listeners--
			if listeners <= 0 {
				return nil
			}
		}
	}

}

func readCCEventRequest(wsConn *websocket.Conn, reqChan chan *ChaincodeEventReq, errChan chan error) {
	for {
		logger.Debug("Begin to waiting for reading chaincode event request")
		reqBody := &ChaincodeEventReq{}

		if err := wsConn.ReadJSON(reqBody); err != nil {
			// TODO To see if it err when client disconnected.
			logger.Errorf("Reading chaincode event reqeust with error: %s.", err.Error())
			errChan <- err
			return
		}

		logger.Errorf("Reading chaincode event reqeust: %s %s %s.", reqBody.ChannelID, reqBody.ChaincodeID, reqBody.EventFilter)
		reqChan <- reqBody
	}
}
