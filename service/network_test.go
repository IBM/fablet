package service

import (
	"testing"

	"github.com/IBM/fablet/util"
)

func TestPeerStatus(t *testing.T) {
	statuses := map[string]util.EndPointStatus{
		"Timeout":     util.EndPointStatus_Timeout,
		"Notfound":    util.EndPointStatus_NotFound,
		"Refused":     util.EndPointStatus_Refused,
		"Connectable": util.EndPointStatus_Connectable,
		"Valid":       util.EndPointStatus_Valid,
	}
	transStatuses := transPeerStatuses(statuses)
	for peer, status := range transStatuses {
		t.Log(peer, status)
	}
}
