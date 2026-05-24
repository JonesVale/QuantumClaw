package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/sessions"
	"github.com/quantumclaw/quantumclaw/model"
)

// MenuResponse is the API response shape for a single menu item.
type MenuResponse struct {
	ID        int    `json:"id"`
	MenuKey   string `json:"menu_key"`
	ParentKey string `json:"parent_key"`
	MenuType  string `json:"menu_type"`
	LabelKey  string `json:"label_key"`
	Icon      string `json:"icon"`
	Path      string `json:"path"`
	SortOrder int    `json:"sort_order"`
	Roles     string `json:"roles"`
	GroupName string `json:"group_name"`
	Enabled   bool   `json:"enabled"`
}

func toMenuResponse(m *model.MenuItem) MenuResponse {
	return MenuResponse{
		ID:        m.Id,
		MenuKey:   m.MenuKey,
		ParentKey: m.ParentKey,
		MenuType:  m.MenuType,
		LabelKey:  m.LabelKey,
		Icon:      m.Icon,
		Path:      m.Path,
		SortOrder: m.SortOrder,
		Roles:     m.Roles,
		GroupName: m.GroupName,
		Enabled:   m.Enabled,
	}
}

// GetMenus returns menus filtered by type and the current user's role.
// GET /api/menus?type=nav|sidebar
func GetMenus(c *gin.Context) {
	menuType := c.DefaultQuery("type", "sidebar")

	// Determine current user's role from session (default to guest)
	role := model.RoleGuestUser
	session := sessions.Default(c)
	if sessionID := session.Get("id"); sessionID != nil {
		if id, ok := sessionID.(int); ok && id > 0 {
			if r := session.Get("role"); r != nil {
				if rv, ok2 := r.(int); ok2 {
					role = rv
				}
			}
		}
	}

	menus, err := model.GetMenusByTypeAndRole(menuType, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	var resp []MenuResponse
	for _, m := range menus {
		resp = append(resp, toMenuResponse(m))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    resp,
	})
}

// AdminGetAllMenus returns all menu items (including disabled) for admin management.
// GET /api/admin/menus
func AdminGetAllMenus(c *gin.Context) {
	menus, err := model.AllMenusAdmin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	var resp []MenuResponse
	for _, m := range menus {
		resp = append(resp, toMenuResponse(m))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    resp,
	})
}

// AdminCreateOrUpdateMenu creates or updates a menu item.
// POST /api/admin/menus
func AdminCreateOrUpdateMenu(c *gin.Context) {
	var req struct {
		ID        int    `json:"id"`
		MenuKey   string `json:"menu_key"`
		ParentKey string `json:"parent_key"`
		MenuType  string `json:"menu_type"`
		LabelKey  string `json:"label_key"`
		Icon      string `json:"icon"`
		Path      string `json:"path"`
		SortOrder int    `json:"sort_order"`
		Roles     string `json:"roles"`
		GroupName string `json:"group_name"`
		Enabled   *bool  `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	// Validate required fields
	if req.MenuKey == "" || req.LabelKey == "" || req.Path == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "menu_key, label_key, and path are required",
		})
		return
	}

	if req.MenuType == "" {
		req.MenuType = "sidebar"
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	if req.Roles == "" {
		req.Roles = "[1]"
	}

	menu := &model.MenuItem{
		Id:        req.ID,
		MenuKey:   req.MenuKey,
		ParentKey: req.ParentKey,
		MenuType:  req.MenuType,
		LabelKey:  req.LabelKey,
		Icon:      req.Icon,
		Path:      req.Path,
		SortOrder: req.SortOrder,
		Roles:     req.Roles,
		GroupName: req.GroupName,
		Enabled:   enabled,
	}

	if err := model.CreateOrUpdateMenu(menu); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Menu saved successfully",
		"data":    toMenuResponse(menu),
	})
}

// AdminDeleteMenu deletes a menu item.
// DELETE /api/admin/menus/:id
func AdminDeleteMenu(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid menu ID",
		})
		return
	}

	if err := model.DeleteMenu(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Menu deleted successfully",
	})
}

// ensure json is used (Go compiler keeps it)
var _ = json.Marshal
