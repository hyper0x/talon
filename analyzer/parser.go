package analyzer

import (
	base "hypermind.cn/talon/base"
	"net/http"
)

// 被用于解析HTTP响应的函数类型。
type ParseResponse func(httpResp *http.Response, respDepth uint32) ([]base.Data, []error)
