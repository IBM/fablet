package api

import (
	"io/ioutil"
	"path/filepath"
)

// Please change this to your own fabric-samples folder.
var fabricSamplePath = "/home/tom/fabric/fabric-samples"

var yamlConfigType = "yaml"

var connConfig, _ = ioutil.ReadFile("../test/connprofiles/conn_profile_simple.yaml")
var connConfig12, _ = ioutil.ReadFile("../test/connprofiles/conn_profile_simple_12.yaml")
var connConfigPre, _ = ioutil.ReadFile("../test/connprofiles/conn_profile_pre.yaml")

var channelTx, _ = ioutil.ReadFile(filepath.Join(fabricSamplePath, "/first-network/channel-artifacts/chtest.tx"))

var testCert, _ = ioutil.ReadFile(filepath.Join(fabricSamplePath, "/first-network/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem"))
var testPrivKey, _ = ioutil.ReadFile(getOnlyFile(filepath.Join(fabricSamplePath, "/first-network/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore")))

var testCert2, _ = ioutil.ReadFile(filepath.Join(fabricSamplePath, "/first-network/crypto-config/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp/signcerts/Admin@org2.example.com-cert.pem"))
var testPrivKey2, _ = ioutil.ReadFile(getOnlyFile(filepath.Join(fabricSamplePath, "/first-network/crypto-config/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp/keystore")))

var target01 = "peer0.org1.example.com:7051"
var target11 = "peer1.org1.example.com:8051"
var target02 = "peer0.org2.example.com:9051"
var target12 = "peer0.org2.example.com:10051"

var mspIDOrg1 = "Org1MSP"
var mspIDOrg2 = "Org2MSP"

var orderer = "orderer.example.com:7050"
var orderer2 = "orderer2.example.com:8050"

var mychannel = "mychannel"

func getOnlyFile(path string) string {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		panic("err read folder: " + path)
	}
	if len(files) < 1 {
		panic("No at least 1 file")
	}
	return filepath.Join(path, files[0].Name())
}

func getConnectionSimple() (*NetworkConnection, error) {
	return NewConnection(
		&ConnectionProfile{connConfig, yamlConfigType},
		&Participant{"TestAdmin", "", mspIDOrg1, testCert, testPrivKey, nil}, true)
}

func getConnectionPre() (*NetworkConnection, error) {
	return NewConnection(
		&ConnectionProfile{connConfigPre, yamlConfigType},
		&Participant{"TestAdmin", "", mspIDOrg1, testCert, testPrivKey, nil}, true)
}
