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
	// var allPeers []*Peer

	// channelIDMap := DiscoverChannels(conn, options...)
	// for channelID, endpointURL := range channelIDMap {
	// 	tempPeers, err := DiscoverChannelPeers(conn, endpointURL, channelID, options...)
	// 	if err != nil {
	// 		logger.Infof("Failed discover chanel peers of endpoint %s, channel ID %s", endpointURL, channelID)
	// 	} else {
	// 		// TODO to merge all peers.
	// 		allPeers = append(allPeers, tempPeers...)
	// 	}
	// }

	// logger.Infof("Totally found %d peers in %d channels of the network.", len(allPeers), len(channelIDMap))

	// opt := generateOption(options...)
	// if opt.IsDetail {
	// 	// Query legder info
	// 	for _, peer := range allPeers {
	// 		ldgs := make(map[string]*Ledger)
	// 		for _, channelID := range peer.Channels.StringList() {
	// 			ldg, err := QueryLedger(conn, channelID, []string{peer.URL})
	// 			if err == nil {
	// 				ldgs[channelID] = ldg
	// 			}
	// 		}
	// 		peer.Ledgers = ldgs
	// 	}
	// }

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
func DiscoverChannels(conn *NetworkConnection, options ...DiscoverOptionFunc) map[string]string {
	// map[channelID]endpointURL
	channelIDMap := make(map[string]string)

	//opt := generateOption(options...)
	// Get all channels.
	// The network peers are corresponding to the config in yaml organizations/<org>/peers.
	// They will be blank if no peers defined. But acutally all the peers can be found by discover service.
	// TODO if the endpoint config is nil, that means the identity is not found, the error should be return.
	if conn.Client.EndpointConfig() != nil {
		for _, endpoint := range conn.Client.EndpointConfig().NetworkPeers() {
			// It doesn't work now, since we still need to retrieve channels from the endpoint configured.
			// if !opt.isTarget(endpoint.URL) {
			// 	continue
			// }

			//if endpoint.MSPID == conn.Participant.MSPID
			channels, err := GetJoinedChannels(conn, endpoint.URL, options...)
			if err != nil {
				logger.Errorf("Getting joined channels got failed for endpoint %s: %s", endpoint.URL, err.Error())
				continue
			}

			// Only the first endpoint will be processed if they are duplicated.
			// TODO to get and use all peers of the channel.
			for _, channel := range channels {
				if _, ok := channelIDMap[channel.GetChannelId()]; !ok {
					channelIDMap[channel.GetChannelId()] = endpoint.URL
				}
			}
		}
	}

	logger.Infof("Found channels and only corresponding endpoints %v.", channelIDMap)
	return channelIDMap
}

func generateOption(options ...DiscoverOptionFunc) *DiscoverOption {
	disOpt := &DiscoverOption{}
	for _, opt := range options {
		opt(disOpt)
	}
	return disOpt
}

// // DiscoverChannelPeers to get all peers, by the specified channelID and peer endpoint.
// func DiscoverChannelPeers(conn *NetworkConnection, endPointURL string, channelID string, options ...DiscoverOptionFunc) ([]*Peer, error) {
// 	ctx := conn.Client
// 	var client *discovery.Client
// 	client, err := discovery.New(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	reqCtx, cancel := context.NewRequest(ctx, context.WithTimeout(10*time.Second))
// 	defer cancel()

// 	req := discovery.NewRequest().OfChannel(channelID).AddPeersQuery()
// 	peerCfg, err := comm.NetworkPeerConfig(ctx.EndpointConfig(), endPointURL)
// 	if err != nil {
// 		return nil, err
// 	}
// 	responses, err := client.Send(reqCtx, req, peerCfg.PeerConfig)
// 	if err != nil {
// 		return nil, err
// 	}
// 	resp := responses[0]
// 	chanResp := resp.ForChannel(channelID)
// 	discPeers, err := chanResp.Peers()
// 	if err != nil {
// 		return nil, err
// 	}

