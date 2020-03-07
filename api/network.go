package api

import (
	"bytes"
	"crypto/md5"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/fablet/util"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
)

// ConnectionProfile basic connection profile.
type ConnectionProfile struct {
	Config     []byte `json:"config"`
	ConfigType string // yaml | json
}

// NewConnection to create a new connection to the Fabric network.
// TODO To add a new paraemter for CognitiveUpdate automatically.
func NewConnection(connProfile *ConnectionProfile, participant *Participant, useDiscovery bool) (*NetworkConnection, error) {
	if len(connProfile.Config) < 1 {
		return nil, errors.New("the connection profile is empty")
	}
	sdk, err := fabsdk.New(config.FromRaw(connProfile.Config, connProfile.ConfigType))
	if err != nil {
		return nil, err
	}
	signID, err := CreateSigningIdentity(sdk, participant)
	if err != nil {
		return nil, err
	}
	participant.SignID = signID
	ctxProvider := sdk.Context(fabsdk.WithIdentity(signID))

	ctx, err := ctxProvider()
	if err != nil {
		return nil, err
	}
	conn := &NetworkConnection{ConnectionProfile: connProfile, Participant: participant, UseDiscovery: useDiscovery}
	conn.SDK = sdk
	conn.ClientProvider = ctxProvider
	conn.Client = ctx

	conn.Identifier = CalConnIdentifier(connProfile, participant, useDiscovery)
	conn.UpdateTime = time.Now()
	conn.ActiveTime = conn.UpdateTime

	conn.initBaseNetwork()

	// TODO refresh connection frequence.
	// TODO separated function for update.
	if conn.UseDiscovery {
		conn.discoverNetwork()

		// TODO condition causes error connection
		//if err := conn.discoverNetwork(); err == nil {
		conn.updateConnection()
		//}
		// If err is not nil, continue to construct connection, to return the configured peers.
	}

	// TODO below might occure peer connection error again.
	conn.updateChannelLedgers()
	conn.updateChannelOrderers()
	conn.updateChannelChaincodes()

	return conn, nil
}

// CalConnIdentifier to calculate the identifier for the connection. Calculated by parameters.
func CalConnIdentifier(connProfile *ConnectionProfile, participant *Participant, useDiscovery bool) string {
	hash := md5.New()
	hash.Write(connProfile.Config)
	hash.Write([]byte(connProfile.ConfigType))
	hash.Write(participant.Cert)
	hash.Write(participant.PrivateKey)
	hash.Write([]byte(participant.MSPID))
	dis := []byte{1}
	if !useDiscovery {
		dis = []byte{0}
	}
	hash.Write(dis)
	return hex.EncodeToString(hash.Sum(nil))
}

// initBaseNetwork to initialize the network based on the base config file, without discovery.
func (conn *NetworkConnection) initBaseNetwork() {
	conn.Channels = make(map[string]*Channel)
	conn.Organizations = make(map[string]*Organization)
	conn.Peers = make(map[string]*Peer)
	conn.Orderers = make(map[string]*Orderer)
	conn.EndpointStatuses = make(map[string]util.EndPointStatus)
	conn.ChannelLedgers = make(map[string]*Ledger)
	conn.ChannelChaincodes = make(map[string][]*Chaincode)
	conn.ChannelOrderers = make(map[string][]*Orderer)

	// Add all peers
	for peerName, peerCfg := range conn.ConfiguredPeers() {
		orgName, MSPID := conn.findConfiguredPeerOrg(peerName)
		conn.Peers[peerName] = &Peer{
			Name:               peerName,
			OrgName:            orgName,
			MSPID:              MSPID,
			URL:                peerCfg.URL,
			TLSCACert:          peerCfg.TLSCACert,
			GRPCOptions:        peerCfg.GRPCOptions,
			Channels:           util.NewSet(),
			PeerChannelConfigs: make(map[string]*fab.PeerChannelConfig),
			IsConfigured:       true,
			UpdateTime:         time.Now().UnixNano() / 1000000,
		}
	}

	// Add all orderers
	for ordName, orderer := range conn.ConfiguredOrderers() {
		conn.Orderers[ordName] = &Orderer{
			Name:        ordName,
			URL:         orderer.URL,
			GRPCOptions: orderer.GRPCOptions,
			TLSCACert:   orderer.TLSCACert,
			Channels:    util.NewSet(),
		}
	}

	// Add all channels
	for channelID, channelEndpointConfig := range conn.ConfiguredChannels() {
		conn.Channels[channelID] = &Channel{ChannelID: channelID, Peers: util.NewSet(), Orderers: util.NewSet(), Policies: &channelEndpointConfig.Policies}

		for peerName, peerChannelConfig := range channelEndpointConfig.Peers {
			conn.addChannelEndpoint(channelID, peerName)
			// All peers should have already been added above.
			if existPeer, ok := conn.Peers[peerName]; ok {
				existPeer.Channels.Add(channelID)
				existPeer.PeerChannelConfigs[channelID] = &peerChannelConfig
			}
		}

		for _, orderer := range channelEndpointConfig.Orderers {
			conn.addChannelOrderer(channelID, orderer)
			if existOrderer, ok := conn.Orderers[orderer]; ok {
				existOrderer.Channels.Add(channelID)
			}
		}
	}

	// Add all organizations
	for orgName, org := range conn.ConfiguredOrganizations() {
		conn.Organizations[orgName] = &Organization{
			Name:       orgName,
			MSPID:      org.MSPID,
			CryptoPath: org.CryptoPath,
			Peers:      util.NewStringSet(org.Peers...),
		}
	}
}

