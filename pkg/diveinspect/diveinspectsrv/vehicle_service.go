package diveinspectsrv

import (
	"context"

	"github.com/Abraxas-365/divi/pkg/diveinspect"
	"github.com/Abraxas-365/divi/pkg/errx"
)

type VehicleService struct {
	vehicleRepo   diveinspect.VehicleRepository
	specsRepo     diveinspect.VehicleSpecsRepository
	equipmentRepo diveinspect.VehicleEquipmentRepository
	listingRepo   diveinspect.GeneratedListingRepository
	inspectionRepo diveinspect.InspectionRepository
	findingRepo   diveinspect.InspectionFindingRepository
	photoRepo     diveinspect.InspectionPhotoRepository
}

func NewVehicleService(
	vehicleRepo diveinspect.VehicleRepository,
	specsRepo diveinspect.VehicleSpecsRepository,
	equipmentRepo diveinspect.VehicleEquipmentRepository,
	listingRepo diveinspect.GeneratedListingRepository,
	inspectionRepo diveinspect.InspectionRepository,
	findingRepo diveinspect.InspectionFindingRepository,
	photoRepo diveinspect.InspectionPhotoRepository,
) *VehicleService {
	return &VehicleService{
		vehicleRepo:    vehicleRepo,
		specsRepo:      specsRepo,
		equipmentRepo:  equipmentRepo,
		listingRepo:    listingRepo,
		inspectionRepo: inspectionRepo,
		findingRepo:    findingRepo,
		photoRepo:      photoRepo,
	}
}

func (s *VehicleService) Create(ctx context.Context, v *diveinspect.Vehicle) error {
	if v.Brand == "" || v.Model == "" {
		return errx.Validation("Brand and model are required")
	}
	if v.Year < 1900 || v.Year > 2100 {
		return errx.Validation("Invalid vehicle year")
	}
	if v.Status == "" {
		v.Status = diveinspect.VehicleStatusDraft
	}
	return s.vehicleRepo.Create(ctx, v)
}

func (s *VehicleService) GetByID(ctx context.Context, id string) (*diveinspect.Vehicle, error) {
	return s.vehicleRepo.GetByID(ctx, id)
}

func (s *VehicleService) Update(ctx context.Context, v *diveinspect.Vehicle) error {
	return s.vehicleRepo.Update(ctx, v)
}

func (s *VehicleService) Delete(ctx context.Context, id string) error {
	return s.vehicleRepo.Delete(ctx, id)
}

func (s *VehicleService) List(ctx context.Context, page, pageSize int) ([]diveinspect.Vehicle, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.vehicleRepo.List(ctx, page, pageSize)
}

func (s *VehicleService) UpdateSpecs(ctx context.Context, specs *diveinspect.VehicleSpecs) error {
	return s.specsRepo.Upsert(ctx, specs)
}

func (s *VehicleService) GetPreview(ctx context.Context, vehicleID string) (*diveinspect.VehiclePreview, error) {
	vehicle, err := s.vehicleRepo.GetByID(ctx, vehicleID)
	if err != nil {
		return nil, err
	}

	preview := &diveinspect.VehiclePreview{
		Vehicle: *vehicle,
	}

	specs, err := s.specsRepo.GetByVehicleID(ctx, vehicleID)
	if err == nil {
		preview.Specs = specs
	}

	equipment, err := s.equipmentRepo.GetByVehicleID(ctx, vehicleID)
	if err == nil {
		preview.Equipment = equipment
	}

	listing, err := s.listingRepo.GetByVehicleID(ctx, vehicleID)
	if err == nil {
		preview.Listing = listing
	}

	inspection, err := s.inspectionRepo.GetByVehicleID(ctx, vehicleID)
	if err == nil {
		findings, _ := s.findingRepo.GetByInspectionID(ctx, inspection.ID)
		photos, _ := s.photoRepo.GetByInspectionID(ctx, inspection.ID)
		preview.Inspection = &diveinspect.InspectionFullView{
			Inspection: *inspection,
			Findings:   findings,
			Photos:     photos,
		}
	}

	return preview, nil
}

func (s *VehicleService) Publish(ctx context.Context, vehicleID string) (*diveinspect.Vehicle, error) {
	vehicle, err := s.vehicleRepo.GetByID(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	vehicle.Status = diveinspect.VehicleStatusPublished
	if err := s.vehicleRepo.Update(ctx, vehicle); err != nil {
		return nil, err
	}
	return vehicle, nil
}
