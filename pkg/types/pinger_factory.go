package types

// PingerFactory 定义了根据配置创建 Pinger 的工厂接口
type PingerFactory interface {
	Create(opts *PingOptions) (Pinger, error)
}
