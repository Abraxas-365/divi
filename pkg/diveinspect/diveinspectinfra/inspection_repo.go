package diveinspectinfra

import (
	"context"

	"github.com/Abraxas-365/divi/pkg/diveinspect"
	"github.com/Abraxas-365/divi/pkg/errx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// ============================================================================
// Inspection Repository
// ============================================================================

type PostgresInspectionRepository struct {
	db *sqlx.DB
}

func NewPostgresInspectionRepository(db *sqlx.DB) *PostgresInspectionRepository {
	return &PostgresInspectionRepository{db: db}
}

func (r *PostgresInspectionRepository) Create(ctx context.Context, i *diveinspect.Inspection) error {
	if i.ID == "" {
		i.ID = uuid.New().String()
	}
	query := `
		INSERT INTO inspections (id, vehicle_id, inspector_name, inspector_branch, score_overall, score_exterior, score_interior, score_mechanical, score_tires, photos_count, findings_count, status, pdf_url, inspected_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING created_at, updated_at`
	return r.db.QueryRowContext(ctx, query,
		i.ID, i.VehicleID, i.InspectorName, i.InspectorBranch,
		i.ScoreOverall, i.ScoreExterior, i.ScoreInterior, i.ScoreMechanical, i.ScoreTires,
		i.PhotosCount, i.FindingsCount, i.Status, i.PDFURL, i.InspectedAt,
	).Scan(&i.CreatedAt, &i.UpdatedAt)
}

func (r *PostgresInspectionRepository) GetByID(ctx context.Context, id string) (*diveinspect.Inspection, error) {
	var i diveinspect.Inspection
	query := `SELECT * FROM inspections WHERE id = $1`
	if err := r.db.GetContext(ctx, &i, query, id); err != nil {
		return nil, errx.NotFound("Inspection not found").WithDetail("id", id)
	}
	return &i, nil
}

func (r *PostgresInspectionRepository) GetByVehicleID(ctx context.Context, vehicleID string) (*diveinspect.Inspection, error) {
	var i diveinspect.Inspection
	query := `SELECT * FROM inspections WHERE vehicle_id = $1 ORDER BY created_at DESC LIMIT 1`
	if err := r.db.GetContext(ctx, &i, query, vehicleID); err != nil {
		return nil, errx.NotFound("Inspection not found for vehicle").WithDetail("vehicle_id", vehicleID)
	}
	return &i, nil
}

func (r *PostgresInspectionRepository) Update(ctx context.Context, i *diveinspect.Inspection) error {
	query := `
		UPDATE inspections SET
			inspector_name = $2, inspector_branch = $3,
			score_overall = $4, score_exterior = $5, score_interior = $6,
			score_mechanical = $7, score_tires = $8,
			photos_count = $9, findings_count = $10, status = $11,
			pdf_url = $12, inspected_at = $13
		WHERE id = $1
		RETURNING updated_at`
	return r.db.QueryRowContext(ctx, query,
		i.ID, i.InspectorName, i.InspectorBranch,
		i.ScoreOverall, i.ScoreExterior, i.ScoreInterior, i.ScoreMechanical, i.ScoreTires,
		i.PhotosCount, i.FindingsCount, i.Status, i.PDFURL, i.InspectedAt,
	).Scan(&i.UpdatedAt)
}

func (r *PostgresInspectionRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM inspections WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errx.NotFound("Inspection not found").WithDetail("id", id)
	}
	return nil
}

// ============================================================================
// Inspection Finding Repository
// ============================================================================

type PostgresInspectionFindingRepository struct {
	db *sqlx.DB
}

func NewPostgresInspectionFindingRepository(db *sqlx.DB) *PostgresInspectionFindingRepository {
	return &PostgresInspectionFindingRepository{db: db}
}

func (r *PostgresInspectionFindingRepository) CreateBatch(ctx context.Context, findings []diveinspect.InspectionFinding) error {
	if len(findings) == 0 {
		return nil
	}
	query := `
		INSERT INTO inspection_findings (id, inspection_id, photo_url, annotated_photo_url, zone, finding_type, severity, description, ai_confidence, confirmed_by_human)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i := range findings {
		if findings[i].ID == "" {
			findings[i].ID = uuid.New().String()
		}
		_, err := tx.ExecContext(ctx, query,
			findings[i].ID, findings[i].InspectionID, findings[i].PhotoURL, findings[i].AnnotatedPhotoURL,
			findings[i].Zone, findings[i].FindingType, findings[i].Severity,
			findings[i].Description, findings[i].AIConfidence, findings[i].ConfirmedByHuman,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *PostgresInspectionFindingRepository) GetByInspectionID(ctx context.Context, inspectionID string) ([]diveinspect.InspectionFinding, error) {
	var findings []diveinspect.InspectionFinding
	query := `SELECT * FROM inspection_findings WHERE inspection_id = $1 ORDER BY severity DESC, zone`
	if err := r.db.SelectContext(ctx, &findings, query, inspectionID); err != nil {
		return nil, err
	}
	return findings, nil
}

func (r *PostgresInspectionFindingRepository) Update(ctx context.Context, f *diveinspect.InspectionFinding) error {
	query := `
		UPDATE inspection_findings SET
			zone = $2, finding_type = $3, severity = $4, description = $5,
			confirmed_by_human = $6
		WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query,
		f.ID, f.Zone, f.FindingType, f.Severity, f.Description, f.ConfirmedByHuman,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errx.NotFound("Finding not found").WithDetail("id", f.ID)
	}
	return nil
}

func (r *PostgresInspectionFindingRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM inspection_findings WHERE id = $1`, id)
	return err
}

// ============================================================================
// Inspection Photo Repository
// ============================================================================

type PostgresInspectionPhotoRepository struct {
	db *sqlx.DB
}

func NewPostgresInspectionPhotoRepository(db *sqlx.DB) *PostgresInspectionPhotoRepository {
	return &PostgresInspectionPhotoRepository{db: db}
}

func (r *PostgresInspectionPhotoRepository) Create(ctx context.Context, p *diveinspect.InspectionPhoto) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	query := `
		INSERT INTO inspection_photos (id, inspection_id, photo_url, zone, sort_order)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING uploaded_at`
	return r.db.QueryRowContext(ctx, query,
		p.ID, p.InspectionID, p.PhotoURL, p.Zone, p.SortOrder,
	).Scan(&p.UploadedAt)
}

func (r *PostgresInspectionPhotoRepository) GetByInspectionID(ctx context.Context, inspectionID string) ([]diveinspect.InspectionPhoto, error) {
	var photos []diveinspect.InspectionPhoto
	query := `SELECT * FROM inspection_photos WHERE inspection_id = $1 ORDER BY sort_order`
	if err := r.db.SelectContext(ctx, &photos, query, inspectionID); err != nil {
		return nil, err
	}
	return photos, nil
}

func (r *PostgresInspectionPhotoRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM inspection_photos WHERE id = $1`, id)
	return err
}
