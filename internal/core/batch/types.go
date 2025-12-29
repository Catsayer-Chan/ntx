package batch

import "time"

// TaskType 任务类型
type TaskType string

const (
	TaskTypePing  TaskType = "ping"
	TaskTypeDNS   TaskType = "dns"
	TaskTypeScan  TaskType = "scan"
	TaskTypeTrace TaskType = "trace"
)

// TaskConfig 任务配置文件结构
type TaskConfig struct {
	Tasks []Task `yaml:"tasks"`
}

// Task 单个任务定义
type Task struct {
	Name        string                 `yaml:"name"`
	Type        TaskType               `yaml:"type"`
	Targets     []string               `yaml:"targets"`
	Options     map[string]interface{} `yaml:"options,omitempty"`
	Schedule    string                 `yaml:"schedule,omitempty"`
	Enabled     bool                   `yaml:"enabled"`
	Concurrency int                    `yaml:"concurrency,omitempty"`
}

// TaskResult 任务执行结果
type TaskResult struct {
	TaskName  string
	TaskType  TaskType
	Success   bool
	Error     error
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Results   []interface{}
}

// BatchResult 批量任务结果
type BatchResult struct {
	TotalTasks    int
	SuccessTasks  int
	FailedTasks   int
	TotalDuration time.Duration
	TaskResults   []*TaskResult
}
