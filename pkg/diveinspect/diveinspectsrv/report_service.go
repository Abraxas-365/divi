package diveinspectsrv

import (
	"context"
	"fmt"
	"strings"

	"github.com/Abraxas-365/divi/pkg/diveinspect"
	"github.com/Abraxas-365/divi/pkg/errx"
	"github.com/Abraxas-365/divi/pkg/fsx"
	"github.com/Abraxas-365/divi/pkg/logx"
)

type ReportService struct {
	vehicleRepo    diveinspect.VehicleRepository
	specsRepo      diveinspect.VehicleSpecsRepository
	equipmentRepo  diveinspect.VehicleEquipmentRepository
	inspectionRepo diveinspect.InspectionRepository
	findingRepo    diveinspect.InspectionFindingRepository
	fs             fsx.FileSystem
}

func NewReportService(
	vehicleRepo diveinspect.VehicleRepository,
	specsRepo diveinspect.VehicleSpecsRepository,
	equipmentRepo diveinspect.VehicleEquipmentRepository,
	inspectionRepo diveinspect.InspectionRepository,
	findingRepo diveinspect.InspectionFindingRepository,
	fs fsx.FileSystem,
) *ReportService {
	return &ReportService{
		vehicleRepo:    vehicleRepo,
		specsRepo:      specsRepo,
		equipmentRepo:  equipmentRepo,
		inspectionRepo: inspectionRepo,
		findingRepo:    findingRepo,
		fs:             fs,
	}
}

// GenerateReport creates a PDF inspection report and returns the storage path
func (s *ReportService) GenerateReport(ctx context.Context, vehicleID string) (string, []byte, error) {
	vehicle, err := s.vehicleRepo.GetByID(ctx, vehicleID)
	if err != nil {
		return "", nil, err
	}

	specs, _ := s.specsRepo.GetByVehicleID(ctx, vehicleID)
	equipment, _ := s.equipmentRepo.GetByVehicleID(ctx, vehicleID)
	inspection, err := s.inspectionRepo.GetByVehicleID(ctx, vehicleID)
	if err != nil {
		return "", nil, errx.Wrap(err, "No inspection found for this vehicle", errx.TypeNotFound)
	}

	findings, _ := s.findingRepo.GetByInspectionID(ctx, inspection.ID)

	// Generate PDF
	pdfBytes := s.buildPDF(vehicle, specs, equipment, inspection, findings)

	// Store PDF
	storagePath := fmt.Sprintf("reports/%s/inspection_report.pdf", vehicleID)
	if err := s.fs.WriteFile(ctx, storagePath, pdfBytes); err != nil {
		logx.Errorf("Failed to store PDF report: %v", err)
	}

	// Update inspection with PDF URL
	inspection.PDFURL = &storagePath
	_ = s.inspectionRepo.Update(ctx, inspection)

	return storagePath, pdfBytes, nil
}

