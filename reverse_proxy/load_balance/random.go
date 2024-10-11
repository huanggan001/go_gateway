package load_balance

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
)

type RandomBalance struct {
	curIndex int
	server   []string

	//观察主体
	conf LoadBalanceConf
}

func (r *RandomBalance) Add(params ...string) error {
	if len(params) == 0 {
		return errors.New("param len 1 at least")
	}
	addr := params[0]
	r.server = append(r.server, addr)
	return nil
}

// 随机负载均衡
func (r *RandomBalance) Random_Next() string {
	if len(r.server) == 0 {
		return ""
	}
	r.curIndex = rand.Intn(len(r.server))
	return r.server[r.curIndex]
}

func (r *RandomBalance) Get(key string) (string, error) {
	return r.Random_Next(), nil
}

func (r *RandomBalance) SetConf(conf LoadBalanceConf) {
	r.conf = conf
}

func (r *RandomBalance) Update() {

	if conf, ok := r.conf.(*LoadBalanceCheckConf); ok {
		fmt.Println("RandomBalance Update get conf:", conf.GetConf())
		r.server = []string{}
		for _, ip := range conf.GetConf() {
			r.Add(strings.Split(ip, ",")...)
		}
	}
}
