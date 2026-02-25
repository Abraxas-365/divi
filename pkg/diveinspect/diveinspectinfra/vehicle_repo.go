package diveinspectinfra

import (
	"context"

	"github.com/Abraxas-365/divi/pkg/diveinspect"
	"github.com/Abraxas-365/divi/pkg/errx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PostgresVehicleRepository struct {
	db *sqlx.DB
}

func NewPostgresVehicleRepository(db *sqlx.DB) *PostgresVehicleRepository {
	return &PostgresVehicleRepository{db: db}
}

func (r *PostgresVehicleRepository) Create(ctx context.Context, v *diveinspect.Vehicle) error {
	if v.ID == "" {
		v.ID = uuid.New().String()
	}
	query := `
		INSERT INTO vehicles (id, plate, brand, model, version, trim, year, mileage_km, color_exterior, color_interior, price_usd, branch, origin, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING created_at, updated_at`
	return r.db.QueryRowContext(ctx, query,
		v.ID, v.Plate, v.Brand, v.Model, v.Version, v.Trim, v.Year, v.MileageKM,
		v.ColorExterior, v.ColorInterior, v.PriceUSD, v.Branch, v.Origin, v.Status,
	).Scan(&v.CreatedAt, &v.UpdatedAt)
}

func (r *PostgresVehicleRepository) GetByID(ctx context.Context, id string) (*diveinspect.Vehicle, error) {
	var v diveinspect.Vehicle
	query := `SELECT * FROM vehicles WHERE id = $1`
	if err := r.db.GetContext(ctx, &v, query, id); err != nil {
		return nil, errx.NotFound("Vehicle not found").WithDetail("id", id)
	}
	return &v, nil
}

func (r *PostgresVehicleRepository) Update(ctx context.Context, v *diveinspect.Vehicle) error {
	query := `
		UPDATE vehicles SET
			plate = $2, brand = $3, model = $4, version = $5, trim = $6, year = $7,
			mileage_km = $8, color_exterior = $9, color_interior = $10, price_usd = $11,
			branch = $12, origin = $13, status = $14
		WHERE id = $1
		RETURNING updated_at`
	return r.db.QueryRowContext(ctx, query,
		v.ID, v.Plate, v.Brand, v.Model, v.Version, v.Trim, v.Year, v.MileageKM,
		v.ColorExterior, v.ColorInterior, v.PriceUSD, v.Branch, v.Origin, v.Status,
	).Scan(&v.UpdatedAt)
}

func (r *PostgresVehicleRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM vehicles WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errx.NotFound("Vehicle not found").WithDetail("id", id)
	}
	return nil
}

func (r *PostgresVehicleRepository) List(ctx context.Context, page, pageSize int) ([]diveinspect.Vehicle, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM vehicles`); err != nil {
		return nil, 0, err
	}

	var vehicles []diveinspect.Vehicle
	query := `SELECT * FROM vehicles ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	offset := (page - 1) * pageSize
	if err := r.db.SelectContext(ctx, &vehicles, query, pageSize, offset); err != nil {
		return nil, 0, err
	}
	return vehicles, total, nil
}

func (r *PostgresVehicleRepository) ListByStatus(ctx context.Context, status diveinspect.VehicleStatus, page, pageSize int) ([]diveinspect.Vehicle, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM vehicles WHERE status = $1`, status); err != nil {
		return nil, 0, err
	}

	var vehicles []diveinspect.Vehicle
	query := `SELECT * FROM vehicles WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	offset := (page - 1) * pageSize
	if err := r.db.SelectContext(ctx, &vehicles, query, status, pageSize, offset); err != nil {
		return nil, 0, err
	}
	return vehicles, total, nil
}

// ============================================================================
// Vehicle Specs Repository
// ============================================================================

type PostgresVehicleSpecsRepository struct {
	db *sqlx.DB
}

func NewPostgresVehicleSpecsRepository(db *sqlx.DB) *PostgresVehicleSpecsRepository {
	return &PostgresVehicleSpecsRepository{db: db}
}

