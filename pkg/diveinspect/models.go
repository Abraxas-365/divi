package diveinspect

import (
	"time"

	"github.com/lib/pq"
)

// ============================================================================
// Vehicle
// ============================================================================

type VehicleStatus string

const (
	VehicleStatusDraft     VehicleStatus = "draft"
	VehicleStatusReview    VehicleStatus = "review"
	VehicleStatusPublished VehicleStatus = "published"
)

type Vehicle struct {
	ID            string        `json:"id" db:"id"`
	Plate         *string       `json:"plate,omitempty" db:"plate"`
	Brand         string        `json:"brand" db:"brand"`
	Model         string        `json:"model" db:"model"`
	Version       *string       `json:"version,omitempty" db:"version"`
	Trim          *string       `json:"trim,omitempty" db:"trim"`
	Year          int           `json:"year" db:"year"`
	MileageKM     int           `json:"mileage_km" db:"mileage_km"`
	ColorExterior *string       `json:"color_exterior,omitempty" db:"color_exterior"`
	ColorInterior *string       `json:"color_interior,omitempty" db:"color_interior"`
	PriceUSD      *float64      `json:"price_usd,omitempty" db:"price_usd"`
	Branch        *string       `json:"branch,omitempty" db:"branch"`
	Origin        *string       `json:"origin,omitempty" db:"origin"`
	Status        VehicleStatus `json:"status" db:"status"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
}

// ============================================================================
// Vehicle Specs
// ============================================================================

type VehicleSpecs struct {
	ID        string `json:"id" db:"id"`
	VehicleID string `json:"vehicle_id" db:"vehicle_id"`

	// Engine
	EngineType     *string  `json:"engine_type,omitempty" db:"engine_type"`
	EngineCC       *int     `json:"engine_cc,omitempty" db:"engine_cc"`
	EngineCylinders *int    `json:"engine_cylinders,omitempty" db:"engine_cylinders"`
	PowerHP        *float64 `json:"power_hp,omitempty" db:"power_hp"`
	PowerKW        *float64 `json:"power_kw,omitempty" db:"power_kw"`
	TorqueNM       *int     `json:"torque_nm,omitempty" db:"torque_nm"`
	TorqueRPMRange *string  `json:"torque_rpm_range,omitempty" db:"torque_rpm_range"`
	FuelType       *string  `json:"fuel_type,omitempty" db:"fuel_type"`
	FuelSystem     *string  `json:"fuel_system,omitempty" db:"fuel_system"`

	// Transmission
	TransmissionType  *string `json:"transmission_type,omitempty" db:"transmission_type"`
	TransmissionGears *int    `json:"transmission_gears,omitempty" db:"transmission_gears"`
	Drivetrain        *string `json:"drivetrain,omitempty" db:"drivetrain"`

	// Performance
	Accel0100   *float64 `json:"accel_0_100,omitempty" db:"accel_0_100"`
	TopSpeedKMH *int     `json:"top_speed_kmh,omitempty" db:"top_speed_kmh"`

	// Fuel consumption
	FuelCityKML     *float64 `json:"fuel_city_kml,omitempty" db:"fuel_city_kml"`
	FuelHighwayKML  *float64 `json:"fuel_highway_kml,omitempty" db:"fuel_highway_kml"`
	FuelCombinedKML *float64 `json:"fuel_combined_kml,omitempty" db:"fuel_combined_kml"`
	FuelTankLiters  *int     `json:"fuel_tank_liters,omitempty" db:"fuel_tank_liters"`

	// Dimensions
	LengthMM      *int `json:"length_mm,omitempty" db:"length_mm"`
	WidthMM       *int `json:"width_mm,omitempty" db:"width_mm"`
	HeightMM      *int `json:"height_mm,omitempty" db:"height_mm"`
	WheelbaseMM   *int `json:"wheelbase_mm,omitempty" db:"wheelbase_mm"`
	CargoLiters   *int `json:"cargo_liters,omitempty" db:"cargo_liters"`
	CargoMaxLiters *int `json:"cargo_max_liters,omitempty" db:"cargo_max_liters"`
	CurbWeightKG  *int `json:"curb_weight_kg,omitempty" db:"curb_weight_kg"`

	// Tires
	TireSize  *string `json:"tire_size,omitempty" db:"tire_size"`
	SpareTire *string `json:"spare_tire,omitempty" db:"spare_tire"`

	// Source
	SpecsSource     *string    `json:"specs_source,omitempty" db:"specs_source"`
	SpecsConfidence *float64   `json:"specs_confidence,omitempty" db:"specs_confidence"`
	EnrichedAt      *time.Time `json:"enriched_at,omitempty" db:"enriched_at"`
}

// ============================================================================
// Vehicle Equipment
// ============================================================================

type EquipmentCategory string

const (
	EquipmentSafety       EquipmentCategory = "safety"
	EquipmentComfort      EquipmentCategory = "comfort"
	EquipmentInfotainment EquipmentCategory = "infotainment"
	EquipmentExterior     EquipmentCategory = "exterior"
	EquipmentInterior     EquipmentCategory = "interior"
)

type EquipmentSource string

const (
	SourceFactorySpec     EquipmentSource = "factory_spec"
	SourceVisualDetection EquipmentSource = "visual_detection"
	SourceManualInput     EquipmentSource = "manual_input"
)

type VehicleEquipment struct {
	ID                 string            `json:"id" db:"id"`
	VehicleID          string            `json:"vehicle_id" db:"vehicle_id"`
	Category           EquipmentCategory `json:"category" db:"category"`
	FeatureName        string            `json:"feature_name" db:"feature_name"`
	FeatureDescription *string           `json:"feature_description,omitempty" db:"feature_description"`
	IsStandard         bool              `json:"is_standard" db:"is_standard"`
	IsConfirmed        bool              `json:"is_confirmed" db:"is_confirmed"`
	Source             EquipmentSource   `json:"source" db:"source"`
}

// ============================================================================
// Inspection
// ============================================================================

type InspectionStatus string

const (
	InspectionPending    InspectionStatus = "pending"
	InspectionProcessing InspectionStatus = "processing"
	InspectionCompleted  InspectionStatus = "completed"
	InspectionApproved   InspectionStatus = "approved"
)

type Inspection struct {
	ID              string           `json:"id" db:"id"`
	VehicleID       string           `json:"vehicle_id" db:"vehicle_id"`
	InspectorName   *string          `json:"inspector_name,omitempty" db:"inspector_name"`
	InspectorBranch *string          `json:"inspector_branch,omitempty" db:"inspector_branch"`
	ScoreOverall    *int             `json:"score_overall,omitempty" db:"score_overall"`
	ScoreExterior   *int             `json:"score_exterior,omitempty" db:"score_exterior"`
	ScoreInterior   *int             `json:"score_interior,omitempty" db:"score_interior"`
	ScoreMechanical *int             `json:"score_mechanical,omitempty" db:"score_mechanical"`
	ScoreTires      *int             `json:"score_tires,omitempty" db:"score_tires"`
	PhotosCount     int              `json:"photos_count" db:"photos_count"`
	FindingsCount   int              `json:"findings_count" db:"findings_count"`
	Status          InspectionStatus `json:"status" db:"status"`
	PDFURL          *string          `json:"pdf_url,omitempty" db:"pdf_url"`
	InspectedAt     *time.Time       `json:"inspected_at,omitempty" db:"inspected_at"`
	CreatedAt       time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at" db:"updated_at"`
}

