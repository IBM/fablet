/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
)

const (
	chaincodeName = "vehiclesharing"
	chaincodePath = "fablet/vehiclesharing"
	gopath        = "../test/chaincodes/vehiclesharing"
	// Chaincode path after 'src', and the Chaincode path will be also used for installation. See below.
	// Gopath path before 'src'
)

// Refined.

func getRandomCCVersion() string {
	return uuid.NewV1().String()[0:8]
}

func getRandCC() *Chaincode {
	ccVersion := getRandomCCVersion()
	ccName := "vs" + ccVersion
	return &Chaincode{
		Name:     ccName,
		Version:  ccVersion,
		Path:     chaincodePath,
		BasePath: gopath,
		Type:     "golang",
	}
}

func TestInstallCCByAPI(t *testing.T) {
	cc := getRandCC()
	t.Log("Begin install chaincode: ", cc.Name, cc.Version)
	conn, err := getConnectionSimple()
	if err != nil {
		t.Fatal(err)
	}
	res, err := InstallChaincode(conn, cc, []string{target01})
	if err != nil {
		t.Log(err.Error())
	}
	for k, m := range res {
		t.Log(k, m)
	}
}

func TestInstallAndInstantiateCCByAPI(t *testing.T) {
	cc := getRandCC()
	t.Log("Begin install and instantiate chaincode: ", cc.Name, cc.Version)
	conn, err := getConnectionSimple()

	res, err := InstallChaincode(conn, cc, []string{target01})
	if err != nil {
		t.Fatal(err.Error())
	}
	for k, m := range res {
		t.Log(k, m)
	}

	t.Log("Wait for 2 seconds.")
	time.Sleep(time.Second * 2)

	cc.Policy = "OR ('Org1MSP.peer','Org2MSP.peer')"
	cc.Constructor = []string{}
	cc.ChannelID = mychannel

	tid, err := InstantiateChaincode(conn, cc, target01, orderer)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Succeed instantiate chaincode %s.", string(tid))
}

func TestInstallAndUpgradeCCByAPI(t *testing.T) {
	ccVersion := getRandomCCVersion()
	ccName := "vehiclesharing"
	cc := &Chaincode{
		Name:     ccName,
		Version:  ccVersion,
		Path:     chaincodePath,
		BasePath: gopath,
		Type:     "golang",
	}
	t.Log("Begin install and upgrade chaincode: ", cc.Name, cc.Version)
	conn, err := getConnectionSimple()

	res, err := InstallChaincode(conn, cc, []string{target01})
	if err != nil {
		t.Fatal(err.Error())
	}
	for k, m := range res {
		t.Log(k, m)
	}

	t.Log("Wait for 2 seconds.")
	time.Sleep(time.Second * 2)

	cc.Policy = "OR ('Org1MSP.peer','Org2MSP.peer')"
	cc.Constructor = []string{}
	cc.ChannelID = mychannel

	tid, err := UpgradeChaincode(conn, cc, target01, orderer)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Succeed upgrade chaincode %s.", string(tid))
}

func TestExecuteCCByAPI(t *testing.T) {
	t.Log("Begin execute chaincode")
	conn, err := getConnectionSimple()

	r := getRandomCCVersion()
	res, err := ExecuteChaincode(conn, mychannel, "vehiclesharing", ChaincodeOperTypeExecute,
		[]string{target01}, "createVehicle", []string{"k_" + r, "b" + r})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Transaction ID:", res.TransactionID)
	t.Log("TxValidationCode", res.TxValidationCode)
	t.Log("ChaincodeStatus", res.ChaincodeStatus)
	t.Log("Payload", string(res.Payload))

	for _, rr := range res.Responses {
		t.Log("<<<<<<<<<<<<<<<<<<<<<", rr, ">>>>>>>>>>>>>>>>>>>>>")
		t.Log(rr.Endorser)
		t.Log(rr.GetVersion())
		t.Log(string(rr.GetResponse().GetPayload()))
		t.Log(rr.GetResponse().GetStatus())
	}
}

func TestExecuteCCRoutine(t *testing.T) {
	t.Log("Begin execute chaincode")
	conn, _ := getConnectionSimple()

	//var wg sync.WaitGroup

	for i := 0; i < 12; i++ {
		//wg.Add(1)
		go func(i int) {
			fmt.Printf("Executing %d", i)
			//defer wg.Done()
			for {
				time.Sleep(time.Second * time.Duration(10*uint(rand.Float32())))
				r := getRandomCCVersion()
				res, err := ExecuteChaincode(conn, mychannel, "vehiclesharing", ChaincodeOperTypeExecute,
					[]string{target01}, "createVehicle", []string{"k_" + r, "b" + r})
				if err != nil {
					t.Fatal(err)
				}
				fmt.Println("Transaction ID:", res.TransactionID)
			}
		}(i)
	}

	time.Sleep(time.Second * 3600)

	//wg.Wait()
}
