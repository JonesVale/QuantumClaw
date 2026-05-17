package model

import "errors"

var (
	ErrTokenNotProvided = errors.New("未提供令牌")
	ErrTokenInvalid     = errors.New("无效的令牌")
	ErrTokenExpired     = errors.New("该令牌已过期")
	ErrTokenExhausted   = errors.New("该令牌额度已用尽")
	ErrQuotaNotEnough   = errors.New("配额不足")
	ErrDatabase         = errors.New("数据库错误")
	ErrChannelNotFound  = errors.New("渠道不存在")
	ErrChannelDisabled  = errors.New("渠道已被禁用")
	ErrPermissionDenied = errors.New("权限不足")
)
