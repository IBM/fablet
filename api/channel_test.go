package api

import (
	"testing"
	// "github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	//"github.com/hyperledger/fabric/protos/utils"
	//"github.com/hyperledger/fabric/protos/utils"
	//protosmsp "github.com/hyperledger/fabric/protos/msp"
)

func TestCreateChannelByAPI(t *testing.T) {
	conn, err := getConnectionSimple()
	if err != nil {
		t.Fatal(err)
	}

	channelID, err := CreateChannel(conn, channelTx, orderer)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Create a new channel: ", channelID)
}

func TestJoinChannelByAPI(t *testing.T) {
	conn, err := getConnectionPre()
	if err != nil {
		t.Fatal(err)
	}

	err = JoinChannel(conn, "chtest2", []string{target01}, orderer)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("Adding channel got succeed")
	}
}
