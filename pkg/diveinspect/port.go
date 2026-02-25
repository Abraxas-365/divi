package diveinspect

import "context"

// ============================================================================
// Vehicle Repository
// ============================================================================

type VehicleRepository interface {
	Create(ctx context.Context, v *Vehicle) error
	GetByID(ctx context.Context, id string) (*Vehicle, error)
	Update(ctx context.Context, v *Vehicle) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, page, pageSize int) ([]Vehicle, int, error)
	ListByStatus(ctx context.Context, status VehicleStatus, page, pageSize int) ([]Vehicle, int, error)
}

// ============================================================================
// Vehicle Specs Repository
// ============================================================================

type VehicleSpecsRepository interface {
	Create(ctx context.Context, s *VehicleSpecs) error
	GetByVehicleID(ctx context.Context, vehicleID string) (*VehicleSpecs, error)
	Upsert(ctx context.Context, s *VehicleSpecs) error
	Delete(ctx context.Context, vehicleID string) error
}

// ============================================================================
// Vehicle Equipment Repository
// ============================================================================

type VehicleEquipmentRepository interface {
	CreateBatch(ctx context.Context, equipment []VehicleEquipment) error
	GetByVehicleID(ctx context.Context, vehicleID string) ([]VehicleEquipment, error)
	GetByVehicleIDAndCategory(ctx context.Context, vehicleID string, category EquipmentCategory) ([]VehicleEquipment, error)
	DeleteByVehicleID(ctx context.Context, vehicleID string) error
}

// ============================================================================
// Inspection Repository
// ============================================================================

type InspectionRepository interface {
	Create(ctx context.Context, i *Inspection) error
	GetByID(ctx context.Context, id string) (*Inspection, error)
	GetByVehicleID(ctx context.Context, vehicleID string) (*Inspection, error)
	Update(ctx context.Context, i *Inspection) error
	Delete(ctx context.Context, id string) error
}

// ============================================================================
// Inspection Finding Repository
// ============================================================================

type InspectionFindingRepository interface {
	CreateBatch(ctx context.Context, findings []InspectionFinding) error
	GetByInspectionID(ctx context.Context, inspectionID string) ([]InspectionFinding, error)
	Update(ctx context.Context, f *InspectionFinding) error
	Delete(ctx context.Context, id string) error
}

// ============================================================================
// Inspection Photo Repository
// ============================================================================

type InspectionPhotoRepository interface {
	Create(ctx context.Context, p *InspectionPhoto) error
	GetByInspectionID(ctx context.Context, inspectionID string) ([]InspectionPhoto, error)
	Delete(ctx context.Context, id string) error
}

// ============================================================================
// Generated Listing Repository
// ============================================================================

type GeneratedListingRepository interface {
	Upsert(ctx context.Context, l *GeneratedListing) error
	GetByVehicleID(ctx context.Context, vehicleID string) (*GeneratedListing, error)
	Delete(ctx context.Context, vehicleID string) error
}