// ============================================================================
// Inspection Finding
// ============================================================================

type FindingZone string

const (
	ZoneFront         FindingZone = "front"
	ZoneRear          FindingZone = "rear"
	ZoneLeft          FindingZone = "left"
	ZoneRight         FindingZone = "right"
	ZoneRoof          FindingZone = "roof"
	ZoneInteriorFront FindingZone = "interior_front"
	ZoneInteriorRear  FindingZone = "interior_rear"
	ZoneEngine        FindingZone = "engine"
	ZoneTrunk         FindingZone = "trunk"
	ZoneTires         FindingZone = "tires"
)

type FindingType string

const (
	FindingScratch       FindingType = "scratch"
	FindingDent          FindingType = "dent"
	FindingRust          FindingType = "rust"
	FindingPaintMismatch FindingType = "paint_mismatch"
	FindingWear          FindingType = "wear"
	FindingCrack         FindingType = "crack"
	FindingStain         FindingType = "stain"
	FindingMissingPart   FindingType = "missing_part"
)

type FindingSeverity string

const (
	SeverityMinor    FindingSeverity = "minor"
	SeverityModerate FindingSeverity = "moderate"
	SeverityMajor    FindingSeverity = "major"
)

type InspectionFinding struct {
	ID               string          `json:"id" db:"id"`
	InspectionID     string          `json:"inspection_id" db:"inspection_id"`
	PhotoURL         *string         `json:"photo_url,omitempty" db:"photo_url"`
	AnnotatedPhotoURL *string        `json:"annotated_photo_url,omitempty" db:"annotated_photo_url"`
	Zone             FindingZone     `json:"zone" db:"zone"`
	FindingType      FindingType     `json:"finding_type" db:"finding_type"`
	Severity         FindingSeverity `json:"severity" db:"severity"`
	Description      *string         `json:"description,omitempty" db:"description"`
	AIConfidence     *float64        `json:"ai_confidence,omitempty" db:"ai_confidence"`
	ConfirmedByHuman bool            `json:"confirmed_by_human" db:"confirmed_by_human"`
}

