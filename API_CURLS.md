# DiveInspect API — curl Reference

> Base URL: `http://localhost:8080/api/v1`
>
> The DiveInspect routes are **open (no auth)**.

---

## 1. Create Vehicle

```bash
curl -s -X POST http://localhost:8080/api/v1/vehicles \
  -H "Content-Type: application/json" \
  -d '{
    "brand": "Toyota",
    "model": "Corolla",
    "version": "SE",
    "trim": "Premium",
    "year": 2022,
    "mileage_km": 35000,
    "plate": "ABC-1234",
    "color_exterior": "White",
    "color_interior": "Black",
    "price_usd": 22000,
    "branch": "Main",
    "origin": "Japan"
  }'
```

**Response** `201 Created`

```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "plate": "ABC-1234",
  "brand": "Toyota",
  "model": "Corolla",
  "version": "SE",
  "trim": "Premium",
  "year": 2022,
  "mileage_km": 35000,
  "color_exterior": "White",
  "color_interior": "Black",
  "price_usd": 22000,
  "branch": "Main",
  "origin": "Japan",
  "status": "draft",
  "created_at": "2026-02-25T12:00:00Z",
  "updated_at": "2026-02-25T12:00:00Z"
}
```

> Save the `id` — all subsequent calls use it as `$VID`.

```bash
VID="a1b2c3d4-e5f6-7890-abcd-ef1234567890"
```

---

## 2. Get Vehicle

```bash
curl -s http://localhost:8080/api/v1/vehicles/$VID
```

**Response** `200 OK`

```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "plate": "ABC-1234",
  "brand": "Toyota",
  "model": "Corolla",
  "version": "SE",
  "trim": "Premium",
  "year": 2022,
  "mileage_km": 35000,
  "color_exterior": "White",
  "color_interior": "Black",
  "price_usd": 22000,
  "branch": "Main",
  "origin": "Japan",
  "status": "draft",
  "created_at": "2026-02-25T12:00:00Z",
  "updated_at": "2026-02-25T12:00:00Z"
}
```

---

## 3. List Vehicles

```bash
curl -s "http://localhost:8080/api/v1/vehicles?page=1&page_size=10"
```

**Response** `200 OK`

```json
{
  "items": [
    {
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "brand": "Toyota",
      "model": "Corolla",
      "year": 2022,
      "mileage_km": 35000,
      "status": "draft",
      "created_at": "2026-02-25T12:00:00Z",
      "updated_at": "2026-02-25T12:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total": 1
  }
}
```

---

## 4. Update Vehicle (partial)

```bash
curl -s -X PATCH http://localhost:8080/api/v1/vehicles/$VID \
  -H "Content-Type: application/json" \
  -d '{
    "mileage_km": 36000,
    "color_interior": "Beige",
    "price_usd": 21500
  }'
```

**Response** `200 OK`

```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "plate": "ABC-1234",
  "brand": "Toyota",
  "model": "Corolla",
  "version": "SE",
  "trim": "Premium",
  "year": 2022,
  "mileage_km": 36000,
  "color_exterior": "White",
  "color_interior": "Beige",
  "price_usd": 21500,
  "branch": "Main",
  "origin": "Japan",
  "status": "draft",
  "created_at": "2026-02-25T12:00:00Z",
  "updated_at": "2026-02-25T12:00:05Z"
}
```

---

## 5. Update Specs

```bash
curl -s -X PATCH http://localhost:8080/api/v1/vehicles/$VID/specs \
  -H "Content-Type: application/json" \
  -d '{
    "engine_type": "inline-4",
    "engine_cc": 1798,
    "engine_cylinders": 4,
    "power_hp": 139,
    "power_kw": 103.7,
    "torque_nm": 172,
    "fuel_type": "gasoline",
    "transmission_type": "CVT",
    "transmission_gears": 10,
    "drivetrain": "FWD",
    "accel_0_100": 9.5,
    "top_speed_kmh": 200,
    "fuel_city_kml": 12.3,
    "fuel_highway_kml": 17.8,
    "fuel_combined_kml": 14.5,
    "fuel_tank_liters": 50,
    "length_mm": 4630,
    "width_mm": 1780,
    "height_mm": 1435,
    "wheelbase_mm": 2700,
    "cargo_liters": 370,
    "curb_weight_kg": 1350,
    "tire_size": "205/55R16"
  }'
```

