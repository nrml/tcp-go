package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

var (
	proxies = make(map[string]int)
)

func main() {
	var listen int64
	var err error

	err = load()

	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 0 {
		listen, err = strconv.ParseInt(os.Args[1], 10, 64)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		listen = 80
	}

	log.Printf("listening on %d\n", listen)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", listen))
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go func(c net.Conn) {
			local := c.LocalAddr().String()
			remote := c.RemoteAddr().String()
			log.Printf("read local address: %v\n", local)
			log.Printf("read remote address: %v\n", remote)
			spl := strings.Split(local, ":")
			nmsl := spl[:len(spl)-1]
			//in case of ipv6
			nm := strings.Join(nmsl, ":")
			localAddr := fmt.Sprintf("%s:%d", nm, proxies[nm])
			log.Printf("wants to forward to: %v\n", localAddr)
			forward(c, localAddr)
		}(conn)
	}
}

///{
///	"127.0.0.1": 8080,
///	"[::1]": 8080
///}
func load() error {
	f, err := os.Open("table.json")
	if err != nil {
		return err
	}
	dec := json.NewDecoder(f)
	dec.Decode(&proxies)
	return err
}

func forward(local net.Conn, remoteAddr string) {
	remote, err := net.Dial("tcp", remoteAddr)
	if remote == nil {
		fmt.Fprintf(os.Stderr, "remote dial failed: %v\n", err)
		return
	}
	go io.Copy(local, remote)
	go io.Copy(remote, local)

}

type proxy struct {
	host string
	port int
}
