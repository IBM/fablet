package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/IBM/fablet/api"
	"github.com/pkg/errors"
)

// LedgerQueryReq to query a ledger
type LedgerQueryReq struct {
	BaseRequest
	ChannelID string   `json:"channelID"`
	Targets   []string `json:"targets"`
}

// BlockQueryReq to query blocks
type BlockQueryReq struct {
	BaseRequest
	ChannelID string   `json:"channelID"`
	Targets   []string `json:"targets"`
	Begin     uint64   `json:"begin"`
	Len       uint64   `json:"len"`
}

// BlockQueryAnyReq to query block with any: tx id, block hash, block number
type BlockQueryAnyReq struct {
	BaseRequest
	ChannelID string   `json:"channelID"`
	Targets   []string `json:"targets"`
	QueryKey  string   `json:"queryKey"`
}

// HandleLedgerQuery to query a ledger of a channel
func HandleLedgerQuery(res http.ResponseWriter, req *http.Request) {
	logger.Info("Service HandleLedgerQuery")

	reqBody := &LedgerQueryReq{}
	conn, err := GetRequest(req, reqBody, true)
	if err != nil {
		ErrorOutput(res, req, RES_CODE_ERR_INTERNAL, errors.WithMessage(err, "Error occurred when parsing from request."))
		return
	}

	logger.Info(fmt.Sprintf("Begin to query ledger at %v of channel %s", reqBody.Targets, reqBody.ChannelID))

	ledgerRes, err := api.QueryLedger(conn, reqBody.ChannelID, reqBody.Targets)
	if err != nil {
		ErrorOutput(res, req, RES_CODE_ERR_INTERNAL, errors.WithMessage(err, "Error occurred when query the ledger."))
		return
	}

	ResultOutput(res, req, map[string]interface{}{
		"ledger": ledgerRes,
	})
}

// HandleBlockQuery to query blocks of a ledger
func HandleBlockQuery(res http.ResponseWriter, req *http.Request) {
	logger.Info("Service HandleBlockQuery")

	reqBody := &BlockQueryReq{}
	conn, err := GetRequest(req, reqBody, true)
	if err != nil {
		ErrorOutput(res, req, RES_CODE_ERR_INTERNAL, errors.WithMessage(err, "Error occurred when parsing from request."))
		return
	}

	blocks, _ := api.QueryBlock(conn, reqBody.ChannelID, reqBody.Targets, reqBody.Begin, reqBody.Len)
	ResultOutput(res, req, map[string]interface{}{
		"blocks": blocks,
	})
}

// HandleBlockQueryAny to query a block of a ledger by any possible kind of key
func HandleBlockQueryAny(res http.ResponseWriter, req *http.Request) {
	logger.Info("Service HandleBlockQueryAny")

	reqBody := &BlockQueryAnyReq{}
	conn, err := GetRequest(req, reqBody, true)
	if err != nil {
		ErrorOutput(res, req, RES_CODE_ERR_INTERNAL, errors.WithMessage(err, "Error occurred when parsing from request."))
		return
	}

	var block *api.Block

	blkNumber, err := strconv.ParseUint(reqBody.QueryKey, 10, 64)
	if err == nil {
		blocks, err := api.QueryBlock(conn, reqBody.ChannelID, reqBody.Targets, blkNumber, 1)
		if err == nil && len(blocks) > 0 {
			block = blocks[0]
		}
	} else {
		block, err = api.QueryBlockByHash(conn, reqBody.ChannelID, reqBody.Targets, reqBody.QueryKey)
		if err != nil {
			block, err = api.QueryBlockByTxID(conn, reqBody.ChannelID, reqBody.Targets, reqBody.QueryKey)
		}
	}

	if err != nil {
		ErrorOutput(res, req, RES_CODE_ERR_INTERNAL, errors.Errorf("Error occurred when query the block, cannot find the block by number, hash or transaction ID."))
		return
	}

	ResultOutput(res, req, map[string]interface{}{
		"block": block,
	})
}