// 	opt := generateOption(options...)

// 	peers := make([]*Peer, 0)
// 	for _, discPeer := range discPeers {
// 		endpoint := discPeer.AliveMessage.GetAliveMsg().GetMembership().GetEndpoint()

// 		// TODO above there is redundant to get all peers, but there is problem in retrieving instantiated chaincodes, so then to be enhanced, see below.
// 		// If there is only one peer to discover, above retrieving all peers is redundant. The whole flow needs to be enhanced.
// 		if !opt.isTarget(endpoint) {
// 			continue
// 		}

// 		var instantiatedCCs []*Chaincode
// 		// TODO it should be more easier if retrieve instantiated chaincodes from resource management, with the isntalled chaincodes together.
// 		// These are the instantiated chaincode and installed on the peer.
// 		for _, cc := range discPeer.StateInfoMessage.GetStateInfo().Properties.GetChaincodes() {
// 			instantiatedCCs = append(instantiatedCCs, &Chaincode{Name: cc.GetName(), Version: cc.GetVersion(), ChannelID: channelID, Installed: true})
// 		}

// 		peers = append(peers, &Peer{
// 			Name:        endpoint,
// 			MSPID:       discPeer.MSPID,
// 			URL:         endpoint,
// 			Channels:    util.NewSet(channelID), // TODO to merge all channels for a peer.
// 			IsConnected: endpoint == endPointURL,
// 			UpdateTime:  time.Now().UnixNano() / 1000000,
// 			Chaincodes:  instantiatedCCs})
// 	}

// 	if opt.IsDetail {
// 		for _, peer := range peers {
// 			DiscoverPeerDetail(conn, channelID, peer)
// 			logger.Debug("Get peer detail", channelID, peer.Name, peer.Chaincodes)
// 		}
// 	}

// 	logger.Debugf("Found %d peers from channel %s.", len(peers), channelID)
// 	return peers, nil
// }

// DiscoverPeerDetail to discover more details of the peer, currently only installed chaincodes.
// peer.Chaincodes are installed and instantiated on the peer.
// func DiscoverPeerDetail(conn *NetworkConnection, channelID string, peer *Peer) {
// 	installedCCs, _ := QueryInstalledChaincodes(conn, peer.URL)
// 	instantiatedPerChannelCCs, _ := QueryInstantiatedChaincodes(conn, channelID)

// 	// To add all chaincodes installed but not instantiated.
// 	for _, installedCC := range installedCCs {
// 		if !existChaincode(peer.Chaincodes, installedCC.GetName(), installedCC.GetVersion()) {
// 			peer.Chaincodes = append(peer.Chaincodes, &Chaincode{
// 				Name:      installedCC.GetName(),
// 				Version:   installedCC.GetVersion(),
// 				Path:      installedCC.GetPath(),
// 				Installed: true})
// 		}
// 	}

// 	// To add all chaincodes instantiated in channel level but not installed.
// 	// Need to check existance with 'installedCCs', as well with peer.Chaincodes, since the installedCC is from resmgmtCliet,
// 	// but it might get nil if the connection is not allowed, for example, Admin@org1 connects to peer0.org, but the peer.Chaincodes will be exact.
// 	for _, instantiatedCC := range instantiatedPerChannelCCs {
// 		if !existChaincodeInstalled(installedCCs, instantiatedCC.GetName(), instantiatedCC.GetVersion()) &&
// 			!existChaincode(peer.Chaincodes, instantiatedCC.GetName(), instantiatedCC.GetVersion()) {
// 			peer.Chaincodes = append(peer.Chaincodes, &Chaincode{
// 				Name:      instantiatedCC.GetName(),
// 				Version:   instantiatedCC.GetVersion(),
// 				ChannelID: channelID,
// 				Path:      instantiatedCC.GetPath(),
// 				Installed: false})
// 		}
// 	}

// }

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
