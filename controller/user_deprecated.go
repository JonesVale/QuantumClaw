package controller

// ============================================================
// ⚠️ DEPRECATED — 此文件已拆分，保留仅为向后兼容
//
// 原文件（1299行）已按功能域拆分为以下文件：
//   user_common.go   — 共享辅助函数 (SetupLogin, ValidatePasswordStrength, getAdminOrgFilter)
//   user_auth.go      — 登录/注册/登出 (Login, Register, Logout)
//   user_admin.go     — 管理员CRUD + 启禁用/升降级 (GetAllUsers, ManageUser, CreateUser, ...)
//   user_self.go      — 用户自助操作 (GetSelf, UpdateSelf, DeleteSelf)
//   user_dashboard.go — 统计面板 (GetUserDashboard, buildModelProviderMap)
//   user_token.go     — Token/推广码 (GenerateAccessToken, GetAffCode)
//   user_profile.go   — 邮箱绑定/升级渠道商/团队 (EmailBind, UpgradeToProvider, GetMyTeam)
//   user_billing.go   — 充值 (TopUp, AdminTopUp)
//
// TODO: 确认路由注册无需修改后，可安全删除此文件。
// ============================================================
