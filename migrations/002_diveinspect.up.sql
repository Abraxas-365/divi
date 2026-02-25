-- ============================================================================
-- DiveInspect AI: Vehicle Inspection & Enrichment Platform
-- ============================================================================

-- ============================================================================
-- VEHICLES
-- ============================================================================

CREATE TABLE vehicles (
    id VARCHAR(255) PRIMARY KEY DEFAULT uuid_generate_v4()::text,
    plate VARCHAR(20),
    brand VARCHAR(100) NOT NULL,
    model VARCHAR(100) NOT NULL,
    version VARCHAR(200),
    trim VARCHAR(100),
    year INTEGER NOT NULL,
    mileage_km INTEGER NOT NULL DEFAULT 0,
    color_exterior VARCHAR(100),
    color_interior VARCHAR(100),
    price_usd DECIMAL(12,2),
    branch VARCHAR(200),
    origin VARCHAR(200),
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_vehicle_status CHECK (status IN ('draft', 'review', 'published')),
    CONSTRAINT chk_vehicle_year CHECK (year >= 1900 AND year <= 2100)
);

CREATE INDEX idx_vehicles_brand_model ON vehicles(brand, model);
CREATE INDEX idx_vehicles_year ON vehicles(year);
CREATE INDEX idx_vehicles_status ON vehicles(status);
CREATE INDEX idx_vehicles_plate ON vehicles(plate);
CREATE INDEX idx_vehicles_created_at ON vehicles(created_at);

-- ============================================================================
-- VEHICLE SPECS
-- ============================================================================

CREATE TABLE vehicle_specs (
    id VARCHAR(255) PRIMARY KEY DEFAULT uuid_generate_v4()::text,
    vehicle_id VARCHAR(255) NOT NULL UNIQUE,

    -- Engine
    engine_type VARCHAR(100),
    engine_cc INTEGER,
    engine_cylinders INTEGER,
    power_hp DECIMAL(8,2),
    power_kw DECIMAL(8,2),
    torque_nm INTEGER,
    torque_rpm_range VARCHAR(50),
    fuel_type VARCHAR(50),
    fuel_system VARCHAR(100),

    -- Transmission
    transmission_type VARCHAR(100),
    transmission_gears INTEGER,
    drivetrain VARCHAR(20),

    -- Performance
    accel_0_100 DECIMAL(5,2),
    top_speed_kmh INTEGER,

    -- Fuel consumption
    fuel_city_kml DECIMAL(5,2),
    fuel_highway_kml DECIMAL(5,2),
    fuel_combined_kml DECIMAL(5,2),
    fuel_tank_liters INTEGER,

    -- Dimensions
    length_mm INTEGER,
    width_mm INTEGER,
    height_mm INTEGER,
    wheelbase_mm INTEGER,
    cargo_liters INTEGER,
    cargo_max_liters INTEGER,
    curb_weight_kg INTEGER,

    -- Tires
    tire_size VARCHAR(50),
    spare_tire VARCHAR(100),

    -- Source info
    specs_source VARCHAR(500),
    specs_confidence DECIMAL(3,2),
    enriched_at TIMESTAMP,

    CONSTRAINT fk_vehicle_specs_vehicle FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE
);

CREATE INDEX idx_vehicle_specs_vehicle_id ON vehicle_specs(vehicle_id);

-- ============================================================================
-- VEHICLE EQUIPMENT
-- ============================================================================

CREATE TABLE vehicle_equipment (
    id VARCHAR(255) PRIMARY KEY DEFAULT uuid_generate_v4()::text,
    vehicle_id VARCHAR(255) NOT NULL,
    category VARCHAR(50) NOT NULL,
    feature_name VARCHAR(200) NOT NULL,
    feature_description TEXT,
    is_standard BOOLEAN NOT NULL DEFAULT TRUE,
    is_confirmed BOOLEAN NOT NULL DEFAULT FALSE,
    source VARCHAR(50) NOT NULL DEFAULT 'factory_spec',

    CONSTRAINT fk_vehicle_equipment_vehicle FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE,
    CONSTRAINT chk_equipment_category CHECK (category IN ('safety', 'comfort', 'infotainment', 'exterior', 'interior')),
    CONSTRAINT chk_equipment_source CHECK (source IN ('factory_spec', 'visual_detection', 'manual_input'))
);

CREATE INDEX idx_vehicle_equipment_vehicle_id ON vehicle_equipment(vehicle_id);
CREATE INDEX idx_vehicle_equipment_category ON vehicle_equipment(category);

-- ============================================================================
-- INSPECTIONS
-- ============================================================================

