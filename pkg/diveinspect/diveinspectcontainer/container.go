package diveinspectcontainer

import (
	"os"

	"github.com/Abraxas-365/divi/pkg/ai/llm"
	"github.com/Abraxas-365/divi/pkg/ai/providers/aiopenai"
	"github.com/Abraxas-365/divi/pkg/diveinspect/diveinspectapi"
	"github.com/Abraxas-365/divi/pkg/diveinspect/diveinspectinfra"
	"github.com/Abraxas-365/divi/pkg/diveinspect/diveinspectsrv"
	"github.com/Abraxas-365/divi/pkg/fsx"
	"github.com/Abraxas-365/divi/pkg/logx"
	"github.com/jmoiron/sqlx"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type Deps struct {
	DB         *sqlx.DB
	FileSystem fsx.FileSystem
}

type Container struct {
	Handlers *diveinspectapi.Handlers
}

func New(deps Deps) *Container {
	logx.Info("Initializing DiveInspect container...")

	c := &Container{}

	// ── Repositories ─────────────────────────────────────────────────────
	vehicleRepo := diveinspectinfra.NewPostgresVehicleRepository(deps.DB)
	specsRepo := diveinspectinfra.NewPostgresVehicleSpecsRepository(deps.DB)
	equipmentRepo := diveinspectinfra.NewPostgresVehicleEquipmentRepository(deps.DB)
	inspectionRepo := diveinspectinfra.NewPostgresInspectionRepository(deps.DB)
	findingRepo := diveinspectinfra.NewPostgresInspectionFindingRepository(deps.DB)
	photoRepo := diveinspectinfra.NewPostgresInspectionPhotoRepository(deps.DB)
	listingRepo := diveinspectinfra.NewPostgresGeneratedListingRepository(deps.DB)

	// ── AI Providers ─────────────────────────────────────────────────────
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")

	// LLM client for enrichment & listing generation
	openaiProvider := aiopenai.NewOpenAIProvider(openaiAPIKey)
	llmClient := llm.NewClient(openaiProvider)

	// Direct OpenAI client for vision (multi-modal)
	openaiClient := openai.NewClient(option.WithAPIKey(openaiAPIKey))

	// ── Services ─────────────────────────────────────────────────────────
	enrichmentSvc := diveinspectsrv.NewEnrichmentService(
		llmClient,
		specsRepo,
		equipmentRepo,
		listingRepo,
	)

	visionSvc := diveinspectsrv.NewVisionService(
		&openaiClient,
		deps.FileSystem,
		inspectionRepo,
		findingRepo,
		photoRepo,
	)

	inspectionSvc := diveinspectsrv.NewInspectionService(
		inspectionRepo,
		findingRepo,
		photoRepo,
		vehicleRepo,
		deps.FileSystem,
		visionSvc,
	)

	vehicleSvc := diveinspectsrv.NewVehicleService(
		vehicleRepo,
		specsRepo,
		equipmentRepo,
		listingRepo,
		inspectionRepo,
		findingRepo,
		photoRepo,
	)

	reportSvc := diveinspectsrv.NewReportService(
		vehicleRepo,
		specsRepo,
		equipmentRepo,
		inspectionRepo,
		findingRepo,
		deps.FileSystem,
	)

	// ── Handlers ─────────────────────────────────────────────────────────
	c.Handlers = diveinspectapi.NewHandlers(
		vehicleSvc,
		enrichmentSvc,
		inspectionSvc,
		reportSvc,
	)

	logx.Info("DiveInspect container initialized")
	return c
}
