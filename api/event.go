package api

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/seek"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

// MonitorBlockEvent to monitor block event
func MonitorBlockEvent(conn *NetworkConnection, channelID string,
	eventChan chan<- *fab.FilteredBlockEvent,
	closeChan <-chan int,
	eventCloseChan chan<- int) error {
	logger.Debug("MonitorBlockEvent begins.")
	channelContext := conn.SDK.ChannelContext(channelID, fabsdk.WithIdentity(conn.SignID))
	eventClient, err := event.New(channelContext, event.WithBlockEvents(), event.WithSeekType(seek.Newest))
	if err != nil {
		eventCloseChan <- 0
		return err
	}
	reg, notifier, err := eventClient.RegisterFilteredBlockEvent()

	// TODO if the event service is closed...

	if err != nil {
		eventCloseChan <- 0
		return err
	}
	defer func() {
		logger.Debugf("MonitorBlockEvent ends.")
		eventClient.Unregister(reg)
		eventCloseChan <- 0
	}()

	for {
		select {
		case event := <-notifier:
			eventChan <- event
		case <-closeChan:
			return nil
		}
	}
}
