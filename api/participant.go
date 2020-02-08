package api

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	fabImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/sw"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	mspImpl "github.com/hyperledger/fabric-sdk-go/pkg/msp"
)

type Participant struct {
	Label      string
	OrgName    string
	MSPID      string
	Cert       []byte
	PrivateKey []byte
	SignID     msp.SigningIdentity
}

// Find the org name definied in the connection profile.
func findOrgNameByMSPID(mspID string, orgs map[string]fab.OrganizationConfig) string {
	for name, org := range orgs {
		if org.MSPID == mspID {
			return name
		}
	}
	return ""
}

// CreateSigningIdentity To create identity by orgname, cert and key.
func CreateSigningIdentity(sdk *fabsdk.FabricSDK, participant *Participant) (msp.SigningIdentity, error) {
	// configBackend, err := confProvider()
	// Reuse the configure provider for sdk.
	configBackend, err := sdk.Config()
	if err != nil {
		return nil, err
	}

	// Use file store
	// userStore, err := msp.NewCertFileUserStore(identityConfig.CredentialStorePath())
	userStore := mspImpl.NewMemoryUserStore()

	cryptSuiteConfig := cryptosuite.ConfigFromBackend(configBackend)
	cryptoSuite, err := sw.GetSuiteByConfig(cryptSuiteConfig)
	if err != nil {
		return nil, err
	}

	endpointConfig, err := fabImpl.ConfigFromBackend(configBackend)

	orgs := endpointConfig.NetworkConfig().Organizations

	if participant.OrgName == "" {
		participant.OrgName = findOrgNameByMSPID(participant.MSPID, orgs)
		if participant.OrgName == "" {
			return nil, errors.Errorf("MSPID %s not found in the connection profile.", participant.MSPID)
		}
	}

	mgr, err := mspImpl.NewIdentityManager(participant.OrgName, userStore, cryptoSuite, endpointConfig)
	if err != nil {
		return nil, err
	}

	id, err := mgr.CreateSigningIdentity(msp.WithCert(participant.Cert), msp.WithPrivateKey(participant.PrivateKey))
	if err != nil {
		return nil, err
	}

	return id, nil
}
