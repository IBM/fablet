package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/IBM/fablet/log"

	"github.com/IBM/fablet/service"
)

// SrvPort The default http listening port
const DefaultSrvPort = 8080

var logger = log.GetLogger()

// TODO
// TODO to persist the connection in session
func getHandlerMap() map[string]service.HTTPHandler {
	handlerMap := map[string]service.HTTPHandler{
		// "/":                      service.HandleRoot,
		//"/network/connect":       service.Post(service.HandleNetworkConnect),
		"/network/discover":      service.Post(service.HandleNetworkDiscover),
		"/network/refresh":       service.Post(service.HandleNetworkRefresh),
		"/peer/details":          service.Post(service.HandlePeerDetails),
		"/chaincode/install":     service.Post(service.HandleChaincodeInstall),
		"/chaincode/instantiate": service.Post(service.HandleChaincodeInstantiate),
		"/chaincode/upgrade":     service.Post(service.HandleChaincodeUpgrade),
		"/chaincode/execute":     service.Post(service.HandleChaincodeExecute),
		"/ledger/query":          service.Post(service.HandleLedgerQuery),
		"/ledger/block":          service.Post(service.HandleBlockQuery),
		"/ledger/blockany":       service.Post(service.HandleBlockQueryAny),
		"/channel/create":        service.Post(service.HandleCreateChannel),
		"/channel/join":          service.Post(service.HandleJoinChannel),
		"/event/blockevent":      service.WS(service.HandleBlockEvent),
	}

	return handlerMap
}

func main() {
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, syscall.SIGTERM)

	port := flag.Int("port", DefaultSrvPort, "Listen on port")
	flag.Parse()

	for url, handler := range getHandlerMap() {
		http.HandleFunc(url, handler)
	}
	http.Handle("/", http.FileServer(http.Dir(filepath.Join(service.ExeFolder, "web"))))

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
			logger.Error(err.Error())
		}
	}()

	logger.Infof("Fablet server starts at %d.", *port)
	<-s
	logger.Info("Fablet server Exits.")
}
