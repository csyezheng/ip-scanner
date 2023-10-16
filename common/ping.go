package common

import (
	"bytes"
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"net"
	"os"
	"strings"
	"time"
)

const (
	ErrorConnRefused = "connection refused"
	ErrorTimeout     = "connection timed out"
)

// pingIcmp performs a ping to a destination. It selects between ipv4 or ipv6 ping based
// on the format of the destination ip.
func pingIcmp(destination string, timeout time.Duration) (err error) {
	var (
		icmpType icmp.Type
		network  string
	)

	if strings.Contains(destination, ":") {
		network = "ip6:ipv6-icmp"
		icmpType = ipv6.ICMPTypeEchoRequest
	} else {
		network = "ip4:icmp"
		icmpType = ipv4.ICMPTypeEcho
	}

	c, err := net.Dial(network, destination)
	if err != nil {
		return fmt.Errorf("dial failed: %v", err)
	}
	err = c.SetDeadline(time.Now().Add(timeout))
	if err != nil {
		return err
	}
	defer func(c net.Conn) {
		err := c.Close()
		if err != nil {
		}
	}(c)

	// xid is the process ID.
	// Get process ID and make sure it fits in 16bits.
	xid := os.Getpid() & 0xffff
	// Sequence number of the packet.
	xseq := 0
	packet := icmp.Message{
		Type: icmpType, // Type of icmp message
		Code: 0,        // icmp query messages use code 0
		Body: &icmp.Echo{
			ID:   xid,  // Packet id
			Seq:  xseq, // Sequence number of the packet
			Data: bytes.Repeat([]byte("Ping!Ping!Ping!"), 3),
		},
	}

	wb, err := packet.Marshal(nil)

	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}

	if _, err := c.Write(wb); err != nil {
		return fmt.Errorf("Conn.Write Error: %v", err)
	}

	rb := make([]byte, 1500)

	if _, err := c.Read(rb); err != nil {
		// If connection timed out, we return ErrorTimeout
		if e := err.(*net.OpError).Timeout(); e {
			return fmt.Errorf(ErrorTimeout)
		}
		if strings.Contains(err.Error(), "connection refused") {
			return fmt.Errorf(ErrorConnRefused)
		}
		return fmt.Errorf("Conn.Read failed: %v", err)
	}

	_, err = icmp.ParseMessage(icmpType.Protocol(), rb)
	if err != nil {
		return fmt.Errorf("ParseICMPMessage failed: %v", err)
	}

	return
}

// pingTcp performs a straightforward connection attempt on a destination ip:port and returns
// an error if the attempt failed
func pingTcp(destination string, destinationPort float64, timeout time.Duration) (err error) {
	conn, err := net.DialTimeout("tcp",
		fmt.Sprintf("%s:%d", destination, int(destinationPort)), timeout)
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	if err != nil {
		// If connection timed out, we return ErrorTimeout
		if e := err.(*net.OpError).Timeout(); e {
			return fmt.Errorf(ErrorTimeout)
		}
		if strings.Contains(err.Error(), "connection refused") {
			return fmt.Errorf(ErrorConnRefused)
		}
		return fmt.Errorf("dial Error: %v", err)
	}
	return
}

// pingUdp sends a UDP packet to a destination ip:port to determine if it is open or closed.
// Because UDP does not reply to connection requests, a lack of response may indicate that the
// port is open, or that the packet got dropped. We chose to be optimistic and treat lack of
// response (connection timeout) as an open port.
func pingUdp(destination string, destinationPort float64, timeout time.Duration) (err error) {
	c, err := net.Dial("udp",
		fmt.Sprintf("%s:%d", destination, int(destinationPort)))
	if err != nil {
		return fmt.Errorf("dial error: %v", err)
	}

	_, err = c.Write([]byte("Ping!Ping!Ping!"))
	if err != nil {
		return err
	}
	err = c.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return err
	}
	defer func(c net.Conn) {
		err := c.Close()
		if err != nil {

		}
	}(c)

	rb := make([]byte, 1500)

	if _, err := c.Read(rb); err != nil {
		// If connection timed out, we return ErrorTimeout
		if e := err.(*net.OpError).Timeout(); e {
			return fmt.Errorf(ErrorTimeout)
		}
		if strings.Contains(err.Error(), "connection refused") {
			return fmt.Errorf(ErrorConnRefused)
		}
		return fmt.Errorf("read error: %v", err.Error())
	}
	return nil
}
