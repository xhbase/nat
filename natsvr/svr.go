package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/mic-yu/nat/natlib"
)

var mUserNatClient map[string][]net.Conn

func init() {
	mUserNatClient = make(map[string][]net.Conn)
}

var m sync.Mutex

func InsertNatCliConn(user string, c net.Conn) {
	m.Lock()
	defer m.Unlock()
	lst, _ := mUserNatClient[user]
	lst = append(lst, c)
	mUserNatClient[user] = lst
	fmt.Println(user, " conn cnt:", len(lst))
}

func PopNatCliConn(user string) net.Conn {
	m.Lock()
	defer m.Unlock()
	lst, ok := mUserNatClient[user]
	if !ok {
		return nil
	}

	if len(lst) > 0 {
		mUserNatClient[user] = lst[1:]
		return lst[0]
	}
	return nil
}

func DealSvrAuth(c net.Conn) (string, error) {
	var req natlib.AUTH_MSG
	var rsp natlib.MSG_RSP
	e := natlib.RecvMsg(c, &req)
	if e != nil {
		return "", e
	}

	//鉴权
	ea := userAuth(req.User, req.Pass)
	fmt.Printf("auth user:%v, e:%v\n", req.User, ea)

	rsp.Cmd = req.Cmd
	rsp.Seq = req.Seq
	if ea != nil {
		rsp.Code = natlib.S_AUTH_FAILED
	} else {
		rsp.Code = natlib.S_OK
	}
	e = natlib.SendMsg(c, &rsp)
	if e == nil && rsp.Code != natlib.S_OK {
		e = ea
	}
	return req.User, e
}

func DealSvrConn(c net.Conn) {
	//进行鉴权
	user, e := DealSvrAuth(c)
	if e != nil {
		c.Close()
		return
	}
	//放到队列，等待连接请求进来
	InsertNatCliConn(user, c)
}

func ListenSvrAddr(l net.Listener) {
	go KeepAlive()
	for {
		c, e := l.Accept()
		if e != nil {
			continue
		}

		go DealSvrConn(c)
	}
}

func KeepAlive() {
	for {
		time.Sleep(time.Second * 30)
		//对连接进行保活
	}

}
