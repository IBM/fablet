package service

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/IBM/fablet/api"
	"github.com/IBM/fablet/log"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

var logger = log.GetLogger()
var ExeFolder = GetExeFolder()

// HTTPHandler To handle all incoming http request
type HTTPHandler func(res http.ResponseWriter, req *http.Request)

// Post to return HTTPHandler only process post method.
func Post(hh HTTPHandler) HTTPHandler {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			PlainOutput(res, req, []byte(""))
			return
		}
		hh(res, req)
	}
}

// GetTmpFolder to get temp folder.
func GetTmpFolder() string {
	return filepath.Join(ExeFolder, "tmp", uuid.NewV1().String())
}

// GetExeFolder to get the folder of current executable binary file.
func GetExeFolder() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(exe)
}

// RequestConnection request with connection
type RequestConnection struct {
	Label         string `json:"label"`
	MSPID         string `json:"MSPID"`
	CertContent   string `json:"certContent"`
	PrvKeyContent string `json:"prvKeyContent"`
	ConnProfile   string `json:"connProfile"`
}

type Request interface {
	GetReqConn() *RequestConnection
}

type BaseRequest struct {
	Connection *RequestConnection `json:"connection"`
}

// GetReqConn interface function
func (req *BaseRequest) GetReqConn() *RequestConnection {
	return req.Connection
}

// RequestOption optios for request.
type RequestOption struct {
	Refresh bool
}

// RequestOptionFunc to handle the request option
type RequestOptionFunc func(opt *RequestOption) error

// WithRefresh only retrieve details of the peer.
func WithRefresh(refresh bool) RequestOptionFunc {
	return func(opt *RequestOption) error {
		opt.Refresh = refresh
		return nil
	}
}

func generateOption(options ...RequestOptionFunc) *RequestOption {
	reqOpt := &RequestOption{}
	for _, opt := range options {
		opt(reqOpt)
	}
	return reqOpt
}

func SetHeader(res http.ResponseWriter, req *http.Request, addHeaders map[string]string) {
	header := res.Header()
	header.Set("Access-Control-Allow-Origin", "*")
	header.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, PATCH, DELETE")
	header.Set("Access-Control-Allow-Headers", "X-Requested-With,content-type")
	header.Set("Access-Control-Allow-Credentials", "true")
	if addHeaders != nil {
		for k, v := range addHeaders {
			header.Set(k, v)
		}
	}
}

func JsonOutput(res http.ResponseWriter, req *http.Request, result map[string]interface{}) {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		logger.Error("Error occurs when marshal result to json.", err.Error())
	} else {
		SetHeader(res, req, map[string]string{"Content-Type": "application/json;charset=utf-8"})
		res.Write(resultJSON)
	}
}

// TODO all function should has return value.
func ErrorOutput(res http.ResponseWriter, req *http.Request, resCode ResCode, err error) {
	result := map[string]interface{}{
		"resCode": resCode,
		"errMsg":  err.Error(),
	}
	logger.Error(err.Error())
	JsonOutput(res, req, result)
}

func ResultOutput(res http.ResponseWriter, req *http.Request, result map[string]interface{}) {
	result["resCode"] = RES_CODE_OK
	JsonOutput(res, req, result)
}

func PlainOutput(res http.ResponseWriter, req *http.Request, content []byte) {
	SetHeader(res, req, map[string]string{"Content-Type": "text/plain;charset=utf-8"})
	res.Write(content)
}

type ResCode int32

const (
	RES_CODE_ERR_INTERNAL = ResCode(500)
	RES_CODE_OK           = ResCode(200)
)

// GetRequest get request from http
// TODO to use GetRequest for all services, and, all connections are not closed after calling, to hold connection in session.
func GetRequest(req *http.Request, reqBody Request, useDiscovery bool, options ...RequestOptionFunc) (*api.NetworkConnection, error) {
	logger.Info("Service common function GetReuest")
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, errors.WithMessage(err, "Error occurred when read from request.")
	}

	err = json.Unmarshal(body, reqBody)
	if err != nil {
		return nil, errors.WithMessage(err, "Error occurred when unmarshal ReqBody from request.")
	}

	reqConn := reqBody.GetReqConn()
	conn, err := getConnOfReq(reqConn, useDiscovery, options...)
	return conn, nil
}

func getConnOfReq(reqConn *RequestConnection, useDiscovery bool, options ...RequestOptionFunc) (*api.NetworkConnection, error) {
	opt := generateOption(options...)
	connFunc := GetConnection
	if opt.Refresh {
		connFunc = NewConnection
	}
	conn, err := connFunc(
		// TODO to support multiple config file type
		&api.ConnectionProfile{Config: []byte(reqConn.ConnProfile), ConfigType: "yaml"},
		&api.Participant{Label: reqConn.Label, OrgName: "", MSPID: reqConn.MSPID,
			Cert: []byte(reqConn.CertContent), PrivateKey: []byte(reqConn.PrvKeyContent), SignID: nil},
		useDiscovery)

	if err != nil {
		return nil, errors.WithMessage(err, "Error occurred when create Fablet connection.")
	}
	return conn, nil
}
