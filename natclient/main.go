package main

import (
	"flag"
	"fmt"
	"net"
	"time"
	"util/nat/natlib"
)

type NATCLICONN struct {
	c net.Conn
}

func (nc *NATCLICONN) Auth() error {
	var req natlib.AUTH_MSG
	req.Cmd = natlib.CMD_AUTH
	req.User = nat_user
	req.Pass = nat_pass
	e := natlib.SendMsg(nc.c, &req)
	if e != nil {
		return e
	}

	var resp natlib.MSG_RSP
	e = natlib.RecvMsg(nc.c, &resp)
	if e != nil {
		return e
	}

	if resp.Code != natlib.S_OK {
		return fmt.Errorf("Code:%v, msg:%v", resp.Code, resp.Msg)
	}

	return nil
}

func (nc *NATCLICONN) DealCreateConnMsg(req *natlib.CREATE_CONN) error {
	var rsp natlib.MSG_RSP
	rsp.Cmd = req.Cmd
	rsp.Seq = req.Seq

	//创建连
	c, e := net.Dial("tcp", req.NatAddr)
	if e != nil {
		rsp.Code = natlib.S_CREATE_CONN_FAILED
		rsp.Msg = fmt.Sprintf("addr:%v, emsg:%v", req.NatAddr, e)
	} else {
		rsp.Code = natlib.S_OK
		rsp.Msg = fmt.Sprintf("succ to create conn")
	}
	e = natlib.SendMsg(nc.c, &rsp)
	if e != nil {
		c.Close()
		nc.c.Close()
		return fmt.Errorf("failed to send msg")
	}

	//连接串起来
	go natlib.JointConn(nc.c, c)

	return nil
}

func (nc *NATCLICONN) DealUnknownMsg(req *natlib.CREATE_CONN) error {
	var rsp natlib.MSG_RSP
	rsp.Cmd = req.Cmd
	rsp.Seq = req.Seq
	rsp.Code = natlib.S_PARAM_ERR
	return natlib.SendMsg(nc.c, &rsp)
}

func (nc *NATCLICONN) DealHelloMsg(req *natlib.CREATE_CONN) error {
	var rsp natlib.MSG_RSP
	rsp.Cmd = req.Cmd
	rsp.Seq = req.Seq
	rsp.Code = natlib.S_OK
	return natlib.SendMsg(nc.c, &rsp)
}
func ConnectSvr(svraddr string) error {

	conn, e := net.Dial("tcp", svraddr)
	if e != nil {
		fmt.Println("failed to conn svr:", e)
		return e
	}

	nc := NATCLICONN{c: conn}

	e = nc.Auth()
	if e != nil {
		fmt.Println("auth failed:", e)
		return e
	}
	fmt.Println("succ to conn from natcli to natsvr.")

	//转变身份，变成监听
	for {
		var req natlib.CREATE_CONN
		e := natlib.RecvMsg(nc.c, &req)
		if e != nil {
			break
		}
		if req.Cmd == natlib.CMD_HELLO {
			if nc.DealHelloMsg(&req) != nil {
				break
			}

		} else if req.Cmd == natlib.CMD_CREATE_CONN {
			nc.DealCreateConnMsg(&req)
			return nil

		} else {
			if nc.DealUnknownMsg(&req) != nil {
				break
			}
		}
	}
	conn.Close()

	return fmt.Errorf(" some error")

}

var nat_svraddr string
var nat_user string
var nat_pass string

func main() {
	flag.StringVar(&nat_svraddr, "addr", "", "service addr of nat, eg: 202.88.1.56:5001")
	flag.StringVar(&nat_user, "user", "", "user for auth")
	flag.StringVar(&nat_pass, "pass", "", "password for auth")

	flag.Parse()
	if len(nat_svraddr) == 0 || len(nat_user) == 0 || len(nat_pass) == 0 {
		flag.Usage()
		return
	}

	for {
		e := ConnectSvr(nat_svraddr)
		if e != nil {
			time.Sleep(time.Second * 5)
		}
	}

}
