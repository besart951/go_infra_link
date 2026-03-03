package facility

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// crudSvc is the minimal service interface required by crudHandler.
// Any entity service that embeds baseService satisfies this automatically.
type crudSvc[T any] interface {
	Create(*T) error
	GetByID(uuid.UUID) (*T, error)
	List(int, int, string) (*domain.PaginatedList[T], error)
	Update(*T) error
	DeleteByID(uuid.UUID) error
}

// crudHandler holds the generic logic for Create, GetByID, List, Update, Delete.
// Compose it into entity-specific handlers to eliminate repeated boilerplate.
type crudHandler[T, CreateReq, UpdateReq any] struct {
	svc         crudSvc[T]
	fromCreate  func(CreateReq) *T
	applyUpdate func(*T, UpdateReq)
	toResp      func(T) any
	toListResp  func(*domain.PaginatedList[T]) any
	notFoundKey string
}

func newCRUD[T, CreateReq, UpdateReq any](
	svc crudSvc[T],
	fromCreate func(CreateReq) *T,
	applyUpdate func(*T, UpdateReq),
	toResp func(T) any,
	toListResp func(*domain.PaginatedList[T]) any,
	notFoundKey string,
) crudHandler[T, CreateReq, UpdateReq] {
	return crudHandler[T, CreateReq, UpdateReq]{
		svc:         svc,
		fromCreate:  fromCreate,
		applyUpdate: applyUpdate,
		toResp:      toResp,
		toListResp:  toListResp,
		notFoundKey: notFoundKey,
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
	item := h.fromCreate(req)
	if err := h.svc.Create(item); respondLocalizedValidationOrError(c, err, "facility.creation_failed") {
		return
	}
	c.JSON(http.StatusCreated, h.toResp(*item))
}

func (h *crudHandler[T, CreateReq, UpdateReq]) handleGetByID(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	item, err := h.svc.GetByID(id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, h.notFoundKey) {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}
	c.JSON(http.StatusOK, h.toResp(*item))
}

func (h *crudHandler[T, CreateReq, UpdateReq]) handleList(c *gin.Context) {
	query, ok := parsePaginationQuery(c)
	if !ok {
		return
	}
	result, err := h.svc.List(query.Page, query.Limit, query.Search)
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
	item, err := h.svc.GetByID(id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, h.notFoundKey) {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}
	h.applyUpdate(item, req)
	if err := h.svc.Update(item); respondLocalizedValidationOrError(c, err, "facility.update_failed") {
		return
	}
	c.JSON(http.StatusOK, h.toResp(*item))
}

func (h *crudHandler[T, CreateReq, UpdateReq]) handleDelete(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteByID(id); err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "deletion_failed", "facility.deletion_failed")
		return
	}
	c.Status(http.StatusNoContent)
}