func (conn *NetworkConnection) discoverNetwork() error {
	// TODO condition causes error connection
	// if err := conn.discoverChannels(); err != nil {
	// 	return err
	// }
	conn.discoverChannels()

	// Discover all peers per channel with corresponding configured endpoints.
	for channelID, channel := range conn.Channels {
		peers, orderers, err := conn.discoverChannelPeers(channelID, channel.Peers.StringList())
		if err != nil {
			logger.Errorf("Discover channel peers got failed: %s.", err)
		}

		for _, peer := range peers {
			conn.addChannelEndpoint(channelID, peer.Name)
			conn.addOrgEndpoint(peer.OrgName, peer.MSPID, peer.Name)

			if existedPeer := conn.findPeer(peer.Name); existedPeer == nil {
				logger.Debugf("%s of %v is new.", peer.Name, peer.Channels.StringList())
				conn.Peers[peer.Name] = peer
			} else {
				existedPeer.Channels.Add(channelID)
				if _, ok := existedPeer.PeerChannelConfigs[channelID]; !ok {
					existedPeer.PeerChannelConfigs[channelID] = peer.PeerChannelConfigs[channelID]
				}
			}
		}

		for _, orderer := range orderers {
			conn.addChannelOrderer(channelID, orderer.Name)

			if existedOrderer := conn.findOrderer(orderer.Name); existedOrderer == nil {
				logger.Debugf("%s of %v is new.", orderer.Name, orderer.Channels.StringList())
				conn.Orderers[orderer.Name] = orderer
			} else {
				existedOrderer.Channels.Add(channelID)
			}
		}
	}

	return nil
}

// discoverChannels to discover all channels via all endpoints from the config.
func (conn *NetworkConnection) discoverChannels() error {
	var wg sync.WaitGroup

	for _, peer := range conn.Peers {
		wg.Add(1)
		go func(peer *Peer) {
			defer wg.Done()
			disChannels, err := getJoinedChannels(conn, peer.URL)
			if err != nil {
				logger.Errorf("Getting joined channels got failed for endpoint %s: %s", peer.URL, err.Error())

				conn.EndpointStatuses[peer.Name] = util.GetEndpointStatus(peer.URL)
				return
			}

			logger.Infof("Find joined channels %s for endpoint %s.", disChannels, peer.Name)

			// Connect fine
			conn.EndpointStatuses[peer.Name] = util.EndPointStatus_Valid

			for _, disChannel := range disChannels {
				peer.Channels.Add(disChannel)
				conn.addChannelEndpoint(disChannel, peer.Name)
			}
		}(peer)

	}

	wg.Wait()

	for _, peer := range conn.Peers {
		if conn.EndpointStatuses[peer.Name] == util.EndPointStatus_Connectable {
			// Any peer is fine
			return nil
		}
	}
	return errors.Errorf("Cannot get response from any peer.")
}

// The peer name is not strict, it might be peer url instead.
func (conn *NetworkConnection) findPeerName(nameOrURL string) string {
	peer := conn.findPeer(nameOrURL)
	if peer == nil {
		return ""
	}
	return peer.Name
}

