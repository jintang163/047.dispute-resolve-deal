package response

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	TraceID string      `json:"traceId,omitempty"`
}

type PageResult struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

const (
	CodeSuccess      = 200
	CodeBadRequest   = 400
	CodeUnauthorized = 401
	CodeForbidden    = 403
	CodeNotFound     = 404
	CodeServerError  = 500
)

var (
	MsgSuccess      = "操作成功"
	MsgBadRequest   = "请求参数错误"
	MsgUnauthorized = "未授权或Token已过期"
	MsgForbidden    = "无权限访问"
	MsgNotFound     = "资源不存在"
	MsgServerError  = "服务器内部错误"
)

func Success(data interface{}) Response {
	return Response{
		Code:    CodeSuccess,
		Message: MsgSuccess,
		Data:    data,
	}
}

func SuccessWithMessage(data interface{}, message string) Response {
	return Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	}
}

func Fail(code int, message string) Response {
	return Response{
		Code:    code,
		Message: message,
	}
}

func BadRequest(message string) Response {
	if message == "" {
		message = MsgBadRequest
	}
	return Fail(CodeBadRequest, message)
}

func Unauthorized(message string) Response {
	if message == "" {
		message = MsgUnauthorized
	}
	return Fail(CodeUnauthorized, message)
}

func Forbidden(message string) Response {
	if message == "" {
		message = MsgForbidden
	}
	return Fail(CodeForbidden, message)
}

func NotFound(message string) Response {
	if message == "" {
		message = MsgNotFound
	}
	return Fail(CodeNotFound, message)
}

func ServerError(message string) Response {
	if message == "" {
		message = MsgServerError
	}
	return Fail(CodeServerError, message)
}

func Page(list interface{}, total int64, page, pageSize int) Response {
	return Success(PageResult{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func PageWithExtra(list interface{}, total int64, page, pageSize int, extra map[string]interface{}) Response {
	result := map[string]interface{}{
		"list":     list,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}
	for k, v := range extra {
		result[k] = v
	}
	return Success(result)
}
