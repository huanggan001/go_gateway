package load_balance

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type WeightRoundRobinBalance struct {
	curIndex int
	server   []*WeightNode
	rsw      []int

	//观察主体
	conf LoadBalanceConf
}

type WeightNode struct {
	addr            string
	weight          int // 权重值
	currentWeight   int //节点当前权重
	effectiveWeight int //有效权重
}

func (r *WeightRoundRobinBalance) Add(params ...string) error {
	if len(params) != 2 {
		return errors.New("params len need 2")
	}
	parInt, err := strconv.ParseInt(params[1], 10, 64)
	if err != nil {
		return err
	}
	node := &WeightNode{addr: params[0], weight: int(parInt)}
	node.effectiveWeight = node.weight
	node.currentWeight = node.weight
	r.server = append(r.server, node)
	return nil
}

func (r *WeightRoundRobinBalance) Next() string {
	total := 0
	index := -1
	var best *WeightNode
	for i := 0; i < len(r.server); i++ {
		server := r.server[i]
		fmt.Printf("%d ", server.currentWeight)
		//统计所有有效值权重之和
		total += server.effectiveWeight
		//变更节点临时权重为节点临时权重 + 节点有效权重
		server.currentWeight += server.effectiveWeight
		//有效权重默认与权重相同，通讯异常时-1，通讯成功+1，直到恢复到weight大小
		//if server.effectiveWeight < server.weight {
		//	server.effectiveWeight++
		//}
		//选择最大临时权重节点
		if best == nil || server.currentWeight > best.currentWeight {
			best = server
			index = i
		}
	}
	fmt.Println()
	if best == nil {
		return ""
	}
	//变更临时权重为 临时权重-有效权重之和
	best.currentWeight -= total
	fmt.Println("best.currentWeight = ", best.currentWeight)
	fmt.Println("r.server", r.server[index].currentWeight)
	return best.addr
}

func (r *WeightRoundRobinBalance) Get(key string) (string, error) {
	return r.Next(), nil
}

func (r *WeightRoundRobinBalance) SetConf(conf LoadBalanceConf) {
	r.conf = conf
}

func (r *WeightRoundRobinBalance) Update() {
	if conf, ok := r.conf.(*LoadBalanceCheckConf); ok {
		fmt.Println("WeightRoundRobinBalance get conf:", conf.GetConf())
		r.server = nil
		for _, ip := range conf.GetConf() {
			r.Add(strings.Split(ip, ",")...)
		}
	}
}
