package api

import (
	"bytes"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"

	"github.com/cloudflare/cfssl/log"
	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
	"github.com/hyperledger/fabric-protos-go/msp"
	protosmsp "github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/core/common/ccprovider"
	"github.com/hyperledger/fabric/common/util"
)

// Endorser who endorses the transaction action
type Endorser struct {
	CommonName string `json:"commonName"`
	Subject    string `json:"subject"`
	IsCA       bool   `json:"isCA"`
	MSPID      string `json:"MSPID"`
	Sign       string `json:"sign"`
	Issuer     string `json:"issuer"`
}

// Action action of a transaction
type Action struct {
	ChaincodeName    string            `json:"chaincodeName"`
	ChaincodeVersion string            `json:"chaincodeVersion"`
	Arguments        []string          `json:"arguments"`
	Endorsers        []*Endorser       `json:"endorsers"`
	ProposalResponse *ProposalResponse `json:"proposalResponse"`
}

// Transaction transaction of a block
type Transaction struct {
	Actions []*Action `json:"actions"`
}

// Block block of a ledger
type Block struct {
	Number       uint64         `json:"number"`
	DataHash     string         `json:"dataHash"`
	PreviousHash string         `json:"previousHash"`
	BlockHash    string         `json:"blockHash"`
	Transactions []*Transaction `json:"transactions"`
	Time         int64          `json:"time"`
}

// Ledger ledger
type Ledger struct {
	Height           uint64 `json:"height"`
	CurrentBlockHash string `json:"currentBlockHash"`
	Endorser         string `json:"endorser"`
	Status           int32  `json:"status"`
}

// ChaincodeData a simplied chaincode data from lscc
type ChaincodeData struct {
	Name       string   `json:"name"`
	Version    string   `json:"version"`
	Principals []string `json:"principals"`
	Rule       string   `json:"rule"`
}

// KVRead key value pair.
type KVRead struct {
	Key         string `json:"key"`
	VerBlockNum uint64 `json:"verBlockNum"`
	VerTxNum    uint64 `json:"verTxNum"`
}

// KVWrite key value pair.
type KVWrite struct {
	Key      string      `json:"key"`
	IsDelete bool        `json:"isDelete"`
	Value    interface{} `json:"value"`
}

// HashReadWriteCol for hashed rw set.
type HashReadWriteCol struct {
	CollectionName string     `json:"collectionName"`
	KVReadSet      []*KVRead  `json:"kvReadSet"`
	KVWriteSet     []*KVWrite `json:"kvWriteSet"`
}

// NSReadWriteSet namespace rw set.
type NSReadWriteSet struct {
	NameSpace         string              `json:"nameSpace"`
	KVReadSet         []*KVRead           `json:"kvReadSet"`
	KVWriteSet        []*KVWrite          `json:"kvWriteSet"`
	HashReadWriteCols []*HashReadWriteCol `json:"hashReadWriteCols"`
}

// TXReadWriteSet transaction rw set.
type TXReadWriteSet struct {
	DataModel       string            `json:"dataModel"`
	NSReadWriteSets []*NSReadWriteSet `json:"nsReadWriteSets"`
}

// ProposalResponse proposal response.
type ProposalResponse struct {
	Chaincode      *Chaincode      `json:"chaincode"`
	Response       interface{}     `json:"response"`
	TXReadWriteSet *TXReadWriteSet `json:"txReadWriteSet"`
}

// QueryLedger to query a ledger from an endpoint.
// TODO the targets can be empty
func QueryLedger(conn *NetworkConnection, channelID string, targets []string) (*Ledger, error) {
	channelContext := conn.SDK.ChannelContext(channelID, fabsdk.WithIdentity(conn.SignID))
	ldgClient, err := ledger.New(channelContext)
	if err != nil {
		return nil, err
	}
	info, err := ldgClient.QueryInfo(ledger.WithTargetEndpoints(targets...))
	if err != nil {
		return nil, err
	}
	ldg := &Ledger{}
	ldg.Height = info.BCI.GetHeight()
	ldg.CurrentBlockHash = hex.EncodeToString(info.BCI.GetCurrentBlockHash())
	ldg.Endorser = info.Endorser
	ldg.Status = info.Status

	return ldg, nil
}

