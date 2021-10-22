package cherryCode

type DataResult struct {
	Code    int32       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func NewDataResult(code int32) DataResult {
	result := DataResult{
		Code:    code,
		Message: GetMessage(code),
		Data:    []string{},
	}
	return result
}

func (p *DataResult) SetCode(code int32) {
	p.Code = code
	p.Message = GetMessage(code)
}