**Response** `200 OK`

```json
{
  "id": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
  "vehicle_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "engine_type": "inline-4",
  "engine_cc": 1798,
  "engine_cylinders": 4,
  "power_hp": 139,
  "power_kw": 103.7,
  "torque_nm": 172,
  "fuel_type": "gasoline",
  "transmission_type": "CVT",
  "transmission_gears": 10,
  "drivetrain": "FWD",
  "accel_0_100": 9.5,
  "top_speed_kmh": 200,
  "fuel_city_kml": 12.3,
  "fuel_highway_kml": 17.8,
  "fuel_combined_kml": 14.5,
  "fuel_tank_liters": 50,
  "length_mm": 4630,
  "width_mm": 1780,
  "height_mm": 1435,
  "wheelbase_mm": 2700,
  "cargo_liters": 370,
  "curb_weight_kg": 1350,
  "tire_size": "205/55R16"
}
```

---

## 6. Enrich Vehicle (AI — calls OpenAI)

Enriches specs and equipment via LLM lookup.

```bash
curl -s -X POST http://localhost:8080/api/v1/vehicles/$VID/enrich
```

**Response** `200 OK`

```json
{
  "vehicle": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "brand": "Toyota",
    "model": "Corolla",
    "year": 2022,
    "status": "draft"
  },
  "specs": {
    "vehicle_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "engine_type": "inline-4",
    "engine_cc": 1798,
    "power_hp": 139,
    "specs_source": "openai",
    "specs_confidence": 0.92,
    "enriched_at": "2026-02-25T12:01:00Z"
  },
  "equipment": [
    {
      "id": "c3d4e5f6-a7b8-9012-cdef-123456789012",
      "vehicle_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "category": "safety",
      "feature_name": "Toyota Safety Sense 2.0",
      "is_standard": true,
      "is_confirmed": false,
      "source": "factory_spec"
    },
    {
      "id": "d4e5f6a7-b8c9-0123-defa-234567890123",
      "vehicle_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "category": "infotainment",
      "feature_name": "8-inch touchscreen",
      "is_standard": true,
      "is_confirmed": false,
      "source": "factory_spec"
    }
  ],
  "listing": null
}
```

---

## 7. Upload Photo (multipart form)

Valid zones: `front`, `rear`, `left`, `right`, `front_left`, `rear_right`, `interior_driver`, `interior_passenger`, `interior_rear`, `dashboard`, `infotainment`, `engine`, `trunk`, `closeup`

```bash
curl -s -X POST http://localhost:8080/api/v1/vehicles/$VID/photos \
  -F "zone=front" \
  -F "photo=@./car-front.jpg" \
  -F "inspector_name=John Doe" \
  -F "inspector_branch=Main"
```

**Response** `201 Created`

```json
{
  "id": "e5f6a7b8-c9d0-1234-efab-345678901234",
  "inspection_id": "f6a7b8c9-d0e1-2345-fabc-456789012345",
  "photo_url": "inspections/f6a7b8c9/front_1234567890.jpg",
  "zone": "front",
  "sort_order": 0,
  "uploaded_at": "2026-02-25T12:02:00Z"
}
```

Upload more zones for a thorough inspection:

