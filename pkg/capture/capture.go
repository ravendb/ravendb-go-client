package capture

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
)

// given ip address finds corresponding device name
func findIterfaceForIPAddress(ipAddr string) (string, error) {
	ip := net.ParseIP(ipAddr)
	if ip == nil {
		// invalid ip address
		return "", fmt.Errorf("'%s' is not a valid ipAddr", ipAddr)
	}
	ifaces, err := pcap.FindAllDevs()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		for _, addr := range iface.Addresses {
			mask := addr.Netmask
			ipMasked := ip.Mask(mask)
			devIPMasked := addr.IP.Mask(mask)
			if ipMasked.Equal(devIPMasked) {
				return iface.Name, nil
			}
		}
	}
	return "", nil
}

type pcapCloser struct {
	askToStop int32
	didStop   int32
}

func (c *pcapCloser) Close() error {
	atomic.StoreInt32(&c.askToStop, 1)
	n := 0
	for {
		time.Sleep(time.Second / 2)
		v := atomic.LoadInt32(&c.didStop)
		if v != 0 {
			break
		}
		n++
		// should stop within second, panic if took more than 5 secs
		if n > 10 {
			panic("didn't stop within 5 seconds")
		}
	}
	return nil
}

func (c *pcapCloser) markStopped() {
	atomic.StoreInt32(&c.didStop, 1)
}

func (c *pcapCloser) shouldStop() bool {
	v := atomic.LoadInt32(&c.askToStop)
	return v != 0
}

// StartCapture starts capture of packets at a given ip address and saves
// the packets to pcap file
// To finish capture, call Close() on returned io.Closer
func StartCapture(ipAddr string, pcapPath string) (io.Closer, error) {
	// addr is "127.0.0.1:3432"
	port := 0
	parts := strings.Split(ipAddr, ":")
	if len(parts) > 2 {
		return nil, fmt.Errorf("ip address '%s' is not valid", ipAddr)
	}
	if len(parts) == 2 {
		var err error
		port, err = strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("ip address '%s' is not valid. Failed to parse port '%s'", ipAddr, parts[1])
		}
		ipAddr = parts[0]
	}

	devName, err := findIterfaceForIPAddress(ipAddr)
	if devName == "" {
		if err == nil {
			return nil, fmt.Errorf("didn't find network interface for ip address '%s'", ipAddr)
		}
		return nil, fmt.Errorf("didn't find network interface for ip address '%s'. Error: %s", ipAddr, err)
	}

	// fmt.Printf("Opening packet capture on '%s' for port '%d'\n", devName, port)
	snaplen := 1600
	handleRead, err := pcap.OpenLive(devName, int32(snaplen), true, time.Second)
	if err != nil {
		return nil, err
	}
	if port != 0 {
		// only capture packets where source or destination port is given
		filter := fmt.Sprintf("tcp port %d", port)
		err = handleRead.SetBPFFilter(filter)
		if err != nil {
			return nil, fmt.Errorf("handle.SetBPFFilter('%s') failed with %s", filter, err)
		}
	}
	pcapFile, err := os.Create(pcapPath)
	if err != nil {
		handleRead.Close()
		return nil, err
	}
	pcapWriter := pcapgo.NewWriter(pcapFile)
	err = pcapWriter.WriteFileHeader(uint32(snaplen), handleRead.LinkType())
	if err != nil {
		handleRead.Close()
		pcapFile.Close()
		os.Remove(pcapPath)
		return nil, err
	}
	res := &pcapCloser{}
	go func() {
		for {
			data, ci, err := handleRead.ReadPacketData()
			if err == nil {
				err = pcapWriter.WritePacket(ci, data)
				if err != nil {
					break
				}
				continue
			}

			if err == pcap.NextErrorTimeoutExpired {
				if res.shouldStop() {
					break
				}
				continue
			}
			break
		}
		handleRead.Close()
		pcapFile.Close()
		res.markStopped()
	}()

	return res, nil
}
