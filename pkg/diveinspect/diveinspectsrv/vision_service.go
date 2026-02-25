package diveinspectsrv

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/Abraxas-365/divi/pkg/diveinspect"
	"github.com/Abraxas-365/divi/pkg/errx"
	"github.com/Abraxas-365/divi/pkg/fsx"
	"github.com/Abraxas-365/divi/pkg/logx"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
)

type VisionService struct {
	openaiClient   *openai.Client
	fs             fsx.FileSystem
	inspectionRepo diveinspect.InspectionRepository
	findingRepo    diveinspect.InspectionFindingRepository
	photoRepo      diveinspect.InspectionPhotoRepository
}

func NewVisionService(
	openaiClient *openai.Client,
	fs fsx.FileSystem,
	inspectionRepo diveinspect.InspectionRepository,
	findingRepo diveinspect.InspectionFindingRepository,
	photoRepo diveinspect.InspectionPhotoRepository,
) *VisionService {
	return &VisionService{
		openaiClient:   openaiClient,
		fs:             fs,
		inspectionRepo: inspectionRepo,
		findingRepo:    findingRepo,
		photoRepo:      photoRepo,
	}
}

type photoAnalysisResult struct {
	Score    int `json:"score"`
	Findings []struct {
		Type        string  `json:"type"`
		Severity    string  `json:"severity"`
		Location    string  `json:"location"`
		Description string  `json:"description"`
		Confidence  float64 `json:"confidence"`
	} `json:"findings"`
}

// RunInspection performs the visual AI inspection on all uploaded photos
func (s *VisionService) RunInspection(ctx context.Context, vehicle *diveinspect.Vehicle, inspectionID string) error {
	inspection, err := s.inspectionRepo.GetByID(ctx, inspectionID)
	if err != nil {
		return err
	}

	// Mark as processing
	inspection.Status = diveinspect.InspectionProcessing
	if err := s.inspectionRepo.Update(ctx, inspection); err != nil {
		return err
	}

	photos, err := s.photoRepo.GetByInspectionID(ctx, inspectionID)
	if err != nil {
		return errx.Wrap(err, "Failed to get inspection photos", errx.TypeInternal)
	}

	if len(photos) == 0 {
		return errx.Validation("No photos uploaded for this inspection")
	}

	logx.Infof("Running vision inspection %s with %d photos", inspectionID, len(photos))

	var allFindings []diveinspect.InspectionFinding
	exteriorScores := []int{}
	interiorScores := []int{}
	mechanicalScores := []int{}
	tireScores := []int{}

	for _, photo := range photos {
		zone := mapPhotoZoneToFindingZone(photo.Zone)
		result, err := s.analyzePhoto(ctx, vehicle, photo, zone)
		if err != nil {
			logx.Errorf("Failed to analyze photo %s: %v", photo.ID, err)
			continue
		}

		// Classify score by zone type
		switch {
		case isExteriorZone(photo.Zone):
			exteriorScores = append(exteriorScores, result.Score)
		case isInteriorZone(photo.Zone):
			interiorScores = append(interiorScores, result.Score)
		case photo.Zone == diveinspect.PhotoZoneEngine:
			mechanicalScores = append(mechanicalScores, result.Score)
		}

		// Convert findings
		for _, f := range result.Findings {
			desc := f.Description
			if f.Location != "" {
				desc = fmt.Sprintf("%s - %s", f.Location, f.Description)
			}
			photoURL := photo.PhotoURL
			confidence := f.Confidence

			finding := diveinspect.InspectionFinding{
				InspectionID: inspectionID,
				PhotoURL:     &photoURL,
				Zone:         zone,
				FindingType:  diveinspect.FindingType(f.Type),
				Severity:     diveinspect.FindingSeverity(f.Severity),
				Description:  &desc,
				AIConfidence: &confidence,
			}
			allFindings = append(allFindings, finding)
		}
	}

	// Calculate scores
	scoreExterior := avgScore(exteriorScores, 10)
	scoreInterior := avgScore(interiorScores, 10)
	scoreMechanical := avgScore(mechanicalScores, calculateMechanicalBase(vehicle))
	scoreTires := avgScore(tireScores, 8)

	// Overall: weighted average scaled to 1-100
	overall := int(math.Round(
		float64(scoreExterior)*0.35*10 +
			float64(scoreInterior)*0.30*10 +
			float64(scoreMechanical)*0.20*10 +
			float64(scoreTires)*0.15*10,
	))
	if overall > 100 {
		overall = 100
	}

	// Save findings
	if len(allFindings) > 0 {
		if err := s.findingRepo.CreateBatch(ctx, allFindings); err != nil {
			logx.Errorf("Failed to save findings: %v", err)
		}
	}

	// Update inspection
	now := time.Now()
	inspection.ScoreOverall = &overall
	inspection.ScoreExterior = &scoreExterior
	inspection.ScoreInterior = &scoreInterior
	inspection.ScoreMechanical = &scoreMechanical
	inspection.ScoreTires = &scoreTires
	inspection.FindingsCount = len(allFindings)
	inspection.PhotosCount = len(photos)
	inspection.Status = diveinspect.InspectionCompleted
	inspection.InspectedAt = &now

	if err := s.inspectionRepo.Update(ctx, inspection); err != nil {
		return errx.Wrap(err, "Failed to update inspection results", errx.TypeInternal)
	}

	logx.Infof("Inspection %s completed: score=%d, findings=%d", inspectionID, overall, len(allFindings))
	return nil
}

