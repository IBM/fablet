package service

import (
	"net/http"

	"github.com/IBM/fablet/api"
	"github.com/IBM/fablet/util"
	"github.com/pkg/errors"
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
		"peers":              networkOverview.Peers,
		"channels":           transChannels(networkOverview.Channels),
		"peerStatuses":       transPeerStatuses(networkOverview.EndpointStatuses),
		"channelLedgers":     networkOverview.ChannelLedgers,
		"channelChaincodes":  networkOverview.ChannelChainCodes,
		"channelOrderers":    networkOverview.ChannelOrderers,
		"channelAnchorPeers": networkOverview.ChannelAnchorPeers,
	})
}

func transChannels(channels []*api.Channel) []string {
	channelIDs := []string{}
	for _, channel := range channels {
		channelIDs = append(channelIDs, channel.ChannelID)
	}
	return channelIDs
}

func transPeerStatuses(statuses map[string]util.EndPointStatus) map[string]api.PeerStatus {
	transStatuses := map[string]api.PeerStatus{}

	for peer, status := range statuses {
		transStatus := api.PeerStatus{
			Ping:  true,
			GRPC:  true,
			Valid: true,
		}

		if status != util.EndPointStatus_Valid {
			transStatus.Valid = false
			if status != util.EndPointStatus_Connectable {
				if status == util.EndPointStatus_Refused {
					transStatus.GRPC = false
				} else {
					// Not found or timeout
					transStatus.GRPC = false
					transStatus.Ping = false
				}
			}
		}

		transStatuses[peer] = transStatus
	}

	return transStatuses
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
