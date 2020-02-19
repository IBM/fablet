package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/IBM/fablet/api"
	"github.com/IBM/fablet/util"
	"github.com/pkg/errors"
)

// ChaincodeInstallReq use fields instead of anonlymous fields, to have a more clear structure.
type ChaincodeInstallReq struct {
	BaseRequest
	Chaincode api.Chaincode `json:"chaincode"`
	Targets   []string      `json:"targets"`
}

// ChaincodeInstantiateReq for instantiate a chaincode.
type ChaincodeInstantiateReq struct {
	BaseRequest
	Chaincode api.Chaincode `json:"chaincode"`
	Target    string        `json:"target"`
	Orderer   string        `json:"orderer"`
	IsUpgrade bool          `json:"isUpgrade"`
}

// ChaincodeExecuteReq to execute a chaincode
type ChaincodeExecuteReq struct {
	BaseRequest
	Chaincode    api.Chaincode `json:"chaincode"`
	ActionType   string        `json:"actionType"`
	FunctionName string        `json:"functionName"`
	Arguments    []string      `json:"arguments"`
	Targets      []string      `json:"targets"`
}

// HandleChaincodeInstall to install a chaincode
func HandleChaincodeInstall(res http.ResponseWriter, req *http.Request) {
	logger.Info("Service HandleChaincodeInstall")

	reqBody := &ChaincodeInstallReq{}
	conn, err := GetRequest(req, reqBody, true)
	if err != nil {
		ErrorOutput(res, req, RES_CODE_ERR_INTERNAL, errors.WithMessage(err, "Error occurred when parsing from request."))
		return
	}
	// TODO to hold on connection in session
	// defer conn.Close()

	// TODO remove tmp folder defer
	tmpFolder := GetTmpFolder()
	defer func() {
		err := os.RemoveAll(tmpFolder)
		if err != nil {
			logger.Errorf("Error in removing temp folder: %s", err.Error())
		}
	}()

	logger.Debugf("Set temp folder %s", tmpFolder)

	chaincode := &reqBody.Chaincode

	// TODO to move this tar related to fablet api.
	err = util.UnTar(chaincode.Package, tmpFolder)
	if err != nil {
		ErrorOutput(res, req, RES_CODE_ERR_INTERNAL, errors.WithMessage(err, "Error occurred when uncompress the chaincode package."))
		return
	}

	if chaincode.Type == api.ChaincodeType_GOLANG {
		chaincode.BasePath = tmpFolder
	} else {
		chaincode.Path = tmpFolder
	}

	logger.Debugf(fmt.Sprintf("Begin to install %s:%s", chaincode.Name, chaincode.Version))
	// installRes length will always be identical to the peers length.
	installRes, err := api.InstallChaincode(conn, chaincode, reqBody.Targets)

	if err != nil && installRes == nil {
		ErrorOutput(res, req, RES_CODE_ERR_INTERNAL, errors.WithMessage(err, "Error occurred when installing the chaincode."))
		return
	}

	// It might be: err is not nil but the reuslt is not nil also, since the installation was executed but got failed.
	ResultOutput(res, req, map[string]interface{}{
		"installRes": installRes,
	})

}

// HandleChaincodeInstantiate to instantiate a chaincode.
func HandleChaincodeInstantiate(res http.ResponseWriter, req *http.Request) {
	logger.Info("Service HandleChaincodeInstantiate")
	reqBody := &ChaincodeInstantiateReq{}
	conn, err := GetRequest(req, reqBody, true)
	if err != nil {
		ErrorOutput(res, req, RES_CODE_ERR_INTERNAL, errors.WithMessage(err, "Error occurred when parsing from request."))
		return
	}
	// TODO to hold on connection in session
	// defer conn.Close()

	logger.Info(fmt.Sprintf("Begin to instantiate %s:%s", reqBody.Chaincode.Name, reqBody.Chaincode.Version))

	transID, err := api.InstantiateChaincode(conn, &reqBody.Chaincode, reqBody.Target, reqBody.Orderer)

	if err != nil {
		ErrorOutput(res, req, RES_CODE_ERR_INTERNAL, errors.WithMessage(err, "Error occurred when instantiating the chaincode."))
		return
	}

	// It might be: err is not nil but the reuslt is not nil also, since the installation was executed but got failed.
	ResultOutput(res, req, map[string]interface{}{
		"instantiateRes": string(transID),
	})
}

