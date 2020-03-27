package api

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

// MonitorBlockEvent to monitor block event
func MonitorBlockEvent(conn *NetworkConnection, channelID string,
	eventChan chan<- *fab.FilteredBlockEvent, closeChan <-chan int, eventCloseChan chan<- int) error {
	logger.Debugf("MonitorBlockEvent of %s begins.", channelID)

	channelContext := conn.SDK.ChannelContext(channelID, fabsdk.WithIdentity(conn.SignID))
	eventClient, err := event.New(channelContext)
	if err != nil {
		eventCloseChan <- 0
		logger.Debugf("Creating event got failed: %s.", err.Error())
		return err
	}

	reg, notifier, err := eventClient.RegisterFilteredBlockEvent()
	// TODO if the event service is closed, i.e., the peer is shut dow.
	if err != nil {
		eventCloseChan <- 0
		logger.Debugf("Registration event got failed: %s.", err.Error())
		return err
	}

	// TODO To unregister the event under all scenarios.
	// Make sure that this goroutine runs under a background context, such as web service, otherwise
	// it might be discarded when the system process exists.
	defer func() {
		eventClient.Unregister(reg)
		eventCloseChan <- 0
		logger.Debugf("MonitorBlockEvent of %s event unregistered.", channelID)
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

// MonitorChaincodeEvent to monitor chaincode event
func MonitorChaincodeEvent(conn *NetworkConnection, channelID string, chaincodeID string, eventFilter string,
	eventChan chan<- *fab.CCEvent, closeChan <-chan int, eventCloseChan chan<- int) error {
	logger.Debugf("MonitorChaincodeEvent of %s %s %s begins.", channelID, chaincodeID, eventFilter)

	channelContext := conn.SDK.ChannelContext(channelID, fabsdk.WithIdentity(conn.SignID))
	eventClient, err := event.New(channelContext, event.WithBlockEvents())
	if err != nil {
		eventCloseChan <- 0
		logger.Debugf("Creating event got failed: %s.", err.Error())
		return err
	}

	reg, notifier, err := eventClient.RegisterChaincodeEvent(chaincodeID, eventFilter)
	// TODO if the event service is closed, i.e., the peer is shut dow.
	if err != nil {
		eventCloseChan <- 0
		logger.Debugf("Registration event got failed: %s.", err.Error())
		return err
	}

	defer func() {
		eventClient.Unregister(reg)
		eventCloseChan <- 0
		logger.Debugf("MonitorChaincodeEvent of %s %s %s unregistered.", channelID, chaincodeID, eventFilter)
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
