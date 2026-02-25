package diveinspectapi

import (
	"strconv"

	"github.com/Abraxas-365/divi/pkg/diveinspect"
	"github.com/Abraxas-365/divi/pkg/diveinspect/diveinspectsrv"
	"github.com/Abraxas-365/divi/pkg/errx"
	"github.com/gofiber/fiber/v2"
)

type Handlers struct {
	vehicleSvc    *diveinspectsrv.VehicleService
	enrichmentSvc *diveinspectsrv.EnrichmentService
	inspectionSvc *diveinspectsrv.InspectionService
	reportSvc     *diveinspectsrv.ReportService
}

func NewHandlers(
	vehicleSvc *diveinspectsrv.VehicleService,
	enrichmentSvc *diveinspectsrv.EnrichmentService,
	inspectionSvc *diveinspectsrv.InspectionService,
	reportSvc *diveinspectsrv.ReportService,
) *Handlers {
	return &Handlers{
		vehicleSvc:    vehicleSvc,
		enrichmentSvc: enrichmentSvc,
		inspectionSvc: inspectionSvc,
		reportSvc:     reportSvc,
	}
}

// RegisterRoutes registers all DiveInspect routes on the given router group
func (h *Handlers) RegisterRoutes(router fiber.Router) {
	vehicles := router.Group("/vehicles")

	// Vehicle CRUD
	vehicles.Post("/", h.CreateVehicle)
	vehicles.Get("/", h.ListVehicles)
	vehicles.Get("/:id", h.GetVehicle)
	vehicles.Patch("/:id", h.UpdateVehicle)
	vehicles.Delete("/:id", h.DeleteVehicle)

	// Enrichment
	vehicles.Post("/:id/enrich", h.EnrichVehicle)

	// Preview & Publish
	vehicles.Get("/:id/preview", h.GetVehiclePreview)
	vehicles.Post("/:id/publish", h.PublishVehicle)

	// Specs
	vehicles.Patch("/:id/specs", h.UpdateSpecs)

	// Photos & Inspection
	vehicles.Post("/:id/photos", h.UploadPhoto)
	vehicles.Post("/:id/inspect", h.RunInspection)

	// Report
	vehicles.Get("/:id/report.pdf", h.GetReport)

	// Listing JSON
	vehicles.Get("/:id/listing.json", h.GetListingJSON)

	// Inspection findings
	findings := router.Group("/findings")
	findings.Patch("/:fid", h.UpdateFinding)
}

// ============================================================================
// Vehicle Handlers
// ============================================================================

type createVehicleRequest struct {
	Plate         *string  `json:"plate"`
	Brand         string   `json:"brand"`
	Model         string   `json:"model"`
	Version       *string  `json:"version"`
	Trim          *string  `json:"trim"`
	Year          int      `json:"year"`
	MileageKM     int      `json:"mileage_km"`
	ColorExterior *string  `json:"color_exterior"`
	ColorInterior *string  `json:"color_interior"`
	PriceUSD      *float64 `json:"price_usd"`
	Branch        *string  `json:"branch"`
	Origin        *string  `json:"origin"`
}

func (h *Handlers) CreateVehicle(c *fiber.Ctx) error {
	var req createVehicleRequest
	if err := c.BodyParser(&req); err != nil {
		return errx.Validation("Invalid request body")
	}

	vehicle := &diveinspect.Vehicle{
		Plate:         req.Plate,
		Brand:         req.Brand,
		Model:         req.Model,
		Version:       req.Version,
		Trim:          req.Trim,
		Year:          req.Year,
		MileageKM:     req.MileageKM,
		ColorExterior: req.ColorExterior,
		ColorInterior: req.ColorInterior,
		PriceUSD:      req.PriceUSD,
		Branch:        req.Branch,
		Origin:        req.Origin,
	}

	if err := h.vehicleSvc.Create(c.Context(), vehicle); err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(vehicle)
}