// HandleChaincodeUpgrade to upgrade chaincode.
func HandleChaincodeUpgrade(res http.ResponseWriter, req *http.Request) {
	// TODO to consolidate all same operations: reqBody, conn...
	logger.Info("Service HandleChaincodeUpgrade")
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		ErrorOutput(res, req, RES_CODE_ERR_INTERNAL, errors.WithMessage(err, "Error occurred when read from request."))
		return
	}

	reqBody := &ChaincodeInstantiateReq{}
	err = json.Unmarshal(body, reqBody)
	if err != nil {
		ErrorOutput(res, req, RES_CODE_ERR_INTERNAL, errors.WithMessage(err, "Error occurred when unmarshal ReqBody from request."))
		return
	}

	reqConn := reqBody.Connection
	conn, err := GetConnection(
		// TODO to support multiple config file type
		&api.ConnectionProfile{Config: []byte(reqConn.ConnProfile), ConfigType: "yaml"},
		&api.Participant{Label: reqConn.Label, OrgName: "", MSPID: reqConn.MSPID,
			Cert: []byte(reqConn.CertContent), PrivateKey: []byte(reqConn.PrvKeyContent), SignID: nil},
		true)
	// TODO useDiscovery

	if err != nil {
		ErrorOutput(res, req, RES_CODE_ERR_INTERNAL, errors.WithMessage(err, "Error occurred when create Fablet connection."))
		return
	}

	defer conn.Close()

	logger.Info(fmt.Sprintf("Begin to upgrade %s:%s", reqBody.Chaincode.Name, reqBody.Chaincode.Version))

	transID, err := api.UpgradeChaincode(conn, &reqBody.Chaincode, reqBody.Target, reqBody.Orderer)

	if err != nil {
		ErrorOutput(res, req, RES_CODE_ERR_INTERNAL, errors.WithMessage(err, "Error occurred when upgrading the chaincode."))
		return
	}

	// It might be: err is not nil but the reuslt is not nil also, since the installation was executed but got failed.
	ResultOutput(res, req, map[string]interface{}{
		"upgradeRes": string(transID),
	})

}

// HandleChaincodeExecute to execute a chaincode
func HandleChaincodeExecute(res http.ResponseWriter, req *http.Request) {
	logger.Info("Service HandleChaincodeExecute")

	// body, err := ioutil.ReadAll(req.Body)
	// if err != nil {
	// 	ErrorOutput(res, req, RES_CODE_ERR_INTERNAL, errors.WithMessage(err, "Error occurred when read from request."))
	// 	return
	// }

	// reqBody := &ChaincodeExecuteReq{}
	// err = json.Unmarshal(body, reqBody)
	// if err != nil {
	// 	ErrorOutput(res, req, RES_CODE_ERR_INTERNAL, errors.WithMessage(err, "Error occurred when unmarshal ReqBody from request."))
	// 	return
	// }

	// reqConn := reqBody.Connection
	// conn, err := api.NewConnection(
	// 	// TODO to support multiple config file type
	// 	&api.ConnectionProfile{Config: []byte(reqConn.ConnProfile), ConfigType: "yaml"},
	// 	&api.Participant{Label: reqConn.Label, OrgName: "", MSPID: reqConn.MSPID,
	// 		Cert: []byte(reqConn.CertContent), PrivateKey: []byte(reqConn.PrvKeyContent), SignID: nil},
	// 	true)
	// TODO useDiscovery

	reqBody := &ChaincodeExecuteReq{}
	conn, err := GetRequest(req, reqBody, true)
	if err != nil {
		ErrorOutput(res, req, RES_CODE_ERR_INTERNAL, errors.WithMessage(err, "Error occurred when parsing from request."))
		return
	}
	// TODO to hold on connection in session
	// defer conn.Close()

	logger.Info(fmt.Sprintf("Begin to execute chaincode %s:%s", reqBody.Chaincode.Name, reqBody.Chaincode.Version))

	ccOperType := api.ChaincodeOperTypeExecute
	if reqBody.ActionType == "query" {
		ccOperType = api.ChaincodeOperTypeQuery
	}

	// TODO target is not supported now
	cceRes, err := api.ExecuteChaincode(conn, reqBody.Chaincode.ChannelID, reqBody.Chaincode.Name, ccOperType, reqBody.Targets, reqBody.FunctionName, reqBody.Arguments)

	if err != nil {
		ErrorOutput(res, req, RES_CODE_ERR_INTERNAL,
			errors.WithMessagef(err, "Error occurred when execute the chaincode %s in channel %s, with arguments %v, with targets %v.",
				reqBody.Chaincode.Name, reqBody.Chaincode.ChannelID, reqBody.Arguments, reqBody.Targets))
		return
	}

	peerRes := []map[string]interface{}{}
	for _, pr := range cceRes.Responses {
		peerRes = append(peerRes, map[string]interface{}{
			"endorser": pr.Endorser,
			"version":  pr.GetVersion(),
			"payload":  string(pr.GetResponse().GetPayload()),
			"status":   pr.GetResponse().GetStatus(),
		})
	}

	ResultOutput(res, req, map[string]interface{}{
		"transactionID":    cceRes.TransactionID,
		"txValidationCode": cceRes.TxValidationCode,
		"chaincodeStatus":  cceRes.ChaincodeStatus,
		"payload":          string(cceRes.Payload),
		"peerResponses":    peerRes,
	})
}
