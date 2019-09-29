package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"util/nat/natlib"
)

func NatCreateConn(c net.Conn, addr string) error {
	var req natlib.CREATE_CONN
	req.Cmd = natlib.CMD_CREATE_CONN
	req.NatAddr = addr //"smtp.richinfo.cn:110"
	e := natlib.SendMsg(c, req)
	if e != nil {
		return e
	}

	var rsp natlib.MSG_RSP
	ersp := natlib.RecvMsg(c, rsp)
	if ersp != nil {
		return e
	}

	if rsp.Cmd != natlib.CMD_CREATE_CONN || rsp.Code != natlib.S_OK {
		return fmt.Errorf("failed to conn:%v, code:%v", req.NatAddr, rsp.Code)
	}
	return nil
}

func Socks5UserPwdCheck(c net.Conn) (string, error) {
	ver, e := natlib.ReadN(c, 1)
	if e != nil {
		return "", e

	}
	if ver[0] != 1 {
		return "", fmt.Errorf("ver 1, actual:%v", ver[0])
	}

	ul, e := natlib.ReadN(c, 1)
	if e != nil {
		return "", e

	}
	ud, e := natlib.ReadN(c, int(ul[0]))
	if e != nil {
		return "", e

	}

	pl, e := natlib.ReadN(c, 1)
	if e != nil {
		return "", e

	}
	pd, e := natlib.ReadN(c, int(pl[0]))
	if e != nil {
		return "", e

	}

	u := string(ud)
	p := string(pd)
	//校验用户名和密码
	//

	var pwrsp []byte
	ea := userAuth(u, p)
	if ea != nil {
		pwrsp = []byte{0x01, 0x01}
	} else {
		pwrsp = []byte{0x01, 0x00}
	}

	fmt.Println("sock5 user=, pass=", u, p, pwrsp, ea)

	//回应鉴定状态
	e = natlib.SendAll(c, pwrsp)
	if e != nil {
		return "", e
	}

	return u, nil
}

type ADDR_RSP struct {
	ver  uint8
	rep  uint8
	rsv  uint8
	atyp uint8
	addr string
	port uint16
}

func (rsp *ADDR_RSP) Encode() []byte {
	d := make([]byte, 1024)
	d[0] = rsp.ver
	d[1] = rsp.rep
	d[2] = rsp.rsv
	d[3] = rsp.atyp
	d[4] = uint8(len(rsp.addr))
	l := 5
	copy(d[l:], []byte(rsp.addr))
	l += len(rsp.addr)
	binary.BigEndian.PutUint16(d[l:], rsp.port)
	l += 2
	return d[:l]

}
func Socks5PickAddr(c net.Conn) (string, error) {
	var addr string
	d1, e := natlib.ReadN(c, 4)
	if e != nil {
		return "", e
	}
	ver := d1[0]
	cmd := d1[1]
	atyp := d1[3]

	var rsp ADDR_RSP

	if ver != 5 || cmd != 1 {
		rsp.ver = ver
		rsp.rep = 0xFF
		natlib.SendAll(c, rsp.Encode())
		return "", fmt.Errorf("not support ver %v", d1[0])
	}

	addrl, e := natlib.ReadN(c, 1)
	if e != nil {
		return "", e
	}
	addrd, e := natlib.ReadN(c, int(addrl[0]))
	if e != nil {
		return "", e
	}

	ipord := ""

	if atyp == 1 { //ipv4
		fmt.Printf("xxxxxxxxx 1")
		ipord = fmt.Sprintf("%v.%v.%v.%v", addrd[0], addrd[1], addrd[2], addrd[3])
	} else if atyp == 3 { //域名
		ipord = string(addrd)
	} else if atyp == 4 {
		//ipv6
		fmt.Print("xxxxxxxxxxxxxx 4")
	} else {
		//bao cuo
		fmt.Print("xxxxxxxxxxxxxx unknown")
	}

	pd, e := natlib.ReadN(c, 2)
	if e != nil {
		return "", e
	}
	port := binary.BigEndian.Uint16(pd)
	addr = fmt.Sprintf("%s:%v", ipord, port)
	fmt.Println("addr :", addr)

	return addr, nil

}
func Socks5MethodCheck(c net.Conn) error {
	d1, e := natlib.ReadN(c, 2)
	if e != nil {
		return e
	}
	if d1[0] != 5 {
		d1[1] = 0xff
		natlib.SendAll(c, d1)
		return fmt.Errorf("not support ver %v", d1[0])
	}

	d2, e := natlib.ReadN(c, int(d1[1]))
	if e != nil {
		return e
	}

	var userpass bool = false
	for i := uint8(0); i < d1[1]; i++ {
		if d2[i] == 0x02 {
			userpass = true
			break
		}
	}

	if !userpass {
		d1[2] = 0xff
		natlib.SendAll(c, d1)
		return fmt.Errorf("not support ver %v", d1[0])
	}
	d1[1] = 0x02
	e = natlib.SendAll(c, d1)
	if e != nil {
		return e
	}

	return nil

}
func Socks5Auth(c net.Conn) (string, string, error) {

	//检查协议版本
	e := Socks5MethodCheck(c)
	if e != nil {
		return "", "", e
	}

	//校验用户名和密码
	u, e := Socks5UserPwdCheck(c)
	if e != nil {
		return "", "", e
	}

	addr, e := Socks5PickAddr(c)
	if e != nil {
		return "", "", e
	}

	return u, addr, nil
}

func Socks5AckAddr(c net.Conn, nc net.Conn) error {
	/*
		var rsp ADDR_RSP
		rsp.ver = 5
		rsp.rep = 0
		rsp.atyp = 3

		laddr := nc.LocalAddr().String()
		idx := strings.LastIndex(laddr, ":")
		rsp.addr = laddr[0:idx]
		port, _ := strconv.Atoi(laddr[idx+1:])
		rsp.port = uint16(port)
	*/

	rd := []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	e := natlib.SendAll(c, rd)
	return e
}

func DealSocks5Conn(c net.Conn) {

	//得到用户和密码,并鉴权
	u, addr, e := Socks5Auth(c)
	if e != nil {
		c.Close()
		return
	}

	a, _ := getUserAccount(u)
	fmt.Println("user:", u, " a:", a)

	for {
		var nc net.Conn
		if a.Atype == NAT_USER_PROXY {
			nc = PopNatCliConn(u)
			if nc == nil {
				//告知没有待处理连接，稍后再试
				fmt.Println("yyyyyyyyyyyyyyyyyyy")
				c.Close()
				return
			}

			//对nc发起创建连接请求
			e := NatCreateConn(nc, addr)
			if e != nil {
				nc.Close()
				continue
			}
		} else {
			nc, e = net.Dial("tcp", addr)
			if e != nil {
				//暂时拒绝，后续完善代理回应
				c.Close()
				return
			}
		}

		//对代理进行回应
		e = Socks5AckAddr(c, nc)
		if e != nil {
			nc.Close()
			c.Close()
			return
		}

		natlib.JointConn(nc, c)
		break
	}

}

func ListenSvrSocks5(l net.Listener) {
	for {
		c, e := l.Accept()
		if e != nil {
			fmt.Println("failed to accept:", e)
			continue
		}

		DealSocks5Conn(c)

	}

}
