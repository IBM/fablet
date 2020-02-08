package service

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/IBM/fablet/api"
)

// PeerDetailsReq to get details of a peer
type PeerDetailsReq struct {
	BaseRequest
	Target   string   `json:"target"`
	Channels []string `json:"channels"`
}

// HandlePeerDetails for peer details
func HandlePeerDetails(res http.ResponseWriter, req *http.Request) {
	logger.Info("Service HandlePeerDetails")

	reqBody := &PeerDetailsReq{}
	conn, err := GetRequest(req, reqBody, true)
	if err != nil {
		ErrorOutput(res, req, RES_CODE_ERR_INTERNAL, errors.WithMessage(err, "Error occurred when parsing from request."))
		return
	}

	// Try to update channels
	channels, err := api.GetJoinedChannels(conn, reqBody.Target)
	channelIDs := reqBody.Channels
	if err == nil {
		channelIDs = []string{}
		for _, channel := range channels {
			channelIDs = append(channelIDs, channel.GetChannelId())
		}
	}

	installedCCQueryError := false
	installedChaincodes, err := api.QueryInstalledChaincodes(conn, reqBody.Target)
	if err != nil {
		installedCCQueryError = true
		//ErrorOutput(res, req, RES_CODE_ERR_INTERNAL, errors.WithMessagef(err, "Error occurs when get peer installed chaincodes of %s.", reqBody.Target))
		//return
	}

	channelChaincodes := make(map[string][]*api.Chaincode)
	for _, channelID := range channelIDs {
		instantiatedChaincodes, err := api.QueryInstantiatedChaincodes(conn, channelID)
		if err == nil {
			channelChaincodes[channelID] = instantiatedChaincodes
		}
		if err != nil {
			ErrorOutput(res, req, RES_CODE_ERR_INTERNAL,
				errors.WithMessagef(err, "Error occurs when get instantiated chaincodes of %s of channel %s.", reqBody.Target, channelID))
			//return
		}
	}

	channelLedgers := make(map[string]*api.Ledger)
	for _, channelID := range channelIDs {
		ledger, err := api.QueryLedger(conn, channelID, []string{reqBody.Target})
		if err == nil {
			channelLedgers[channelID] = ledger
		}
		if err != nil {
			ErrorOutput(res, req, RES_CODE_ERR_INTERNAL,
				errors.WithMessagef(err, "Error occurs when get channels of %s.", reqBody.Target))
			//return
		}
	}

	ResultOutput(res, req, map[string]interface{}{
		"installedChaincodes":   installedChaincodes, // Per peer
		"channelChaincodes":     channelChaincodes,   // Per channel
		"installedCCQueryError": installedCCQueryError,
		"channelLedgers":        channelLedgers,
		"channels":              channelIDs,
	})
}