// CalBlockHash to calculate block hash
func CalBlockHash(block *common.Block) ([]byte, error) {

	asn1Bytes, err := asn1.Marshal(struct {
		Number       int64
		PreviousHash []byte
		DataHash     []byte
	}{
		Number:       int64(block.GetHeader().GetNumber()),
		PreviousHash: []byte(block.GetHeader().GetPreviousHash()),
		DataHash:     block.GetHeader().GetDataHash(),
	})
	if err != nil {
		return nil, err
	}

	return util.ComputeSHA256(asn1Bytes), nil
}

func getStrings(bytes [][]byte) []string {
	strings := make([]string, len(bytes))
	for idx, bs := range bytes {
		strings[idx] = string(bs)
	}
	return strings
}

func fixArgs(chaincodeName string, bytes [][]byte) []string {
	args := getStrings(bytes)
	// TODO lscc code
	if chaincodeName != LSCC {
		return args
	}

	// Fix args 2, 3
	ccds := &peer.ChaincodeDeploymentSpec{}
	if err := proto.Unmarshal(bytes[2], ccds); err == nil {
		args[2] = ccds.GetChaincodeSpec().String()

	}
	spe := &common.SignaturePolicyEnvelope{}
	if err := proto.Unmarshal(bytes[3], spe); err != nil {
		args[3] = spe.String()
	}
	return args
}

func getCert(bytes []byte) *x509.Certificate {
	block, _ := pem.Decode(bytes)

	if block != nil {
		pub, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			// TOOD no panic
			// panic(err)
			return nil
		}
		return pub
	}
	return nil
}

func translateProposalResponse(capl *peer.ChaincodeActionPayload) *ProposalResponse {
	_txrwset := &TXReadWriteSet{NSReadWriteSets: []*NSReadWriteSet{}}
	_pr := &ProposalResponse{TXReadWriteSet: _txrwset}

	prpl := &peer.ProposalResponsePayload{}
	ccac := &peer.ChaincodeAction{}
	txrwst := &rwset.TxReadWriteSet{}
	kvst := &kvrwset.KVRWSet{}

	if proto.Unmarshal(capl.GetAction().GetProposalResponsePayload(), prpl) != nil ||
		proto.Unmarshal(prpl.GetExtension(), ccac) != nil ||
		proto.Unmarshal(ccac.GetResults(), txrwst) != nil {
		return _pr
	}

	_pr.Chaincode = &Chaincode{
		Name:    ccac.GetChaincodeId().GetName(),
		Version: ccac.GetChaincodeId().GetVersion(),
		Path:    ccac.GetChaincodeId().GetPath()}

	if _pr.Chaincode.Name == LSCC {
		_pr.Response = translateChaincodeData(ccac.GetResponse().GetPayload())
	} else {
		_pr.Response = string(ccac.GetResponse().GetPayload())
	}

	_txrwset.DataModel = txrwst.GetDataModel().String()

	for _, nsrwst := range txrwst.GetNsRwset() {
		_nsrwst := &NSReadWriteSet{
			NameSpace:         nsrwst.GetNamespace(),
			KVReadSet:         []*KVRead{},
			KVWriteSet:        []*KVWrite{},
			HashReadWriteCols: []*HashReadWriteCol{},
		}
		_txrwset.NSReadWriteSets = append(_txrwset.NSReadWriteSets, _nsrwst)

		for _, hhrw := range nsrwst.GetCollectionHashedRwset() {
			_hhrw := &HashReadWriteCol{
				CollectionName: hhrw.GetCollectionName(),
				KVReadSet:      []*KVRead{},
				KVWriteSet:     []*KVWrite{},
			}
			_nsrwst.HashReadWriteCols = append(_nsrwst.HashReadWriteCols, _hhrw)

			hkvst := &kvrwset.HashedRWSet{}
			err := proto.Unmarshal(hhrw.GetHashedRwset(), hkvst)
			if err == nil {
				for _, read := range hkvst.GetHashedReads() {
					_hhrw.KVReadSet = append(_hhrw.KVReadSet, &KVRead{
						Key:         string(read.GetKeyHash()),
						VerBlockNum: read.GetVersion().GetBlockNum(),
						VerTxNum:    read.GetVersion().GetTxNum(),
					})
				}
				for _, write := range hkvst.GetHashedWrites() {
					_hhrw.KVWriteSet = append(_hhrw.KVWriteSet, &KVWrite{
						Key:      string(write.GetKeyHash()),
						IsDelete: write.GetIsDelete(),
						Value:    string(write.GetValueHash()),
					})
				}
			}
		}

		err := proto.Unmarshal(nsrwst.GetRwset(), kvst)
		if err == nil {
			// for _, mtwt := range kvst.GetMetadataWrites() {
			// 	log.Info("        meta write key: ", mtwt.GetKey())
			// 	for _, entr := range mtwt.GetEntries() {
			// 		log.Info("        meta entry: ", entr.GetName(), "  ", string(entr.GetValue()))
			// 	}
			// }
			for _, read := range kvst.GetReads() {
				_nsrwst.KVReadSet = append(_nsrwst.KVReadSet, &KVRead{
					Key:         read.GetKey(),
					VerBlockNum: read.GetVersion().GetBlockNum(),
					VerTxNum:    read.GetVersion().GetTxNum(),
				})
			}
			for _, write := range kvst.GetWrites() {
				_kvw := &KVWrite{
					Key:      write.GetKey(),
					IsDelete: write.GetIsDelete(),
				}
				if ccac.GetChaincodeId().GetName() == LSCC {
					_kvw.Value = translateChaincodeData(write.GetValue())
				} else {
					_kvw.Value = string(write.GetValue())
				}
				_nsrwst.KVWriteSet = append(_nsrwst.KVWriteSet, _kvw)
			}
		}
	}

	return _pr
}

