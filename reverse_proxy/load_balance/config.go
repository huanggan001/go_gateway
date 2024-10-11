package load_balance

type Oberver interface {
	Update()
}

// 配置主题
type LoadBalanceConf interface {
	Attach(o Oberver)
	GetConf() []string
	WatchConf()
	UpdateConf(conf []string)
}
