package service

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/IBM/fablet/api"
)

// NetworkDiscoverReq discover request
type NetworkDiscoverReq struct {
	BaseRequest
}

// NetworkRefreshReq discover request. Same with NetworkDiscoverReq, but it will includes refresh=true.
type NetworkRefreshReq struct {
	BaseRequest
}

// HandleNetworkDiscover to discover all network
func HandleNetworkDiscover(res http.ResponseWriter, req *http.Request) {
	logger.Info("Service HandleNetworkDiscover")

	reqBody := &NetworkDiscoverReq{}
	conn, err := GetRequest(req, reqBody, true)
	if err != nil {
		ErrorOutput(res, req, RES_CODE_ERR_INTERNAL, errors.WithMessage(err, "Error occurred when parsing from request."))
		return
	}

	networkOverview, err := api.DiscoverNetworkOverview(conn)
	if err != nil {
		ErrorOutput(res, req, RES_CODE_ERR_INTERNAL, errors.WithMessage(err, "Error occurs when get network overview."))
		return
	}

	ResultOutput(res, req, map[string]interface{}{
		"peers":             networkOverview.Peers,
		"channelLedgers":    networkOverview.ChannelLedgers,
		"channelChaincodes": networkOverview.ChannelChainCodes,
		"channelOrderers":   networkOverview.ChannelOrderers,
	})
}

// HandleNetworkRefresh to discover all network
func HandleNetworkRefresh(res http.ResponseWriter, req *http.Request) {
	logger.Info("Service HandleNetworkRefresh")

	reqBody := &NetworkRefreshReq{}
	conn, err := GetRequest(req, reqBody, true, WithRefresh(true))
	if err != nil {
		ErrorOutput(res, req, RES_CODE_ERR_INTERNAL, errors.WithMessage(err, "Error occurred when parsing from request."))
		return
	}
	ResultOutput(res, req, map[string]interface{}{
		"conn": conn.Identifier,
	})
}
