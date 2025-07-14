package monitor

// Result 包含了监控检查的结果
type Result struct {
	Success bool   // 是否成功
	Message string // 附带信息
	Err     error  // 错误信息
}

// Monitor 定义了所有监控器的通用接口
type Monitor interface {
	// Check 执行一次监控检查
	Check() Result
	// Name 返回监控器的名称
	Name() string
}