func (conn *NetworkConnection) findPeer(nameOrURL string) *Peer {
	for _, peer := range conn.Peers {
		if peer.Name == nameOrURL || peer.URL == nameOrURL {
			return peer
		}
	}
	return nil
}

// The peer name is not strict, it might be peer url instead.
func (conn *NetworkConnection) findOrdererName(nameOrURL string) string {
	orderer := conn.findOrderer(nameOrURL)
	if orderer == nil {
		return ""
	}
	return orderer.Name
}

func (conn *NetworkConnection) findOrderer(nameOrURL string) *Orderer {
	for _, ord := range conn.Orderers {
		if ord.Name == nameOrURL || ord.URL == nameOrURL {
			return ord
		}
	}
	return nil
}

func (conn *NetworkConnection) addChannelEndpoint(channelID string, endpoint string) {
	if channel, ok := conn.Channels[channelID]; !ok {
		conn.Channels[channelID] = &Channel{ChannelID: channelID, Peers: util.NewSet(endpoint), Orderers: util.NewSet()}
	} else if !channel.Peers.Exist(endpoint) && !channel.Peers.Exist(conn.findPeerName(endpoint)) {
		channel.Peers.Add(endpoint)
	}
}

func (conn *NetworkConnection) addChannelOrderer(channelID string, orderer string) {
	if channel, ok := conn.Channels[channelID]; !ok {
		conn.Channels[channelID] = &Channel{ChannelID: channelID, Peers: util.NewSet(), Orderers: util.NewSet(orderer)}
	} else if !channel.Orderers.Exist(orderer) && !channel.Orderers.Exist(conn.findOrdererName(orderer)) {
		channel.Orderers.Add(orderer)
	}
}

func (conn *NetworkConnection) addOrgEndpoint(orgName string, MSPID string, endpoint string) {
	if org, ok := conn.Organizations[orgName]; !ok {
		conn.Organizations[orgName] = &Organization{
			Name:       orgName,
			MSPID:      MSPID,
			CryptoPath: "./tmp", // TODO
			Peers:      util.NewSet(endpoint),
		}
	} else if !org.Peers.Exist(endpoint) && !org.Peers.Exist(conn.findPeerName(endpoint)) {
		conn.Organizations[orgName].Peers.Add(endpoint)
	}
}

// DiscoverChannelPeers to get all peers, by the specified channelID and peer endpoint.
func (conn *NetworkConnection) discoverChannelPeers(channelID string, endPoints []string, options ...DiscoverOptionFunc) ([]*Peer, []*Orderer, error) {
	peers := []*Peer{}
	orderers := []*Orderer{}

	ctx := conn.Client
	client, err := discovery.New(ctx)
	if err != nil {
		return peers, orderers, err
	}
	reqCtx, cancel := context.NewRequest(ctx, context.WithTimeout(DiscoverTimeOut))
	defer cancel()

	req := discovery.NewRequest().OfChannel(channelID).AddPeersQuery().AddConfigQuery()

	var wg sync.WaitGroup
	var locker sync.RWMutex

	for _, endpoint := range endPoints {
		peerCfg, err := comm.NetworkPeerConfig(ctx.EndpointConfig(), endpoint)
		if err != nil {
			continue
		}

		wg.Add(1)

		// Query per endpoint to void exception.
		go func(peerCfg *fab.NetworkPeer) {
			defer wg.Done()

			responses, err := client.Send(reqCtx, req, peerCfg.PeerConfig)
			if err != nil {
				return
			}

			locker.Lock()
			defer locker.Unlock()

			for _, resp := range responses {
				chanResp := resp.ForChannel(channelID)
				discPeers, err := chanResp.Peers()
				if err != nil {
					return
				}
				cfg, cfgErr := chanResp.Config()
				for _, discPeer := range discPeers {
					endpoint := discPeer.AliveMessage.GetAliveMsg().GetMembership().GetEndpoint()
					var cert *x509.Certificate
					if cfgErr == nil {
						if certBytes, ok := cfg.GetMsps()[discPeer.MSPID]; ok {
							certs, err := TLSAllCertsByBytes(certBytes.GetTlsRootCerts())
							if err == nil {
								// TODO take the first cert now.
								cert = certs[0]
							}
						}
					}
					peers = append(peers, &Peer{
						Name:        endpoint,
						OrgName:     conn.findOrgName(discPeer.MSPID),
						MSPID:       discPeer.MSPID,
						URL:         endpoint,
						TLSCACert:   cert,
						GRPCOptions: make(map[string]interface{}), // Blank options
						Channels:    util.NewSet(channelID),
						PeerChannelConfigs: map[string]*fab.PeerChannelConfig{channelID: &fab.PeerChannelConfig{
							EndorsingPeer:  true,
							ChaincodeQuery: true,
							LedgerQuery:    true,
							EventSource:    true,
						}}, // TODO *** PeerChannelConfig should be fetched from the config.
						IsConfigured: false,
						UpdateTime:   time.Now().UnixNano() / 1000000,
					})
				}

				for MSPID, ordererCfg := range cfg.GetOrderers() {
					var cert *x509.Certificate
					if certBytes, ok := cfg.GetMsps()[MSPID]; ok {
						certs, err := TLSAllCertsByBytes(certBytes.GetTlsRootCerts())
						if err == nil {
							// TODO take the first cert now.
							cert = certs[0]
						}
					}
					for _, endpoint := range ordererCfg.GetEndpoint() {
						URL := fmt.Sprintf("%s:%d", endpoint.GetHost(), endpoint.GetPort())
						orderers = append(orderers, &Orderer{
							Name:        URL,
							URL:         URL,
							GRPCOptions: make(map[string]interface{}),
							TLSCACert:   cert,
							Channels:    util.NewSet(channelID),
						})
					}
				}
			}
		}(peerCfg)

	}

	wg.Wait()

	logger.Infof("Found %d peers from channel %s via endpoints %v.", len(peers), channelID, endPoints)
	return peers, orderers, nil
}

