package capture

import (
	"errors"
	"io"
)

// This is only a stub because on windows gopacket requires cgo and winpcap

var (
	// RequestsSummaryWriter is where we log summary line of http requests
	// must be set before StartCapture
	// if not set, we don't print summaries
	RequestsSummaryWriter io.Writer
)

type PacketCapturer struct {
}

// StartCapture starts capture of packets at a given ip address and saves
// the packets to pcap file
// To finish capture, call Close() on returned io.Closer
func StartCapture(ipAddr string, pcapPath string) (*PacketCapturer, error) {
	return nil, errors.New("NYI")
}
