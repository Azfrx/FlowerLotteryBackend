package common

import "errors"

type AppError struct {
	HTTPStatus int
	Code       int
	Message    string
}

func (e *AppError) Error() string { return e.Message }
func NewError(status, code int, message string) error {
	return &AppError{HTTPStatus: status, Code: code, Message: message}
}

var ErrInvalidCredentials = NewError(401, 11001, "用户ID或密码错误")
var ErrAccountDisabled = NewError(403, 11002, "账号已禁用")
var ErrUnauthorized = NewError(401, 11003, "登录状态已失效")
var ErrUserExists = NewError(409, 11004, "用户ID已存在")
var ErrCurrentPassword = NewError(400, 11005, "当前密码错误")
var ErrPasswordUnchanged = NewError(400, 11006, "新密码不能与当前密码相同")
var ErrNicknameRequired = NewError(400, 11007, "昵称不能为空")
var ErrUserIDInvalid = NewError(400, 11008, "用户ID需为3至64个字符")
var ErrCoinInsufficient = NewError(409, 12001, "金币余额不足")
var ErrExchangeOption = NewError(400, 12002, "兑换档位无效")
var ErrDuplicateRequest = NewError(409, 10003, "请勿重复提交")

func AsApp(err error) (*AppError, bool) {
	var target *AppError
	ok := errors.As(err, &target)
	return target, ok
}
