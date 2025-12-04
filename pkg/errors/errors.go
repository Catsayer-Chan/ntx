// Package errors 提供 NTX 的错误类型定义
//
// 本模块定义了项目中使用的所有错误类型，包括：
// - 网络错误
// - 解析错误
// - 权限错误
// - 配置错误
//
// 作者: Catsayer
package errors

import (
	"errors"
	"fmt"
)

// 预定义错误
var (
	// ErrInvalidArgument 无效参数
	ErrInvalidArgument = errors.New("invalid argument")
	// ErrInvalidHost 无效主机
	ErrInvalidHost = errors.New("invalid host")
	// ErrInvalidPort 无效端口
	ErrInvalidPort = errors.New("invalid port")
	// ErrInvalidProtocol 无效协议
	ErrInvalidProtocol = errors.New("invalid protocol")
	// ErrInvalidIPVersion 无效 IP 版本
	ErrInvalidIPVersion = errors.New("invalid ip version")

	// ErrTimeout 超时
	ErrTimeout = errors.New("operation timeout")
	// ErrConnectionRefused 连接被拒绝
	ErrConnectionRefused = errors.New("connection refused")
	// ErrNetworkUnreachable 网络不可达
	ErrNetworkUnreachable = errors.New("network unreachable")
	// ErrHostUnreachable 主机不可达
	ErrHostUnreachable = errors.New("host unreachable")
	// ErrPortUnreachable 端口不可达
	ErrPortUnreachable = errors.New("port unreachable")

	// ErrPermissionDenied 权限被拒绝
	ErrPermissionDenied = errors.New("permission denied")
	// ErrNotSupported 不支持的操作
	ErrNotSupported = errors.New("operation not supported")

	// ErrDNSResolution DNS 解析失败
	ErrDNSResolution = errors.New("dns resolution failed")
	// ErrNoAddress 无可用地址
	ErrNoAddress = errors.New("no address available")

	// ErrCanceled 操作被取消
	ErrCanceled = errors.New("operation canceled")
	// ErrClosed 资源已关闭
	ErrClosed = errors.New("resource closed")
)

// NetworkError 网络错误
type NetworkError struct {
	// Op 操作名称
	Op string
	// Target 目标地址
	Target string
	// Err 底层错误
	Err error
}

// Error 实现 error 接口
func (e *NetworkError) Error() string {
	if e.Target != "" {
		return fmt.Sprintf("%s %s: %v", e.Op, e.Target, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

// Unwrap 返回底层错误
func (e *NetworkError) Unwrap() error {
	return e.Err
}

// NewNetworkError 创建网络错误
func NewNetworkError(op, target string, err error) *NetworkError {
	return &NetworkError{
		Op:     op,
		Target: target,
		Err:    err,
	}
}

// TimeoutError 超时错误
type TimeoutError struct {
	// Op 操作名称
	Op string
	// Duration 超时时长
	Duration string
}

// Error 实现 error 接口
func (e *TimeoutError) Error() string {
	return fmt.Sprintf("%s: timeout after %s", e.Op, e.Duration)
}

// Timeout 标记为超时错误
func (e *TimeoutError) Timeout() bool {
	return true
}

// NewTimeoutError 创建超时错误
func NewTimeoutError(op, duration string) *TimeoutError {
	return &TimeoutError{
		Op:       op,
		Duration: duration,
	}
}

// PermissionError 权限错误
type PermissionError struct {
	// Op 操作名称
	Op string
	// Resource 资源名称
	Resource string
	// Reason 原因
	Reason string
}

// Error 实现 error 接口
func (e *PermissionError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("%s %s: permission denied (%s)", e.Op, e.Resource, e.Reason)
	}
	return fmt.Sprintf("%s %s: permission denied", e.Op, e.Resource)
}

// NewPermissionError 创建权限错误
func NewPermissionError(op, resource, reason string) *PermissionError {
	return &PermissionError{
		Op:       op,
		Resource: resource,
		Reason:   reason,
	}
}

// ValidationError 验证错误
type ValidationError struct {
	// Field 字段名称
	Field string
	// Value 值
	Value interface{}
	// Message 错误消息
	Message string
}

// Error 实现 error 接口
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s: %s (value: %v)", e.Field, e.Message, e.Value)
}

// NewValidationError 创建验证错误
func NewValidationError(field string, value interface{}, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// NotSupportedError 不支持的错误
type NotSupportedError struct {
	// Feature 功能名称
	Feature string
	// Platform 平台名称
	Platform string
}

// Error 实现 error 接口
func (e *NotSupportedError) Error() string {
	if e.Platform != "" {
		return fmt.Sprintf("%s is not supported on %s", e.Feature, e.Platform)
	}
	return fmt.Sprintf("%s is not supported", e.Feature)
}

// NewNotSupportedError 创建不支持错误
func NewNotSupportedError(feature, platform string) *NotSupportedError {
	return &NotSupportedError{
		Feature:  feature,
		Platform: platform,
	}
}

// IsTimeout 判断是否为超时错误
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrTimeout) {
		return true
	}
	type timeout interface {
		Timeout() bool
	}
	te, ok := err.(timeout)
	return ok && te.Timeout()
}

// IsPermissionDenied 判断是否为权限错误
func IsPermissionDenied(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrPermissionDenied) {
		return true
	}
	_, ok := err.(*PermissionError)
	return ok
}

// IsNetworkError 判断是否为网络错误
func IsNetworkError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*NetworkError)
	return ok
}