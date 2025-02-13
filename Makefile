.PHONY: tomo tomo-cross evm all test clean
.PHONY: tomo-linux tomo-linux-386 tomo-linux-amd64 tomo-linux-mips64 tomo-linux-mips64le
.PHONY: tomo-darwin tomo-darwin-386 tomo-darwin-amd64

GOBIN = $(shell pwd)/build/bin
GOFMT = gofmt
GO ?= 1.18.10
GO_PACKAGES = .
GO_FILES := $(shell find $(shell go list -f '{{.Dir}}' $(GO_PACKAGES)) -name \*.go)

GIT = git

NETWORK_ID="2569"

# Change these settings
BASE_DIR=$(shell pwd)
TOMO_DIR=${BASE_DIR}/build/bin
WORKSPACE=${BASE_DIR}/build/_workspace
GOBIN=./build/bin
KEYSTORE_DIR=${BASE_DIR}/keystore
PASSWORD_DIR=${BASE_DIR}
ETHERBASE=0xF7E6258432CDA2b44b013D6b67cED090ec4bf78f
PATH_TO_GENESIS_FILE=${BASE_DIR}/genesis.json
tomo=${TOMO_DIR}/tomo
bootnode=${TOMO_DIR}/bootnode
BOOTNODE_KEY_DIR=${BASE_DIR}/bootnode.key
BOOTNODE_KEY="enode://284b032ce57ed824f69eb44da10e0a55aec3a9fd6b9dd7f11ce594f1bee4314196f3f6e247eee52deb26ee5abd1d4190a218d0ab0264f90b757d85afb6f86818@[127.0.0.1]:30301"

tomo:
	go run build/ci.go install ./cmd/tomo
	@echo "Done building."
	@echo "Run \"$(GOBIN)/tomo\" to launch tomo."

gc:
	go run build/ci.go install ./cmd/gc
	@echo "Done building."
	@echo "Run \"$(GOBIN)/gc\" to launch gc."

bootnode:
	go run build/ci.go install ./cmd/bootnode
	@echo "Done building."
	@echo "Run \"$(GOBIN)/bootnode\" to launch a bootnode."

puppeth:
	go run build/ci.go install ./cmd/puppeth
	@echo "Done building."
	@echo "Run \"$(GOBIN)/puppeth\" to launch puppeth."

all:
	go run build/ci.go install

test: all
	go run build/ci.go test

clean:
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

# Cross Compilation Targets (xgo)

tomo-cross: tomo-windows-amd64 tomo-darwin-amd64 tomo-linux
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/tomo-*

tomo-linux: tomo-linux-386 tomo-linux-amd64 tomo-linux-mips64 tomo-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/tomo-linux-*

tomo-linux-386:
	go run build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/tomo
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/tomo-linux-* | grep 386

tomo-linux-amd64:
	go run build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/tomo
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/tomo-linux-* | grep amd64

tomo-linux-mips:
	go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/tomo
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/tomo-linux-* | grep mips

tomo-linux-mipsle:
	go run build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/tomo
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/tomo-linux-* | grep mipsle

tomo-linux-mips64:
	go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/tomo
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/tomo-linux-* | grep mips64

tomo-linux-mips64le:
	go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/tomo
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/tomo-linux-* | grep mips64le

tomo-darwin: tomo-darwin-386 tomo-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/tomo-darwin-*

tomo-darwin-386:
	go run build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/tomo
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/tomo-darwin-* | grep 386

tomo-darwin-amd64:
	go run build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/tomo
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/tomo-darwin-* | grep amd64

tomo-windows-amd64:
	go run build/ci.go xgo -- --go=$(GO) -buildmode=mode -x --targets=windows/amd64 -v ./cmd/tomo
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/tomo-windows-* | grep amd64
gofmt:
	$(GOFMT) -s -w $(GO_FILES)
	$(GIT) checkout vendor

