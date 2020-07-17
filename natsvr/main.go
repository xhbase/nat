package main

import (
	"flag"
	"fmt"
	"net"
)

var nat_svraddr string = ":1082"
var socks5_svraddr string = ":1080"

func main() {
	flag.StringVar(&nat_svraddr, "svraddr", ":1082", "nat service addr 202.36.25.188:1082")
	flag.StringVar(&socks5_svraddr, "socks5addr", ":1080", "sock5 service addr 202.36.25.188:1080")
	flag.Parse()

	if len(nat_svraddr) == 0 || len(socks5_svraddr) == 0 || !flag.Parsed() {
		flag.Usage()
		return
	}

	svrl, e := net.Listen("tcp", nat_svraddr)
	if e != nil {
		fmt.Println("failed to listen:", e)
		return
	}

	svrs, e := net.Listen("tcp", socks5_svraddr)
	if e != nil {
		fmt.Println("failed to socks5 listen:", e)
		return
	}

	go ListenSvrAddr(svrl)
	ListenSvrSocks5(svrs)
}