func translateChaincodeData(payload []byte) *ChaincodeData {
	_ccd := &ChaincodeData{}

	ccd := &ccprovider.ChaincodeData{}
	if err := proto.Unmarshal(payload, ccd); err == nil {
		_ccd.Name = ccd.Name
		_ccd.Version = ccd.Version

		spe := &common.SignaturePolicyEnvelope{}
		if err := proto.Unmarshal(ccd.Policy, spe); err == nil {
			for _, id := range spe.GetIdentities() {
				buffer := &bytes.Buffer{}
				buffer.WriteString(id.GetPrincipalClassification().String())
				buffer.WriteString("(")
				// See Fabric cauthdsl_builder.go signedByAnyOfGivenRole
				role := &msp.MSPRole{}
				if err := proto.Unmarshal(id.GetPrincipal(), role); err == nil {
					buffer.WriteString(role.GetMspIdentifier() + "." + role.GetRole().String())
				}
				buffer.WriteString("); ")
				_ccd.Principals = append(_ccd.Principals, buffer.String())
			}
			_ccd.Rule = spe.GetRule().String()
		}

	}
	return _ccd
}

// Translate common.Block to Fablet Block
func translateBlock(block *common.Block) *Block {
	blockNumber := block.GetHeader().GetNumber()
	blockHash, err := CalBlockHash(block)
	if err != nil {
		log.Errorf("Error occurred when calculate block hash %d of %s: %s", blockNumber, err.Error())
		blockHash = []byte{}
	}

	transactions := []*Transaction{}

	env := &common.Envelope{}
	pl := &common.Payload{} //common.Header, byte of peer.Transaction
	ch := &common.ChannelHeader{}
	tx := &peer.Transaction{}
	capl := &peer.ChaincodeActionPayload{}
	cppl := &peer.ChaincodeProposalPayload{}
	edr := &protosmsp.SerializedIdentity{} //Not SigningIdentityInfo{}
	input := &peer.ChaincodeInvocationSpec{}

	for _, d := range block.GetData().GetData() {
		// No error handling
		// TODO
		proto.Unmarshal(d, env)
		proto.Unmarshal(env.GetPayload(), pl)
		proto.Unmarshal(pl.GetHeader().GetChannelHeader(), ch)
		proto.Unmarshal(pl.GetData(), tx)

		actions := []*Action{}
		for _, txa := range tx.GetActions() {
			// No error handling
			proto.Unmarshal(txa.GetPayload(), capl)
			proto.Unmarshal(capl.GetChaincodeProposalPayload(), cppl)
			proto.Unmarshal(cppl.GetInput(), input)

			endorsers := []*Endorser{}
			for _, edrsm := range capl.GetAction().GetEndorsements() {
				// No error handling
				proto.Unmarshal(edrsm.GetEndorser(), edr)
				endorser := &Endorser{
					MSPID: edr.GetMspid(),
					Sign:  hex.EncodeToString(edrsm.GetSignature()),
				}
				if edr.GetIdBytes() != nil {
					cert := getCert(edr.GetIdBytes())
					endorser.CommonName = cert.Subject.CommonName
					endorser.IsCA = cert.IsCA
					endorser.Subject = cert.Subject.String()
					endorser.Issuer = cert.Issuer.String()
				}
				endorsers = append(endorsers, endorser)
			}

			chaincodeName := input.GetChaincodeSpec().GetChaincodeId().GetName()
			action := &Action{
				ChaincodeName:    chaincodeName,
				ChaincodeVersion: input.GetChaincodeSpec().GetChaincodeId().GetVersion(),
				Arguments:        fixArgs(chaincodeName, input.GetChaincodeSpec().GetInput().GetArgs()),
				Endorsers:        endorsers,
				ProposalResponse: translateProposalResponse(capl),
			}
			actions = append(actions, action)
		}

		transaction := &Transaction{Actions: actions}
		transactions = append(transactions, transaction)
	}

	blk := &Block{
		Number:       blockNumber,
		DataHash:     hex.EncodeToString(block.GetHeader().GetDataHash()),
		PreviousHash: hex.EncodeToString(block.GetHeader().GetPreviousHash()),
		BlockHash:    hex.EncodeToString(blockHash),
		Transactions: transactions,
		Time:         ch.GetTimestamp().GetSeconds()*1000 + int64(ch.GetTimestamp().GetNanos())/1000,
	}

	return blk
}

