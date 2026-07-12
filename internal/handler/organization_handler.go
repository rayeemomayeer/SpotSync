package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/dto"
	appmw "github.com/rayeemomayeer/SpotSync/internal/middleware"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"github.com/rayeemomayeer/SpotSync/internal/service"
)

type OrganizationHandler struct {
	svc *service.OrganizationService
}

func NewOrganizationHandler(svc *service.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{svc: svc}
}

func (h *OrganizationHandler) Create(c echo.Context) error {
	var req dto.CreateOrganizationRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}
	actorID, ok := appmw.UserID(c)
	if !ok {
		return domain.ErrUnauthorized
	}
	org, err := h.svc.Create(c.Request().Context(), actorID, req.Name, req.Slug)
	if err != nil {
		return err
	}
	return JSONSuccess(c, http.StatusCreated, "Organization created", toOrgResponse(org))
}

func (h *OrganizationHandler) List(c echo.Context) error {
	list, err := h.svc.List(c.Request().Context())
	if err != nil {
		return err
	}
	out := make([]dto.OrganizationResponse, 0, len(list))
	for i := range list {
		out = append(out, toOrgResponse(&list[i]))
	}
	return JSONSuccess(c, http.StatusOK, "Organizations retrieved", out)
}

func (h *OrganizationHandler) Get(c echo.Context) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	org, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return JSONSuccess(c, http.StatusOK, "Organization retrieved", toOrgResponse(org))
}

func (h *OrganizationHandler) SetStatus(c echo.Context) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	var req dto.SuspendOrganizationRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}
	actorID, ok := appmw.UserID(c)
	if !ok {
		return domain.ErrUnauthorized
	}
	org, err := h.svc.SetStatus(c.Request().Context(), actorID, id, req.Status)
	if err != nil {
		return err
	}
	return JSONSuccess(c, http.StatusOK, "Organization updated", toOrgResponse(org))
}

func (h *OrganizationHandler) ListAudit(c echo.Context) error {
	var orgFilter *uint
	if raw := c.QueryParam("organization_id"); raw != "" {
		n, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return domain.NewValidationError("Validation failed", map[string]string{"organization_id": "Must be a positive integer"})
		}
		id := uint(n)
		orgFilter = &id
		if appmw.IsOrgAdmin(c) && !appmw.IsPlatformAdmin(c) {
			actorID, _ := appmw.UserID(c)
			if err := h.svc.EnsureOrgAccess(c.Request().Context(), models.RoleOrgAdmin, actorID, id); err != nil {
				return err
			}
		}
	} else if !appmw.IsPlatformAdmin(c) {
		return domain.ErrForbidden
	}
	list, err := h.svc.ListAudit(c.Request().Context(), orgFilter, 50)
	if err != nil {
		return err
	}
	out := make([]dto.AuditLogResponse, 0, len(list))
	for _, e := range list {
		out = append(out, dto.AuditLogResponse{
			ID:             e.ID,
			ActorUserID:    e.ActorUserID,
			OrganizationID: e.OrganizationID,
			Action:         e.Action,
			ResourceType:   e.ResourceType,
			ResourceID:     e.ResourceID,
			CreatedAt:      e.CreatedAt,
		})
	}
	return JSONSuccess(c, http.StatusOK, "Audit logs retrieved", out)
}

func toOrgResponse(org *models.Organization) dto.OrganizationResponse {
	return dto.OrganizationResponse{
		ID:        org.ID,
		Name:      org.Name,
		Slug:      org.Slug,
		Status:    org.Status,
		CreatedAt: org.CreatedAt,
		UpdatedAt: org.UpdatedAt,
	}
}
