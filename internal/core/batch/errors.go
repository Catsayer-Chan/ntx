package batch

import "fmt"

// ErrUnsupportedTaskType 返回标准化的任务类型错误
func ErrUnsupportedTaskType(taskType TaskType) error {
	return fmt.Errorf("不支持的任务类型: %s", taskType)
}