// buildPDF generates a minimal but valid PDF using raw PDF format.
// No external libraries required.
func (s *ReportService) buildPDF(vehicle *diveinspect.Vehicle, specs *diveinspect.VehicleSpecs, equipment []diveinspect.VehicleEquipment, inspection *diveinspect.Inspection, findings []diveinspect.InspectionFinding) []byte {
	p := newPDFWriter()

	// --- Page 1: Cover + Vehicle Info + Inspection Summary ---
	var coverLines []string
	coverLines = append(coverLines, "DIVEMOTOR - Reporte de Inspeccion Vehicular")
	coverLines = append(coverLines, strings.Repeat("=", 60))
	coverLines = append(coverLines, "")

	version := ""
	if vehicle.Version != nil {
		version = *vehicle.Version
	}
	coverLines = append(coverLines, fmt.Sprintf("Vehiculo: %s %s %s %d", vehicle.Brand, vehicle.Model, version, vehicle.Year))

	if inspection.ScoreOverall != nil {
		coverLines = append(coverLines, fmt.Sprintf("SCORE GENERAL: %d / 100", *inspection.ScoreOverall))
	}
	coverLines = append(coverLines, "")
	coverLines = append(coverLines, "--- Datos del Vehiculo ---")

	addField := func(label string, value *string) {
		if value != nil && *value != "" {
			coverLines = append(coverLines, fmt.Sprintf("  %-20s %s", label+":", *value))
		}
	}
	coverLines = append(coverLines, fmt.Sprintf("  %-20s %s", "Marca:", vehicle.Brand))
	coverLines = append(coverLines, fmt.Sprintf("  %-20s %s", "Modelo:", vehicle.Model))
	coverLines = append(coverLines, fmt.Sprintf("  %-20s %d", "Ano:", vehicle.Year))
	coverLines = append(coverLines, fmt.Sprintf("  %-20s %d km", "Kilometraje:", vehicle.MileageKM))
	addField("Placa", vehicle.Plate)
	addField("Color", vehicle.ColorExterior)
	if vehicle.PriceUSD != nil {
		coverLines = append(coverLines, fmt.Sprintf("  %-20s USD $%.0f", "Precio:", *vehicle.PriceUSD))
	}
	addField("Sede", vehicle.Branch)
	addField("Origen", vehicle.Origin)

	coverLines = append(coverLines, "")
	coverLines = append(coverLines, "--- Informacion de Inspeccion ---")
	if inspection.InspectorName != nil {
		coverLines = append(coverLines, fmt.Sprintf("  Inspector: %s", *inspection.InspectorName))
	}
	if inspection.InspectedAt != nil {
		coverLines = append(coverLines, fmt.Sprintf("  Fecha: %s", inspection.InspectedAt.Format("02/01/2006 15:04")))
	}
	coverLines = append(coverLines, fmt.Sprintf("  Fotos analizadas: %d", inspection.PhotosCount))
	coverLines = append(coverLines, fmt.Sprintf("  Hallazgos: %d", inspection.FindingsCount))

	coverLines = append(coverLines, "")
	coverLines = append(coverLines, "--- Puntajes por Area ---")
	if inspection.ScoreExterior != nil {
		coverLines = append(coverLines, fmt.Sprintf("  Exterior:     %d / 10", *inspection.ScoreExterior))
	}
	if inspection.ScoreInterior != nil {
		coverLines = append(coverLines, fmt.Sprintf("  Interior:     %d / 10", *inspection.ScoreInterior))
	}
	if inspection.ScoreMechanical != nil {
		coverLines = append(coverLines, fmt.Sprintf("  Mecanica:     %d / 10", *inspection.ScoreMechanical))
	}
	if inspection.ScoreTires != nil {
		coverLines = append(coverLines, fmt.Sprintf("  Neumaticos:   %d / 10", *inspection.ScoreTires))
	}

	p.addPage(strings.Join(coverLines, "\n"))

	// --- Page 2: Technical Specs ---
	if specs != nil {
		var specLines []string
		specLines = append(specLines, "FICHA TECNICA COMPLETA")
		specLines = append(specLines, strings.Repeat("=", 60))

		addSpec := func(label string, val *string) {
			if val != nil && *val != "" {
				specLines = append(specLines, fmt.Sprintf("  %-28s %s", label+":", *val))
			}
		}
		addSpecInt := func(label string, val *int, suffix string) {
			if val != nil {
				specLines = append(specLines, fmt.Sprintf("  %-28s %d%s", label+":", *val, suffix))
			}
		}
		addSpecFloat := func(label string, val *float64, suffix string) {
			if val != nil {
				specLines = append(specLines, fmt.Sprintf("  %-28s %.1f%s", label+":", *val, suffix))
			}
		}

		specLines = append(specLines, "")
		specLines = append(specLines, "Motor & Performance")
		specLines = append(specLines, strings.Repeat("-", 40))
		addSpec("Tipo de motor", specs.EngineType)
		addSpecInt("Cilindrada", specs.EngineCC, " cc")
		addSpecInt("Cilindros", specs.EngineCylinders, "")
		addSpecFloat("Potencia", specs.PowerHP, " HP")
		addSpecFloat("Potencia", specs.PowerKW, " kW")
		addSpecInt("Torque", specs.TorqueNM, " Nm")
		addSpec("Rango RPM torque", specs.TorqueRPMRange)
		addSpec("Combustible", specs.FuelType)
		addSpec("Sistema de combustible", specs.FuelSystem)
		addSpecFloat("0-100 km/h", specs.Accel0100, " s")
		addSpecInt("Velocidad maxima", specs.TopSpeedKMH, " km/h")

		specLines = append(specLines, "")
		specLines = append(specLines, "Transmision & Tren Motriz")
		specLines = append(specLines, strings.Repeat("-", 40))
		addSpec("Tipo de transmision", specs.TransmissionType)
		addSpecInt("Marchas", specs.TransmissionGears, "")
		addSpec("Traccion", specs.Drivetrain)

		specLines = append(specLines, "")
		specLines = append(specLines, "Dimensiones & Capacidades")
		specLines = append(specLines, strings.Repeat("-", 40))
		addSpecInt("Largo", specs.LengthMM, " mm")
		addSpecInt("Ancho", specs.WidthMM, " mm")
		addSpecInt("Alto", specs.HeightMM, " mm")
		addSpecInt("Distancia entre ejes", specs.WheelbaseMM, " mm")
		addSpecInt("Maletero", specs.CargoLiters, " litros")
		addSpecInt("Maletero max.", specs.CargoMaxLiters, " litros")
		addSpecInt("Peso en vacio", specs.CurbWeightKG, " kg")
		addSpec("Neumaticos", specs.TireSize)
		addSpec("Llanta de repuesto", specs.SpareTire)

		specLines = append(specLines, "")
		specLines = append(specLines, "Consumo de Combustible")
		specLines = append(specLines, strings.Repeat("-", 40))
		addSpecFloat("Ciudad", specs.FuelCityKML, " km/L")
		addSpecFloat("Carretera", specs.FuelHighwayKML, " km/L")
		addSpecFloat("Combinado", specs.FuelCombinedKML, " km/L")
		addSpecInt("Tanque", specs.FuelTankLiters, " litros")

		p.addPage(strings.Join(specLines, "\n"))
	}

	// --- Page 3: Equipment ---
	if len(equipment) > 0 {
		var eqLines []string
		eqLines = append(eqLines, "EQUIPAMIENTO DE SERIE")
		eqLines = append(eqLines, strings.Repeat("=", 60))

		categories := []struct {
			cat  diveinspect.EquipmentCategory
			name string
		}{
			{diveinspect.EquipmentSafety, "Seguridad"},
			{diveinspect.EquipmentComfort, "Confort"},
			{diveinspect.EquipmentInfotainment, "Infotainment & Conectividad"},
			{diveinspect.EquipmentExterior, "Exterior"},
			{diveinspect.EquipmentInterior, "Interior"},
		}

		for _, cat := range categories {
			var items []diveinspect.VehicleEquipment
			for _, eq := range equipment {
				if eq.Category == cat.cat {
					items = append(items, eq)
				}
			}
			if len(items) == 0 {
				continue
			}

			eqLines = append(eqLines, "")
			eqLines = append(eqLines, cat.name)
			eqLines = append(eqLines, strings.Repeat("-", 40))
			for _, item := range items {
				mark := "[OK]"
				if item.IsConfirmed {
					mark = "[OK*]"
				}
				eqLines = append(eqLines, fmt.Sprintf("  %s %s", mark, item.FeatureName))
			}
		}

		p.addPage(strings.Join(eqLines, "\n"))
	}

	// --- Page 4: Inspection Findings ---
	{
		var findLines []string
		findLines = append(findLines, "RESULTADOS DE INSPECCION VISUAL")
		findLines = append(findLines, strings.Repeat("=", 60))

		if len(findings) == 0 {
			findLines = append(findLines, "")
			findLines = append(findLines, "No se encontraron hallazgos significativos.")
			findLines = append(findLines, "El vehiculo se encuentra en excelente estado general.")
		} else {
			minorCount, moderateCount, majorCount := 0, 0, 0
			for _, f := range findings {
				switch f.Severity {
				case diveinspect.SeverityMinor:
					minorCount++
				case diveinspect.SeverityModerate:
					moderateCount++
				case diveinspect.SeverityMajor:
					majorCount++
				}
			}

			findLines = append(findLines, "")
			findLines = append(findLines, "Resumen de Hallazgos:")
			if majorCount > 0 {
				findLines = append(findLines, fmt.Sprintf("  MAYOR:    %d hallazgo(s)", majorCount))
			}
			if moderateCount > 0 {
				findLines = append(findLines, fmt.Sprintf("  MODERADO: %d hallazgo(s)", moderateCount))
			}
			if minorCount > 0 {
				findLines = append(findLines, fmt.Sprintf("  MENOR:    %d hallazgo(s)", minorCount))
			}

			findLines = append(findLines, "")
			findLines = append(findLines, "Detalle de Hallazgos:")
			findLines = append(findLines, strings.Repeat("-", 60))

			for i, f := range findings {
				findLines = append(findLines, "")
				header := fmt.Sprintf("#%d | Zona: %s | Tipo: %s | Severidad: %s",
					i+1, strings.ToUpper(string(f.Zone)), f.FindingType, strings.ToUpper(string(f.Severity)))
				findLines = append(findLines, header)
				if f.AIConfidence != nil {
					findLines = append(findLines, fmt.Sprintf("   Confianza IA: %.0f%%", *f.AIConfidence*100))
				}
				if f.Description != nil {
					findLines = append(findLines, fmt.Sprintf("   %s", *f.Description))
				}
				if f.ConfirmedByHuman {
					findLines = append(findLines, "   [Confirmado por inspector]")
				}
			}
		}

		findLines = append(findLines, "")
		findLines = append(findLines, strings.Repeat("-", 60))
		findLines = append(findLines, "Reporte generado por DiveInspect AI - Divemotor")

		p.addPage(strings.Join(findLines, "\n"))
	}

	return p.bytes()
}

