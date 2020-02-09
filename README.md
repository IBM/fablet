# Fablet
Fablet is a client tools set for Hyperledger Fabric blockchain platform.  
It can help blockchain participants to connect to Fabric network, and perform operations of chaincode, channel, ledger...  

# Playground
An example Fablet service was set up with a Fabric network, you can try it via:  
https://bctest01.fablet.pub:8081/  

## Initial connection

If it is the first time you access the service, you need to provices some materials to create the connectdion, we provides some examples corresponding to an example Fabric network (Fabric 1.4.3 first-network with 4 nodes, 5 etcdraft orderer nodes). Please download these accordinginly:  

* Connection Profile  
  https://bctest01.fablet.pub:8081/test/conn_profile_simple.yaml

* Identity certificate:  
  https://bctest01.fablet.pub:8081/test/Admin@org1.example.com-cert.pem  

* Identity private key:  
  https://bctest01.fablet.pub:8081/test/Admin@org1.example.com.key

* MSP ID  
  The corresponding MSP ID is `Org1MSP`

## Chaincode

Please find this chaincode (Go) example can be installed via Fablet page.

### Installation
* Tar file  
  https://bctest01.fablet.pub:8081/test/vs_src.tar

* Chaincode path  
  The corresponding chaincode path is `fablet/vs`.

### Instantiation
*This is a lower machine, it might take several minutes...*
* Policy  
  ```
  OR ('Org1MSP.peer','Org2MSP.peer')
  ```

* Constructor parameters  
  Please leave it as blank.

### Execution
* Function name
  ```
  invoke 
  ```

* Arguments
  ```
  v001,brand001
  ```
  
*No more document for user now. I think that user should get all points from the UI directly, instead of documentation.*

# Build

## Prerequisite 

* Go ^1.13.4  
  If there is network issues, please try:  
  `go env -w GOPROXY=https://goproxy.io,direct`

* Node.js ^12.13.0

* yarn ^1.19.1  
  `npm install -g yarn`

* Hyperledger Fabric 1.4.3  
  Currently, Fablet supports Fabric 1.4.3, we are working to adapt to 2.0.0. Please refer to Fabric installation document for details.  
  *You can 

## Download repository

```
git clone https://github.com/IBM/fablet.git
```

## Build

* Build all
  ```
  ./build.sh
  ```

* Build service (go) project only
  ```
  ./build.sh service
  ```

* Build web (js/react) project only
  ```
  ./build web
  ```

## Start

The build output will be found at ./release/<OS_Arch>/fablet.

* Start Fablet as default with http on port 8080:
  ```
  ./release/<OS_Arch>/fablet
  ```

* Start Fablet with https on customized port 8081, and TCP address:  
  ```
  ./release/<OS_Arch>/fablet -addr localhost -port 8081 -cert <tls_cert> -key <tls_private_key>
  ```

When Fablet start, you can access it via browser (We tested it on Chrome and Firefox).

# Development

It is composed of 2 projects: service project and web project. These 2 projects folders can be opened as individual project, by MS Code or other IDE.

## Service project
Under folder `./`.  
It is developed in Go language, it provides web service, and html/js/image host.
* Run in development
  ```
  go run ./main
  ```

## Web project
Under folder `./web`.  
It is developed in Javascript with React.
* Run in development
  ```
  yarn start
  ```
