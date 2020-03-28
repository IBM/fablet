package api

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
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

	// Mock client termination.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	// Mock background goroutine of web service.
	ctxBg, cancelBg := context.WithTimeout(context.Background(), time.Second*50)

	go func() {
		defer func() {
			logger.Info("Event end")
			closeChan <- 0
			cancel()
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
			case <-ctx.Done():
				return
			}
		}
	}()

	<-ctxBg.Done()
	cancelBg()
	logger.Debugf("Background context ends.")
}

func TestChaincodeEvent(t *testing.T) {
	t.Log("Begin event test")
	conn, err := getConnectionSimple()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("Begin event...")

	eventChan := make(chan *fab.CCEvent, 1)
	closeChan := make(chan int, 1)
	eventCloseChan := make(chan error, 1)

	go MonitorChaincodeEvent(conn, mychannel, vehiclesharing, ".*(create|update).*", eventChan, closeChan, eventCloseChan)

	go func() {
		i := 0
		for {
			fmt.Println(i, " second ...")
			time.Sleep(time.Second)
			i++
		}
	}()

	// Mock client termination.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*300)

	defer func() {
		logger.Info("Event end")
		closeChan <- 0
		cancel()
	}()
	for {
		select {
		case event := <-eventChan:
			fmt.Println(event.BlockNumber, event.ChaincodeID, event.EventName, event.SourceURL, event.SourceURL, string(event.Payload))
		case <-eventCloseChan:
			return
		case <-ctx.Done():
			return
		}
	}

}