// ============================================================================
// Minimal PDF Writer (no external dependency)
// ============================================================================

type pdfWriter struct {
	pages []string
}

func newPDFWriter() *pdfWriter {
	return &pdfWriter{}
}

func (w *pdfWriter) addPage(content string) {
	w.pages = append(w.pages, content)
}

// bytes produces a valid PDF 1.4 document with embedded text using Courier font.
func (w *pdfWriter) bytes() []byte {
	var b strings.Builder

	// Track object byte offsets for xref table
	offsets := []int{}

	// Header
	b.WriteString("%PDF-1.4\n")

	// Object 1: Catalog
	offsets = append(offsets, b.Len())
	b.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	// Object 2: Pages (will reference page objects)
	pageRefs := ""
	for i := range w.pages {
		objNum := 4 + i*2 // page objects start at 4, content at 5, next page at 6, etc.
		if i > 0 {
			pageRefs += " "
		}
		pageRefs += fmt.Sprintf("%d 0 R", objNum)
	}
	offsets = append(offsets, b.Len())
	b.WriteString(fmt.Sprintf("2 0 obj\n<< /Type /Pages /Kids [ %s ] /Count %d >>\nendobj\n", pageRefs, len(w.pages)))

	// Object 3: Font (Courier - built-in, no embedding needed)
	offsets = append(offsets, b.Len())
	b.WriteString("3 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Courier >>\nendobj\n")

	// Pages and their content streams
	nextObj := 4
	for _, pageContent := range w.pages {
		pageObj := nextObj
		contentObj := nextObj + 1
		nextObj += 2

		// Escape special PDF characters in text and build text stream
		stream := w.buildTextStream(pageContent)

		// Page object
		offsets = append(offsets, b.Len())
		b.WriteString(fmt.Sprintf("%d 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Contents %d 0 R /Resources << /Font << /F1 3 0 R >> >> >>\nendobj\n", pageObj, contentObj))

		// Content stream
		offsets = append(offsets, b.Len())
		b.WriteString(fmt.Sprintf("%d 0 obj\n<< /Length %d >>\nstream\n%s\nendstream\nendobj\n", contentObj, len(stream), stream))
	}

	// Cross-reference table
	xrefOffset := b.Len()
	totalObjects := len(offsets) + 1 // +1 for the free entry
	b.WriteString(fmt.Sprintf("xref\n0 %d\n", totalObjects))
	b.WriteString("0000000000 65535 f \n")
	for _, offset := range offsets {
		b.WriteString(fmt.Sprintf("%010d 00000 n \n", offset))
	}

	// Trailer
	b.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", totalObjects, xrefOffset))

	return []byte(b.String())
}

func (w *pdfWriter) buildTextStream(content string) string {
	var stream strings.Builder
	stream.WriteString("BT\n")
	stream.WriteString("/F1 9 Tf\n")       // Courier 9pt
	stream.WriteString("1 0 0 1 40 800 Tm\n") // Start position
	stream.WriteString("12 TL\n")           // Line spacing

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		escaped := pdfEscapeString(line)
		stream.WriteString(fmt.Sprintf("(%s) Tj T*\n", escaped))
	}

	stream.WriteString("ET")
	return stream.String()
}

func pdfEscapeString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	// Replace non-ASCII characters with ? to avoid encoding issues in basic PDF
	var result strings.Builder
	for _, r := range s {
		if r > 127 {
			result.WriteRune('?')
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