func (r *PostgresVehicleSpecsRepository) Create(ctx context.Context, s *diveinspect.VehicleSpecs) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	query := `
		INSERT INTO vehicle_specs (
			id, vehicle_id, engine_type, engine_cc, engine_cylinders, power_hp, power_kw,
			torque_nm, torque_rpm_range, fuel_type, fuel_system, transmission_type, transmission_gears,
			drivetrain, accel_0_100, top_speed_kmh, fuel_city_kml, fuel_highway_kml, fuel_combined_kml,
			fuel_tank_liters, length_mm, width_mm, height_mm, wheelbase_mm, cargo_liters,
			cargo_max_liters, curb_weight_kg, tire_size, spare_tire, specs_source, specs_confidence, enriched_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19,
			$20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32
		)`
	_, err := r.db.ExecContext(ctx, query,
		s.ID, s.VehicleID, s.EngineType, s.EngineCC, s.EngineCylinders, s.PowerHP, s.PowerKW,
		s.TorqueNM, s.TorqueRPMRange, s.FuelType, s.FuelSystem, s.TransmissionType, s.TransmissionGears,
		s.Drivetrain, s.Accel0100, s.TopSpeedKMH, s.FuelCityKML, s.FuelHighwayKML, s.FuelCombinedKML,
		s.FuelTankLiters, s.LengthMM, s.WidthMM, s.HeightMM, s.WheelbaseMM, s.CargoLiters,
		s.CargoMaxLiters, s.CurbWeightKG, s.TireSize, s.SpareTire, s.SpecsSource, s.SpecsConfidence, s.EnrichedAt,
	)
	return err
}

func (r *PostgresVehicleSpecsRepository) GetByVehicleID(ctx context.Context, vehicleID string) (*diveinspect.VehicleSpecs, error) {
	var s diveinspect.VehicleSpecs
	query := `SELECT * FROM vehicle_specs WHERE vehicle_id = $1`
	if err := r.db.GetContext(ctx, &s, query, vehicleID); err != nil {
		return nil, errx.NotFound("Vehicle specs not found").WithDetail("vehicle_id", vehicleID)
	}
	return &s, nil
}

func (r *PostgresVehicleSpecsRepository) Upsert(ctx context.Context, s *diveinspect.VehicleSpecs) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	query := `
		INSERT INTO vehicle_specs (
			id, vehicle_id, engine_type, engine_cc, engine_cylinders, power_hp, power_kw,
			torque_nm, torque_rpm_range, fuel_type, fuel_system, transmission_type, transmission_gears,
			drivetrain, accel_0_100, top_speed_kmh, fuel_city_kml, fuel_highway_kml, fuel_combined_kml,
			fuel_tank_liters, length_mm, width_mm, height_mm, wheelbase_mm, cargo_liters,
			cargo_max_liters, curb_weight_kg, tire_size, spare_tire, specs_source, specs_confidence, enriched_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19,
			$20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32
		)
		ON CONFLICT (vehicle_id) DO UPDATE SET
			engine_type = EXCLUDED.engine_type, engine_cc = EXCLUDED.engine_cc,
			engine_cylinders = EXCLUDED.engine_cylinders, power_hp = EXCLUDED.power_hp,
			power_kw = EXCLUDED.power_kw, torque_nm = EXCLUDED.torque_nm,
			torque_rpm_range = EXCLUDED.torque_rpm_range, fuel_type = EXCLUDED.fuel_type,
			fuel_system = EXCLUDED.fuel_system, transmission_type = EXCLUDED.transmission_type,
			transmission_gears = EXCLUDED.transmission_gears, drivetrain = EXCLUDED.drivetrain,
			accel_0_100 = EXCLUDED.accel_0_100, top_speed_kmh = EXCLUDED.top_speed_kmh,
			fuel_city_kml = EXCLUDED.fuel_city_kml, fuel_highway_kml = EXCLUDED.fuel_highway_kml,
			fuel_combined_kml = EXCLUDED.fuel_combined_kml, fuel_tank_liters = EXCLUDED.fuel_tank_liters,
			length_mm = EXCLUDED.length_mm, width_mm = EXCLUDED.width_mm,
			height_mm = EXCLUDED.height_mm, wheelbase_mm = EXCLUDED.wheelbase_mm,
			cargo_liters = EXCLUDED.cargo_liters, cargo_max_liters = EXCLUDED.cargo_max_liters,
			curb_weight_kg = EXCLUDED.curb_weight_kg, tire_size = EXCLUDED.tire_size,
			spare_tire = EXCLUDED.spare_tire, specs_source = EXCLUDED.specs_source,
			specs_confidence = EXCLUDED.specs_confidence, enriched_at = EXCLUDED.enriched_at`
	_, err := r.db.ExecContext(ctx, query,
		s.ID, s.VehicleID, s.EngineType, s.EngineCC, s.EngineCylinders, s.PowerHP, s.PowerKW,
		s.TorqueNM, s.TorqueRPMRange, s.FuelType, s.FuelSystem, s.TransmissionType, s.TransmissionGears,
		s.Drivetrain, s.Accel0100, s.TopSpeedKMH, s.FuelCityKML, s.FuelHighwayKML, s.FuelCombinedKML,
		s.FuelTankLiters, s.LengthMM, s.WidthMM, s.HeightMM, s.WheelbaseMM, s.CargoLiters,
		s.CargoMaxLiters, s.CurbWeightKG, s.TireSize, s.SpareTire, s.SpecsSource, s.SpecsConfidence, s.EnrichedAt,
	)
	return err
}