CREATE TABLE inspections (
    id VARCHAR(255) PRIMARY KEY DEFAULT uuid_generate_v4()::text,
    vehicle_id VARCHAR(255) NOT NULL,
    inspector_name VARCHAR(200),
    inspector_branch VARCHAR(200),

    -- Scores
    score_overall INTEGER,
    score_exterior INTEGER,
    score_interior INTEGER,
    score_mechanical INTEGER,
    score_tires INTEGER,

    -- Meta
    photos_count INTEGER NOT NULL DEFAULT 0,
    findings_count INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    pdf_url TEXT,
    inspected_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_inspections_vehicle FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE,
    CONSTRAINT chk_inspection_status CHECK (status IN ('pending', 'processing', 'completed', 'approved'))
);

CREATE INDEX idx_inspections_vehicle_id ON inspections(vehicle_id);
CREATE INDEX idx_inspections_status ON inspections(status);
CREATE INDEX idx_inspections_created_at ON inspections(created_at);

-- ============================================================================
-- INSPECTION FINDINGS
-- ============================================================================

CREATE TABLE inspection_findings (
    id VARCHAR(255) PRIMARY KEY DEFAULT uuid_generate_v4()::text,
    inspection_id VARCHAR(255) NOT NULL,
    photo_url TEXT,
    annotated_photo_url TEXT,
    zone VARCHAR(50) NOT NULL,
    finding_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    description TEXT,
    ai_confidence DECIMAL(3,2),
    confirmed_by_human BOOLEAN NOT NULL DEFAULT FALSE,

    CONSTRAINT fk_findings_inspection FOREIGN KEY (inspection_id) REFERENCES inspections(id) ON DELETE CASCADE,
    CONSTRAINT chk_finding_zone CHECK (zone IN ('front', 'rear', 'left', 'right', 'roof', 'interior_front', 'interior_rear', 'engine', 'trunk', 'tires')),
    CONSTRAINT chk_finding_type CHECK (finding_type IN ('scratch', 'dent', 'rust', 'paint_mismatch', 'wear', 'crack', 'stain', 'missing_part')),
    CONSTRAINT chk_finding_severity CHECK (severity IN ('minor', 'moderate', 'major'))
);

CREATE INDEX idx_findings_inspection_id ON inspection_findings(inspection_id);
CREATE INDEX idx_findings_zone ON inspection_findings(zone);
CREATE INDEX idx_findings_severity ON inspection_findings(severity);

-- ============================================================================
-- GENERATED LISTINGS
-- ============================================================================

CREATE TABLE generated_listings (
    id VARCHAR(255) PRIMARY KEY DEFAULT uuid_generate_v4()::text,
    vehicle_id VARCHAR(255) NOT NULL UNIQUE,
    title VARCHAR(500),
    description_es TEXT,
    description_en TEXT,
    seo_keywords TEXT[],
    schema_json_ld JSONB,
    generated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_listings_vehicle FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE
);

CREATE INDEX idx_listings_vehicle_id ON generated_listings(vehicle_id);

-- ============================================================================
-- INSPECTION PHOTOS (track uploaded photos)
-- ============================================================================

CREATE TABLE inspection_photos (
    id VARCHAR(255) PRIMARY KEY DEFAULT uuid_generate_v4()::text,
    inspection_id VARCHAR(255) NOT NULL,
    photo_url TEXT NOT NULL,
    zone VARCHAR(50) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    uploaded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_photos_inspection FOREIGN KEY (inspection_id) REFERENCES inspections(id) ON DELETE CASCADE,
    CONSTRAINT chk_photo_zone CHECK (zone IN ('front', 'rear', 'left', 'right', 'front_left', 'rear_right', 'interior_driver', 'interior_passenger', 'interior_rear', 'dashboard', 'infotainment', 'engine', 'trunk', 'closeup'))
);

CREATE INDEX idx_photos_inspection_id ON inspection_photos(inspection_id);

-- ============================================================================
-- TRIGGERS for updated_at
-- ============================================================================

CREATE TRIGGER update_vehicles_updated_at BEFORE UPDATE ON vehicles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_inspections_updated_at BEFORE UPDATE ON inspections
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE vehicles IS 'Vehicles registered for inspection and listing enrichment';
COMMENT ON TABLE vehicle_specs IS 'Factory specifications enriched via AI for each vehicle';
COMMENT ON TABLE vehicle_equipment IS 'Equipment features per vehicle, categorized';
COMMENT ON TABLE inspections IS 'Visual inspections performed on vehicles';
COMMENT ON TABLE inspection_findings IS 'Individual findings from visual inspection';
COMMENT ON TABLE generated_listings IS 'AI-generated listing content for vehicles';
COMMENT ON TABLE inspection_photos IS 'Photos uploaded for vehicle inspection';
