package schedule

type PipelineInterface interface {
	Invoke()
	Next(context *PipelineContext)
}

//PipelineStatus 表示当前 pipeline 的运行状态
type PipelineStatus int

const (
	//SuccessStatus 表示 当前task 成功运行的状态
	SuccessStatus PipelineStatus = 1
	//FailStatus 表示当前 task 运行失败的状态
	FailStatus PipelineStatus = 2
)

//Phase 描述 当前task 的生命周期
type Phase string

//Phase 运行顺序为 WaitingPhase ---> CreatingPhase -----> RunningPhase -----> TerminationPhase
//               |------------------------------------> StoppedPhase 停止状态 可以将所有的状态停止
const (
	//RunningPhase 当前task 正在运行
	RunningPhase Phase = "Running"
	//TerminationPhase 销毁当前 task
	TerminationPhase Phase = "Termination"
	//CreatingPhase 创建
	CreatingPhase Phase = "Creating"
	//WaitingPhase 等待创建
	WaitingPhase Phase = "Waiting"
	//StoppedPhase 暂停
	StoppedPhase Phase = "Stopped"
)

type Pipeline struct {
}

type PipelineContext struct {
	Context []*PipelineInterface `json:"context"`
	Status  PipelineStatus       `json:"status"`
	Phase   Phase                `json:"phase"`
	Reason  string               `json:"reason"`
}

type Task struct {
	Phase `json:"_"`
}

func NewPipeline() PipelineInterface {
	return &Pipeline{}
}
