package api

import (
	"testing"

	"github.com/pkg/errors"
)

// Refined.

func TestNetwork(t *testing.T) {
	conn, err := NewConnection(
		&ConnectionProfile{connConfigPre, yamlConfigType},
		&Participant{"TestAdmin", "", mspIDOrg1, testCert, testPrivKey, nil},
		true,
	)
	if err != nil {
		t.Fatal(errors.WithStack(err).Error())
	}
	defer conn.Close()
	t.Log(conn.Show())
}
