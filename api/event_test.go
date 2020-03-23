package api

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

func TestBlockEvent(t *testing.T) {
	t.Log("Begin event test")
	conn, err := getConnectionSimple()
	if err != nil {
		t.Fatal(err)
	}

	eventChan := make(chan *fab.FilteredBlockEvent, 1)
	closeChan := make(chan int, 1)
	eventCloseChan := make(chan int, 1)
	go MonitorBlockEvent(conn, mychannel, eventChan, closeChan, eventCloseChan)

	defer func() {
		logger.Info("Event end")
		closeChan <- 0
	}()

	for {
		select {
		case event := <-eventChan:
			logger.Debugf("Get an event from %s.", event.SourceURL)
			if event == nil {
				// return errors.Errorf("Event is nil")
			} else {
				fmt.Println(event.FilteredBlock.GetNumber())
			}
		case <-eventCloseChan:
			return
		}
	}
}

func TestChaincodeEvent(t *testing.T) {
	t.Log("Begin event test")
	conn, err := getConnectionSimple()
	if err != nil {
		t.Fatal(err)
	}

	channelContext := conn.SDK.ChannelContext("chtest1", fabsdk.WithIdentity(conn.SignID))
	channelClient, err := channel.New(channelContext)
	if err != nil {
		t.Fatal(err)
	}

	// eventClient, err := event.New(channelContext, event.WithBlockEvents(), event.WithSeekType(seek.Newest))

	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("Begin event...")

	ccBlockEvent := func() {
		fmt.Println(" ===================== chaincode event =======================")
		reg, notifier, err := channelClient.RegisterChaincodeEvent("vseventtest", ".*")

		if err != nil {
			fmt.Println(err)
			return
		}
		defer channelClient.UnregisterChaincodeEvent(reg)
		for {
			select {
			case event := <-notifier:
				fmt.Println(event.EventName, event.ChaincodeID, event.BlockNumber, event.SourceURL, string(event.Payload), event.TxID)
			}
			time.Sleep(time.Second * 2)
		}
	}

	go ccBlockEvent()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	<-ctx.Done()
}
