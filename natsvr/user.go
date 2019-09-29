package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

const (
	NAT_SVR_LOCAL  = 1 //服务器内部网络代理
	NAT_USER_PROXY = 2 //客户端环境代理
)

type ACCOUNT struct {
	User  string //用户账号
	Pass  string //用户密码
	Atype int    //账户类型
}

var mAccount map[string]ACCOUNT
var mu sync.RWMutex

func init() {
	//加载账户
	ad, e := ioutil.ReadFile("./account.json")
	if e != nil {
		fmt.Println("failt to load ./account.json:", e)
		os.Exit(1)
	}
	mAccount = make(map[string]ACCOUNT)
	e = json.Unmarshal(ad, &mAccount)
	if e != nil {
		fmt.Println("failt to Unmarshal ./account.json:", e)
		os.Exit(1)
	}
	if len(mAccount) == 0 {
		fmt.Println("account is null of ./account.json:")
		os.Exit(1)
	}
}
func getUserAccount(user string) (ACCOUNT, error) {
	mu.RLock()
	mu.RUnlock()

	a, ok := mAccount[user]
	if !ok {
		return a, fmt.Errorf("not found account.")
	}
	return a, nil
}

func userAuth(user string, pass string) error {
	a, e := getUserAccount(user)
	if e != nil {
		return e
	}
	if a.Atype != NAT_USER_PROXY && a.Atype != NAT_SVR_LOCAL {
		return fmt.Errorf("user account not support proxy")
	}

	if a.Pass != pass {
		return fmt.Errorf("account auth failed.")
	}

	return nil
}
