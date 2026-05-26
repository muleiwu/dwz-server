package middleware

import (
	"strconv"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/constants"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/model"
	"cnb.cool/mliev/dwz/dwz-server/v2/app/service"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/helper"
	httpInterfaces "cnb.cool/mliev/open/go-web/pkg/server/http_server/interfaces"
)

const (
	HeaderWorkspaceID = "X-Workspace-Id"

	contextWorkspaceID   = "current_workspace_id"
	contextWorkspaceRole = "current_workspace_role"
)

func WorkspaceMiddleware() httpInterfaces.HandlerFunc {
	return func(c httpInterfaces.RouterContextInterface) {
		user := GetCurrentUser(c)
		if user == nil {
			respondForbidden(c, "用户未登录")
			return
		}

		var requestedID uint64
		if headerValue := c.GetHeader(HeaderWorkspaceID); headerValue != "" {
			parsedID, err := strconv.ParseUint(headerValue, 10, 64)
			if err != nil {
				respondForbidden(c, "无效的工作区")
				return
			}
			requestedID = parsedID
		}

		member, err := service.NewWorkspaceService(helper.GetHelper()).ResolveWorkspace(user.ID, requestedID)
		if err != nil || member == nil || member.WorkspaceID == 0 {
			respondForbidden(c, "无可用工作区")
			return
		}

		c.Set(contextWorkspaceID, member.WorkspaceID)
		c.Set(contextWorkspaceRole, member.Role)
		c.Next()
	}
}

func GetCurrentWorkspaceID(c httpInterfaces.RouterContextInterface) uint64 {
	if v := c.Get(contextWorkspaceID); v != nil {
		switch value := v.(type) {
		case uint64:
			return value
		case int:
			return uint64(value)
		case string:
			id, _ := strconv.ParseUint(value, 10, 64)
			return id
		}
	}
	return 1
}

func GetCurrentWorkspaceRole(c httpInterfaces.RouterContextInterface) string {
	if v := c.Get(contextWorkspaceRole); v != nil {
		if role, ok := v.(string); ok {
			return role
		}
	}
	return ""
}

func HasAnyWorkspaceRole(c httpInterfaces.RouterContextInterface, roles ...string) bool {
	current := GetCurrentWorkspaceRole(c)
	for _, role := range roles {
		if current == role {
			return true
		}
	}
	return false
}

func CanManageWorkspace(c httpInterfaces.RouterContextInterface) bool {
	return HasAnyWorkspaceRole(c, model.WorkspaceRoleOwner, model.WorkspaceRoleAdmin)
}

func CanManageBusinessResource(c httpInterfaces.RouterContextInterface) bool {
	return HasAnyWorkspaceRole(c, model.WorkspaceRoleOwner, model.WorkspaceRoleAdmin, model.WorkspaceRoleMember)
}

func CanManageAdminResource(c httpInterfaces.RouterContextInterface) bool {
	return HasAnyWorkspaceRole(c, model.WorkspaceRoleOwner, model.WorkspaceRoleAdmin)
}

func respondForbidden(c httpInterfaces.RouterContextInterface, message string) {
	c.AbortWithStatusJSON(constants.ErrCodeForbidden, map[string]any{
		"code":    constants.ErrCodeForbidden,
		"message": message,
	})
}