init-genesis:
	@echo "Init genesis"
	rm -rf ${WORKSPACE}/1
	rm -rf ${WORKSPACE}/2
	rm -rf ${WORKSPACE}/3
	rm -rf ${WORKSPACE}/4
	rm -rf ${WORKSPACE}/5
	rm -rf ${WORKSPACE}/6
	rm -rf ${WORKSPACE}/7
	rm -rf ${WORKSPACE}/8
	rm -rf ${WORKSPACE}/9

#	${tomo} --datadir ${WORKSPACE}/1 init genesis.json
#	${tomo} --datadir ${WORKSPACE}/2 init genesis.json
#	${tomo} --datadir ${WORKSPACE}/3 init genesis.json
#	${tomo} --datadir ${WORKSPACE}/4 init genesis.json
#	${tomo} --datadir ${WORKSPACE}/5 init genesis.json
#	${tomo} --datadir ${WORKSPACE}/6 init genesis.json
#	${tomo} --datadir ${WORKSPACE}/7 init genesis.json
#	${tomo} --datadir ${WORKSPACE}/8 init genesis.json
#	${tomo} --datadir ${WORKSPACE}/9 init genesis.json

# tomo --datadir build/_workspace/1 init genesis.json
# tomo --datadir build/_workspace/2 init genesis.json
# tomo --datadir build/_workspace/3 init genesis.json
# tomo --datadir build/_workspace/4 init genesis.json
# tomo --datadir build/_workspace/5 init genesis.json
# tomo --datadir build/_workspace/6 init genesis.json
# tomo --datadir build/_workspace/7 init genesis.json
# tomo --datadir build/_workspace/8 init genesis.json
# tomo --datadir build/_workspace/9 init genesis.json

#tomo account new --password keystore/password_1 --keystore keystore/1
#tomo account new --password keystore/password_2 --keystore keystore/2
#tomo account new --password keystore/password_3 --keystore keystore/3
#tomo account new --password keystore/password_4 --keystore keystore/4
#tomo account new --password keystore/password_5 --keystore keystore/5
#tomo account new --password keystore/password_6 --keystore keystore/6
#tomo account new --password keystore/password_6 --keystore keystore/7
#tomo account new --password keystore/password_6 --keystore keystore/8
#tomo account new --password keystore/password_6 --keystore keystore/9


#address_1:8cb1883564db887a1d2e52311bed01664f921c21
#address_2:c7db2f9977488dca3a17ce3b83e417c03468dbf6
#address_3:255fca57249b5747e11be0b4fc9ad2d0ecfcda88
#address_4:615ad0d8e40d709af9b015be131e4e81637208c3
#address_5:1c4006de9cb80bae42c8921933285b819f3bcf0a
#address_6:596571d3f8b5903d908a7bd6bb9e96bcde691581
#address_7:118db4a6718e79a8ca2a05a5def0e6afeaaf24f4
#address_8:322a8d78955774256a3174db0abb9cfc75211759
#address_9:ce55bf99666fbba399260c53a05863f8adc7b121
#fd: D01d903011fd0DfD473220994C6Fca9755848C22
#address_10:a08eac777b0a3cf3b7876db1788bd17392ee5ba4


bootnode-local:
	@echo "Start bootnode"
	${bootnode} -nodekey ${BOOTNODE_KEY_DIR}

init_node10:
	rm -rf ${WORKSPACE}/7
	${tomo} --datadir ${WORKSPACE}/7 init genesis.json

init_node12:
	rm -rf ${WORKSPACE}/1
	rm -rf ${WORKSPACE}/2
	rm -rf ${WORKSPACE}/3
	rm -rf ${WORKSPACE}/4

	${tomo} --datadir ${WORKSPACE}/1 init genesis.json
	${tomo} --datadir ${WORKSPACE}/2 init genesis.json
	${tomo} --datadir ${WORKSPACE}/3 init genesis.json
	${tomo} --datadir ${WORKSPACE}/4 init genesis.json

delete_node:
	rm -rf ${WORKSPACE}/1

	${tomo} --datadir ${WORKSPACE}/1 init genesis.json



