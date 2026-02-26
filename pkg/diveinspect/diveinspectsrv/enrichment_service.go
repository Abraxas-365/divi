package diveinspectsrv

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Abraxas-365/divi/pkg/ai/llm"
	"github.com/Abraxas-365/divi/pkg/diveinspect"
	"github.com/Abraxas-365/divi/pkg/errx"
	"github.com/Abraxas-365/divi/pkg/logx"
)

type EnrichmentService struct {
	llmClient     *llm.Client
	specsRepo     diveinspect.VehicleSpecsRepository
	equipmentRepo diveinspect.VehicleEquipmentRepository
	listingRepo   diveinspect.GeneratedListingRepository
}

func NewEnrichmentService(
	llmClient *llm.Client,
	specsRepo diveinspect.VehicleSpecsRepository,
	equipmentRepo diveinspect.VehicleEquipmentRepository,
	listingRepo diveinspect.GeneratedListingRepository,
) *EnrichmentService {
	return &EnrichmentService{
		llmClient:     llmClient,
		specsRepo:     specsRepo,
		equipmentRepo: equipmentRepo,
		listingRepo:   listingRepo,
	}
}

// EnrichVehicle runs the full enrichment pipeline: specs + equipment + listing description
func (s *EnrichmentService) EnrichVehicle(ctx context.Context, vehicle *diveinspect.Vehicle) error {
	logx.Infof("Starting enrichment for vehicle %s: %s %s %d", vehicle.ID, vehicle.Brand, vehicle.Model, vehicle.Year)

	// Step 1: Enrich specs
	specs, err := s.enrichSpecs(ctx, vehicle)
	if err != nil {
		logx.Errorf("Failed to enrich specs for vehicle %s: %v", vehicle.ID, err)
		return errx.Wrap(err, "Failed to enrich vehicle specs", errx.TypeExternal)
	}

	if err := s.specsRepo.Upsert(ctx, specs); err != nil {
		return errx.Wrap(err, "Failed to save enriched specs", errx.TypeInternal)
	}
	logx.Infof("Specs enriched for vehicle %s", vehicle.ID)

	// Step 2: Enrich equipment
	equipment, err := s.enrichEquipment(ctx, vehicle)
	if err != nil {
		logx.Errorf("Failed to enrich equipment for vehicle %s: %v", vehicle.ID, err)
		return errx.Wrap(err, "Failed to enrich vehicle equipment", errx.TypeExternal)
	}

	// Replace existing equipment
	_ = s.equipmentRepo.DeleteByVehicleID(ctx, vehicle.ID)
	if err := s.equipmentRepo.CreateBatch(ctx, equipment); err != nil {
		return errx.Wrap(err, "Failed to save enriched equipment", errx.TypeInternal)
	}
	logx.Infof("Equipment enriched for vehicle %s: %d features", vehicle.ID, len(equipment))

	// Step 3: Generate listing description
	listing, err := s.generateListing(ctx, vehicle, specs, equipment)
	if err != nil {
		logx.Errorf("Failed to generate listing for vehicle %s: %v", vehicle.ID, err)
		return errx.Wrap(err, "Failed to generate listing", errx.TypeExternal)
	}

	if err := s.listingRepo.Upsert(ctx, listing); err != nil {
		return errx.Wrap(err, "Failed to save generated listing", errx.TypeInternal)
	}
	logx.Infof("Listing generated for vehicle %s", vehicle.ID)

	return nil
}

