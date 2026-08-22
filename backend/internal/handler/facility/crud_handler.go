package facility

import (
	"context"
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// crudSvc is the minimal service interface required by crudHandler.
// Any entity service that embeds baseService satisfies this automatically.
type crudSvc[T any] interface {
	Create(context.Context, *T) error
	GetByID(context.Context, uuid.UUID) (*T, error)
	List(context.Context, int, int, string) (*domain.PaginatedList[T], error)
	Update(context.Context, *T) error
	DeleteAtVersion(context.Context, uuid.UUID, uint64) error
}

type versionedUpdateRequest interface {
	ExpectedVersion() uint64
}

// crudHandler holds the generic logic for Create, GetByID, List, Update, Delete.
// Compose it into entity-specific handlers to eliminate repeated boilerplate.
type crudHandler[T, CreateReq any, UpdateReq versionedUpdateRequest] struct {
	svc          crudSvc[T]
	fromCreate   func(CreateReq) *T
	applyUpdate  func(*T, UpdateReq)
	toResp       func(T) any
	toListResp   func(*domain.PaginatedList[T]) any
	resourceKind string
	notFoundKey  string
}

func newCRUD[T, CreateReq any, UpdateReq versionedUpdateRequest](
	svc crudSvc[T],
	fromCreate func(CreateReq) *T,
	applyUpdate func(*T, UpdateReq),
	toResp func(T) any,
	toListResp func(*domain.PaginatedList[T]) any,
	resourceKind string,
	notFoundKey string,
) crudHandler[T, CreateReq, UpdateReq] {
	return crudHandler[T, CreateReq, UpdateReq]{
		svc:          svc,
		fromCreate:   fromCreate,
		applyUpdate:  applyUpdate,
		toResp:       toResp,
		toListResp:   toListResp,
		resourceKind: resourceKind,
		notFoundKey:  notFoundKey,
	}
}

// respFn adapts a typed response function to func(T) any.
func respFn[T, R any](fn func(T) R) func(T) any {
	return func(t T) any { return fn(t) }
}

// listRespFn adapts a typed list-response function to func(*domain.PaginatedList[T]) any.
func listRespFn[T, R any](fn func(*domain.PaginatedList[T]) R) func(*domain.PaginatedList[T]) any {
	return func(l *domain.PaginatedList[T]) any { return fn(l) }
}

func (h *crudHandler[T, CreateReq, UpdateReq]) handleCreate(c *gin.Context) {
	var req CreateReq
	if !bindJSON(c, &req) {
		return
	}
	ctx := c.Request.Context()
	item := h.fromCreate(req)
	if err := h.svc.Create(ctx, item); respondLocalizedValidationOrError(c, err, "facility.creation_failed") {
		return
	}
	c.JSON(http.StatusCreated, h.toResp(*item))
}

func (h *crudHandler[T, CreateReq, UpdateReq]) handleGetByID(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	ctx := c.Request.Context()
	item, err := h.svc.GetByID(ctx, id)
	if err != nil {
		respondLocalizedDomainError(c, err, "fetch_failed", "facility.fetch_failed",
			localizedNotFound(h.notFoundKey),
		)
		return
	}
	c.JSON(http.StatusOK, h.toResp(*item))
}

func (h *crudHandler[T, CreateReq, UpdateReq]) handleList(c *gin.Context) {
	query, ok := parsePaginationQuery(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()
	result, err := h.svc.List(ctx, query.Page, query.Limit, query.Search)
	if err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}
	c.JSON(http.StatusOK, h.toListResp(result))
}

func (h *crudHandler[T, CreateReq, UpdateReq]) handleUpdate(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req UpdateReq
	if !bindJSON(c, &req) {
		return
	}
	ctx := c.Request.Context()
	item, err := h.svc.GetByID(ctx, id)
	if err != nil {
		respondLocalizedDomainError(c, err, "fetch_failed", "facility.fetch_failed",
			localizedNotFound(h.notFoundKey),
		)
		return
	}
	h.applyUpdate(item, req)
	base, ok := any(item).(interface{ GetBase() *domain.Base })
	if !ok {
		respondInvalidArgument(c, "aggregate version is unavailable")
		return
	}
	base.GetBase().Version = req.ExpectedVersion()
	if err := h.svc.Update(ctx, item); err != nil {
		if h.respondCurrentConflict(c, id, req.ExpectedVersion(), err) {
			return
		}
		respondLocalizedValidationOrError(c, err, "facility.update_failed")
		return
	}
	c.JSON(http.StatusOK, h.toResp(*item))
}

func (h *crudHandler[T, CreateReq, UpdateReq]) handleDelete(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var query requiredBaseVersionQuery
	if !bindQuery(c, &query) {
		return
	}
	ctx := c.Request.Context()
	if err := h.svc.DeleteAtVersion(ctx, id, query.BaseVersion); err != nil {
		if h.respondCurrentConflict(c, id, query.BaseVersion, err) {
			return
		}
		respondLocalizedDomainError(c, err, "deletion_failed", "facility.deletion_failed",
			localizedNotFound(h.notFoundKey),
			localizedReferenceInUse(),
			localizedBacnetReferenceInUse(),
		)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *crudHandler[T, CreateReq, UpdateReq]) respondCurrentConflict(c *gin.Context, id uuid.UUID, expected uint64, err error) bool {
	if !errors.Is(err, domain.ErrConflict) {
		return false
	}
	current, getErr := h.svc.GetByID(c.Request.Context(), id)
	if getErr != nil {
		return false
	}
	base, ok := any(current).(interface{ GetBase() *domain.Base })
	if !ok {
		return false
	}
	handlerutil.RespondWriteConflict(c, h.resourceKind, id.String(), expected, base.GetBase().Version, nil, h.toResp(*current))
	return true
}

type requiredBaseVersionQuery struct {
	BaseVersion uint64 `form:"base_version" binding:"required,min=1"`
}
