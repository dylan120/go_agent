package defaults

var (
	Success = 0
	Failure = 1
	TimeOut = 2003
	Run     = 1001
	Wait    = 1002
	Stop    = 1003
)

type Status struct {
	Code       int
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

func (status *Status) Set(code int, desc string, isFinished bool) {
	status.Code = code
	status.Desc = desc
	status.IsFinished = isFinished
}
