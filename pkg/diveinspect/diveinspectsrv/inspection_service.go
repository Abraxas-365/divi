package diveinspectsrv

import (
	"context"
	"fmt"
	"io"

	"github.com/Abraxas-365/divi/pkg/diveinspect"
	"github.com/Abraxas-365/divi/pkg/errx"
	"github.com/Abraxas-365/divi/pkg/fsx"
	"github.com/google/uuid"
)

type InspectionService struct {
	inspectionRepo diveinspect.InspectionRepository
	findingRepo    diveinspect.InspectionFindingRepository
	photoRepo      diveinspect.InspectionPhotoRepository
	vehicleRepo    diveinspect.VehicleRepository
	fs             fsx.FileSystem
	visionService  *VisionService
}

func NewInspectionService(
	inspectionRepo diveinspect.InspectionRepository,
	findingRepo diveinspect.InspectionFindingRepository,
	photoRepo diveinspect.InspectionPhotoRepository,
	vehicleRepo diveinspect.VehicleRepository,
	fs fsx.FileSystem,
	visionService *VisionService,
) *InspectionService {
	return &InspectionService{
		inspectionRepo: inspectionRepo,
		findingRepo:    findingRepo,
		photoRepo:      photoRepo,
		vehicleRepo:    vehicleRepo,
		fs:             fs,
		visionService:  visionService,
	}
}

func (s *InspectionService) CreateInspection(ctx context.Context, vehicleID string, inspectorName, inspectorBranch *string) (*diveinspect.Inspection, error) {
	// Verify vehicle exists
	if _, err := s.vehicleRepo.GetByID(ctx, vehicleID); err != nil {
		return nil, err
	}

	inspection := &diveinspect.Inspection{
		VehicleID:       vehicleID,
		InspectorName:   inspectorName,
		InspectorBranch: inspectorBranch,
		Status:          diveinspect.InspectionPending,
	}

	if err := s.inspectionRepo.Create(ctx, inspection); err != nil {
		return nil, errx.Wrap(err, "Failed to create inspection", errx.TypeInternal)
	}

	return inspection, nil
}

func (s *InspectionService) GetByID(ctx context.Context, id string) (*diveinspect.InspectionFullView, error) {
	inspection, err := s.inspectionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	findings, _ := s.findingRepo.GetByInspectionID(ctx, id)
	photos, _ := s.photoRepo.GetByInspectionID(ctx, id)

	return &diveinspect.InspectionFullView{
		Inspection: *inspection,
		Findings:   findings,
		Photos:     photos,
	}, nil
}

func (s *InspectionService) UploadPhoto(ctx context.Context, inspectionID string, zone diveinspect.PhotoZone, fileData io.Reader, filename string) (*diveinspect.InspectionPhoto, error) {
	// Verify inspection exists
	inspection, err := s.inspectionRepo.GetByID(ctx, inspectionID)
	if err != nil {
		return nil, err
	}

	// Store the photo
	photoID := uuid.New().String()
	storagePath := fmt.Sprintf("inspections/%s/photos/%s_%s", inspectionID, photoID, filename)

	if err := s.fs.WriteFileStream(ctx, storagePath, fileData); err != nil {
		return nil, errx.Wrap(err, "Failed to upload photo", errx.TypeInternal)
	}

	// Get current photo count for sort order
	photos, _ := s.photoRepo.GetByInspectionID(ctx, inspectionID)
	sortOrder := len(photos)

	photo := &diveinspect.InspectionPhoto{
		InspectionID: inspectionID,
		PhotoURL:     storagePath,
		Zone:         zone,
		SortOrder:    sortOrder,
	}

	if err := s.photoRepo.Create(ctx, photo); err != nil {
		return nil, errx.Wrap(err, "Failed to save photo record", errx.TypeInternal)
	}

	// Update photos count
	inspection.PhotosCount = sortOrder + 1
	_ = s.inspectionRepo.Update(ctx, inspection)

	return photo, nil
}

func (s *InspectionService) RunInspection(ctx context.Context, inspectionID string) error {
	inspection, err := s.inspectionRepo.GetByID(ctx, inspectionID)
	if err != nil {
		return err
	}

	vehicle, err := s.vehicleRepo.GetByID(ctx, inspection.VehicleID)
	if err != nil {
		return err
	}

	return s.visionService.RunInspection(ctx, vehicle, inspectionID)
}

func (s *InspectionService) UpdateFinding(ctx context.Context, finding *diveinspect.InspectionFinding) error {
	return s.findingRepo.Update(ctx, finding)
}

func (s *InspectionService) GetByVehicleID(ctx context.Context, vehicleID string) (*diveinspect.InspectionFullView, error) {
	inspection, err := s.inspectionRepo.GetByVehicleID(ctx, vehicleID)
	if err != nil {
		return nil, err
	}

	findings, _ := s.findingRepo.GetByInspectionID(ctx, inspection.ID)
	photos, _ := s.photoRepo.GetByInspectionID(ctx, inspection.ID)

	return &diveinspect.InspectionFullView{
		Inspection: *inspection,
		Findings:   findings,
		Photos:     photos,
	}, nil
}
