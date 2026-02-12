package facility

import (
	"net/http"
	"os"

	domainExport "github.com/besart951/go_infra_link/backend/internal/domain/exporting"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type ExportHandler struct {
	service ExportService
}

func NewExportHandler(service ExportService) *ExportHandler {
	return &ExportHandler{service: service}
}

func (h *ExportHandler) CreateFieldDeviceExport(c *gin.Context) {
	var req dto.CreateFieldDeviceExportRequest
	if !bindJSON(c, &req) {
		return
	}

	job, err := h.service.Create(c.Request.Context(), domainExport.Request{
		ProjectIDs:        req.ProjectIDs,
		BuildingIDs:       req.BuildingIDs,
		ControlCabinetIDs: req.ControlCabinetIDs,
		SPSControllerIDs:  req.SPSControllerIDs,
		ForceAsync:        req.ForceAsync,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "export_creation_failed", err.Error())
		return
	}

	c.JSON(http.StatusAccepted, toExportJobResponse(job))
}

func (h *ExportHandler) GetExportStatus(c *gin.Context) {
	jobID, ok := parseUUIDParam(c, "jobId")
	if !ok {
		return
	}

	job, err := h.service.Get(c.Request.Context(), jobID)
	if err != nil {
		respondNotFound(c, "Export job not found")
		return
	}

	c.JSON(http.StatusOK, toExportJobResponse(job))
}

func (h *ExportHandler) DownloadExport(c *gin.Context) {
	jobID, ok := parseUUIDParam(c, "jobId")
	if !ok {
		return
	}

	job, err := h.service.Get(c.Request.Context(), jobID)
	if err != nil {
		respondNotFound(c, "Export job not found")
		return
	}

	if job.Status != domainExport.StatusCompleted {
		respondError(c, http.StatusConflict, "export_not_ready", "Export is not ready for download")
		return
	}

	if _, statErr := os.Stat(job.FilePath); statErr != nil {
		respondError(c, http.StatusNotFound, "export_file_missing", "Export file does not exist")
		return
	}

	c.Header("Content-Type", job.ContentType)
	c.Header("Content-Disposition", "attachment; filename=\""+job.FileName+"\"")
	c.File(job.FilePath)
}

func toExportJobResponse(job domainExport.Job) dto.FieldDeviceExportJobResponse {
	res := dto.FieldDeviceExportJobResponse{
		JobID:       job.ID,
		Status:      string(job.Status),
		Progress:    job.Progress,
		Message:     job.Message,
		OutputType:  string(job.OutputType),
		FileName:    job.FileName,
		ContentType: job.ContentType,
		Error:       job.Error,
		CreatedAt:   job.CreatedAt,
		UpdatedAt:   job.UpdatedAt,
	}
	if job.Status == domainExport.StatusCompleted {
		res.DownloadURL = "/api/v1/facility/exports/jobs/" + job.ID.String() + "/download"
	}
	return res
}
