package api

import (
	"testing"

	"github.com/pkg/errors"
)

// Refined.

func TestNetwork(t *testing.T) {
	conn, err := NewConnection(
		&ConnectionProfile{connConfig, yamlConfigType},
		&Participant{"TestAdmin", "", mspIDOrg1, testCert, testPrivKey, nil},
		true,
	)
	if err != nil {
		t.Fatal(errors.WithStack(err).Error())
	}
	defer conn.Close()
	t.Log(conn.Show())

}

func TestQueryInstalledChaincodes(t *testing.T) {
	conn, err := NewConnection(
		&ConnectionProfile{connConfig, yamlConfigType},
		&Participant{"TestAdmin", "", mspIDOrg1, testCert, testPrivKey, nil},
		true,
	)
	if err != nil {
		t.Fatal(errors.WithStack(err).Error())
	}
	defer conn.Close()

	ccs, err := QueryInstalledChaincodes(conn, target01)
	if err != nil {
		t.Fatal(err)
	}
	for _, cc := range ccs {
		t.Log(cc)
	}
}