func (r *PostgresVehicleSpecsRepository) Delete(ctx context.Context, vehicleID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM vehicle_specs WHERE vehicle_id = $1`, vehicleID)
	return err
}

// ============================================================================
// Vehicle Equipment Repository
// ============================================================================

type PostgresVehicleEquipmentRepository struct {
	db *sqlx.DB
}

func NewPostgresVehicleEquipmentRepository(db *sqlx.DB) *PostgresVehicleEquipmentRepository {
	return &PostgresVehicleEquipmentRepository{db: db}
}

func (r *PostgresVehicleEquipmentRepository) CreateBatch(ctx context.Context, equipment []diveinspect.VehicleEquipment) error {
	if len(equipment) == 0 {
		return nil
	}
	query := `
		INSERT INTO vehicle_equipment (id, vehicle_id, category, feature_name, feature_description, is_standard, is_confirmed, source)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i := range equipment {
		if equipment[i].ID == "" {
			equipment[i].ID = uuid.New().String()
		}
		_, err := tx.ExecContext(ctx, query,
			equipment[i].ID, equipment[i].VehicleID, equipment[i].Category,
			equipment[i].FeatureName, equipment[i].FeatureDescription,
			equipment[i].IsStandard, equipment[i].IsConfirmed, equipment[i].Source,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *PostgresVehicleEquipmentRepository) GetByVehicleID(ctx context.Context, vehicleID string) ([]diveinspect.VehicleEquipment, error) {
	var equipment []diveinspect.VehicleEquipment
	query := `SELECT * FROM vehicle_equipment WHERE vehicle_id = $1 ORDER BY category, feature_name`
	if err := r.db.SelectContext(ctx, &equipment, query, vehicleID); err != nil {
		return nil, err
	}
	return equipment, nil
}

func (r *PostgresVehicleEquipmentRepository) GetByVehicleIDAndCategory(ctx context.Context, vehicleID string, category diveinspect.EquipmentCategory) ([]diveinspect.VehicleEquipment, error) {
	var equipment []diveinspect.VehicleEquipment
	query := `SELECT * FROM vehicle_equipment WHERE vehicle_id = $1 AND category = $2 ORDER BY feature_name`
	if err := r.db.SelectContext(ctx, &equipment, query, vehicleID, category); err != nil {
		return nil, err
	}
	return equipment, nil
}

func (r *PostgresVehicleEquipmentRepository) DeleteByVehicleID(ctx context.Context, vehicleID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM vehicle_equipment WHERE vehicle_id = $1`, vehicleID)
	return err
}
