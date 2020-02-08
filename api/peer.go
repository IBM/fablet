package api

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
)

// QueryInstalledChaincodes to get all installed chaincodes
func QueryInstalledChaincodes(conn *NetworkConnection, endpointURL string, options ...DiscoverOptionFunc) ([]*Chaincode, error) {
	resMgmtClient, err := resmgmt.New(conn.ClientProvider)
	if err != nil {
		return nil, err
	}
	ccRes, err := resMgmtClient.QueryInstalledChaincodes(resmgmt.WithTargetEndpoints(endpointURL))
	if err != nil {
		return nil, err
	}
	ccs := []*Chaincode{}
	for _, installedCC := range ccRes.GetChaincodes() {
		ccs = append(ccs, &Chaincode{
			Name:      installedCC.GetName(),
			Version:   installedCC.GetVersion(),
			Path:      installedCC.GetPath(),
			Installed: true})
	}
	return ccs, nil
}