func (h *Handlers) GetVehicle(c *fiber.Ctx) error {
	id := c.Params("id")
	vehicle, err := h.vehicleSvc.GetByID(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(vehicle)
}

func (h *Handlers) ListVehicles(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	vehicles, total, err := h.vehicleSvc.List(c.Context(), page, pageSize)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"items": vehicles,
		"pagination": fiber.Map{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

func (h *Handlers) UpdateVehicle(c *fiber.Ctx) error {
	id := c.Params("id")
	vehicle, err := h.vehicleSvc.GetByID(c.Context(), id)
	if err != nil {
		return err
	}

	// Parse partial update
	var updates map[string]interface{}
	if err := c.BodyParser(&updates); err != nil {
		return errx.Validation("Invalid request body")
	}

	if v, ok := updates["plate"]; ok {
		if s, ok := v.(string); ok {
			vehicle.Plate = &s
		}
	}
	if v, ok := updates["brand"]; ok {
		if s, ok := v.(string); ok {
			vehicle.Brand = s
		}
	}
	if v, ok := updates["model"]; ok {
		if s, ok := v.(string); ok {
			vehicle.Model = s
		}
	}
	if v, ok := updates["version"]; ok {
		if s, ok := v.(string); ok {
			vehicle.Version = &s
		}
	}
	if v, ok := updates["trim"]; ok {
		if s, ok := v.(string); ok {
			vehicle.Trim = &s
		}
	}
	if v, ok := updates["year"]; ok {
		if f, ok := v.(float64); ok {
			vehicle.Year = int(f)
		}
	}
	if v, ok := updates["mileage_km"]; ok {
		if f, ok := v.(float64); ok {
			vehicle.MileageKM = int(f)
		}
	}
	if v, ok := updates["color_exterior"]; ok {
		if s, ok := v.(string); ok {
			vehicle.ColorExterior = &s
		}
	}
	if v, ok := updates["color_interior"]; ok {
		if s, ok := v.(string); ok {
			vehicle.ColorInterior = &s
		}
	}
	if v, ok := updates["price_usd"]; ok {
		if f, ok := v.(float64); ok {
			vehicle.PriceUSD = &f
		}
	}
	if v, ok := updates["branch"]; ok {
		if s, ok := v.(string); ok {
			vehicle.Branch = &s
		}
	}
	if v, ok := updates["origin"]; ok {
		if s, ok := v.(string); ok {
			vehicle.Origin = &s
		}
	}
	if v, ok := updates["status"]; ok {
		if s, ok := v.(string); ok {
			vehicle.Status = diveinspect.VehicleStatus(s)
		}
	}

	if err := h.vehicleSvc.Update(c.Context(), vehicle); err != nil {
		return err
	}

	return c.JSON(vehicle)
}

func (h *Handlers) DeleteVehicle(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.vehicleSvc.Delete(c.Context(), id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ============================================================================
// Enrichment Handlers
// ============================================================================

func (h *Handlers) EnrichVehicle(c *fiber.Ctx) error {
	id := c.Params("id")
	vehicle, err := h.vehicleSvc.GetByID(c.Context(), id)
	if err != nil {
		return err
	}

	if err := h.enrichmentSvc.EnrichVehicle(c.Context(), vehicle); err != nil {
		return err
	}

	// Return the full preview after enrichment
	preview, err := h.vehicleSvc.GetPreview(c.Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(preview)
}

// ============================================================================
// Preview & Publish
// ============================================================================

func (h *Handlers) GetVehiclePreview(c *fiber.Ctx) error {
	id := c.Params("id")
	preview, err := h.vehicleSvc.GetPreview(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(preview)
}

func (h *Handlers) PublishVehicle(c *fiber.Ctx) error {
	id := c.Params("id")
	vehicle, err := h.vehicleSvc.Publish(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(vehicle)
}

// ============================================================================
// Specs
// ============================================================================

func (h *Handlers) UpdateSpecs(c *fiber.Ctx) error {
	id := c.Params("id")

	// Verify vehicle exists
	if _, err := h.vehicleSvc.GetByID(c.Context(), id); err != nil {
		return err
	}

	var specs diveinspect.VehicleSpecs
	if err := c.BodyParser(&specs); err != nil {
		return errx.Validation("Invalid request body")
	}
	specs.VehicleID = id

	if err := h.vehicleSvc.UpdateSpecs(c.Context(), &specs); err != nil {
		return err
	}

	return c.JSON(specs)
}

// ============================================================================
// Photo Upload & Inspection
// ============================================================================

func (h *Handlers) UploadPhoto(c *fiber.Ctx) error {
	vehicleID := c.Params("id")

	// Verify vehicle exists
	if _, err := h.vehicleSvc.GetByID(c.Context(), vehicleID); err != nil {
		return err
	}

	zone := c.FormValue("zone", "closeup")
	photoZone := diveinspect.PhotoZone(zone)

	// Get or create inspection for this vehicle
	inspection, err := h.inspectionSvc.GetByVehicleID(c.Context(), vehicleID)
	if err != nil {
		// Create a new inspection
		inspectorName := c.FormValue("inspector_name", "")
		inspectorBranch := c.FormValue("inspector_branch", "")
		var namePtr, branchPtr *string
		if inspectorName != "" {
			namePtr = &inspectorName
		}
		if inspectorBranch != "" {
			branchPtr = &inspectorBranch
		}
		newInspection, createErr := h.inspectionSvc.CreateInspection(c.Context(), vehicleID, namePtr, branchPtr)
		if createErr != nil {
			return createErr
		}
		inspection = &diveinspect.InspectionFullView{
			Inspection: *newInspection,
		}
	}

	// Get the uploaded file
	file, err := c.FormFile("photo")
	if err != nil {
		return errx.Validation("Photo file is required")
	}

	fileReader, err := file.Open()
	if err != nil {
		return errx.Internal("Failed to read uploaded file")
	}
	defer fileReader.Close()

	photo, err := h.inspectionSvc.UploadPhoto(c.Context(), inspection.Inspection.ID, photoZone, fileReader, file.Filename)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(photo)
}

func (h *Handlers) RunInspection(c *fiber.Ctx) error {
	vehicleID := c.Params("id")

	// Get existing inspection
	inspection, err := h.inspectionSvc.GetByVehicleID(c.Context(), vehicleID)
	if err != nil {
		return errx.NotFound("No inspection found for this vehicle. Upload photos first.")
	}

	if err := h.inspectionSvc.RunInspection(c.Context(), inspection.Inspection.ID); err != nil {
		return err
	}

	// Return updated inspection
	updated, err := h.inspectionSvc.GetByID(c.Context(), inspection.Inspection.ID)
	if err != nil {
		return err
	}

	return c.JSON(updated)
}

// ============================================================================
// Report
// ============================================================================

func (h *Handlers) GetReport(c *fiber.Ctx) error {
	vehicleID := c.Params("id")

	_, pdfBytes, err := h.reportSvc.GenerateReport(c.Context(), vehicleID)
	if err != nil {
		return err
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "inline; filename=inspection_report.pdf")
	return c.Send(pdfBytes)
}

// ============================================================================
// Listing JSON
// ============================================================================

func (h *Handlers) GetListingJSON(c *fiber.Ctx) error {
	vehicleID := c.Params("id")
	preview, err := h.vehicleSvc.GetPreview(c.Context(), vehicleID)
	if err != nil {
		return err
	}
	return c.JSON(preview)
}

// ============================================================================
// Finding Update
// ============================================================================

type updateFindingRequest struct {
	Zone            *string `json:"zone"`
	FindingType     *string `json:"finding_type"`
	Severity        *string `json:"severity"`
	Description     *string `json:"description"`
	ConfirmedByHuman *bool  `json:"confirmed_by_human"`
}

func (h *Handlers) UpdateFinding(c *fiber.Ctx) error {
	fid := c.Params("fid")

	var req updateFindingRequest
	if err := c.BodyParser(&req); err != nil {
		return errx.Validation("Invalid request body")
	}

	finding := &diveinspect.InspectionFinding{ID: fid}

	if req.Zone != nil {
		finding.Zone = diveinspect.FindingZone(*req.Zone)
	}
	if req.FindingType != nil {
		finding.FindingType = diveinspect.FindingType(*req.FindingType)
	}
	if req.Severity != nil {
		finding.Severity = diveinspect.FindingSeverity(*req.Severity)
	}
	if req.Description != nil {
		finding.Description = req.Description
	}
	if req.ConfirmedByHuman != nil {
		finding.ConfirmedByHuman = *req.ConfirmedByHuman
	}

	if err := h.inspectionSvc.UpdateFinding(c.Context(), finding); err != nil {
		return err
	}

	return c.JSON(finding)
}