```bash
curl -s -X POST http://localhost:8080/api/v1/vehicles/$VID/photos \
  -F "zone=rear" \
  -F "photo=@./car-rear.jpg"

curl -s -X POST http://localhost:8080/api/v1/vehicles/$VID/photos \
  -F "zone=left" \
  -F "photo=@./car-left.jpg"

curl -s -X POST http://localhost:8080/api/v1/vehicles/$VID/photos \
  -F "zone=interior_driver" \
  -F "photo=@./car-interior.jpg"

curl -s -X POST http://localhost:8080/api/v1/vehicles/$VID/photos \
  -F "zone=engine" \
  -F "photo=@./car-engine.jpg"
```

---

## 8. Run Inspection (AI Vision — calls OpenAI GPT-4o)

Analyzes all uploaded photos and scores the vehicle condition.

> **Prerequisite:** At least one photo must be uploaded first.

```bash
curl -s -X POST http://localhost:8080/api/v1/vehicles/$VID/inspect
```

**Response** `200 OK`

```json
{
  "inspection": {
    "id": "f6a7b8c9-d0e1-2345-fabc-456789012345",
    "vehicle_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "inspector_name": "John Doe",
    "inspector_branch": "Main",
    "score_overall": 8,
    "score_exterior": 7,
    "score_interior": 9,
    "score_mechanical": 8,
    "score_tires": 8,
    "photos_count": 5,
    "findings_count": 2,
    "status": "completed",
    "inspected_at": "2026-02-25T12:03:00Z",
    "created_at": "2026-02-25T12:02:00Z",
    "updated_at": "2026-02-25T12:03:00Z"
  },
  "findings": [
    {
      "id": "a7b8c9d0-e1f2-3456-abcd-567890123456",
      "inspection_id": "f6a7b8c9-d0e1-2345-fabc-456789012345",
      "zone": "front",
      "finding_type": "scratch",
      "severity": "minor",
      "description": "Light scratch on front bumper, approximately 10cm",
      "ai_confidence": 0.87,
      "confirmed_by_human": false
    },
    {
      "id": "b8c9d0e1-f2a3-4567-bcde-678901234567",
      "inspection_id": "f6a7b8c9-d0e1-2345-fabc-456789012345",
      "zone": "left",
      "finding_type": "dent",
      "severity": "moderate",
      "description": "Small dent on left rear door panel",
      "ai_confidence": 0.92,
      "confirmed_by_human": false
    }
  ],
  "photos": [
    {
      "id": "e5f6a7b8-c9d0-1234-efab-345678901234",
      "inspection_id": "f6a7b8c9-d0e1-2345-fabc-456789012345",
      "photo_url": "inspections/f6a7b8c9/front_1234567890.jpg",
      "zone": "front",
      "sort_order": 0,
      "uploaded_at": "2026-02-25T12:02:00Z"
    }
  ]
}
```

---

## 9. Update a Finding

> Use a finding `id` from the inspection response.

```bash
FID="a7b8c9d0-e1f2-3456-abcd-567890123456"

curl -s -X PATCH http://localhost:8080/api/v1/findings/$FID \
  -H "Content-Type: application/json" \
  -d '{
    "severity": "moderate",
    "description": "Confirmed: light scratch on front bumper, 12cm",
    "confirmed_by_human": true
  }'
```

**Response** `200 OK`

```json
{
  "id": "a7b8c9d0-e1f2-3456-abcd-567890123456",
  "inspection_id": "f6a7b8c9-d0e1-2345-fabc-456789012345",
  "zone": "front",
  "finding_type": "scratch",
  "severity": "moderate",
  "description": "Confirmed: light scratch on front bumper, 12cm",
  "ai_confidence": 0.87,
  "confirmed_by_human": true
}
```

**All updatable fields:**

