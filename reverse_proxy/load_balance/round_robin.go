package load_balance

import (
	"errors"
	"fmt"
	"strings"
)

type RoundRobinBalance struct {
	curIndex int
	server   []string

	//观察主体
	conf LoadBalanceConf
}

func (r *RoundRobinBalance) Add(params ...string) error {
	if len(params) == 0 {
		return errors.New("param len 1 at least")
	}
	addr := params[0]
	r.server = append(r.server, addr)
	return nil
}

// 轮询负载均衡
func (r *RoundRobinBalance) Round_robin_Next() string {
	if len(r.server) == 0 {
		return ""
	}
	lens := len(r.server)
	if r.curIndex >= lens {
		r.curIndex = 0
	}
	curServer := r.server[r.curIndex]
	r.curIndex = (r.curIndex + 1) % lens
	return curServer
}

func (r *RoundRobinBalance) Get(key string) (string, error) {
	return r.Round_robin_Next(), nil
}

func (r *RoundRobinBalance) SetConf(conf LoadBalanceConf) {
	r.conf = conf
}

func (r *RoundRobinBalance) Update() {
	if conf, ok := r.conf.(*LoadBalanceCheckConf); ok {
		fmt.Println("RoundRobinBalance Update get conf:", conf.GetConf())
		r.server = []string{}
		for _, ip := range conf.GetConf() {
			r.Add(strings.Split(ip, ",")...)
		}
	}
}
