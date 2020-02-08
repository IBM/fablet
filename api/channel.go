package api

import (
	"bytes"
	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
)

// getJoinedChannels to get all joined channels of an endpoint.
func getJoinedChannels(conn *NetworkConnection, endpointURL string) ([]string, error) {
	resMgmtClient, err := resmgmt.New(conn.ClientProvider)
	if err != nil {
		return nil, err
	}
	chresp, err := resMgmtClient.QueryChannels(resmgmt.WithTargetEndpoints(endpointURL))
	if err != nil {
		return nil, err
	}

	channelIDs := []string{}
	for _, channelInfo := range chresp.Channels {
		channelIDs = append(channelIDs, channelInfo.GetChannelId())
	}
	return channelIDs, nil
}

// CreateChannel to create a channel
func CreateChannel(conn *NetworkConnection, txContent []byte, orderer string) (string, error) {
	cub, _ := resource.ExtractChannelConfig(txContent)
	cu := &common.ConfigUpdate{}
	err := proto.Unmarshal(cub, cu)
	if err != nil {
		return "", err
	}
	channelID := cu.GetChannelId()
	resMgmtClient, err := resmgmt.New(conn.ClientProvider)
	req := resmgmt.SaveChannelRequest{
		ChannelID:     channelID,
		ChannelConfig: bytes.NewReader(txContent)}

	_, err = resMgmtClient.SaveChannel(req, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint(orderer))
	if err != nil {
		return "", err
	}
	return channelID, nil
}

// JoinChannel to join a peer into channel
func JoinChannel(conn *NetworkConnection, channelID string, targets []string, orderer string) error {
	resMgmtClient, err := resmgmt.New(conn.ClientProvider)
	if err != nil {
		return err
	}
	return resMgmtClient.JoinChannel(channelID,
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
		resmgmt.WithOrdererEndpoint(orderer),
		resmgmt.WithTargetEndpoints(targets...))
}