func (s *EnrichmentService) enrichSpecs(ctx context.Context, vehicle *diveinspect.Vehicle) (*diveinspect.VehicleSpecs, error) {
	version := ""
	if vehicle.Version != nil {
		version = *vehicle.Version
	}
	trim := ""
	if vehicle.Trim != nil {
		trim = *vehicle.Trim
	}

	prompt := fmt.Sprintf(`You are a vehicle specifications expert. Given the following vehicle identification, provide the complete factory specifications.

Vehicle: %s %s %s %s %d

Return a JSON object with these fields (use null for unknown values):
{
  "engine_type": "string (e.g. 'Inline-4 Turbo')",
  "engine_cc": number,
  "engine_cylinders": number,
  "power_hp": number,
  "power_kw": number,
  "torque_nm": number,
  "torque_rpm_range": "string (e.g. '1620-2600')",
  "fuel_type": "string",
  "fuel_system": "string",
  "transmission_type": "string (e.g. '7G-DCT Doble Embrague')",
  "transmission_gears": number,
  "drivetrain": "string (FWD/RWD/AWD/4WD)",
  "accel_0_100": number,
  "top_speed_kmh": number,
  "fuel_city_kml": number,
  "fuel_highway_kml": number,
  "fuel_combined_kml": number,
  "fuel_tank_liters": number,
  "length_mm": number,
  "width_mm": number,
  "height_mm": number,
  "wheelbase_mm": number,
  "cargo_liters": number,
  "cargo_max_liters": number,
  "curb_weight_kg": number,
  "tire_size": "string",
  "spare_tire": "string"
}

Only return the JSON object, no additional text.`, vehicle.Brand, vehicle.Model, version, trim, vehicle.Year)

	resp, err := s.llmClient.Chat(ctx, []llm.Message{
		llm.NewSystemMessage("You are a precise vehicle specifications database. Return only valid JSON with factory specs for the given vehicle. Be accurate and use real data from manufacturer catalogs."),
		llm.NewUserMessage(prompt),
	}, llm.WithJSONMode())
	if err != nil {
		logx.Errorf("AAAAAAAAaaa %v", err)
		return nil, err
	}
	logx.Infof("Received specs response for vehicle %s: %s", vehicle.ID, resp.Message.Content)

	var specsData struct {
		EngineType        *string  `json:"engine_type"`
		EngineCC          *int     `json:"engine_cc"`
		EngineCylinders   *int     `json:"engine_cylinders"`
		PowerHP           *float64 `json:"power_hp"`
		PowerKW           *float64 `json:"power_kw"`
		TorqueNM          *int     `json:"torque_nm"`
		TorqueRPMRange    *string  `json:"torque_rpm_range"`
		FuelType          *string  `json:"fuel_type"`
		FuelSystem        *string  `json:"fuel_system"`
		TransmissionType  *string  `json:"transmission_type"`
		TransmissionGears *int     `json:"transmission_gears"`
		Drivetrain        *string  `json:"drivetrain"`
		Accel0100         *float64 `json:"accel_0_100"`
		TopSpeedKMH       *int     `json:"top_speed_kmh"`
		FuelCityKML       *float64 `json:"fuel_city_kml"`
		FuelHighwayKML    *float64 `json:"fuel_highway_kml"`
		FuelCombinedKML   *float64 `json:"fuel_combined_kml"`
		FuelTankLiters    *int     `json:"fuel_tank_liters"`
		LengthMM          *int     `json:"length_mm"`
		WidthMM           *int     `json:"width_mm"`
		HeightMM          *int     `json:"height_mm"`
		WheelbaseMM       *int     `json:"wheelbase_mm"`
		CargoLiters       *int     `json:"cargo_liters"`
		CargoMaxLiters    *int     `json:"cargo_max_liters"`
		CurbWeightKG      *int     `json:"curb_weight_kg"`
		TireSize          *string  `json:"tire_size"`
		SpareTire         *string  `json:"spare_tire"`
	}

	if err := json.Unmarshal([]byte(resp.Message.Content), &specsData); err != nil {
		logx.Errorf("Failed to parse specs response for vehicle %s: %v. Response content: %s", vehicle.ID, err, resp.Message.Content)
		return nil, fmt.Errorf("failed to parse specs response: %w", err)
	}

	now := time.Now()
	source := "llm-enrichment"
	confidence := 0.85

	return &diveinspect.VehicleSpecs{
		VehicleID:         vehicle.ID,
		EngineType:        specsData.EngineType,
		EngineCC:          specsData.EngineCC,
		EngineCylinders:   specsData.EngineCylinders,
		PowerHP:           specsData.PowerHP,
		PowerKW:           specsData.PowerKW,
		TorqueNM:          specsData.TorqueNM,
		TorqueRPMRange:    specsData.TorqueRPMRange,
		FuelType:          specsData.FuelType,
		FuelSystem:        specsData.FuelSystem,
		TransmissionType:  specsData.TransmissionType,
		TransmissionGears: specsData.TransmissionGears,
		Drivetrain:        specsData.Drivetrain,
		Accel0100:         specsData.Accel0100,
		TopSpeedKMH:       specsData.TopSpeedKMH,
		FuelCityKML:       specsData.FuelCityKML,
		FuelHighwayKML:    specsData.FuelHighwayKML,
		FuelCombinedKML:   specsData.FuelCombinedKML,
		FuelTankLiters:    specsData.FuelTankLiters,
		LengthMM:          specsData.LengthMM,
		WidthMM:           specsData.WidthMM,
		HeightMM:          specsData.HeightMM,
		WheelbaseMM:       specsData.WheelbaseMM,
		CargoLiters:       specsData.CargoLiters,
		CargoMaxLiters:    specsData.CargoMaxLiters,
		CurbWeightKG:      specsData.CurbWeightKG,
		TireSize:          specsData.TireSize,
		SpareTire:         specsData.SpareTire,
		SpecsSource:       &source,
		SpecsConfidence:   &confidence,
		EnrichedAt:        &now,
	}, nil
}

