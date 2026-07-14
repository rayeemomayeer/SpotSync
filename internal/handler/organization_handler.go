package handler

import (
	"net/http"
	"strconv"
	"strings"

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
	q := strings.TrimSpace(c.QueryParam("q"))
	var list []models.Organization
	var err error
	if q != "" {
		list, err = h.svc.Search(c.Request().Context(), q)
	} else {
		list, err = h.svc.List(c.Request().Context())
	}
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

func (h *OrganizationHandler) Me(c echo.Context) error {
	actorID, ok := appmw.UserID(c)
	if !ok {
		return domain.ErrUnauthorized
	}
	if !appmw.IsOrgAdmin(c) && !appmw.IsPlatformAdmin(c) {
		return domain.ErrForbidden
	}
	orgID, err := h.svc.PrimaryOrgID(c.Request().Context(), actorID)
	if err != nil {
		return err
	}
	if orgID == nil {
		return domain.ErrNotFound
	}
	org, err := h.svc.Get(c.Request().Context(), *orgID)
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

func (h *OrganizationHandler) Approve(c echo.Context) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	actorID, ok := appmw.UserID(c)
	if !ok {
		return domain.ErrUnauthorized
	}
	org, err := h.svc.Approve(c.Request().Context(), actorID, id)
	if err != nil {
		return err
	}
	return JSONSuccess(c, http.StatusOK, "Organization approved", toOrgResponse(org))
}

func (h *OrganizationHandler) Reject(c echo.Context) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	actorID, ok := appmw.UserID(c)
	if !ok {
		return domain.ErrUnauthorized
	}
	org, err := h.svc.Reject(c.Request().Context(), actorID, id)
	if err != nil {
		return err
	}
	return JSONSuccess(c, http.StatusOK, "Organization rejected", toOrgResponse(org))
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

func (h *OrganizationHandler) ListMembers(c echo.Context) error {
	orgID, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	actorID, ok := appmw.UserID(c)
	if !ok {
		return domain.ErrUnauthorized
	}
	if !appmw.IsPlatformAdmin(c) {
		if err := h.svc.EnsureOrgAccess(c.Request().Context(), models.RoleOrgAdmin, actorID, orgID); err != nil {
			return err
		}
	}
	list, err := h.svc.ListMembers(c.Request().Context(), orgID)
	if err != nil {
		return err
	}
	out := make([]dto.OrgMemberResponse, 0, len(list))
	for _, m := range list {
		out = append(out, dto.OrgMemberResponse{
			UserID:    m.UserID,
			Email:     m.Email,
			Name:      m.Name,
			Role:      m.Role,
			CreatedAt: m.CreatedAt,
		})
	}
	return JSONSuccess(c, http.StatusOK, "Organization members retrieved", out)
}

func (h *OrganizationHandler) AssignMember(c echo.Context) error {
	orgID, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	var req dto.AssignOrgMemberRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}
	actorID, ok := appmw.UserID(c)
	if !ok {
		return domain.ErrUnauthorized
	}
	if !appmw.IsPlatformAdmin(c) {
		if err := h.svc.EnsureOrgAccess(c.Request().Context(), models.RoleOrgAdmin, actorID, orgID); err != nil {
			return err
		}
	}
	if err := h.svc.AssignOrgAdminByEmail(c.Request().Context(), actorID, orgID, req.Email); err != nil {
		return err
	}
	return JSONSuccess(c, http.StatusCreated, "Organization admin assigned", nil)
}

func (h *OrganizationHandler) RemoveMember(c echo.Context) error {
	orgID, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	userID, err := parseUintParam(c, "userId")
	if err != nil {
		return err
	}
	actorID, ok := appmw.UserID(c)
	if !ok {
		return domain.ErrUnauthorized
	}
	if !appmw.IsPlatformAdmin(c) {
		if err := h.svc.EnsureOrgAccess(c.Request().Context(), models.RoleOrgAdmin, actorID, orgID); err != nil {
			return err
		}
	}
	if err := h.svc.RemoveMember(c.Request().Context(), actorID, orgID, userID); err != nil {
		return err
	}
	return JSONSuccess(c, http.StatusOK, "Organization member removed", nil)
}

func (h *OrganizationHandler) SetPlan(c echo.Context) error {
	orgID, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	var req dto.SetOrgPlanRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}
	actorID, ok := appmw.UserID(c)
	if !ok {
		return domain.ErrUnauthorized
	}
	org, err := h.svc.SetBillingPlan(c.Request().Context(), actorID, orgID, req.Plan, req.StripeCustomerID)
	if err != nil {
		return err
	}
	return JSONSuccess(c, http.StatusOK, "Organization billing plan updated", toOrgResponse(org))
}

func toOrgResponse(org *models.Organization) dto.OrganizationResponse {
	return dto.OrganizationResponse{
		ID:               org.ID,
		Name:             org.Name,
		Slug:             org.Slug,
		Status:           org.Status,
		BillingPlan:      org.BillingPlan,
		StripeCustomerID: org.StripeCustomerID,
		CreatedAt:        org.CreatedAt,
		UpdatedAt:        org.UpdatedAt,
	}
}
