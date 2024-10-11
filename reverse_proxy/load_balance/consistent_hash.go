package load_balance

import (
	"errors"
	"fmt"
	"hash/crc32"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// 声明新切片类型
type Hash func(data []byte) uint32

type UInt32Slice []uint32

// 返回切片长度
func (x UInt32Slice) Len() int {
	return len(x)
}

// 对比两个数大小
func (x UInt32Slice) Less(i, j int) bool {
	return x[i] < x[j]
}

// 切片中两个值的交换
func (x UInt32Slice) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

// 当hash环上没有数据时
//var errEmpty = errors.New("hash环上没有数据")

type ConsistentHashBalance struct {
	mux      sync.RWMutex
	hash     Hash
	replicas int               //复制因子
	keys     UInt32Slice       //已排序的节点hash切片
	hasMap   map[uint32]string //节点哈希和key的map，键是hash值，值是节点key

	//观察主体
	conf LoadBalanceConf
}

func NewConsistentHashBalance(replicas int, fn Hash) *ConsistentHashBalance {
	m := &ConsistentHashBalance{
		replicas: replicas,
		hash:     fn,
		hasMap:   make(map[uint32]string),
	}
	if m.hash == nil {
		//最多32位，保证是一个2^32-1环
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// 验证是否为空
func (c *ConsistentHashBalance) IsEmpty() bool {
	return len(c.keys) == 0
}

// Add 方法用来添加缓存节点，参数为节点key，比如使用IP
func (c *ConsistentHashBalance) Add(params ...string) error {
	if len(params) == 0 {
		return errors.New("param len 1 at least")
	}
	addr := params[0]
	c.mux.Lock()
	defer c.mux.Unlock()
	//结合复制因子计算所有虚拟节点的hash值，并存入m.keys中，同时在m.hashMap中保存哈希值和key的映射
	for i := 0; i < c.replicas; i++ {
		hash := c.hash([]byte(strconv.Itoa(i) + addr))
		c.keys = append(c.keys, hash)
		c.hasMap[hash] = addr
	}
	//对所有虚拟节点的哈希值进行排序，方便之后进行二分查找
	sort.Sort(c.keys)
	return nil
}

// Get方法根据给定的对象获取最靠近它的那个节点
func (c *ConsistentHashBalance) Get(key string) (string, error) {
	if c.IsEmpty() {
		return "", errors.New("node is empty")
	}
	hash := c.hash([]byte(key))

	//通过二分查找获取最优节点，第一个服务器hash值大于数据hash值的就是最优服务器节点
	idx := sort.Search(len(c.keys), func(i int) bool {
		return c.keys[i] >= hash
	})

	//如果查找结果大于服务器节点哈希数组的最大索引，表示此时该对象哈希值位于最后一个节点之后，那么放入第一个节点中
	if idx == len(c.keys) {
		idx = 0
	}
	c.mux.RLock()
	defer c.mux.Unlock()
	return c.hasMap[c.keys[idx]], nil
}

func (c *ConsistentHashBalance) SetConf(conf LoadBalanceConf) {
	c.conf = conf
}

func (c *ConsistentHashBalance) Update() {

	if conf, ok := c.conf.(*LoadBalanceCheckConf); ok {
		fmt.Println("ConsistentHashBalance Update get conf:", conf.GetConf())
		c.keys = nil
		c.hasMap = map[uint32]string{}
		for _, ip := range conf.GetConf() {
			c.Add(strings.Split(ip, ",")...)
		}
	}

}

// 创建结构体保存一致性哈希信息
//type HashConsistent struct {
//	//hash环，key为哈希值，值存放节点信息
//	circle map[uint32]string
//	//已经排序的节点的hash切片，方便后期查看key落在hash环上的哪个节点上，使用二分查找
//	//因为key是通过顺时针查找最近的节点
//	sortedHashes units
//	//虚拟节点个数，用来增加hash的平衡性
//	VirtualNode int
//	//map 读写锁
//	sync.RWMutex
//}

// 创建一致性hash算法结构体，设置默认虚拟节点个数
//func NewHashConsistent() *HashConsistent {
//	return &HashConsistent{
//		//初始化变量
//		circle: make(map[uint32]string),
//		//设置虚拟节点个数
//		VirtualNode: 10,
//	}
//}

// 自动生成key值
//func (c *HashConsistent) generateKey(element string, i int) string {
//	//副本key生成逻辑
//	return element + strconv.Itoa(i)
//}
//
//// 获取hash位置，计算key的hash值
//func (c *HashConsistent) hashKey(key string) uint32 {
//	if len(key) < 64 {
//		//声明一个数组长度为64
//		var srcatch [64]byte
//		//拷贝数据到数组
//		copy(srcatch[:], key)
//		//使用IEEE多项式返回数据的CRC-32校验和
//		return crc32.ChecksumIEEE(srcatch[:len(key)])
//	}
//	return crc32.ChecksumIEEE([]byte(key))
//}
//
//// 更新hash排序，方便查找
//func (c *HashConsistent) updateSortedHashes() {
//	hashes := c.sortedHashes[:0]
//	//判断切片容量,是否过大，如果过大则重置
//	if cap(c.sortedHashes)/(c.VirtualNode*4) > len(c.circle) {
//		hashes = nil
//	}
//
//	//添加host
//	for k := range c.circle {
//		hashes = append(hashes, k)
//	}
//	//对所有节点hash进行排序
//	//方便之后进行二分查找
//	sort.Sort(hashes)
//	//重新赋值
//	c.sortedHashes = hashes
//}
//
//// 向hash环中添加节点
//func (c *HashConsistent) Add(params ...string) error {
//	c.Lock()
//	defer c.Unlock()
//	element := params[0]
//	c.add(element)
//	return nil
//}
//
//// 添加节点
//func (c *HashConsistent) add(element string) {
//	//循环虚拟节点，设置副本
//	for i := 0; i < c.VirtualNode; i++ {
//		//根据生成的节点添加到hash环中，先生成key，再计算key的hash值
//		c.circle[c.hashKey(c.generateKey(element, i))] = element
//	}
//	//更新排序
//	c.updateSortedHashes()
//}
//
//// 删除节点对应的虚拟节点
//func (c *HashConsistent) remove(element string) {
//	for i := 0; i < c.VirtualNode; i++ {
//		delete(c.circle, c.hashKey(c.generateKey(element, i)))
//	}
//	c.updateSortedHashes()
//}
//
//// 删除一个节点
//func (c *HashConsistent) Remove(element string) {
//	c.Lock()
//	defer c.Unlock()
//	c.remove(element)
//}
//
//// 顺时针查找最近的节点
//func (c *HashConsistent) search(key uint32) int {
//	//查找算法
//	f := func(x int) bool {
//		return c.sortedHashes[x] > key
//	}
//	//使用“二分法”来搜索指定切片满足条件的最小值
//	//返回值为c.sortedHashes[x] > key的最小索引x
//	i := sort.Search(len(c.sortedHashes), f)
//	//如果超出范围则设置i=0
//	if i >= len(c.sortedHashes) {
//		i = 0
//	}
//	return i
//}
//
//// 根据数据标示获取最近的服务器节点信息
//func (c *HashConsistent) GetServer(name string) (string, error) {
//	c.RLock()
//	defer c.RUnlock()
//
//	//如果为0则返回错误
//	if len(c.circle) == 0 {
//		return "", errEmpty
//	}
//	//计算hash值
//	key := c.hashKey(name)
//	//根据数据标示获取最近的服务器
//	i := c.search(key)
//	return c.circle[c.sortedHashes[i]], nil
//}
//
//func (c *HashConsistent) Get(name string) (string, error) {
//	return c.GetServer(name)
//}
