package ping

import (
	"fmt"

	"github.com/catsayer/ntx/internal/logger"
	pkgerrors "github.com/catsayer/ntx/pkg/errors"
	"github.com/catsayer/ntx/pkg/types"
	"go.uber.org/zap"
)

// Factory 提供根据配置创建 Pinger 的默认实现
type Factory struct{}

// NewFactory 返回默认 Factory
func NewFactory() *Factory {
	return &Factory{}
}

// Create 根据协议创建 Pinger，并在必要时执行回退逻辑
func (Factory) Create(opts *types.PingOptions) (types.Pinger, error) {
	if opts == nil {
		opts = types.DefaultPingOptions()
	}

	switch opts.Protocol {
	case types.ProtocolICMP:
		pinger, err := NewICMPPinger(opts)
		if err != nil {
			if pkgerrors.IsPermissionDenied(err) {
				logger.Warn("ICMP 需要权限，自动切换到 TCP", zap.Error(err))
				opts.Protocol = types.ProtocolTCP
				return NewTCPPinger(opts), nil
			}
			return nil, err
		}
		return pinger, nil
	case types.ProtocolTCP:
		return NewTCPPinger(opts), nil
	case types.ProtocolHTTP:
		return NewHTTPPinger(opts), nil
	}

	return nil, fmt.Errorf("未知的协议: %s", opts.Protocol)
}
