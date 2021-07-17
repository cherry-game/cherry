package cherryResult

type Result struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func New(code ...int) *Result {
	c := 0
	if len(code) > 0 {
		c = code[0]
	}

	return &Result{
		Code:    c,
		Message: "",
		Data:    []string{},
	}
}

func NewResult(code int, message string) *Result {
	return &Result{
		Code:    code,
		Message: message,
		Data:    []string{},
	}
}