node10: 
	${tomo} --syncmode "full" \
	--datadir "${WORKSPACE}/10" --networkid ${NETWORK_ID} --port 10310 \
	--saigon "[596571d3f8b5903d908a7bd6bb9e96bcde691581, 118db4a6718e79a8ca2a05a5def0e6afeaaf24f4,322a8d78955774256a3174db0abb9cfc75211759,ce55bf99666fbba399260c53a05863f8adc7b121]" \
	--keystore "${BASE_DIR}/keystore/10" --password "${BASE_DIR}/keystore/password_10" \
	--rpc --rpccorsdomain "*" --rpcaddr 0.0.0.0 --rpcport 8550 --rpcvhosts "*" \
	--rpcapi "admin,db,eth,net,web3,personal,debug" \
	--gcmode "archive" \
	--ws --wsaddr 0.0.0.0 --wsport 2010 --wsorigins "*" --unlock "a08eac777b0a3cf3b7876db1788bd17392ee5ba4" \
	--identity "NODE10" \
	--mine --gasprice 2500 \ --bootnodes ${BOOTNODE_KEY} \
	console

node1:
	${tomo} --syncmode "full" \
	--datadir "${WORKSPACE}/1" --networkid ${NETWORK_ID} --port 10301 \
	--keystore "${BASE_DIR}/keystore/1" --password "${BASE_DIR}/keystore/password_1" \
	--rpc --rpccorsdomain "*" --rpcaddr 0.0.0.0 --rpcport 8541 --rpcvhosts "*" \
	--rpcapi "admin,db,eth,net,web3,personal,debug" \
	--gcmode "archive" \
	--ws --wsaddr 0.0.0.0 --wsport 2001 --wsorigins "*" --unlock "8cb1883564db887a1d2e52311bed01664f921c21" \
	--identity "NODE1" \
	--mine --gasprice 2500 \ --bootnodes ${BOOTNODE_KEY} \
	console

#		tomo --syncmode "full" \
#    	--datadir "/Users/lenguyenngocvien/Desktop/workspace/victionchain/build/_workspace/1" --networkid "2569" --port 10301 \
#    	--saigon "[levien, 7s62]" \
#    	--keystore "/Users/lenguyenngocvien/Desktop/workspace/victionchain/keystore/1" --password "/Users/lenguyenngocvien/Desktop/workspace/victionchain/keystore/password_1" \
#    	--rpc --rpccorsdomain "*" --rpcaddr 0.0.0.0 --rpcport 8541 --rpcvhosts "*" \
#    	--rpcapi "admin,db,eth,net,web3,personal,debug" \
#    	--gcmode "archive" \
#    	--ws --wsaddr 0.0.0.0 --wsport 2001 --wsorigins "*" --unlock "8cb1883564db887a1d2e52311bed01664f921c21" \
#    	--identity "NODE1" \
#    	--mine --gasprice 2500 \ --bootnodes "enode://284b032ce57ed824f69eb44da10e0a55aec3a9fd6b9dd7f11ce594f1bee4314196f3f6e247eee52deb26ee5abd1d4190a218d0ab0264f90b757d85afb6f86818@[127.0.0.1]:30301" \
#    	console

node2:
	${tomo} --syncmode "full" \
	--datadir "${WORKSPACE}/2" --networkid ${NETWORK_ID} --port 10302 \
	--keystore "${BASE_DIR}/keystore/2" --password "${BASE_DIR}/keystore/password_2" \
	--rpc --rpccorsdomain "*" --rpcaddr 0.0.0.0 --rpcport 8542 --rpcvhosts "*" \
	--rpcapi "admin,db,eth,net,web3,personal,debug" \
	--gcmode "archive" \
	--ws --wsaddr 0.0.0.0 --wsport 2002 --wsorigins "*" --unlock "c7db2f9977488dca3a17ce3b83e417c03468dbf6" \
	--identity "NODE2" \
	--mine --gasprice 2500 \ --bootnodes ${BOOTNODE_KEY} \
	console


