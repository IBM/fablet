package service

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/IBM/fablet/api"
)

// JoinChannelReq to join a channel
type JoinChannelReq struct {
	BaseRequest
	ChannelID string   `json:"channelID"`
	Targets   []string `json:"targets"`
	Orderer   string   `json:"orderer"`
}

// CreateChannelReq to create a channel
type CreateChannelReq struct {
	BaseRequest
	TxContent []byte `json:"txContent"`
	Orderer   string `json:"orderer"`
}

// HandleCreateChannel to create a channle via orderer
func HandleCreateChannel(res http.ResponseWriter, req *http.Request) {
	logger.Info("Service HandleCreateChannel")

	reqBody := &CreateChannelReq{}
	conn, err := GetRequest(req, reqBody, true)
	if err != nil {
		ErrorOutput(res, req, RES_CODE_ERR_INTERNAL, errors.WithMessage(err, "Error occurred when parsing from request."))
		return
	}

	channelID, err := api.CreateChannel(conn, reqBody.TxContent, reqBody.Orderer)
	if err != nil {
		ErrorOutput(res, req, RES_CODE_ERR_INTERNAL,
			errors.WithMessagef(err, "Error occurs when create channel %s via orderer %s.", channelID, reqBody.Orderer))
		return
	}

	ResultOutput(res, req, map[string]interface{}{
		"channelID": channelID,
	})
}

// HandleJoinChannel for peer to join a channel
func HandleJoinChannel(res http.ResponseWriter, req *http.Request) {
	logger.Info("Service HandleJoinChannel")

	reqBody := &JoinChannelReq{}
	conn, err := GetRequest(req, reqBody, true)
	if err != nil {
		ErrorOutput(res, req, RES_CODE_ERR_INTERNAL, errors.WithMessage(err, "Error occurred when parsing from request."))
		return
	}

	err = api.JoinChannel(conn, reqBody.ChannelID, reqBody.Targets, reqBody.Orderer)

	if err != nil {
		ErrorOutput(res, req, RES_CODE_ERR_INTERNAL,
			errors.WithMessagef(err, "Error occurs when peer %v join channel %s.", reqBody.Targets, reqBody.ChannelID))
		return
	}

	ResultOutput(res, req, map[string]interface{}{
		"channelID": reqBody.ChannelID,
	})
}