func (s *EnrichmentService) enrichEquipment(ctx context.Context, vehicle *diveinspect.Vehicle) ([]diveinspect.VehicleEquipment, error) {
	version := ""
	if vehicle.Version != nil {
		version = *vehicle.Version
	}
	trim := ""
	if vehicle.Trim != nil {
		trim = *vehicle.Trim
	}

	// FIX 1: Update prompt to explicitly ask for a "features" key inside an object.
	// This is much more stable for LLM JSON mode than asking for a raw array.
	prompt := fmt.Sprintf(`You are a vehicle equipment expert. List ALL standard equipment features for:

Vehicle: %s %s %s %s %d

Return a JSON object containing a "features" array. The structure must be:
{
  "features": [
    {
      "category": "safety|comfort|infotainment|exterior|interior",
      "feature_name": "string",
      "feature_description": "string (brief description in Spanish)"
    }
  ]
}

Include ALL standard features for this specific trim level. Be comprehensive.
Only return the JSON object, no additional text.`, vehicle.Brand, vehicle.Model, version, trim, vehicle.Year)

	resp, err := s.llmClient.Chat(ctx, []llm.Message{
		llm.NewSystemMessage("You are a comprehensive vehicle equipment database. Return a valid JSON object with a 'features' list. Write descriptions in Spanish."),
		llm.NewUserMessage(prompt),
	}, llm.WithJSONMode(), llm.WithTemperature(0.1))
	if err != nil {
		return nil, err
	}

	// FIX 2: Create a wrapper struct to handle the root object
	var responseWrapper struct {
		Features []struct {
			Category    string `json:"category"`
			FeatureName string `json:"feature_name"`
			Description string `json:"feature_description"`
		} `json:"features"`
	}

	// Try to unmarshal into the wrapper first (standard behavior)
	if err := json.Unmarshal([]byte(resp.Message.Content), &responseWrapper); err != nil {
		// FALLBACK: If the LLM ignored instructions and sent a raw array, try unmarshalling directly
		// This handles the edge case where the LLM is stubborn.
		var rawFeatures []struct {
			Category    string `json:"category"`
			FeatureName string `json:"feature_name"`
			Description string `json:"feature_description"`
		}
		if errArray := json.Unmarshal([]byte(resp.Message.Content), &rawFeatures); errArray != nil {
			// Return original error if both fail
			logx.Errorf("Failed to parse equipment JSON. Content: %s", resp.Message.Content)
			return nil, fmt.Errorf("failed to parse equipment response: %w", err)
		}
		responseWrapper.Features = rawFeatures
	}

	equipment := make([]diveinspect.VehicleEquipment, 0, len(responseWrapper.Features))
	for _, item := range responseWrapper.Features {
		desc := item.Description
		eq := diveinspect.VehicleEquipment{
			VehicleID:          vehicle.ID,
			Category:           diveinspect.EquipmentCategory(item.Category),
			FeatureName:        item.FeatureName,
			FeatureDescription: &desc,
			IsStandard:         true,
			IsConfirmed:        false,
			Source:             diveinspect.SourceFactorySpec,
		}
		equipment = append(equipment, eq)
	}

	return equipment, nil
}