node3:
	${tomo} --syncmode "full" \
	--datadir ${WORKSPACE}/3 --networkid ${NETWORK_ID} --port 10303 \
	--keystore "${BASE_DIR}/keystore/3" --password "${BASE_DIR}/keystore/password_3" \
	--rpc --rpccorsdomain "*" --rpcaddr 0.0.0.0 --rpcport 8543 --rpcvhosts "*" \
	--rpcapi "admin,db,eth,net,web3,personal,debug" \
	--gcmode "archive" \
	--ws --wsaddr 0.0.0.0 --wsport 2003 --wsorigins "*" --unlock "255fca57249b5747e11be0b4fc9ad2d0ecfcda88" \
	--identity "NODE3" \
	--mine --gasprice 2500 \ --bootnodes ${BOOTNODE_KEY} \
	console

node4:
	${tomo} --syncmode "full" \
	--datadir ${WORKSPACE}/4 --networkid ${NETWORK_ID} --port 10304 \
	--keystore "${BASE_DIR}/keystore/4" --password "${BASE_DIR}/keystore/password_4" \
	--rpc --rpccorsdomain "*" --rpcaddr 0.0.0.0 --rpcport 8544 --rpcvhosts "*" \
	--rpcapi "admin,db,eth,net,web3,personal,debug" \
	--gcmode "archive" \
	--ws --wsaddr 0.0.0.0 --wsport 2004 --wsorigins "*" --unlock "615ad0d8e40d709af9b015be131e4e81637208c3" \
	--identity "NODE4" \
	--mine --gasprice 2500 \ --bootnodes ${BOOTNODE_KEY} \
	console

node5:
	${tomo} --syncmode "full" \
	--datadir ${WORKSPACE}/5 --networkid ${NETWORK_ID} --port 10305 \
	--keystore "${BASE_DIR}/keystore/5" --password "${BASE_DIR}/keystore/password_5" \
	--rpc --rpccorsdomain "*" --rpcaddr 0.0.0.0 --rpcport 8545 --rpcvhosts "*" \
	--rpcapi "admin,db,eth,net,web3,personal,debug" \
	--gcmode "archive" \
	--ws --wsaddr 0.0.0.0 --wsport 2005 --wsorigins "*" --unlock "1c4006de9cb80bae42c8921933285b819f3bcf0a" \
	--identity "NODE5" \
	--mine --gasprice 2500 \ --bootnodes ${BOOTNODE_KEY} \
	console

	-#-saigon "[596571d3f8b5903d908a7bd6bb9e96bcde691581, 118db4a6718e79a8ca2a05a5def0e6afeaaf24f4,322a8d78955774256a3174db0abb9cfc75211759,ce55bf99666fbba399260c53a05863f8adc7b121]" \

node6:
	${tomo} --syncmode "full" \
	--datadir ${WORKSPACE}/6 --networkid ${NETWORK_ID} --port 10306 \
	--keystore "${BASE_DIR}/keystore/6" --password "${BASE_DIR}/keystore/password_6" \
	--rpc --rpccorsdomain "*" --rpcaddr 0.0.0.0 --rpcport 8546 --rpcvhosts "*" \
	--rpcapi "admin,db,eth,net,web3,personal,debug" \
	--gcmode "archive" \
	--ws --wsaddr 0.0.0.0 --wsport 2006 --wsorigins "*" --unlock "596571d3f8b5903d908a7bd6bb9e96bcde691581" \
	--identity "NODE6" \
	--mine --gasprice 2500 \ --bootnodes ${BOOTNODE_KEY} \
	console

node7:
	${tomo} --syncmode "full" \
	--datadir ${WORKSPACE}/7 --networkid ${NETWORK_ID} --port 10307 \
	--keystore "${BASE_DIR}/keystore/7" --password "${BASE_DIR}/keystore/password_7" \
	--rpc --rpccorsdomain "*" --rpcaddr 0.0.0.0 --rpcport 8547 --rpcvhosts "*" \
	--rpcapi "admin,db,eth,net,web3,personal,debug" \
	--gcmode "archive" \
	--ws --wsaddr 0.0.0.0 --wsport 2007 --wsorigins "*" --unlock "118db4a6718e79a8ca2a05a5def0e6afeaaf24f4" \
	--identity "NODE7" \
	--mine --gasprice 2500 \ --bootnodes ${BOOTNODE_KEY} \
	console

