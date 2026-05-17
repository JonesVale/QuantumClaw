package system_setting

import "github.com/quantumclaw/quantumclaw/common/logger"

var (
	OidcEnabled      = false
	OidcIssuer       = ""
	OidcClientId     = ""
	OidcClientSecret = ""
	OidcScopes       = "openid profile email"
)

func InitOidcSettings() {
	logger.SysLog("OIDC settings initialized")
}
