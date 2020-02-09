# Fablet
A client tools set for Hyperledger Fabric blockchain platform.  
It is composed of 2 project:   
* Service project  
  Under folder `./`.  
  It is developed in Go language, it provides web service, and html/js/image host.

* Web project  
  Under folder `./web`.  
  It is developed in Javascript on React.

  These 2 projects folders can be opened as individual project, by MS Code or other IDE.

# Usage
*Please*

# Development
``` 
git clone https://github.com/IBM/fablet.git  
cd fablet  
./build.sh
 ```


go env -w GOPROXY=https://goproxy.io,direct

# Run development
go run github.com/IBM/fablet/main  
go run ./main

# Build
go build -o fablet github.com/IBM/fablet/main

# Import Fabric SDK go
go get github.com/hyperledger/fabric-sdk-go

# Test
go test ./service

# Build
./build.sh