// QueryBlock to query blocks of the given numbers.
// TODO the targets can be empty
func QueryBlock(conn *NetworkConnection, channelID string, targets []string, begin uint64, len uint64) ([]*Block, error) {
	blocks := []*Block{}

	// TODO to use a common getChannelContext
	channelContext := conn.SDK.ChannelContext(channelID, fabsdk.WithIdentity(conn.SignID))
	ldgClient, err := ledger.New(channelContext)
	if err != nil {
		return nil, err
	}

	if len > 512 {
		len = 512
	}

	for n := uint64(0); n < len; n++ {
		_begin := begin + n
		block, err := ldgClient.QueryBlock(_begin, ledger.WithTargetEndpoints(targets...))
		if err != nil {
			log.Errorf("Failed to query block %d of %s: %s", _begin, channelID, err.Error())
			continue
		}
		blocks = append(blocks, translateBlock(block))
	}

	return blocks, nil
}

// QueryBlockByHash to query blocks of the given hash.
func QueryBlockByHash(conn *NetworkConnection, channelID string, targets []string, blockHash string) (*Block, error) {
	// TODO to use a common getChannelContext
	channelContext := conn.SDK.ChannelContext(channelID, fabsdk.WithIdentity(conn.SignID))
	ldgClient, err := ledger.New(channelContext)
	if err != nil {
		return nil, err
	}
	hashBytes, err := hex.DecodeString(blockHash)
	if err != nil {
		return nil, err
	}
	block, err := ldgClient.QueryBlockByHash(hashBytes, ledger.WithTargetEndpoints(targets...))
	if err != nil {
		return nil, err
	}
	return translateBlock(block), nil
}

// QueryBlockByTxID to query blocks of the given tx id.
func QueryBlockByTxID(conn *NetworkConnection, channelID string, targets []string, txID string) (*Block, error) {
	// TODO to use a common getChannelContext
	channelContext := conn.SDK.ChannelContext(channelID, fabsdk.WithIdentity(conn.SignID))
	ldgClient, err := ledger.New(channelContext)
	if err != nil {
		return nil, err
	}

	block, err := ldgClient.QueryBlockByTxID(fab.TransactionID(txID), ledger.WithTargetEndpoints(targets...))
	if err != nil {
		return nil, err
	}
	return translateBlock(block), nil
}
