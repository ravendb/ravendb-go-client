// +build !windows

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

	"github.com/ga0/netgraph/ngnet"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
	"github.com/google/gopacket/tcpassembly"
)

var (
	// RequestsSummaryWriter is where we log summary line of http requests
	// must be set before StartCapture
	// if not set, we don't print summaries
	RequestsSummaryWriter io.Writer
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

type PacketCapturer struct {
	wasAskedToStop int32
	didStop        int32
	pcapFile       *os.File
	pcapWriter     *pcapgo.Writer
	handleRead     *pcap.Handle
	packetSource   *gopacket.PacketSource

	eventChan chan interface{}
	assembler *tcpassembly.Assembler
}

func (c *PacketCapturer) Close() error {
	atomic.StoreInt32(&c.wasAskedToStop, 1)
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

func (c *PacketCapturer) markStopped() {
	atomic.StoreInt32(&c.didStop, 1)
}

func (c *PacketCapturer) shouldStop() bool {
	v := atomic.LoadInt32(&c.wasAskedToStop)
	return v != 0
}

func (c *PacketCapturer) handlePacket(packet gopacket.Packet) {
	data := packet.Data()
	ci := packet.Metadata().CaptureInfo
	// ignore writing errors
	_ = c.pcapWriter.WritePacket(ci, data)
	if c.assembler == nil {
		return
	}
	netLayer := packet.NetworkLayer()
	if netLayer == nil {
		return
	}
	transLayer := packet.TransportLayer()
	if transLayer == nil {
		return
	}
	tcp, _ := transLayer.(*layers.TCP)
	if tcp == nil {
		return
	}
	c.assembler.AssembleWithTimestamp(
		netLayer.NetworkFlow(),
		tcp,
		ci.Timestamp)
}

func (c *PacketCapturer) readPackets() {
	for {
		packet, err := c.packetSource.NextPacket()
		if err == pcap.NextErrorTimeoutExpired {
			// fmt.Printf("readPackets(): timeout error\n")
			if c.shouldStop() {
				break
			}
			continue
		}
		if err != nil {
			fmt.Printf("readPackets(): NextPacket() failed with %s\n", err)
			break
		}
		if c.shouldStop() {
			break
		}
		c.handlePacket(packet)
	}
	c.markStopped()
	c.handleRead.Close()
	c.pcapFile.Close()
}

func printRequestSummary(req *ngnet.HTTPRequestEvent) {
	fmt.Fprintf(RequestsSummaryWriter, "%s %s %s\n", req.Method, req.URI, req.Version)
}

func runEvents(eventChan <-chan interface{}) {
	for e := range eventChan {
		switch v := e.(type) {
		case ngnet.HTTPRequestEvent:
			printRequestSummary(&v)
		case ngnet.HTTPResponseEvent:
			// ignore
		default:
			panic(fmt.Sprintf("Unsupported event %T", e))
		}
	}
}

// StartCapture starts capture of packets at a given ip address and saves
// the packets to pcap file
// To finish capture, call Close() on returned io.Closer
func StartCapture(ipAddr string, pcapPath string) (*PacketCapturer, error) {
	// addr is "127.0.0.1:3432"
	var err error
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
	res := &PacketCapturer{}
	if RequestsSummaryWriter != nil {
		res.eventChan = make(chan interface{}, 1024)
		go runEvents(res.eventChan)
		streamFactory := ngnet.NewHTTPStreamFactory(res.eventChan)
		pool := tcpassembly.NewStreamPool(streamFactory)
		res.assembler = tcpassembly.NewAssembler(pool)
	}

	snaplen := 65536
	res.handleRead, err = pcap.OpenLive(devName, int32(snaplen), true, time.Second)
	if err != nil {
		return nil, err
	}
	res.packetSource = gopacket.NewPacketSource(res.handleRead, res.handleRead.LinkType())
	if port != 0 {
		// only capture packets where source or destination port is given
		filter := fmt.Sprintf("tcp port %d", port)
		err = res.handleRead.SetBPFFilter(filter)
		if err != nil {
			return nil, fmt.Errorf("handle.SetBPFFilter('%s') failed with %s", filter, err)
		}
	}
	res.pcapFile, err = os.Create(pcapPath)
	if err != nil {
		res.handleRead.Close()
		return nil, err
	}
	res.pcapWriter = pcapgo.NewWriter(res.pcapFile)
	linkType := res.handleRead.LinkType()
	err = res.pcapWriter.WriteFileHeader(uint32(snaplen), linkType)
	if err != nil {
		res.handleRead.Close()
		res.pcapFile.Close()
		os.Remove(pcapPath)
		return nil, err
	}
	go res.readPackets()

	return res, nil
}
