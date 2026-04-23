package controller

import (
	"net/http"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/constants"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/dto"
	httpInterfaces "cnb.cool/mliev/open/go-web/pkg/server/http_server/interfaces"
)

type BaseResponse struct{}

func (BaseResponse) Success(c httpInterfaces.RouterContextInterface, data any) {
	c.JSON(http.StatusOK, dto.Response{
		Code:    constants.ErrCodeSuccess,
		Message: constants.GetErrMessage(constants.ErrCodeSuccess),
		Data:    data,
	})
}

func (BaseResponse) SuccessWithMessage(c httpInterfaces.RouterContextInterface, message string, data any) {
	c.JSON(http.StatusOK, dto.Response{
		Code:    constants.ErrCodeSuccess,
		Message: message,
		Data:    data,
	})
}

func (r BaseResponse) Error(c httpInterfaces.RouterContextInterface, code int, message string) {
	httpStatus := r.getHTTPStatus(code)
	if message == "" {
		message = constants.GetErrMessage(code)
	}
	c.JSON(httpStatus, dto.Response{
		Code:    code,
		Message: message,
	})
}

func (r BaseResponse) ErrorWithData(c httpInterfaces.RouterContextInterface, code int, message string, data any) {
	httpStatus := r.getHTTPStatus(code)
	if message == "" {
		message = constants.GetErrMessage(code)
	}
	c.JSON(httpStatus, dto.Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

func (BaseResponse) getHTTPStatus(code int) int {
	if code >= 400 && code < 600 {
		return code
	}
	return http.StatusOK
}
