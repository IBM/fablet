package api

import (
	"crypto/x509"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/IBM/fablet/log"

	"github.com/IBM/fablet/util"
)

var logger = log.GetLogger()

const DiscoverTimeOut = 10 * time.Second

// ResultCode uint32
type ResultCode uint32

const (
	// ResultSuccess success
	ResultSuccess ResultCode = 0
	// ResultFailure failure
	ResultFailure ResultCode = 1
)

const (
	// LSCC code of lifecycle chaincode
	LSCC = "lscc"
)

// ExecutionResult execution result
type ExecutionResult struct {
	Code    ResultCode  `json:"code"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

// Peer common peer instance
type Peer struct {
	Name               string                            `json:"name"`
	OrgName            string                            `json:"orgName"`
	MSPID              string                            `json:"MSPID"`
	URL                string                            `json:"URL"`         // from PeerConfig
	TLSCACert          *x509.Certificate                 `json:"TLSCACert"`   // from PeerConfig
	GRPCOptions        map[string]interface{}            `json:"GRPCOptions"` // from PeerConfig
	Channels           util.Set                          `json:"channels"`
	PeerChannelConfigs map[string]*fab.PeerChannelConfig `json:"peerChannelConfigs"` // map[channelID]
	IsConfigured       bool                              `json:"isConfigured"`
	UpdateTime         int64                             `json:"updateTime"`
	// Chaincodes         []*Chaincode                      `json:"chaincodes"` //TODO @deprecated
}

// AddChannel add channel ID
// func (peer *Peer) AddChannel(channelID string) {
// 	if !util.ExistString(peer.Channels, channelID) {
// 		peer.Channels = append(peer.Channels, channelID)
// 	}
// }

// Orderer common orderer instance
type Orderer struct {
	Name        string                 `json:"name"`
	URL         string                 `json:"URL"`         // from OrdererConfig
	GRPCOptions map[string]interface{} `json:"TLSCACert"`   // from OrdererConfig
	TLSCACert   *x509.Certificate      `json:"GRPCOptions"` // from OrdererConfig
	Channels    util.Set               `json:"channels"`
}

// Organization org
type Organization struct {
	Name       string   `json:"name"`
	MSPID      string   `json:"MSPID"`
	CryptoPath string   `json:"cryptoPath"`
	Peers      util.Set `json:"peers"`
}

// Channel channel
type Channel struct {
	ChannelID string               `json:"channelID"`
	Peers     util.Set             `json:"peers"`
	Orderers  util.Set             `json:"orderers"`
	Policies  *fab.ChannelPolicies `json:"policies"`
}

// NetworkConnection the entry to the Fabric network.
// TODO to have lock for update, and a safe disconnection when update/close.
type NetworkConnection struct {
	// Materials to initialize the connection
	*ConnectionProfile
	*Participant
	UseDiscovery bool

	// To identify the connection from others. Normally be hash of connection profile, participant and useDiscovery.
	Identifier string
	UpdateTime time.Time
	ActiveTime time.Time

	// Some intermediate/context based on the sdk.
	SDK            *fabsdk.FabricSDK
	Client         context.Client
	ClientProvider context.ClientProvider

	// Conifguration and discovered result, they are only for indication or presentation, not for the network directly.
	Channels      map[string]*Channel
	Organizations map[string]*Organization
	Peers         map[string]*Peer
	Orderers      map[string]*Orderer

	ChannelLedgers    map[string]*Ledger
	ChannelChaincodes map[string][]*Chaincode
	ChannelOrderers   map[string][]*Orderer
}

// NetworkOverview for whole network
type NetworkOverview struct {
	Peers             []*Peer                 `json:"peers"`
	ChannelOrderers   map[string][]*Orderer   `json:"channelOrderers"`
	ChannelLedgers    map[string]*Ledger      `json:"channelLedgers"`
	ChannelChainCodes map[string][]*Chaincode `json:"channelChaincodes"`
}