func (s *VisionService) analyzePhoto(ctx context.Context, vehicle *diveinspect.Vehicle, photo diveinspect.InspectionPhoto, zone diveinspect.FindingZone) (*photoAnalysisResult, error) {
	// Read the photo from filesystem
	photoData, err := s.fs.ReadFile(ctx, photo.PhotoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to read photo: %w", err)
	}

	base64Image := base64.StdEncoding.EncodeToString(photoData)
	imageURL := fmt.Sprintf("data:image/jpeg;base64,%s", base64Image)

	version := ""
	if vehicle.Version != nil {
		version = *vehicle.Version
	}

	systemPrompt := fmt.Sprintf(`You are a professional vehicle inspector for Divemotor, the leading automotive dealer in Peru.

Analyze this photo of the %s zone of a %s %s %s %d.

Evaluate and return JSON:
{
  "score": <1-10 integer, 10 being perfect condition>,
  "findings": [
    {
      "type": "scratch|dent|rust|paint_mismatch|wear|crack|stain|missing_part",
      "severity": "minor|moderate|major",
      "location": "descriptive location within the zone",
      "description": "detailed description of the finding in Spanish",
      "confidence": <0.0-1.0>
    }
  ]
}

Rules:
- Be precise but do not invent damage you cannot clearly see
- If the zone looks perfect, return score 10 with empty findings array
- Score 8-10: Excellent/Like new
- Score 6-7: Good with minor cosmetic issues
- Score 4-5: Fair with visible wear
- Score 1-3: Poor with significant damage
- Only return the JSON object`, zone, vehicle.Brand, vehicle.Model, version, vehicle.Year)

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemPrompt),
		{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfArrayOfContentParts: []openai.ChatCompletionContentPartUnionParam{
						{
							OfImageURL: &openai.ChatCompletionContentPartImageParam{
								ImageURL: openai.ChatCompletionContentPartImageImageURLParam{
									URL:    imageURL,
									Detail: "auto",
								},
							},
						},
						{
							OfText: &openai.ChatCompletionContentPartTextParam{
								Text: "Analyze this vehicle photo and provide your inspection findings as JSON.",
							},
						},
					},
				},
			},
		},
	}

	completion, err := s.openaiClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    "gpt-4o",
		Messages: messages,
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &shared.ResponseFormatJSONObjectParam{},
		},
		MaxTokens:   openai.Int(1024),
		Temperature: openai.Float(0.1),
	})
	if err != nil {
		return nil, fmt.Errorf("vision API call failed: %w", err)
	}

	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("no response from vision API")
	}

	var result photoAnalysisResult
	if err := json.Unmarshal([]byte(completion.Choices[0].Message.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse vision response: %w", err)
	}

	return &result, nil
}

func mapPhotoZoneToFindingZone(pz diveinspect.PhotoZone) diveinspect.FindingZone {
	switch pz {
	case diveinspect.PhotoZoneFront, diveinspect.PhotoZoneFrontLeft:
		return diveinspect.ZoneFront
	case diveinspect.PhotoZoneRear, diveinspect.PhotoZoneRearRight:
		return diveinspect.ZoneRear
	case diveinspect.PhotoZoneLeft:
		return diveinspect.ZoneLeft
	case diveinspect.PhotoZoneRight:
		return diveinspect.ZoneRight
	case diveinspect.PhotoZoneInteriorDriver, diveinspect.PhotoZoneDashboard, diveinspect.PhotoZoneInfotainment:
		return diveinspect.ZoneInteriorFront
	case diveinspect.PhotoZoneInteriorPassenger, diveinspect.PhotoZoneInteriorRear:
		return diveinspect.ZoneInteriorRear
	case diveinspect.PhotoZoneEngine:
		return diveinspect.ZoneEngine
	case diveinspect.PhotoZoneTrunk:
		return diveinspect.ZoneTrunk
	default:
		return diveinspect.ZoneFront
	}
}

func isExteriorZone(pz diveinspect.PhotoZone) bool {
	switch pz {
	case diveinspect.PhotoZoneFront, diveinspect.PhotoZoneRear,
		diveinspect.PhotoZoneLeft, diveinspect.PhotoZoneRight,
		diveinspect.PhotoZoneFrontLeft, diveinspect.PhotoZoneRearRight:
		return true
	}
	return false
}

func isInteriorZone(pz diveinspect.PhotoZone) bool {
	switch pz {
	case diveinspect.PhotoZoneInteriorDriver, diveinspect.PhotoZoneInteriorPassenger,
		diveinspect.PhotoZoneInteriorRear, diveinspect.PhotoZoneDashboard,
		diveinspect.PhotoZoneInfotainment:
		return true
	}
	return false
}

func avgScore(scores []int, defaultVal int) int {
	if len(scores) == 0 {
		return defaultVal
	}
	sum := 0
	for _, s := range scores {
		sum += s
	}
	return sum / len(scores)
}

func calculateMechanicalBase(vehicle *diveinspect.Vehicle) int {
	age := time.Now().Year() - vehicle.Year
	km := vehicle.MileageKM

	score := 10
	// Deduct for age
	if age > 5 {
		score -= 2
	} else if age > 2 {
		score -= 1
	}
	// Deduct for high mileage
	if km > 100000 {
		score -= 3
	} else if km > 50000 {
		score -= 2
	} else if km > 20000 {
		score -= 1
	}
	if score < 1 {
		score = 1
	}
	return score
}