| Field              | Type    | Values                                                                                 |
|--------------------|---------|----------------------------------------------------------------------------------------|
| `zone`             | string  | `front`, `rear`, `left`, `right`, `roof`, `interior_front`, `interior_rear`, `engine`, `trunk`, `tires` |
| `finding_type`     | string  | `scratch`, `dent`, `rust`, `paint_mismatch`, `wear`, `crack`, `stain`, `missing_part`  |
| `severity`         | string  | `minor`, `moderate`, `major`                                                           |
| `description`      | string  | Free text                                                                              |
| `confirmed_by_human` | bool | `true` / `false`                                                                       |

---

## 10. Get Vehicle Preview

Full composite view: vehicle + specs + equipment + listing + inspection.

```bash
curl -s http://localhost:8080/api/v1/vehicles/$VID/preview
```

**Response** `200 OK`

```json
{
  "vehicle": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "brand": "Toyota",
    "model": "Corolla",
    "version": "SE",
    "year": 2022,
    "mileage_km": 36000,
    "status": "draft",
    "created_at": "2026-02-25T12:00:00Z",
    "updated_at": "2026-02-25T12:00:05Z"
  },
  "specs": {
    "vehicle_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "engine_type": "inline-4",
    "engine_cc": 1798,
    "power_hp": 139,
    "transmission_type": "CVT",
    "drivetrain": "FWD"
  },
  "equipment": [
    {
      "category": "safety",
      "feature_name": "Toyota Safety Sense 2.0",
      "is_standard": true,
      "source": "factory_spec"
    }
  ],
  "listing": null,
  "inspection": {
    "inspection": {
      "id": "f6a7b8c9-d0e1-2345-fabc-456789012345",
      "score_overall": 8,
      "status": "completed"
    },
    "findings": [],
    "photos": []
  }
}
```

---

## 11. Get Listing JSON

Same data as preview — useful for marketplace integrations.

```bash
curl -s http://localhost:8080/api/v1/vehicles/$VID/listing.json
```

**Response** `200 OK` — same shape as [Get Vehicle Preview](#10-get-vehicle-preview).

---

## 12. Get PDF Report

```bash
curl -s http://localhost:8080/api/v1/vehicles/$VID/report.pdf -o report.pdf
```

**Response** `200 OK` — binary PDF file.

```
Content-Type: application/pdf
Content-Disposition: inline; filename=inspection_report.pdf
```

---

## 13. Publish Vehicle

Changes status from `draft` → `published`.

```bash
curl -s -X POST http://localhost:8080/api/v1/vehicles/$VID/publish
```

**Response** `200 OK`

```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "brand": "Toyota",
  "model": "Corolla",
  "year": 2022,
  "status": "published",
  "created_at": "2026-02-25T12:00:00Z",
  "updated_at": "2026-02-25T12:05:00Z"
}
```

---

## 14. Delete Vehicle

```bash
curl -s -X DELETE http://localhost:8080/api/v1/vehicles/$VID -w "\n%{http_code}\n"
```

**Response** `204 No Content` — empty body.

---

## Quick Full Workflow

```bash
BASE="http://localhost:8080/api/v1"

# 1. Create
VID=$(curl -s -X POST $BASE/vehicles \
  -H "Content-Type: application/json" \
  -d '{"brand":"Toyota","model":"Corolla","year":2022,"mileage_km":35000}' \
  | jq -r '.id')
echo "Vehicle: $VID"

# 2. Enrich specs + equipment via AI
curl -s -X POST $BASE/vehicles/$VID/enrich | jq .

# 3. Upload photos
for zone in front rear left right interior_driver engine; do
  curl -s -X POST $BASE/vehicles/$VID/photos \
    -F "zone=$zone" \
    -F "photo=@./photos/${zone}.jpg" | jq .
done

# 4. Run AI inspection
curl -s -X POST $BASE/vehicles/$VID/inspect | jq .

# 5. Preview everything
curl -s $BASE/vehicles/$VID/preview | jq .

# 6. Download PDF report
curl -s $BASE/vehicles/$VID/report.pdf -o report.pdf

# 7. Publish
curl -s -X POST $BASE/vehicles/$VID/publish | jq .
```