// Show to show details
func (conn *NetworkConnection) Show() string {
	buffer := &bytes.Buffer{}
	cert, err := TLSCertByBytes(conn.Participant.Cert)

	buffer.WriteString(fmt.Sprintf("Participant: %s\n", conn.Participant.Label))
	if err == nil {
		buffer.WriteString(fmt.Sprintf("Cert subject: %s\n", cert.Subject.CommonName))
	}

	buffer.WriteString("================ Org ================\n")
	for orgName, org := range conn.Organizations {
		buffer.WriteString(fmt.Sprintf("Org name: %s, MSPID: %s\n", orgName, org.MSPID))
		buffer.WriteString(fmt.Sprintf("Org peers: %s\n", org.Peers.StringList()))
	}
	buffer.WriteString("================ Orderer ================\n")
	for ordName, ord := range conn.Orderers {
		buffer.WriteString(fmt.Sprintf("Orderer name: %s, URL: %s\n", ordName, ord.URL))
		buffer.WriteString(fmt.Sprintf("Orderer options: %v\n", ord.GRPCOptions))
		buffer.WriteString(fmt.Sprintf("Orderer cert: %v\n", ord.TLSCACert.Subject.CommonName))
		buffer.WriteString(fmt.Sprintf("Orderer channels: %v\n", ord.Channels.StringList()))
	}
	buffer.WriteString("================ Channel ================\n")
	for channelID, channel := range conn.Channels {
		buffer.WriteString(fmt.Sprintf("Channel name: %s\n", channelID))
		buffer.WriteString(fmt.Sprintf("Channel peers: %s\n", channel.Peers.StringList()))
		buffer.WriteString(fmt.Sprintf("Channel orderers: %s\n", channel.Orderers.StringList()))
		buffer.WriteString(fmt.Sprintf("Channel policies: %v\n", channel.Policies))
	}

	buffer.WriteString("================ Peer ================\n")
	for peerName, peer := range conn.Peers {
		buffer.WriteString("-------------------------------------\n")
		buffer.WriteString(fmt.Sprintf("Peer name: %s\n", peerName))
		buffer.WriteString(fmt.Sprintf("Peer: %s\n", peer.OrgName))
		if peer.TLSCACert != nil {
			buffer.WriteString(fmt.Sprintf("Peer: %s\n", peer.TLSCACert.Subject.CommonName))
		}
		buffer.WriteString(fmt.Sprintf("Peer: %s\n", peer.MSPID))
		buffer.WriteString(fmt.Sprintf("Peer: %s\n", peer.URL))
		buffer.WriteString(fmt.Sprintf("Peer: %v\n", peer.GRPCOptions))
		buffer.WriteString(fmt.Sprintf("Peer: %s\n", peer.Channels.StringList()))
		buffer.WriteString(fmt.Sprintf("Peer: %v\n", peer.IsConfigured))
		for channel, pcc := range peer.PeerChannelConfigs {
			buffer.WriteString(fmt.Sprintf("Peer: channel %s %v\n", channel, pcc))
		}
		buffer.WriteString(fmt.Sprintf("Peer Status: %v\n", conn.EndpointStatuses[peer.Name]))
	}

	buffer.WriteString("================ Channel Orderers ================\n")
	for channelID, channelOrderers := range conn.ChannelOrderers {
		buffer.WriteString("-------------------------------------\n")
		buffer.WriteString(fmt.Sprintf("Channel name: %s\n", channelID))
		for _, orderer := range channelOrderers {
			buffer.WriteString(fmt.Sprintf("Orderer name: %s\n", orderer.Name))
			buffer.WriteString(fmt.Sprintf("Orderer url: %s\n", orderer.URL))
		}
	}

	buffer.WriteString("================ Channel ledgers ================\n")
	for channelID, ledger := range conn.ChannelLedgers {
		buffer.WriteString("-------------------------------------\n")
		buffer.WriteString(fmt.Sprintf("Channel name: %s\n", channelID))
		buffer.WriteString(fmt.Sprintf("Ledger: %v\n", ledger))
	}

	buffer.WriteString("================ Channel Chaincodes ================\n")
	for channelID, ccs := range conn.ChannelChaincodes {
		buffer.WriteString("-------------------------------------\n")
		buffer.WriteString(fmt.Sprintf("Channel name: %s\n", channelID))
		for _, cc := range ccs {
			buffer.WriteString(fmt.Sprintf("Chaincode: %s - %s\n", cc.Name, cc.Version))
		}
	}

	return buffer.String()
}

