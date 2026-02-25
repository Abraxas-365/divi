package diveinspectinfra

import (
	"context"

	"github.com/Abraxas-365/divi/pkg/diveinspect"
	"github.com/Abraxas-365/divi/pkg/errx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PostgresGeneratedListingRepository struct {
	db *sqlx.DB
}

func NewPostgresGeneratedListingRepository(db *sqlx.DB) *PostgresGeneratedListingRepository {
	return &PostgresGeneratedListingRepository{db: db}
}

func (r *PostgresGeneratedListingRepository) Upsert(ctx context.Context, l *diveinspect.GeneratedListing) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	query := `
		INSERT INTO generated_listings (id, vehicle_id, title, description_es, description_en, seo_keywords, schema_json_ld, generated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (vehicle_id) DO UPDATE SET
			title = EXCLUDED.title,
			description_es = EXCLUDED.description_es,
			description_en = EXCLUDED.description_en,
			seo_keywords = EXCLUDED.seo_keywords,
			schema_json_ld = EXCLUDED.schema_json_ld,
			generated_at = EXCLUDED.generated_at`
	_, err := r.db.ExecContext(ctx, query,
		l.ID, l.VehicleID, l.Title, l.DescriptionES, l.DescriptionEN,
		l.SEOKeywords, l.SchemaJSONLD, l.GeneratedAt,
	)
	return err
}

func (r *PostgresGeneratedListingRepository) GetByVehicleID(ctx context.Context, vehicleID string) (*diveinspect.GeneratedListing, error) {
	var l diveinspect.GeneratedListing
	query := `SELECT * FROM generated_listings WHERE vehicle_id = $1`
	if err := r.db.GetContext(ctx, &l, query, vehicleID); err != nil {
		return nil, errx.NotFound("Generated listing not found").WithDetail("vehicle_id", vehicleID)
	}
	return &l, nil
}

func (r *PostgresGeneratedListingRepository) Delete(ctx context.Context, vehicleID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM generated_listings WHERE vehicle_id = $1`, vehicleID)
	return err
}