node8:
	${tomo} --syncmode "full" \
	--datadir ${WORKSPACE}/8 --networkid ${NETWORK_ID} --port 10308 \
	--keystore "${BASE_DIR}/keystore/8" --password "${BASE_DIR}/keystore/password_8" \
	--rpc --rpccorsdomain "*" --rpcaddr 0.0.0.0 --rpcport 8548 --rpcvhosts "*" \
	--rpcapi "admin,db,eth,net,web3,personal,debug" \
	--gcmode "archive" \
	--ws --wsaddr 0.0.0.0 --wsport 2008 --wsorigins "*" --unlock "322a8d78955774256a3174db0abb9cfc75211759" \
	--identity "NODE8" \
	--mine --gasprice 2500 \ --bootnodes ${BOOTNODE_KEY} \
	console

node9:
	${tomo} --syncmode "full" \
	--datadir ${WORKSPACE}/9 --networkid ${NETWORK_ID} --port 10309 \
	--keystore "${BASE_DIR}/keystore/9" --password "${BASE_DIR}/keystore/password_9" \
	--rpc --rpccorsdomain "*" --rpcaddr 0.0.0.0 --rpcport 8549 --rpcvhosts "*" \
	--rpcapi "admin,db,eth,net,web3,personal,debug" \
	--gcmode "archive" \
	--ws --wsaddr 0.0.0.0 --wsport 2009 --wsorigins "*" --unlock "ce55bf99666fbba399260c53a05863f8adc7b121" \
	--identity "NODE9" \
	--mine --gasprice 2500 \ --bootnodes ${BOOTNODE_KEY} \
	console


add-peers-1:
	${tomo} attach ${BASE_DIR}/nodes/1/tomo.ipc
	admin.addPeer("enode://6a25904b4934568f15a335b277101460c22cddfb3d8881ab23e2b04f5664a726501cffbb7c840dc230f9f9b7fd5c79a7ff2dbf5779b95f27675fa038fa50edb9@[::]:10301")
	admin.addPeer("enode://b7db3ca3d8c17271abaec8318ffbdb974bc53365714b6d5af0faa26c291a73474ebdf98461daea9a2501630df01701b5af153e6837b9351ecba36fbb9d39f716@[::]:10302")
	admin.addPeer("enode://af9421f0ce6b525d877d833294bb92747b9e79a6822c0ddb865279e8111b43c74afb0253f65c9d1086fa50e5d5081ed1399c3b175e5a5ceff457adea9c103991@[::]:10303")
	admin.addPeer("enode://54d58706acdb231ca1e503e9a44699141f876380533d88386c275f3ba6c817911004510fc103290c2f56a58dff1c7294be7053ec8ac1beca29f37881cce4c417@[::]:10304")
	exit
	admin.addPeer("enode://4a2d104d2cf153de10a56858fd53b4342021708fc3b9ce8290b7829ab94da97674cfb41db0f2eb398678049d9a4a87511a55a338e98aeee887ff86b32797112e@[::]:10305")
	admin.addPeer("enode://2d4042a4e264fe67ef4493f102be30adeb39e66ac4bc22f782392c20a3a106f33100d043847db836cfb0edc97dc9621a6ff9a8b770862ad5a78f9fa80ae83e45@[::]:10306")
	admin.addPeer("enode://46103df2a152f09cdbaf902a9cea1bd88a4c4d92486df7ac88b7d01863d8b1f82cfe7c1d227aa1ad0095cd9462793e36429d732d31a0aac4697dec5c68c68fd9@[::]:10307")
	admin.addPeer("enode://f23bedfc9b7598d79504d42bc844daf1e122710901a4f3084faaf9f0d4328028b0ab95449ec7a4617a43627b8aa2ddd84a27bf1a14d998237b5a9fd7ce86dfb5@[::]:10308")
	admin.addPeer("enode://44928db28d54879cf7074f16bead4cee64ba6a7bc41bda3743fe95b1c1b9b020660d65bcf2b3054515e36e1296f593cd2f0fb5afd7c1806f8e43881c28738aec@[::]:10309")
	exit