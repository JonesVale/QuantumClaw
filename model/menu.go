package model

import (
	"fmt"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
)

// MenuItem represents a navigation/sidebar menu item with role-based access control.
type MenuItem struct {
	Id        int       `json:"id"`
	MenuKey   string    `json:"menu_key" gorm:"type:varchar(64);uniqueIndex;not null"`
	ParentKey string    `json:"parent_key" gorm:"type:varchar(64);default:'';index"`
	MenuType  string    `json:"menu_type" gorm:"type:varchar(20);not null;default:'sidebar';index"` // nav or sidebar
	LabelKey  string    `json:"label_key" gorm:"type:varchar(64);not null"`
	Icon      string    `json:"icon" gorm:"type:varchar(64);default:''"`
	Path      string    `json:"path" gorm:"type:varchar(256);not null"`
	SortOrder int       `json:"sort_order" gorm:"type:int;default:0"`
	Roles     string    `json:"roles" gorm:"type:text"`                        // JSON array string e.g. "[1,2,10,100]"
	GroupName string    `json:"group_name" gorm:"type:varchar(64);default:''"`                // Sidebar group title
	Enabled   bool      `json:"enabled" gorm:"type:boolean;default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AllMenus returns all enabled menu items ordered by sort_order.
func AllMenus() ([]*MenuItem, error) {
	var menus []*MenuItem
	err := DB.Where("enabled = ?", true).Order("sort_order asc").Find(&menus).Error
	return menus, err
}

// AllMenusAdmin returns all menu items including disabled ones for admin management.
func AllMenusAdmin() ([]*MenuItem, error) {
	var menus []*MenuItem
	err := DB.Order("sort_order asc").Find(&menus).Error
	return menus, err
}

// GetMenusByTypeAndRole returns menus filtered by type and user role.
func GetMenusByTypeAndRole(menuType string, role int) ([]*MenuItem, error) {
	allMenus, err := AllMenus()
	if err != nil {
		return nil, err
	}

	roleStr := roleToString(role)
	var filtered []*MenuItem
	for _, m := range allMenus {
		if m.MenuType != menuType {
			continue
		}
		if !m.Enabled {
			continue
		}
		if roleMatches(m.Roles, role, roleStr) {
			filtered = append(filtered, m)
		}
	}
	return filtered, nil
}

// roleMatches checks if the given role integer is in the Roles JSON array string.
func roleMatches(rolesJSON string, role int, roleStr string) bool {
	if rolesJSON == "" || rolesJSON == "[]" {
		return true
	}
	return containsRole(rolesJSON, roleStr)
}

// containsRole checks if a role value appears in the JSON array string.
func containsRole(rolesJSON, role string) bool {
	cleaned := rolesJSON
	if len(cleaned) > 0 && cleaned[0] == '[' {
		cleaned = cleaned[1:]
	}
	if len(cleaned) > 0 && cleaned[len(cleaned)-1] == ']' {
		cleaned = cleaned[:len(cleaned)-1]
	}
	parts := splitAndTrim(cleaned, ",")
	for _, p := range parts {
		if p == role {
			return true
		}
	}
	return false
}

func splitAndTrim(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if string(s[i]) == sep {
			part := trimSpace(s[start:i])
			if part != "" {
				result = append(result, part)
			}
			start = i + 1
		}
	}
	part := trimSpace(s[start:])
	if part != "" {
		result = append(result, part)
	}
	return result
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

func roleToString(role int) string {
	switch role {
	case RoleGuestUser:
		return "0"
	case RoleCommonUser:
		return "1"
	case RoleSupplier:
		return "2"
	case RoleAdminUser:
		return "10"
	case RoleRootUser:
		return "100"
	default:
		return "1"
	}
}

// CreateOrUpdateMenu creates or updates a menu item.
func CreateOrUpdateMenu(menu *MenuItem) error {
	if menu.Id > 0 {
		return DB.Model(menu).Updates(menu).Error
	}
	return DB.Create(menu).Error
}

// DeleteMenu deletes a menu item by ID.
func DeleteMenu(id int) error {
	return DB.Delete(&MenuItem{}, id).Error
}

// GetMenuByID returns a single menu item by ID.
func GetMenuByID(id int) (*MenuItem, error) {
	var menu MenuItem
	err := DB.First(&menu, id).Error
	return &menu, err
}

// SeedDefaultMenus always upserts all default menu items by MenuKey.
// Existing items are updated, new items are inserted, stale items are preserved.
func SeedDefaultMenus() error {
	// Complete list of all seed menu items.
	// GroupName "" (empty) = main sidebar (no collapsible label)
	// GroupName "management" = "Management" collapsible section
	// GroupName "account" = "Account" collapsible section
	seedMenus := []MenuItem{
		// ===== Nav items (top navigation bar) =====
		{MenuKey: "nav-models", ParentKey: "", MenuType: "nav", LabelKey: "Models", Icon: "LayoutDashboard", Path: "/models", SortOrder: 10, Roles: "[0,1,2,10,100]", GroupName: "", Enabled: true},
		{MenuKey: "nav-pricing", ParentKey: "", MenuType: "nav", LabelKey: "Pricing", Icon: "DollarSign", Path: "/pricing", SortOrder: 20, Roles: "[0,1,2,10,100]", GroupName: "", Enabled: true},
		{MenuKey: "nav-rankings", ParentKey: "", MenuType: "nav", LabelKey: "Rankings", Icon: "TrendingUp", Path: "/rankings", SortOrder: 30, Roles: "[0,1,2,10,100]", GroupName: "", Enabled: true},
		{MenuKey: "nav-apps", ParentKey: "", MenuType: "nav", LabelKey: "Apps", Icon: "Sparkles", Path: "/apps", SortOrder: 40, Roles: "[0,1,2,10,100]", GroupName: "", Enabled: true},
		{MenuKey: "nav-enterprise", ParentKey: "", MenuType: "nav", LabelKey: "Enterprise", Icon: "Building2", Path: "/enterprise", SortOrder: 50, Roles: "[0,1,2,10,100]", GroupName: "", Enabled: true},
		{MenuKey: "nav-dashboard", ParentKey: "", MenuType: "nav", LabelKey: "Dashboard", Icon: "LayoutDashboard", Path: "/dashboard", SortOrder: 5, Roles: "[1,2,10,100]", GroupName: "", Enabled: true},
		{MenuKey: "nav-news", ParentKey: "", MenuType: "nav", LabelKey: "AI News", Icon: "Newspaper", Path: "/news", SortOrder: 60, Roles: "[0,1,2,10,100]", GroupName: "", Enabled: true},
		{MenuKey: "nav-api-docs", ParentKey: "", MenuType: "nav", LabelKey: "API Docs", Icon: "BookOpen", Path: "/api-docs", SortOrder: 70, Roles: "[1,2,10,100]", GroupName: "", Enabled: true},

		// ===== Sidebar items (group: "" — main sidebar, no collapsible label) =====
		{MenuKey: "sidebar-dashboard", ParentKey: "", MenuType: "sidebar", LabelKey: "Dashboard", Icon: "LayoutDashboard", Path: "/dashboard", SortOrder: 10, Roles: "[1,2,10,100]", GroupName: "", Enabled: true},
		{MenuKey: "sidebar-chat", ParentKey: "", MenuType: "sidebar", LabelKey: "AI Chat", Icon: "MessageSquare", Path: "/chat", SortOrder: 20, Roles: "[0,1,2,10,100]", GroupName: "", Enabled: true},
		{MenuKey: "sidebar-models", ParentKey: "", MenuType: "sidebar", LabelKey: "Models", Icon: "Box", Path: "/models", SortOrder: 30, Roles: "[0,1]", GroupName: "", Enabled: true},
		{MenuKey: "sidebar-rankings", ParentKey: "", MenuType: "sidebar", LabelKey: "Rankings", Icon: "TrendingUp", Path: "/rankings", SortOrder: 40, Roles: "[0,1]", GroupName: "", Enabled: true},
		{MenuKey: "sidebar-pricing", ParentKey: "", MenuType: "sidebar", LabelKey: "Pricing", Icon: "DollarSign", Path: "/pricing", SortOrder: 50, Roles: "[0,1]", GroupName: "", Enabled: true},
		{MenuKey: "sidebar-quantum", ParentKey: "", MenuType: "sidebar", LabelKey: "Quantum", Icon: "Atom", Path: "/quantum", SortOrder: 55, Roles: "[0,1]", GroupName: "", Enabled: true},
		{MenuKey: "sidebar-fusion", ParentKey: "", MenuType: "sidebar", LabelKey: "Fusion", Icon: "GitCompare", Path: "/fusion", SortOrder: 57, Roles: "[0,1]", GroupName: "", Enabled: true},

		{MenuKey: "sidebar-apps", ParentKey: "", MenuType: "sidebar", LabelKey: "Apps", Icon: "Sparkles", Path: "/apps", SortOrder: 60, Roles: "[0,1]", GroupName: "", Enabled: true},
		{MenuKey: "sidebar-enterprise", ParentKey: "", MenuType: "sidebar", LabelKey: "Enterprise", Icon: "Building2", Path: "/enterprise", SortOrder: 70, Roles: "[0,1]", GroupName: "", Enabled: true},


// ===== Sidebar items (group: "management") =====
		{MenuKey: "sidebar-keys", ParentKey: "", MenuType: "sidebar", LabelKey: "API Keys", Icon: "Key", Path: "/keys", SortOrder: 10, Roles: "[1,2,10,100]", GroupName: "management", Enabled: true},
		{MenuKey: "sidebar-users", ParentKey: "", MenuType: "sidebar", LabelKey: "Users", Icon: "Users", Path: "/users", SortOrder: 20, Roles: "[10,100]", GroupName: "management", Enabled: true},
		{MenuKey: "sidebar-logs", ParentKey: "", MenuType: "sidebar", LabelKey: "Usage Logs", Icon: "ScrollText", Path: "/logs", SortOrder: 30, Roles: "[1,2,10,100]", GroupName: "management", Enabled: true},
		{MenuKey: "sidebar-redemption", ParentKey: "", MenuType: "sidebar", LabelKey: "Redemption Codes", Icon: "Ticket", Path: "/redemption", SortOrder: 40, Roles: "[10,100]", GroupName: "management", Enabled: true},
		{MenuKey: "sidebar-distributors", ParentKey: "", MenuType: "sidebar", LabelKey: "Distributors", Icon: "Truck", Path: "/distributors", SortOrder: 50, Roles: "[10,100]", GroupName: "management", Enabled: true},
		{MenuKey: "sidebar-admin-tools", ParentKey: "", MenuType: "sidebar", LabelKey: "Admin Tools", Icon: "Wrench", Path: "/admin-tools", SortOrder: 60, Roles: "[10,100]", GroupName: "management", Enabled: true},
		{MenuKey: "sidebar-monitoring", ParentKey: "", MenuType: "sidebar", LabelKey: "Monitoring", Icon: "Activity", Path: "/monitoring", SortOrder: 70, Roles: "[1,2,10,100]", GroupName: "management", Enabled: true},
		{MenuKey: "sidebar-profit", ParentKey: "", MenuType: "sidebar", LabelKey: "Channel Profit", Icon: "TrendingUp", Path: "/profit", SortOrder: 80, Roles: "[10,100]", GroupName: "management", Enabled: true},
		{MenuKey: "sidebar-news", ParentKey: "", MenuType: "sidebar", LabelKey: "AI News", Icon: "Newspaper", Path: "/news", SortOrder: 90, Roles: "[1,2,10,100]", GroupName: "management", Enabled: true},
		{MenuKey: "sidebar-channels", ParentKey: "", MenuType: "sidebar", LabelKey: "Channels", Icon: "Network", Path: "/channels", SortOrder: 95, Roles: "[10,100]", GroupName: "management", Enabled: true},
		{MenuKey: "sidebar-reseller-admin", ParentKey: "", MenuType: "sidebar", LabelKey: "Reseller Management", Icon: "Store", Path: "/reseller-admin", SortOrder: 100, Roles: "[10,100]", GroupName: "management", Enabled: true},
		{MenuKey: "sidebar-settlement", ParentKey: "", MenuType: "sidebar", LabelKey: "Settlement Config", Icon: "Percent", Path: "/settlement", SortOrder: 110, Roles: "[10,100]", GroupName: "management", Enabled: true},
		{MenuKey: "sidebar-transactions", ParentKey: "", MenuType: "sidebar", LabelKey: "Transactions", Icon: "Receipt", Path: "/transactions", SortOrder: 120, Roles: "[10,100]", GroupName: "management", Enabled: true},
		{MenuKey: "sidebar-platform-settings", ParentKey: "", MenuType: "sidebar", LabelKey: "Platform Settings", Icon: "Settings", Path: "/platform-settings", SortOrder: 130, Roles: "[10,100]", GroupName: "management", Enabled: true},
		{MenuKey: "sidebar-promo-ads", ParentKey: "", MenuType: "sidebar", LabelKey: "Promo Ads", Icon: "Megaphone", Path: "/promo-ads", SortOrder: 135, Roles: "[10,100]", GroupName: "management", Enabled: true},
		{MenuKey: "sidebar-reseller", ParentKey: "", MenuType: "sidebar", LabelKey: "Reseller Portal", Icon: "Store", Path: "/reseller", SortOrder: 140, Roles: "[1,2,10,100]", GroupName: "management", Enabled: true},
		{MenuKey: "sidebar-reseller-keys", ParentKey: "", MenuType: "sidebar", LabelKey: "My Keys", Icon: "Key", Path: "/reseller-keys", SortOrder: 150, Roles: "[1,2,10,100]", GroupName: "management", Enabled: true},
		{MenuKey: "sidebar-team", ParentKey: "", MenuType: "sidebar", LabelKey: "My Team", Icon: "Users", Path: "/team", SortOrder: 155, Roles: "[1,2,10,100]", GroupName: "management", Enabled: true},
		{MenuKey: "sidebar-menu-permissions", ParentKey: "", MenuType: "sidebar", LabelKey: "Menu Permissions", Icon: "Settings", Path: "/menu-permissions", SortOrder: 160, Roles: "[10,100]", GroupName: "management", Enabled: true},

		// ===== Sidebar items (group: "account") =====
		{MenuKey: "sidebar-profile", ParentKey: "", MenuType: "sidebar", LabelKey: "Profile", Icon: "User", Path: "/profile", SortOrder: 10, Roles: "[1,2,10,100]", GroupName: "account", Enabled: true},
		{MenuKey: "sidebar-wallet", ParentKey: "", MenuType: "sidebar", LabelKey: "Wallet", Icon: "Wallet", Path: "/wallet", SortOrder: 20, Roles: "[1,2,10,100]", GroupName: "account", Enabled: true},
		{MenuKey: "sidebar-billing", ParentKey: "", MenuType: "sidebar", LabelKey: "Billing", Icon: "DollarSign", Path: "/billing", SortOrder: 30, Roles: "[1,2,10,100]", GroupName: "account", Enabled: true},
		{MenuKey: "sidebar-checkin", ParentKey: "", MenuType: "sidebar", LabelKey: "Daily Check-in", Icon: "Gift", Path: "/checkin", SortOrder: 40, Roles: "[1,2,10,100]", GroupName: "account", Enabled: true},
		{MenuKey: "sidebar-subscription", ParentKey: "", MenuType: "sidebar", LabelKey: "Subscriptions", Icon: "CreditCard", Path: "/subscription", SortOrder: 50, Roles: "[1,2,10,100]", GroupName: "account", Enabled: true},
		{MenuKey: "sidebar-tasks", ParentKey: "", MenuType: "sidebar", LabelKey: "Task Logs", Icon: "ClipboardList", Path: "/tasks", SortOrder: 60, Roles: "[10,100]", GroupName: "account", Enabled: true},
		{MenuKey: "sidebar-settings", ParentKey: "", MenuType: "sidebar", LabelKey: "Settings", Icon: "Settings", Path: "/settings", SortOrder: 70, Roles: "[10,100]", GroupName: "account", Enabled: true},
		{MenuKey: "sidebar-api-docs", ParentKey: "", MenuType: "sidebar", LabelKey: "API Docs", Icon: "BookOpen", Path: "/api-docs", SortOrder: 80, Roles: "[1,2,10,100]", GroupName: "account", Enabled: true},
		{MenuKey: "sidebar-about", ParentKey: "", MenuType: "sidebar", LabelKey: "About", Icon: "Info", Path: "/about", SortOrder: 90, Roles: "[0,1,2,10,100]", GroupName: "account", Enabled: true},
		{MenuKey: "sidebar-connections", ParentKey: "", MenuType: "sidebar", LabelKey: "OAuth Connections", Icon: "Link2", Path: "/connections", SortOrder: 92, Roles: "[1,2,10,100]", GroupName: "account", Enabled: true},
		{MenuKey: "sidebar-notifications", ParentKey: "", MenuType: "sidebar", LabelKey: "Notifications", Icon: "Bell", Path: "/notifications", SortOrder: 94, Roles: "[1,2,10,100]", GroupName: "account", Enabled: true},
		{MenuKey: "sidebar-password", ParentKey: "", MenuType: "sidebar", LabelKey: "Password & Security", Icon: "Lock", Path: "/password", SortOrder: 96, Roles: "[1,2,10,100]", GroupName: "account", Enabled: true},
	}

	inserted, updated := 0, 0
	for _, m := range seedMenus {
		var existing MenuItem
		result := DB.Where("menu_key = ?", m.MenuKey).First(&existing)
		if result.Error != nil {
			// Not found — INSERT
			m.Id = 0 // ensure zero ID for INSERT
			if err := DB.Create(&m).Error; err != nil {
				logger.SysError("[SeedDefaultMenus] insert failed for " + m.MenuKey + ": " + err.Error())
			} else {
				inserted++
			}
		} else {
			// Found — UPDATE all fields
			if err := DB.Model(&existing).Updates(map[string]interface{}{
				"parent_key": m.ParentKey,
				"menu_type":  m.MenuType,
				"label_key":  m.LabelKey,
				"icon":       m.Icon,
				"path":       m.Path,
				"sort_order": m.SortOrder,
				"roles":      m.Roles,
				"group_name": m.GroupName,
				"enabled":    m.Enabled,
			}).Error; err != nil {
				logger.SysError("[SeedDefaultMenus] update failed for " + m.MenuKey + ": " + err.Error())
			} else {
				updated++
			}
		}
	}
	// Delete orphaned seed-prefix items (in DB but not in seed list)
	// Only cleans up sidebar-* and nav-* items, leaves admin-created menus intact
	seedKeys := make([]string, 0, len(seedMenus))
	for _, m := range seedMenus {
		seedKeys = append(seedKeys, m.MenuKey)
	}
	if err := DB.Where("menu_key NOT IN ? AND (menu_key LIKE 'sidebar-%' OR menu_key LIKE 'nav-%')", seedKeys).Delete(&MenuItem{}).Error; err != nil {
		logger.SysError("[SeedDefaultMenus] delete orphans: " + err.Error())
	} else {
		logger.SysLog("[SeedDefaultMenus] " + fmt.Sprintf("%d/%d", inserted, len(seedMenus)-updated) + " inserted, " + fmt.Sprintf("%d/%d", updated, len(seedMenus)-inserted) + " updated, orphans cleaned")
	}
	return nil
}