// UpdateWithDiscovery update the network connection based on the discovery result, if useDiscovery option is true.
func (conn *NetworkConnection) updateConnection() error {
	occr := &OrdererConfigCognRes{OrdererConfigs: make(map[string]*fab.OrdererConfig)}
	cccr := &ChannelConfigCongnRes{ChannelConfigs: make(map[string]*fab.ChannelEndpointConfig)}
	cpcr := &ChannelPeerCongnRes{ChannelPeersList: make(map[string][]fab.ChannelPeer)}

	conn.updateOrdererConfig(occr)
	conn.updateChannelConfig(cccr)
	conn.updateChannelPeers(cpcr)

	sdk, err := fabsdk.New(config.FromRaw(conn.ConnectionProfile.Config, conn.ConnectionProfile.ConfigType),
		fabsdk.WithEndpointConfig(cpcr, cccr, occr))

	if err != nil {
		return err
	}
	signID, err := CreateSigningIdentity(sdk, conn.Participant)
	if err != nil {
		return err
	}
	conn.Participant.SignID = signID
	ctxProvider := sdk.Context(fabsdk.WithIdentity(signID))

	ctx, err := ctxProvider()
	if err != nil {
		return err
	}

	// Close the old sdk connection.
	conn.SDK.Close()

	conn.SDK = sdk
	conn.ClientProvider = ctxProvider
	conn.Client = ctx
	return nil
}

func (conn *NetworkConnection) updateChannelLedgers() {
	for channelID := range conn.Channels {
		// Auto target
		ledger, err := QueryLedger(conn, channelID, nil)
		if err == nil {
			conn.ChannelLedgers[channelID] = ledger
		}
	}
}

// GetChannelOrderers to get all orderers per channel.
func (conn *NetworkConnection) updateChannelOrderers() {
	for channelID, channel := range conn.Channels {
		channelOrderers := []*Orderer{}
		for _, ordName := range channel.Orderers.StringList() {
			if orderer := conn.findOrderer(ordName); orderer != nil {
				channelOrderers = append(channelOrderers, orderer)
			}
		}
		conn.ChannelOrderers[channelID] = channelOrderers
	}
}

func (conn *NetworkConnection) updateChannelChaincodes() {
	for channelID := range conn.Channels {
		// Auto target
		instantiatedPerChannelCCs, err := QueryInstantiatedChaincodes(conn, channelID)
		if err == nil {
			conn.ChannelChaincodes[channelID] = instantiatedPerChannelCCs
		}
	}
}

