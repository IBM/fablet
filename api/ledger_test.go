package api

import (
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

	blocks, err := QueryBlock(conn, "chtest1", []string{target01}, 0, 8)
	if err != nil {
		t.Fatal(err)
	}
	for idx, block := range blocks {
		t.Log(idx, block.Time)
	}

}
