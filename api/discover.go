package api

import (
	"sort"

	"github.com/hyperledger/fabric-protos-go/peer"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
)

// DiscoverOption to collect all options in discovery
type DiscoverOption struct {
	// List of endpoint urls, only discover/monitor these peers.
	Targets  []string `json:"discoverPeers"`
	IsDetail bool     `json:"isDetail"`
}

func (opt *DiscoverOption) isTarget(endpointURL string) bool {
	// Default empty allows all.
	if len(opt.Targets) == 0 {
		return true
	}
	for _, p := range opt.Targets {
		if p == endpointURL {
			return true
		}
	}
	return false
}

// DiscoverOptionFunc to handle the discovery option
type DiscoverOptionFunc func(opt *DiscoverOption) error

// WithTargets only retrieve details of the peer.
func WithTargets(targets ...string) DiscoverOptionFunc {
	return func(opt *DiscoverOption) error {
		opt.Targets = targets
		return nil
	}
}

// WithIsDetail only retrieve details of the peer.
func WithIsDetail(isDetail bool) DiscoverOptionFunc {
	return func(opt *DiscoverOption) error {
		opt.IsDetail = isDetail
		return nil
	}
}

// DiscoverNetworkOverview To get all peers by discover.
// TODO for now, only 1 endpoint config is well supported.
// Fault tolerant.
func DiscoverNetworkOverview(conn *NetworkConnection, options ...DiscoverOptionFunc) (*NetworkOverview, error) {
	// return allPeers, nil
	peers := []*Peer{}
	for _, peer := range conn.Peers {
		peers = append(peers, peer)
	}
	sort.SliceStable(peers, func(i, j int) bool {
		if peers[i].MSPID != peers[j].MSPID {
			return peers[i].MSPID < peers[j].MSPID
		}
		return peers[i].Name < peers[j].Name
	})

	return &NetworkOverview{
		Peers:             peers,
		ChannelOrderers:   conn.ChannelOrderers,
		ChannelLedgers:    conn.ChannelLedgers,
		ChannelChainCodes: conn.ChannelChaincodes,
	}, nil
}

/////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////

// DiscoverChannels to discover all channels via all endpoints from the config.
// func DiscoverChannels(conn *NetworkConnection, options ...DiscoverOptionFunc) map[string]string {
// 	// map[channelID]endpointURL
// 	channelIDMap := make(map[string]string)

// 	//opt := generateOption(options...)
// 	// Get all channels.
// 	// The network peers are corresponding to the config in yaml organizations/<org>/peers.
// 	// They will be blank if no peers defined. But acutally all the peers can be found by discover service.
// 	// TODO if the endpoint config is nil, that means the identity is not found, the error should be return.
// 	if conn.Client.EndpointConfig() != nil {
// 		for _, endpoint := range conn.Client.EndpointConfig().NetworkPeers() {
// 			// It doesn't work now, since we still need to retrieve channels from the endpoint configured.
// 			// if !opt.isTarget(endpoint.URL) {
// 			// 	continue
// 			// }

// 			//if endpoint.MSPID == conn.Participant.MSPID
// 			channels, err := GetJoinedChannels(conn, endpoint.URL, options...)
// 			if err != nil {
// 				logger.Errorf("Getting joined channels got failed for endpoint %s: %s", endpoint.URL, err.Error())
// 				continue
// 			}

// 			// Only the first endpoint will be processed if they are duplicated.
// 			// TODO to get and use all peers of the channel.
// 			for _, channel := range channels {
// 				if _, ok := channelIDMap[channel.GetChannelId()]; !ok {
// 					channelIDMap[channel.GetChannelId()] = endpoint.URL
// 				}
// 			}
// 		}
// 	}

// 	logger.Infof("Found channels and only corresponding endpoints %v.", channelIDMap)
// 	return channelIDMap
// }

func generateOption(options ...DiscoverOptionFunc) *DiscoverOption {
	disOpt := &DiscoverOption{}
	for _, opt := range options {
		opt(disOpt)
	}
	return disOpt
}

func existChaincode(ccs []*Chaincode, name string, version string) bool {
	for _, cc := range ccs {
		if cc.Name == name && cc.Version == version {
			return true
		}
	}
	return false
}

func existChaincodeInstalled(ccs []*peer.ChaincodeInfo, name string, version string) bool {
	for _, cc := range ccs {
		if cc.Name == name && cc.Version == version {
			return true
		}
	}
	return false
}

// GetJoinedChannels to get all joined channels of an endpoint.
func GetJoinedChannels(conn *NetworkConnection, endpointURL string, options ...DiscoverOptionFunc) ([]*peer.ChannelInfo, error) {
	ctx := conn.Client

	peerCfg, err := comm.NetworkPeerConfig(ctx.EndpointConfig(), endpointURL)
	if err != nil {
		return nil, err
	}
	resMgmtClient, err := resmgmt.New(conn.ClientProvider)
	if err != nil {
		return nil, err
	}
	p, err := ctx.InfraProvider().CreatePeerFromConfig(peerCfg)
	if err != nil {
		return nil, err
	}
	// TODO to use resmgmt.WithTargetEndpoints to be more easier.
	chresp, err := resMgmtClient.QueryChannels(resmgmt.WithTargets(p))
	if err != nil {
		return nil, err
	}

	logger.Infof("Get %d joined channels from: %s", len(chresp.Channels), endpointURL)

	return chresp.Channels, nil
}
