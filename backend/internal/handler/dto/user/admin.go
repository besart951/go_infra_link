package user

type AdminSetUserRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=superadmin admin_fzag fzag admin_planer planer admin_entrepreneur entrepreneur"`
}
