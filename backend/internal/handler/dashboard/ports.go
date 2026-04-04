package dashboard

import (
	"context"

	dashboardsvc "github.com/besart951/go_infra_link/backend/internal/service/dashboard"
	"github.com/google/uuid"
)

type DashboardService interface {
	GetUserDashboard(ctx context.Context, userID uuid.UUID) (*dashboardsvc.DashboardResponse, error)
}
