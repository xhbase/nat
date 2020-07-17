package natlib

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
)

func SendAll(c net.Conn, d []byte) error {
	dlen := len(d)
	for dlen > 0 {
		n, e := c.Write(d[len(d)-dlen:])
		if e != nil {
			return e
		}
		dlen -= n
	}
	return nil
}

func ReadN(c net.Conn, n int) ([]byte, error) {
	if n <= 0 {
		return nil, fmt.Errorf("param error")
	}
	d := make([]byte, n)
	for n > 0 {
		rn, e := c.Read(d[len(d)-n:])
		if e != nil {
			return nil, e
		}
		n -= rn
	}

	return d, nil
}

func SendMsg(c net.Conn, msg interface{}) error {
	d, _ := json.Marshal(msg)
	pkg := make([]byte, 2)
	binary.BigEndian.PutUint16(pkg, uint16(len(d)))
	pkg = append(pkg, d...)
	e := SendAll(c, pkg)
	if e != nil {
		return e
	}
	return nil
}

func RecvMsg(c net.Conn, msg interface{}) error {
	l, e := ReadN(c, 2)
	if e != nil {
		return e
	}
	lpkg := binary.BigEndian.Uint16(l)
	d, e := ReadN(c, int(lpkg))
	if e != nil {
		return e
	}

	e = json.Unmarshal(d, msg)
	return e
}

func JointConn(c1 net.Conn, c2 net.Conn) {

	//两个连接串起来
	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(c1, c2) //读取原地址请求（conn），然后将读取到的数据发送给目标主机。这里建议用"r",不建议用conn哟！因为它有重传机制！
		c1.Close()
	}()

	go func() {
		defer wg.Done()
		io.Copy(c2, c1) //与上面相反，就是讲目标主机的数据返回给客户端。
		c2.Close()
	}()
	wg.Wait()

	fmt.Println("exit joint")
}
