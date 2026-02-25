package diveinspect

import "github.com/Abraxas-365/divi/pkg/errx"

var errorRegistry = errx.NewRegistry("DIVEINSPECT")

var (
	ErrVehicleNotFound    = errorRegistry.Register("VEHICLE_NOT_FOUND", errx.TypeNotFound, 404, "Vehicle not found")
	ErrInspectionNotFound = errorRegistry.Register("INSPECTION_NOT_FOUND", errx.TypeNotFound, 404, "Inspection not found")
	ErrFindingNotFound    = errorRegistry.Register("FINDING_NOT_FOUND", errx.TypeNotFound, 404, "Inspection finding not found")
	ErrListingNotFound    = errorRegistry.Register("LISTING_NOT_FOUND", errx.TypeNotFound, 404, "Generated listing not found")
	ErrSpecsNotFound      = errorRegistry.Register("SPECS_NOT_FOUND", errx.TypeNotFound, 404, "Vehicle specs not found")

	ErrInvalidInput   = errorRegistry.Register("INVALID_INPUT", errx.TypeValidation, 400, "Invalid input data")
	ErrMissingField   = errorRegistry.Register("MISSING_FIELD", errx.TypeValidation, 400, "Required field is missing")
	ErrInvalidStatus  = errorRegistry.Register("INVALID_STATUS", errx.TypeValidation, 400, "Invalid status value")
	ErrInvalidYear    = errorRegistry.Register("INVALID_YEAR", errx.TypeValidation, 400, "Invalid vehicle year")

	ErrEnrichmentFailed  = errorRegistry.Register("ENRICHMENT_FAILED", errx.TypeExternal, 502, "Vehicle enrichment failed")
	ErrVisionFailed      = errorRegistry.Register("VISION_FAILED", errx.TypeExternal, 502, "Vision analysis failed")
	ErrPDFGenFailed      = errorRegistry.Register("PDF_GEN_FAILED", errx.TypeInternal, 500, "PDF generation failed")

	ErrPhotoUploadFailed = errorRegistry.Register("PHOTO_UPLOAD_FAILED", errx.TypeInternal, 500, "Photo upload failed")
	ErrDBOperation       = errorRegistry.Register("DB_OPERATION", errx.TypeInternal, 500, "Database operation failed")
)
