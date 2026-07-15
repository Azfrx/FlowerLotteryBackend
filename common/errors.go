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

var ErrInvalidCredentials = NewError(401, 11001, "账号或密码错误")
var ErrAccountDisabled = NewError(403, 11002, "账号已禁用")
var ErrUnauthorized = NewError(401, 11003, "登录状态已失效")
var ErrAccountExists = NewError(409, 11004, "账号已存在")
var ErrCurrentPassword = NewError(400, 11005, "当前密码错误")
var ErrPasswordUnchanged = NewError(400, 11006, "新密码不能与当前密码相同")
var ErrNicknameRequired = NewError(400, 11007, "昵称不能为空")
var ErrAccountInvalid = NewError(400, 11008, "账号需为3至64位英文或数字")
var ErrAvatarRequired = NewError(400, 11009, "请选择头像文件")
var ErrAvatarTooLarge = NewError(413, 11010, "头像文件不能超过8MB")
var ErrAvatarType = NewError(400, 11011, "头像仅支持JPEG或PNG格式")
var ErrAvatarDecode = NewError(400, 11012, "头像文件无法识别或已损坏")
var ErrAvatarDimensions = NewError(400, 11013, "头像图片像素尺寸过大")
var ErrCoinInsufficient = NewError(409, 12001, "金币余额不足")
var ErrExchangeOption = NewError(400, 12002, "兑换档位无效")
var ErrPetalGiftPackPurchased = NewError(409, 12003, "花瓣特惠礼包每人限购一次")
var ErrExchangeOptionChanged = NewError(409, 12004, "兑换档位已更新，请重新打开购买花瓣页面")
var ErrActivityReadOnly = NewError(409, 13018, "活动已结束，当前仅支持查看")
var ErrActivityUnavailable = NewError(404, 13019, "暂无可查看的活动")
var ErrDuplicateRequest = NewError(409, 10003, "请勿重复提交")

func AsApp(err error) (*AppError, bool) {
	var target *AppError
	ok := errors.As(err, &target)
	return target, ok
}
