package api

import (
	"fmt"
	"sync"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	packager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/pkg/errors"
)

// Chaincode chaincode
type Chaincode struct {
	// For installation and instantiation
	Name    string `json:"name"`
	Version string `json:"version"`
	Path    string `json:"path"` // for go, it is package
	// For installtion only
	BasePath string `json:"basePath"` // for go, it is go path. <BasePath>/src/<Path> will be the go chaincode real path.
	Type     string `json:"type"`     // see pb.ChaincodeSpec_Type
	Package  []byte `json:"package"`
	// For instantiation only
	ChannelID   string   `json:"channelID"`   // instantiated in channnel
	Policy      string   `json:"policy"`      // Endorsement policy
	Constructor []string `json:"constructor"` // Arguments for instantiation
	// For status. The chaincode might be instantiated (on channel) by not installed (on peer).
	Installed bool `json:"installed"` // It might be false while channelID is not empty, since it is instantiated in channel but not installed in current peer.
}

const (
	ChaincodeType_GOLANG = "golang"
	ChaincodeType_NODE   = "node"
	ChaincodeType_JAVA   = "java"
)

func (cc *Chaincode) String() string {
	return fmt.Sprintf("%s:%s", cc.Name, cc.Version)
}

// InstallChaincode to handle the detailed corresponding message, and multiple peers - some failed issue, the installation will be executed for multiple times.
// TODO to use resmgmt.InstallCC
func InstallChaincode(conn *NetworkConnection, cc *Chaincode, targets []string) (map[string]ExecutionResult, error) {
	if len(targets) < 1 {
		return nil, errors.New("no any targets to install chaincode")
	}

	var ccPkg *resource.CCPackage
	var err error
	if cc.Type == ChaincodeType_GOLANG {
		ccPkg, err = packager.NewCCPackage(cc.Path, cc.BasePath)
	} else if cc.Type == ChaincodeType_NODE {
		ccPkg, err = NewNodeCCPackage(cc.Path)
	} else if cc.Type == ChaincodeType_JAVA {
		ccPkg, err = NewJavaCCPackage(cc.Path)
	} else {
		return nil, errors.Errorf("%s is not a valid chaincode type.", cc.Type)
	}

	if err != nil {
		return nil, errors.WithMessagef(err, "Error occurred when generating chaincode package of \"%s\"", cc.String())
	}
	icr := resource.InstallChaincodeRequest{Name: cc.Name, Path: cc.Path, Version: cc.Version, Package: ccPkg}
	ctx := conn.Client
	reqCtx, cancel := context.NewRequest(ctx, context.WithTimeoutType(fab.PeerResponse))
	defer cancel()

	var peers []fab.ProposalProcessor
	for _, target := range targets {
		peerCfg, err := comm.NetworkPeerConfig(ctx.EndpointConfig(), target)
		if err != nil {
			return nil, errors.WithMessagef(err, "Error occurred when finding target \"%s\"", target)
		}
		peer, err := ctx.InfraProvider().CreatePeerFromConfig(peerCfg)
		if err != nil {
			return nil, errors.WithMessagef(err, "Error occurred when getting network peer from target \"%s\"", target)
		}
		peers = append(peers, peer)
	}

	var errAll error = nil
	var results = make(map[string]ExecutionResult)
	var wg sync.WaitGroup

	// Run per peer, to get accurate result mapping.
	for _, peer := range peers {
		wg.Add(1)
		go func(processor fab.ProposalProcessor) {
			defer wg.Done()
			peerURL := processor.(fab.Peer).URL()
			logger.Info(fmt.Sprintf("Sending chaincode installation proposal request to %s", peerURL))
			r, _, err := resource.InstallChaincode(reqCtx, icr, []fab.ProposalProcessor{processor}, resource.WithRetry(retry.DefaultResMgmtOpts))
			if err != nil {
				errAll = errors.New("There is at least one chaincode installation got failed")
				results[peerURL] = ExecutionResult{
					Code:    ResultFailure,
					Message: errors.WithMessage(err, "Error occurred when installing chaincode").Error()}
				return
			}
			// Must be 1 response
			response := r[0]
			if response.Status != int32(common.Status_SUCCESS) {
				errAll = errors.New("There is at least one chaincode installation got failed")
				results[peerURL] = ExecutionResult{
					Code:    ResultFailure,
					Message: errors.WithMessagef(err, "Error occurred when installing chaincode, the original status is %d", response.Status).Error()}
			} else {
				results[peerURL] = ExecutionResult{Code: ResultSuccess}
			}
		}(peer)
	}

	wg.Wait()
	return results, errAll
}

