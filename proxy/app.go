package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	//"strings"
)

var (
	proxies = make(map[string]int)
)

func main() {
	var listen int64
	var err error

	err = load()

	fmt.Printf("found %d proxies\n", len(proxies))

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

	listeners := make(map[int]net.Listener)

	for h, p := range proxies {

		log.Printf("range on %s:%d\n", h, p)
		log.Printf("trying to listen on: %v", fmt.Sprintf("%s:%d", h, listen))
		l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", h, listen))

		if err != nil {
			log.Fatalf("fatal error listening: %s,  %v", h, err)
		} else {
			log.Printf("listening on %d\n", listen)
		}

		listeners[p] = l

		defer l.Close()
	}

	for {
		for p, lstn := range listeners {
			conn, err := lstn.Accept()
			if err != nil {
				log.Printf("failed to listen for internal port %d", p)
			}
			go func(c net.Conn, port int) {
				listenAndForward(c, port)
			}(conn, p)

		}

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
func listenAndForward(conn net.Conn, p int) {
	go func(c net.Conn, port int) {

		localAddr := fmt.Sprintf("127.0.0.1:%d", port)
		log.Printf("wants to forward to: %v\n", localAddr)
		forward(c, localAddr)
	}(conn, p)

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