// ============================================================================
// Inspection Photo
// ============================================================================

type PhotoZone string

const (
	PhotoZoneFront             PhotoZone = "front"
	PhotoZoneRear              PhotoZone = "rear"
	PhotoZoneLeft              PhotoZone = "left"
	PhotoZoneRight             PhotoZone = "right"
	PhotoZoneFrontLeft         PhotoZone = "front_left"
	PhotoZoneRearRight         PhotoZone = "rear_right"
	PhotoZoneInteriorDriver    PhotoZone = "interior_driver"
	PhotoZoneInteriorPassenger PhotoZone = "interior_passenger"
	PhotoZoneInteriorRear      PhotoZone = "interior_rear"
	PhotoZoneDashboard         PhotoZone = "dashboard"
	PhotoZoneInfotainment      PhotoZone = "infotainment"
	PhotoZoneEngine            PhotoZone = "engine"
	PhotoZoneTrunk             PhotoZone = "trunk"
	PhotoZoneCloseup           PhotoZone = "closeup"
)

type InspectionPhoto struct {
	ID           string    `json:"id" db:"id"`
	InspectionID string    `json:"inspection_id" db:"inspection_id"`
	PhotoURL     string    `json:"photo_url" db:"photo_url"`
	Zone         PhotoZone `json:"zone" db:"zone"`
	SortOrder    int       `json:"sort_order" db:"sort_order"`
	UploadedAt   time.Time `json:"uploaded_at" db:"uploaded_at"`
}

// ============================================================================
// Generated Listing
// ============================================================================

type GeneratedListing struct {
	ID             string         `json:"id" db:"id"`
	VehicleID      string         `json:"vehicle_id" db:"vehicle_id"`
	Title          *string        `json:"title,omitempty" db:"title"`
	DescriptionES  *string        `json:"description_es,omitempty" db:"description_es"`
	DescriptionEN  *string        `json:"description_en,omitempty" db:"description_en"`
	SEOKeywords    pq.StringArray `json:"seo_keywords,omitempty" db:"seo_keywords"`
	SchemaJSONLD   *string        `json:"schema_json_ld,omitempty" db:"schema_json_ld"`
	GeneratedAt    time.Time      `json:"generated_at" db:"generated_at"`
}

// ============================================================================
// Composite Views (for API responses)
// ============================================================================

type VehicleFullView struct {
	Vehicle   Vehicle            `json:"vehicle"`
	Specs     *VehicleSpecs      `json:"specs,omitempty"`
	Equipment []VehicleEquipment `json:"equipment,omitempty"`
	Listing   *GeneratedListing  `json:"listing,omitempty"`
}

type InspectionFullView struct {
	Inspection Inspection          `json:"inspection"`
	Findings   []InspectionFinding `json:"findings,omitempty"`
	Photos     []InspectionPhoto   `json:"photos,omitempty"`
}

type VehiclePreview struct {
	Vehicle    Vehicle             `json:"vehicle"`
	Specs      *VehicleSpecs       `json:"specs,omitempty"`
	Equipment  []VehicleEquipment  `json:"equipment,omitempty"`
	Listing    *GeneratedListing   `json:"listing,omitempty"`
	Inspection *InspectionFullView `json:"inspection,omitempty"`
}