// InstantiateChaincode to instantiate chaincode.
func InstantiateChaincode(conn *NetworkConnection, cc *Chaincode, target string, orderer string) (fab.TransactionID, error) {
	resMgmtClient, err := resmgmt.New(conn.ClientProvider)
	if err != nil {
		return "", errors.WithMessagef(err, "Failed to create new resource management client.")
	}

	ccPolicy, err := cauthdsl.FromString(cc.Policy)
	if err != nil {
		return "", errors.WithMessagef(err, "Failed to sing the policy for instantiation.")
	}

	argBytes := [][]byte{}
	for _, arg := range cc.Constructor {
		argBytes = append(argBytes, []byte(arg))
	}

	insRes, err := resMgmtClient.InstantiateCC(
		cc.ChannelID,
		resmgmt.InstantiateCCRequest{
			Name:    cc.Name,
			Path:    cc.Path,
			Version: cc.Version,
			Args:    argBytes,
			Policy:  ccPolicy,
			// TODO CollConfig and some other options.
			//CollConfig: collConfigs,
		},
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
		resmgmt.WithTargetEndpoints(target),
		resmgmt.WithOrdererEndpoint(orderer),
	)

	if err != nil {
		return "", errors.WithMessagef(err, "Failed to instantiate the chaincode %s:%s on channel %s.", cc.Name, cc.Version, cc.ChannelID)
	}
	logger.Infof("Succeed instantiated the chaincode %s:%s on channel %s.", cc.Name, cc.Version, cc.ChannelID)

	return insRes.TransactionID, nil
}

// UpgradeChaincode to upgrade chaincode.
func UpgradeChaincode(conn *NetworkConnection, cc *Chaincode, target string, orderer string) (fab.TransactionID, error) {
	resMgmtClient, err := resmgmt.New(conn.ClientProvider)
	if err != nil {
		return "", errors.WithMessagef(err, "Failed to create new resource management client.")
	}

	ccPolicy, err := cauthdsl.FromString(cc.Policy)
	if err != nil {
		return "", errors.WithMessagef(err, "Failed to sing the policy for instantiation.")
	}

	argBytes := [][]byte{}
	for _, arg := range cc.Constructor {
		argBytes = append(argBytes, []byte(arg))
	}

	updRes, err := resMgmtClient.UpgradeCC(
		cc.ChannelID,
		resmgmt.UpgradeCCRequest{
			Name:    cc.Name,
			Path:    cc.Path,
			Version: cc.Version,
			Args:    argBytes,
			Policy:  ccPolicy,
			// TODO CollConfig and some other options.
			//CollConfig: collConfigs,
		},
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
		resmgmt.WithTargetEndpoints(target),
		resmgmt.WithOrdererEndpoint(orderer),
	)

	if err != nil {
		return "", errors.WithMessagef(err, "Failed to upgrade the chaincode %s:%s on channel %s.", cc.Name, cc.Version, cc.ChannelID)
	}
	logger.Infof("Succeed upgrade the chaincode %s:%s on channel %s.", cc.Name, cc.Version, cc.ChannelID)

	return updRes.TransactionID, nil
}

// ChaincodeOperType chaincode operation type of
type ChaincodeOperType int

const (
	// ChaincodeOperTypeExecute execute a chaincode
	ChaincodeOperTypeExecute ChaincodeOperType = iota
	// ChaincodeOperTypeQuery query a chaincode
	ChaincodeOperTypeQuery
)

// ExecuteChaincode to invoke a chaincode
// TODO to determine the targets
func ExecuteChaincode(conn *NetworkConnection, channelID string, chaincodeID string,
	operType ChaincodeOperType, targets []string,
	funcName string, args []string,
	options ...channel.RequestOption) (*channel.Response, error) {
	channelContext := conn.SDK.ChannelContext(channelID, fabsdk.WithIdentity(conn.SignID))
	channelClient, err := channel.New(channelContext)

	if err != nil {
		return nil, errors.WithMessagef(err, "Error occurred when creating a new client for channel %s.", channelID)
	}

	argsByte := make([][]byte, len(args))
	for idx, arg := range args {
		argsByte[idx] = []byte(arg)
	}

	reqOpts := []channel.RequestOption{}
	reqOpts = append(reqOpts, channel.WithTargetEndpoints(targets...))
	reqOpts = append(reqOpts, channel.WithRetry(retry.DefaultChannelOpts))
	reqOpts = append(reqOpts, options...)

	oper := channelClient.Execute
	if operType == ChaincodeOperTypeQuery {
		oper = channelClient.Query
	}

	response, err := oper(
		channel.Request{
			ChaincodeID: chaincodeID,
			Fcn:         funcName,
			Args:        argsByte,
		},
		reqOpts...,
	)

	if err != nil {
		return nil, errors.WithMessagef(err, "Error occurred when executing the chaincode %s.", chaincodeID)
	}

	return &response, nil
}