func (s *EnrichmentService) generateListing(ctx context.Context, vehicle *diveinspect.Vehicle, specs *diveinspect.VehicleSpecs, equipment []diveinspect.VehicleEquipment) (*diveinspect.GeneratedListing, error) {
	version := ""
	if vehicle.Version != nil {
		version = *vehicle.Version
	}
	color := ""
	if vehicle.ColorExterior != nil {
		color = *vehicle.ColorExterior
	}
	origin := ""
	if vehicle.Origin != nil {
		origin = *vehicle.Origin
	}
	branch := ""
	if vehicle.Branch != nil {
		branch = *vehicle.Branch
	}
	price := ""
	if vehicle.PriceUSD != nil {
		price = fmt.Sprintf("USD $%.0f", *vehicle.PriceUSD)
	}

	// Build equipment summary
	equipSummary := ""
	for _, eq := range equipment {
		equipSummary += fmt.Sprintf("- [%s] %s\n", eq.Category, eq.FeatureName)
	}

	// Build specs summary
	specsSummary := ""
	if specs != nil {
		if specs.PowerHP != nil {
			specsSummary += fmt.Sprintf("Motor: %.0f HP", *specs.PowerHP)
		}
		if specs.TransmissionType != nil {
			specsSummary += fmt.Sprintf(", Transmisión: %s", *specs.TransmissionType)
		}
		if specs.EngineCC != nil {
			specsSummary += fmt.Sprintf(", %d cc", *specs.EngineCC)
		}
	}

	prompt := fmt.Sprintf(`Generate a professional vehicle sales listing in Spanish for:

Vehicle: %s %s %s %d
Kilometraje: %d km
Color: %s
Origen: %s
Sede: %s
Precio: %s

Specs: %s

Equipment highlights:
%s

Generate a JSON response:
{
  "title": "SEO-optimized title (max 100 chars, include brand, model, year, key differentiator)",
  "description_es": "Professional sales description in Spanish (150-250 words). Highlight key features, low mileage if applicable, mention Divemotor warranty. Professional but attractive tone. SEO-friendly.",
  "description_en": "Same description in English",
  "seo_keywords": ["keyword1", "keyword2", ...]
}

Only return the JSON object.`, vehicle.Brand, vehicle.Model, version, vehicle.Year,
		vehicle.MileageKM, color, origin, branch, price, specsSummary, equipSummary)

	resp, err := s.llmClient.Chat(ctx, []llm.Message{
		llm.NewSystemMessage("You are an expert automotive copywriter for Divemotor, the leading Mercedes-Benz dealer in Peru. Write compelling, SEO-optimized vehicle listings. Always mention Garantía Divemotor. Return only valid JSON."),
		llm.NewUserMessage(prompt),
	}, llm.WithJSONMode(), llm.WithTemperature(0.5))
	if err != nil {
		return nil, err
	}

	var listingData struct {
		Title         string   `json:"title"`
		DescriptionES string   `json:"description_es"`
		DescriptionEN string   `json:"description_en"`
		SEOKeywords   []string `json:"seo_keywords"`
	}

	if err := json.Unmarshal([]byte(resp.Message.Content), &listingData); err != nil {
		return nil, fmt.Errorf("failed to parse listing response: %w", err)
	}

	return &diveinspect.GeneratedListing{
		VehicleID:     vehicle.ID,
		Title:         &listingData.Title,
		DescriptionES: &listingData.DescriptionES,
		DescriptionEN: &listingData.DescriptionEN,
		SEOKeywords:   listingData.SEOKeywords,
		GeneratedAt:   time.Now(),
	}, nil
}
