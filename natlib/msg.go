package natlib

const (
	CMD_AUTH        uint8 = 1
	CMD_CREATE_CONN uint8 = 2
	CMD_HELLO       uint8 = 3
)
const (
	S_OK                 uint16 = 0
	S_PARAM_ERR          uint16 = 1
	S_CREATE_CONN_FAILED uint16 = 2
	S_AUTH_FAILED        uint16 = 3
)

type MSG_RSP struct {
	Cmd  uint8
	Seq  uint32
	Code uint16
	Msg  string
}

//-------------- nat客户端到服务端
type AUTH_MSG struct {
	Cmd  uint8
	Seq  uint32
	User string //鉴权用户名
	Pass string //鉴权密码
}

//--------------- 服务端到nat客户端
type CREATE_CONN struct {
	Cmd     uint8
	Seq     uint32
	NatAddr string //此值为空，表示心跳
}
