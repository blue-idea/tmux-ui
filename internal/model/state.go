package model

// State 表示智能体当前运行状态
type State string

const (
	StateWorking      State = "working"       // 智能体正在执行或生成内容
	StateIdle         State = "idle"          // 处于空闲状态，无活跃命令或纯 shell
	StateWaitingInput State = "waiting_input" // 等待用户确认授权或交互提问
	StateError        State = "error"         // 执行出错或进程异常退出
	StateExited       State = "exited"        // Pane 已销毁或进程彻底退出
)

// IsValid 校验状态枚举是否合法
func (s State) IsValid() bool {
	switch s {
	case StateWorking, StateIdle, StateWaitingInput, StateError, StateExited:
		return true
	default:
		return false
	}
}

