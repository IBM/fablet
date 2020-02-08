package api

import (
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/seek"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

func TestEvent(t *testing.T) {
	t.Log("Begin event test")
	conn, err := getConnectionSimple()
	if err != nil {
		t.Fatal(err)
	}

	channelContext := conn.SDK.ChannelContext(mychannel, fabsdk.WithIdentity(conn.SignID))
	eventClient, err := event.New(channelContext, event.WithBlockEvents(), event.WithSeekType(seek.Newest))
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("Begin event...")

	// reg, notifier, err := eventClient.RegisterChaincodeEvent("vehiclesharing", ".*")
	// defer eventClient.Unregister(reg)
	// select {
	// case event := <-notifier:
	// 	t.Log(event)
	// case <-time.After(time.Second * 20):
	// 	t.Log("No events, so then quit.")
	// }

	blockEvent := func() {
		fmt.Println(" ===================== block event =======================")
		reg, notifier, err := eventClient.RegisterBlockEvent()
		if err != nil {
			return
		}
		defer eventClient.Unregister(reg)
		for {
			select {
			case event := <-notifier:
				fmt.Println(" ===================== block event =======================")
				fmt.Println(event.SourceURL)
				fmt.Println(hex.EncodeToString(event.Block.GetHeader().GetDataHash()))
			}
			time.Sleep(time.Second * 2)
		}
	}

	filteredBlockEvent := func() {
		fmt.Println(" ===================== filter block event =======================")
		reg, notifier, err := eventClient.RegisterFilteredBlockEvent()
		if err != nil {
			return
		}
		defer eventClient.Unregister(reg)
		for {
			select {
			case event := <-notifier:
				fmt.Println(" ===================== filter block event =======================")
				fmt.Println(event.SourceURL)
				fmt.Println(event.FilteredBlock.GetChannelId(), event.FilteredBlock.GetNumber())
			}
			time.Sleep(time.Second * 2)
		}
	}

	go blockEvent()
	go filteredBlockEvent()

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, syscall.SIGTERM)
	<-s
}
