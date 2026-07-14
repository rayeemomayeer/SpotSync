package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/dto"
	appmw "github.com/rayeemomayeer/SpotSync/internal/middleware"
	"github.com/rayeemomayeer/SpotSync/internal/models"
)

type ZoneService interface {
	Create(ctx context.Context, req dto.CreateZoneRequest, orgID *uint) (dto.ZoneResponse, error)
	List(ctx context.Context, q dto.ZoneListQuery) ([]dto.ZoneResponse, error)
	GetByID(ctx context.Context, id uint) (dto.ZoneResponse, error)
	Update(ctx context.Context, id uint, req dto.UpdateZoneRequest) (dto.ZoneResponse, error)
	Delete(ctx context.Context, id uint) error
}

type OrgResolver interface {
	PrimaryOrgID(ctx context.Context, userID uint) (*uint, error)
	EnsureOrgEntitled(ctx context.Context, orgID uint) error
}

type ZoneHandler struct {
	zones ZoneService
	orgs  OrgResolver
}

func NewZoneHandler(zones ZoneService, orgs OrgResolver) *ZoneHandler {
	return &ZoneHandler{zones: zones, orgs: orgs}
}

func (h *ZoneHandler) Create(c echo.Context) error {
	var req dto.CreateZoneRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}

	var orgID *uint
	userID, _ := appmw.UserID(c)
	role := appmw.Role(c)
	if role == models.RoleOrgAdmin && h.orgs != nil {
		id, err := h.orgs.PrimaryOrgID(c.Request().Context(), userID)
		if err != nil {
			return err
		}
		if id == nil {
			return domain.NewValidationError("Validation failed", map[string]string{
				"organization": "Org admin must belong to an organization",
			})
		}
		if err := h.orgs.EnsureOrgEntitled(c.Request().Context(), *id); err != nil {
			return err
		}
		orgID = id
	}

	if appmw.IsDemoMode(c) {
		if sid := strings.TrimSpace(appmw.DemoSessionID(c)); sid != "" {
			req.DemoSessionID = &sid
		}
	}

	zone, err := h.zones.Create(c.Request().Context(), req, orgID)
	if err != nil {
		return err
	}

	return JSONSuccess(c, http.StatusCreated, "Zone created successfully", zone)
}

func (h *ZoneHandler) List(c echo.Context) error {
	var q dto.ZoneListQuery
	if err := BindAndValidate(c, &q); err != nil {
		return err
	}
	q.DemoMode = appmw.IsDemoMode(c)
	q.DemoSessionID = appmw.DemoSessionID(c)

	zones, err := h.zones.List(c.Request().Context(), q)
	if err != nil {
		return err
	}
	if zones == nil {
		zones = []dto.ZoneResponse{}
	}
	return JSONSuccess(c, http.StatusOK, "Zones retrieved successfully", zones)
}

func (h *ZoneHandler) GetByID(c echo.Context) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}

	zone, err := h.zones.GetByID(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return JSONSuccess(c, http.StatusOK, "Zone retrieved successfully", zone)
}

func (h *ZoneHandler) requireOrgEntitledForZone(c echo.Context, zoneID uint) error {
	role := appmw.Role(c)
	if role != models.RoleOrgAdmin || h.orgs == nil {
		return nil
	}
	userID, _ := appmw.UserID(c)
	orgID, err := h.orgs.PrimaryOrgID(c.Request().Context(), userID)
	if err != nil {
		return err
	}
	if orgID == nil {
		return domain.ErrForbidden
	}
	if err := h.orgs.EnsureOrgEntitled(c.Request().Context(), *orgID); err != nil {
		return err
	}
	zone, err := h.zones.GetByID(c.Request().Context(), zoneID)
	if err != nil {
		return err
	}
	if zone.OrganizationID == nil || *zone.OrganizationID != *orgID {
		return domain.ErrForbidden
	}
	return nil
}

func (h *ZoneHandler) Update(c echo.Context) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}

	if err := h.requireOrgEntitledForZone(c, id); err != nil {
		return err
	}

	var req dto.UpdateZoneRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}

	zone, err := h.zones.Update(c.Request().Context(), id, req)
	if err != nil {
		return err
	}

	return JSONSuccess(c, http.StatusOK, "Zone updated successfully", zone)
}

func (h *ZoneHandler) Delete(c echo.Context) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}

	if err := h.requireOrgEntitledForZone(c, id); err != nil {
		return err
	}

	if err := h.zones.Delete(c.Request().Context(), id); err != nil {
		return err
	}

	return JSONSuccess(c, http.StatusOK, "Zone deleted successfully", nil)
}