// QueryInstantiatedChaincodes to get all instantiated chaincodes per channel.
func QueryInstantiatedChaincodes(conn *NetworkConnection, channelID string) ([]*Chaincode, error) {
	resMgmtClient, err := resmgmt.New(conn.ClientProvider)
	if err != nil {
		return nil, err
	}
	// All instantiated chaincodes for channel. The channel must been configured in profile, or be update by cognition.
	ccRes, err := resMgmtClient.QueryInstantiatedChaincodes(channelID)
	if err != nil {
		return nil, err
	}
	ccs := []*Chaincode{}
	for _, instantiatedCC := range ccRes.GetChaincodes() {
		ccs = append(ccs, &Chaincode{
			Name:      instantiatedCC.GetName(),
			Version:   instantiatedCC.GetVersion(),
			ChannelID: channelID,
			Path:      instantiatedCC.GetPath(),
			// Installed: not sure now.
		})
	}
	return ccs, nil

}

// ConfiguredPeers return the configured peers of the network.
func (conn *NetworkConnection) ConfiguredPeers() map[string]fab.PeerConfig {
	return conn.Client.EndpointConfig().NetworkConfig().Peers
}

// ConfiguredChannels return the configured channels of the network.
func (conn *NetworkConnection) ConfiguredChannels() map[string]fab.ChannelEndpointConfig {
	return conn.Client.EndpointConfig().NetworkConfig().Channels
}

// ConfiguredOrderers return the configured orderers of the network.
func (conn *NetworkConnection) ConfiguredOrderers() map[string]fab.OrdererConfig {
	return conn.Client.EndpointConfig().NetworkConfig().Orderers
}

// ConfiguredOrganizations return the configured organizations of the network.
func (conn *NetworkConnection) ConfiguredOrganizations() map[string]fab.OrganizationConfig {
	return conn.Client.EndpointConfig().NetworkConfig().Organizations
}

// Return org name, MSPID
func (conn *NetworkConnection) findConfiguredPeerOrg(peerName string) (string, string) {
	for orgName, org := range conn.ConfiguredOrganizations() {
		for _, tmpPeer := range org.Peers {
			if tmpPeer == peerName {
				return orgName, org.MSPID
			}
		}
	}
	return "", ""
}

func (conn *NetworkConnection) findOrgName(MSPID string) string {
	for orgName, org := range conn.ConfiguredOrganizations() {
		if org.MSPID == MSPID {
			return orgName
		}
	}
	return MSPID
}

// Close to close the underlying connection.
// TODO if use session, it should not be closed.
func (conn *NetworkConnection) Close() {
	if conn != nil && conn.SDK != nil {
		conn.SDK.Close()
	}
}

// OrdererConfigCognRes option config interface, See fabric-sdk-go/pkg/fab/opts.go
type OrdererConfigCognRes struct {
	OrdererConfigs map[string]*fab.OrdererConfig
}

//OrdererConfig overrides EndpointConfig's OrdererConfig function which returns the ordererConfig instance for the name/URL arg
func (occr *OrdererConfigCognRes) OrdererConfig(nameOrURL string) (*fab.OrdererConfig, bool) {
	for name, ord := range occr.OrdererConfigs {
		if name == nameOrURL || ord.URL == nameOrURL {
			return ord, true
		}
	}
	return nil, false
}

// CognOrdererConfig to cognize the orderers.
func (conn *NetworkConnection) updateOrdererConfig(occr *OrdererConfigCognRes) {
	for _, ord := range conn.Orderers {
		occr.OrdererConfigs[ord.Name] = &fab.OrdererConfig{
			URL:         ord.URL,
			GRPCOptions: ord.GRPCOptions,
			// map[string]interface{}{
			// 	"ssl-target-name-override": edp.GetHost(),
			// 	// TODO properties
			// 	// "keep-alive-time":          0 * time.Second,
			// 	// "keep-alive-timeout":       200 * time.Second,
			// 	// "keep-alive-permit":        false,
			// 	// "fail-fast":                false,
			// 	// "allow-insecure":           false,
			// },
			// TODO msp certs might have multiple, but this TLSCACert has only one.
			TLSCACert: ord.TLSCACert,
		}
	}
}

