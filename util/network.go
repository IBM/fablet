package util

import (
	"net"
	"strings"
	"time"
)

type EndPointStatus int

const (
	EndPointStatus_Valid EndPointStatus = iota
	EndPointStatus_Connectable
	EndPointStatus_Refused
	EndPointStatus_Timeout
	EndPointStatus_NotFound
)

func resolveAddress(address string) error {
	_, err := net.ResolveTCPAddr("tcp", address)
	return err
}

func connectAddress(address string) error {
	conn, err := net.DialTimeout("tcp", address, time.Second*3)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}

// GetEndpointStatus to get status of the peer endpoint.
// The result must not be EndpointStatus_Valid, since that is determined by peer query, instead of this network method,
func GetEndpointStatus(address string) EndPointStatus {
	_, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return EndPointStatus_NotFound
	}
	conn, err := net.DialTimeout("tcp", address, time.Second*2)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "connection refused") {
			return EndPointStatus_Refused
		}
		// What might includes "i/o timeout"
		return EndPointStatus_Timeout
	}
	defer conn.Close()
	return EndPointStatus_Connectable
}
