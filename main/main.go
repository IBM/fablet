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
		"/event/chaincodeevent":  service.WS(service.HandleChaincodeEvent),
	}

	return handlerMap
}

func main() {
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, syscall.SIGTERM)

	addr := flag.String("addr", "", "Listen on the TCP address (default all)")
	port := flag.Int("port", DefaultSrvPort, "Listen on port")
	cert := flag.String("cert", "", "TLS cert (default non-https)")
	key := flag.String("key", "", "TLS key (default non-https)")
	flag.Parse()

	addrPort := fmt.Sprintf("%s:%d", *addr, *port)

	for url, handler := range getHandlerMap() {
		http.HandleFunc(url, handler)
	}
	http.Handle("/", http.FileServer(http.Dir(filepath.Join(service.ExeFolder, "web"))))

	go func() {
		var err error
		if *cert != "" && *key != "" {
			err = http.ListenAndServeTLS(addrPort, *cert, *key, nil)
		} else {
			err = http.ListenAndServe(addrPort, nil)
		}
		if err != nil {
			logger.Error(err.Error())
			s <- syscall.SIGTERM
		}
	}()

	logger.Infof("Fablet serve at %s.", addrPort)
	<-s
	logger.Info("Fablet server Exits.")
}