// ChannelPeerCongnRes option config interface, See fabric-sdk/go/pkg/fab/opts.go
type ChannelPeerCongnRes struct {
	ChannelPeersList map[string][]fab.ChannelPeer
}

// ChannelPeers to find peers of channel.
func (cpcr *ChannelPeerCongnRes) ChannelPeers(channelID string) []fab.ChannelPeer {
	peers, ok := cpcr.ChannelPeersList[channelID]
	if !ok {
		return []fab.ChannelPeer{}
	}
	return peers
}

// CognChannelPeers overrides EndpointConfig's ChannelPeers function which returns the list of peers for the channel name arg
// TODO to client config look to find the peer name in config.
func (conn *NetworkConnection) updateChannelPeers(cpcr *ChannelPeerCongnRes) {
	for channelID, channel := range conn.Channels {
		channelPeers := []fab.ChannelPeer{}
		for _, peerName := range channel.Peers.StringList() {
			if peer := conn.findPeer(peerName); peer != nil {
				channelPeers = append(channelPeers, fab.ChannelPeer{
					PeerChannelConfig: *peer.PeerChannelConfigs[channelID],
					NetworkPeer: fab.NetworkPeer{
						PeerConfig: fab.PeerConfig{
							URL:         peer.URL,
							GRPCOptions: peer.GRPCOptions,
							TLSCACert:   peer.TLSCACert,
						},
						MSPID: peer.MSPID,
					},
				})
			}
		}
		cpcr.ChannelPeersList[channelID] = channelPeers
	}

}

// ChannelConfigCongnRes implementation of option config, ordererConfig interface. See fabric-sdk/go/pkg/fab/opts.go
type ChannelConfigCongnRes struct {
	ChannelConfigs map[string]*fab.ChannelEndpointConfig
}

// ChannelConfig to find config of chanel.
func (cccr *ChannelConfigCongnRes) ChannelConfig(channelID string) *fab.ChannelEndpointConfig {
	ch, ok := cccr.ChannelConfigs[channelID]
	if !ok {
		return &fab.ChannelEndpointConfig{}
	}
	return ch
}

// UpdateChannelConfig to congnize channel config.
// cec configured in connection profile.
func (conn *NetworkConnection) updateChannelConfig(cccr *ChannelConfigCongnRes) {
	for channelID, channel := range conn.Channels {
		peerChannelConfigs := make(map[string]fab.PeerChannelConfig)
		for _, peerName := range channel.Peers.StringList() {
			if peer := conn.findPeer(peerName); peer != nil {
				peerChannelConfigs[peerName] = *peer.PeerChannelConfigs[channelID]
			}
		}

		channelEndpointConfig := &fab.ChannelEndpointConfig{
			Peers: peerChannelConfigs,
			// TODO *** policies
			Policies: fab.ChannelPolicies{
				QueryChannelConfig: fab.QueryChannelConfigPolicy{
					// Below 2 very important. Otherwise will be error with no target.
					MinResponses: 1,
					MaxTargets:   1,
					/////////////////////////////////////////////////////////////////
					RetryOpts: retry.Opts{
						Attempts:       5,
						InitialBackoff: 500 * time.Millisecond,
						MaxBackoff:     5 * time.Second,
						BackoffFactor:  2.0,
					},
				},
				EventService: fab.EventServicePolicy{
					ResolverStrategy:                 fab.MinBlockHeightStrategy,
					MinBlockHeightResolverMode:       fab.ResolveByThreshold,
					BlockHeightLagThreshold:          5,
					ReconnectBlockHeightLagThreshold: 10,
					PeerMonitorPeriod:                5 * time.Second,
				},
			},
		}
		cccr.ChannelConfigs[channelID] = channelEndpointConfig
	}
}

// TLSAllCertsByBytes to make cert from bytes.
func TLSAllCertsByBytes(certBytes [][]byte) ([]*x509.Certificate, error) {
	certs := []*x509.Certificate{}
	for _, certBytes := range certBytes {
		cert, err := TLSCertByBytes(certBytes)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
	}

	return certs, nil
}

// TLSCertByBytes to make cert from bytes.
func TLSCertByBytes(bytes []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(bytes)
	if block == nil {
		return nil, errors.Errorf("Decode pem certificate got failed.")
	}

	pub, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, errors.WithMessage(err, "Parse x509 certificate got failed.")
	}
	return pub, nil
}
