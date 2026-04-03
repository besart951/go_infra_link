package dashboard

import (
	dashboardsvc "github.com/besart951/go_infra_link/backend/internal/service/dashboard"
	"github.com/google/uuid"
)

type DashboardService interface {
	GetUserDashboard(userID uuid.UUID) (*dashboardsvc.DashboardResponse, error)
}
