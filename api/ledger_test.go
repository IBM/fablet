package api

import (
	"encoding/json"
	"testing"
)

// TestUtilLedgerQuery to test query ledger
func TestLedgerQueryByAPI(t *testing.T) {
	conn, err := getConnectionSimple()
	if err != nil {
		t.Fatal(err)
	}

	///////////////////////////////////

	ledger, err := QueryLedger(conn, mychannel, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(">>>>>>>>>>>>>> ", ledger)

	blocks, err := QueryBlock(conn, mychannel, nil, 10, 1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(">>>>>>>>>>>>>> ", blocks)
}

func TestBlockLSCCAPI(t *testing.T) {
	conn, err := getConnectionSimple()
	if err != nil {
		t.Fatal(err)
	}

	defer conn.Close()
	///////////////////////////////////
	t.Log(conn.Show())

	blocks, err := QueryBlock(conn, "mychannel", []string{target01}, 10, 1)
	if err != nil {
		t.Fatal(err)
	}
	js, _ := json.Marshal(blocks[0])
	t.Log(string(js))

}
