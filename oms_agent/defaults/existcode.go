package defaults

var (
	Success int32 = 0
	Failure int32 = 1
	TimeOut int32 = 2003
	Run     int32 = 1001
	Wait    int32 = 1002
	Stop    int32 = 1003
)

type Status struct {
	Code       int32
	Desc       string
	IsFinished bool
}

func NewStatus() *Status {
	return &Status{
		Desc:       "",
		IsFinished: false,
		Code:       Run,
	}
}

func (status *Status) Set(code int32, desc string, isFinished bool) {
	status.Code = code
	status.Desc = desc
	status.IsFinished = isFinished
}
