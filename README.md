# Fablet
A management tools set for Hyperledger Fabric blockchain platform.

cd ~
git clone https://github.com/IBM/fablet.git
cd fablet


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