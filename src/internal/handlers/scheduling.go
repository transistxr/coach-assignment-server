package handlers

import (
	"database/sql"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/transistxr/coach-assignment-server/src/internal/clients"
	"github.com/transistxr/coach-assignment-server/src/internal/db"
	"github.com/transistxr/coach-assignment-server/src/internal/structs"

	"log"
)

type HandlerDeps struct {
	DB                 *sql.DB
	RDB                *db.RedisClient
	AvailabilityClient *clients.AvailabilityClient
	CRMClient          *clients.CRMClient
	AuthClient         *clients.AuthClient
}

type SchedulingHandler struct {
	Deps *HandlerDeps
}

func splitInto15MinStarts(start, end time.Time) []time.Time {
	var slots []time.Time
	t := start.Truncate(time.Minute).UTC()
	end = end.UTC()
	for t.Before(end) {
		slots = append(slots, t)
		t = t.Add(15 * time.Minute)
	}
	return slots
}

// GetAvailability handles GET /api/availability.
// It calls the external calendar API to fetch availability for each coach,
// breaks the availability into 15-minute slots, and stores them in the
// `coach_slots` table. Returns available slots for the requested coach.
func (h *SchedulingHandler) GetAvailability(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	log.Println("Received GET Request: /api/availability")

	getAvailabilityResponse := &structs.AvailabilityResponse{}
	w.Header().Set("Content-Type", "application/json")

	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		w.WriteHeader(http.StatusBadRequest)
		getAvailabilityResponse.Error = "VALIDATION_ERROR"
		getAvailabilityResponse.Message = "X-API-Key Header is missing"
		json.NewEncoder(w).Encode(getAvailabilityResponse)
		return

	}
	authApiValidationResponse, err := h.Deps.AuthClient.ValidateKey(ctx, apiKey)

	log.Println("Auth,", authApiValidationResponse.Valid)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		getAvailabilityResponse.Error = "DOWNSTREAM_ERROR"
		getAvailabilityResponse.Message = "Downstream service failure"
		getAvailabilityResponse.ErrorDetails = err.Error()
		json.NewEncoder(w).Encode(getAvailabilityResponse)
		return
	}

	if !authApiValidationResponse.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		getAvailabilityResponse.Error = "VALIDATION_ERROR"
		getAvailabilityResponse.Message = "Invalid API Key"
		json.NewEncoder(w).Encode(getAvailabilityResponse)
		return
	}

	// parse days param
	days := 7
	if ds := r.URL.Query().Get("days"); ds != "" {
		if d, err := strconv.Atoi(ds); err == nil && d > 0 {
			days = d
		}
	}

	now := time.Now().UTC()
	windowEnd := now.AddDate(0, 0, days)

	// 1) load all coaches from DB
	rows, err := h.Deps.DB.QueryContext(ctx, `SELECT id FROM coaches`)
	if err != nil {
		log.Printf("GetAvailability: failed to query coaches: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		getAvailabilityResponse.Error = "INTERNAL_ERROR"
		getAvailabilityResponse.Message = "Unhandled error"
		getAvailabilityResponse.ErrorDetails = err.Error()
		json.NewEncoder(w).Encode(getAvailabilityResponse)
		return
	}
	defer rows.Close()

	var coachIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			getAvailabilityResponse.Error = "INTERNAL_ERROR"
			getAvailabilityResponse.Message = "Unhandled error"
			getAvailabilityResponse.ErrorDetails = err.Error()
			json.NewEncoder(w).Encode(getAvailabilityResponse)
			log.Printf("GetAvailability: scan coach id: %v", err)
			return
		}
		coachIDs = append(coachIDs, id)
	}

	for _, coachID := range coachIDs {
		avail, err := h.Deps.AvailabilityClient.GetAvailability(coachID, days)
		if err != nil {
			log.Printf("GetAvailability: calendar API failed for coach %s: %v", coachID, err)
			continue
		}

		for _, s := range avail.Slots {
			start := s.StartTime.UTC()
			end := s.EndTime.UTC()

			subStarts := splitInto15MinStarts(start, end)
			for _, st := range subStarts {
				_, err := h.Deps.DB.ExecContext(ctx, `
INSERT INTO coach_slots (coach_id, start_time, available)
VALUES ($1, $2, $3)
ON CONFLICT (coach_id, start_time) DO UPDATE
  SET available = CASE
    WHEN EXISTS (
      SELECT 1 FROM coach_appointments ca
      WHERE ca.coach_id = EXCLUDED.coach_id
        AND ca.start_time = EXCLUDED.start_time
        AND ca.status = 'scheduled'
    ) THEN false
    ELSE EXCLUDED.available
  END,
  updated_at = NOW()
`, coachID, st, s.Available)
				if err != nil {
					log.Printf("GetAvailability: upsert slot failed coach=%s start=%s: %v", coachID, st.Format(time.RFC3339), err)
					continue
				}
			}
		}
	}

	selRows, err := h.Deps.DB.QueryContext(ctx, `
SELECT coach_id, start_time
FROM coach_slots
WHERE start_time >= $1 AND start_time < $2 AND available = true
ORDER BY start_time, coach_id
`, now, windowEnd)
	if err != nil {
		http.Error(w, "failed to query slots", http.StatusInternalServerError)
		log.Printf("GetAvailability: select coach_slots: %v", err)
		return
	}
	defer selRows.Close()

	var respSlots []structs.AvailabilitySlot
	for selRows.Next() {
		var cid string
		var st time.Time
		if err := selRows.Scan(&cid, &st); err != nil {
			http.Error(w, "failed to scan slot", http.StatusInternalServerError)
			log.Printf("GetAvailability: scan slot: %v", err)
			return
		}
		respSlots = append(respSlots, structs.AvailabilitySlot{
			CoachID:   cid,
			StartTime: st.UTC(),
			EndTime:   st.Add(15 * time.Minute).UTC(),
		})
	}

	resp := structs.AvailabilityResponse{
		Slots:          respSlots,
		TotalAvailable: len(respSlots),
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("GetAvailability: encode response: %v", err)
	}
}

// BookAppointment handles POST /api/appointments.
// It checks coach eligibility by validating slot duration, appointment conflicts,
// and daily appointment limits. Then it selects the coach with the highest score,
// books the appointment in a transaction lock, updates slots, and writes to
// the distribution log. Also triggers downstream webhooks (CRM, Auth).
func (h *SchedulingHandler) BookAppointment(w http.ResponseWriter, r *http.Request) {


	log.Println("Received POST Request: /api/appointments")
	ctx := r.Context()
	var selectionReason string

	var req structs.BookAppointmentRequest

	appointmentBookingResponse := &structs.BookAppointmentResponse{}

	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		w.WriteHeader(http.StatusBadRequest)
		appointmentBookingResponse.Error = "VALIDATION_ERROR"
		appointmentBookingResponse.Message = "X-API-Key Header is missing"
		json.NewEncoder(w).Encode(appointmentBookingResponse)
		return

	}
	authApiValidationResponse, err := h.Deps.AuthClient.ValidateKey(ctx, apiKey)

	log.Println("Auth,", authApiValidationResponse.Valid)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		appointmentBookingResponse.Error = "DOWNSTREAM_ERROR"
		appointmentBookingResponse.Message = "Downstream service failure"
		appointmentBookingResponse.ErrorDetails = err.Error()
		json.NewEncoder(w).Encode(appointmentBookingResponse)
		return
	}

	if !authApiValidationResponse.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		appointmentBookingResponse.Error = "VALIDATION_ERROR"
		appointmentBookingResponse.Message = "Invalid API Key"
		json.NewEncoder(w).Encode(appointmentBookingResponse)
		return
	}


	w.Header().Add("Content-Type", "application/json")
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		appointmentBookingResponse.Error = "VALIDATION_ERROR"
		appointmentBookingResponse.Message = "Invalid request body"
		appointmentBookingResponse.ErrorDetails = err.Error()
		return
	}

	log.Println("Request:", req)

	tx, err := h.Deps.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		w.WriteHeader(http.StatusConflict)
		appointmentBookingResponse.Error = "TRANSACTION_ERROR"
		appointmentBookingResponse.Message = "Unable to create appointment"
		appointmentBookingResponse.ErrorDetails = err.Error()
		return
	}
	defer tx.Rollback()

	var calendarName string
	var slotDuration int
	err = tx.QueryRowContext(ctx,
		`SELECT name, slot_duration FROM calendars where id = $1`,
		req.CalendarID).Scan(&calendarName, &slotDuration)
	if err != nil {
		log.Println("No such calendar")
		w.WriteHeader(http.StatusBadRequest)
		appointmentBookingResponse.Error = "VALIDATION_ERROR"
		appointmentBookingResponse.Message = "Invalid calendar information"
		appointmentBookingResponse.ErrorDetails = err.Error()
		return
	}

	log.Printf("For calendar %s, with slot duration %d \n", calendarName, slotDuration)

	query := `
		SELECT
			c.id,
			c.name,
			c.email,
			c.score,
			c.max_daily_appointments,
			c.working_hours_start,
			c.working_hours_end,
			c.timezone
		FROM coaches c
		WHERE c.id IN (
			SELECT coach_id
			FROM coach_calendars cc
			WHERE cc.calendar_id = $1
		)
	`

	rows, err := tx.QueryContext(ctx, query, req.CalendarID)
	if err != nil {
	}
	defer rows.Close()

	var coaches []structs.Coach

	for rows.Next() {
		var coach structs.Coach
		err = rows.Scan(
			&coach.ID,
			&coach.Name,
			&coach.Email,
			&coach.Score,
			&coach.MaxDailyAppointments,
			&coach.WorkingHoursStart,
			&coach.WorkingHoursEnd,
			&coach.Timezone,
		)
		coaches = append(coaches, coach)
	}

	log.Println("Coaches with slot and calendar: ", coaches)

	if len(coaches) == 1 {
		selectionReason = "Only coach with this slot duration"
	}

	appointmentDay := req.StartTime.Format("2006-01-02")
	endTime := req.StartTime.Add(time.Duration(slotDuration) * time.Minute)

	query = `SELECT COUNT(*), c.max_daily_appointments
	FROM coach_appointments ca
	JOIN coaches c ON ca.coach_id = c.id
	WHERE ca.coach_id = $1
	  AND ca.start_time >= ($2::date + c.working_hours_start::interval)
	  AND ca.end_time <= ($2::date + c.working_hours_end::interval)
	  AND ca.status = 'scheduled'
	GROUP BY c.max_daily_appointments;
	`

	var coachLister []structs.Coach

	for i := 0; i < len(coaches); i++ {

		var currentAppointments, maxDaily int
		log.Printf("Iter %d, For %s", i, coaches[i].ID)
		err := tx.QueryRowContext(ctx, query, coaches[i].ID, appointmentDay).Scan(&currentAppointments, &maxDaily)
		if err != nil {
			if err == sql.ErrNoRows {
				coachLister = append(coachLister, coaches[i])
				continue
			}
			continue
		}
		if currentAppointments < maxDaily {
			coachLister = append(coachLister, coaches[i])
		}

	}

	log.Println("Coaches with slot and calendar: ", coachLister)

	if len(coachLister) == 1 {
		selectionReason = "Only coach not reaching daily appointment limit"
	}

	var newCoachList []structs.Coach
	for i := 0; i < len(coachLister); i++ {
		var isAvailable bool
		err := tx.QueryRowContext(ctx,
			`
    SELECT
        EXISTS (
            SELECT 1
            FROM coach_slots s
            WHERE s.coach_id = $1
              AND s.start_time = $2
              AND s.available = TRUE
        )
        AND NOT EXISTS (
            SELECT 1
            FROM coAch_appointments a
            WHERE a.coach_id = $1
              AND a.start_time = $2
              AND a.status = 'scheduled'
        ) AS is_available
`, coachLister[i].ID, req.StartTime).Scan(&isAvailable)

		if err != nil {
			// test
		}

		if isAvailable {
			newCoachList = append(newCoachList, coaches[i])
		}

	}

	log.Println("Coaches with slot and calendar: ", newCoachList)

	if len(newCoachList) == 0 {
		appointmentBookingResponse.Error = "NO_SLOT_ERROR"
		appointmentBookingResponse.Message = "No slot available at this time for any coach"
		json.NewEncoder(w).Encode(appointmentBookingResponse)
		return

	}

	if len(newCoachList) == 1 {
		selectionReason = "Only coach available at this time"
	}

	top := newCoachList[0]
	for _, c := range newCoachList[1:] {
		if c.Score > top.Score {
			top = c
		}
	}

	log.Printf("Best coach %s with score %f", top.Name, top.Score)

	selectionReason = "Selected coach with highest score"

	appointmentID := uuid.New().String()

	userContactID := "user-" + uuid.New().String()

	idemKey := r.Header.Get("X-Idempotency-Key")
	if idemKey == "" {
		idemKey = uuid.NewString()
	}

	appointmentCreationRequest := &structs.AppointmentCreatedRequest{
		AppointmentID: appointmentID,
		CoachID:       top.ID,
		StartTime:     req.StartTime,
		EndTime:       endTime,
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO coach_appointments (
			id, coach_id, calendar_id, contact_id, title, start_time, end_time, status, source
		) VALUES ($1,$2,$3,$4,$5, $6, $7, 'scheduled','api')
	`, appointmentID, top.ID, req.CalendarID, userContactID, req.Notes, req.StartTime, endTime)
	if err != nil {
		w.WriteHeader(http.StatusConflict)
		appointmentBookingResponse.Error = "NO_SLOT_ERROR"
		appointmentBookingResponse.Message = "This slot has already been scheduled."
		appointmentBookingResponse.ErrorDetails = err.Error()
		json.NewEncoder(w).Encode(appointmentBookingResponse)

		return
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE coach_slots SET available = false, updated_at = NOW() WHERE coach_id = $1
AND start_time >= $2 AND start_time < $3`, top.ID, req.StartTime, endTime)

	_, _ = tx.ExecContext(ctx, `
		INSERT INTO distribution_log (
			appointment_id, coaches_considered, selected_coach_id, selection_reason, distribution_score
		) VALUES ($1, $2, $3, $4, $5)
	`, appointmentID, `[]`, top.ID, selectionReason, 1.0)

	if err := tx.Commit(); err != nil {
		w.WriteHeader(http.StatusConflict)
		appointmentBookingResponse.Error = "TRANSACTION_ERROR"
		appointmentBookingResponse.Message = "Failed to create appointment"
		appointmentBookingResponse.ErrorDetails = err.Error()
		json.NewEncoder(w).Encode(appointmentBookingResponse)

		return
	}

	response, err := h.Deps.CRMClient.SendAppointmentCreated(ctx, appointmentCreationRequest, idemKey)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		appointmentBookingResponse.Error = "GATEWAY_ERROR"
		appointmentBookingResponse.Message = "Failed to send appointment creation request to CRM"
		appointmentBookingResponse.ErrorDetails = err.Error()
		json.NewEncoder(w).Encode(appointmentBookingResponse)

		return
	}

	_, err = h.Deps.DB.ExecContext(ctx, `
       UPDATE coach_appointments SET crm_contact_id = $1 WHERE id = $2
`, response.CrmID, appointmentID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		appointmentBookingResponse.Error = "INTERNAL_ERROR"
		appointmentBookingResponse.Message = "Database failure"
		appointmentBookingResponse.ErrorDetails = err.Error()
		json.NewEncoder(w).Encode(appointmentBookingResponse)

		return

	}

	blockSlotResponse, err := h.Deps.AvailabilityClient.BlockSlot(top.ID, req.StartTime, endTime)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		appointmentBookingResponse.Error = "GATEWAY_ERROR"
		appointmentBookingResponse.Message = "Failed to send a request to block slot to calendar"
		appointmentBookingResponse.ErrorDetails = err.Error()
		json.NewEncoder(w).Encode(appointmentBookingResponse)
		return

	}

	_, err = h.Deps.DB.ExecContext(ctx, `
       UPDATE coach_appointments SET external_calendar_id = $1 WHERE id = $2
`, blockSlotResponse.BlockId, appointmentID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		appointmentBookingResponse.Error = "INTERNAL_ERROR"
		appointmentBookingResponse.Message = "Database failure"
		appointmentBookingResponse.ErrorDetails = err.Error()
		json.NewEncoder(w).Encode(appointmentBookingResponse)

		return
	}

	resp := structs.BookAppointmentResponse{
		AppointmentID: appointmentID,
		CoachID:       top.ID,
		StartTime:     req.StartTime,
		EndTime:       endTime,
		Status:        "scheduled",
	}
	json.NewEncoder(w).Encode(resp)
}


// CalendarWebhook handles POST /api/webhooks/calendar.
// It receives calendar updates from the external Calendar API (e.g., slot blocked or freed).
// Updates the `coach_slots` and `coach_appointments`table accordingly to reflect real-time changes.
func (h *SchedulingHandler) WebhookHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	webHookResponse := &structs.WebHookResponse{}

	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		w.WriteHeader(http.StatusBadRequest)
		webHookResponse.Error = "VALIDATION_ERROR"
		webHookResponse.Message = "X-API-Key Header is missing"
		json.NewEncoder(w).Encode(webHookResponse)
		return

	}
	authApiValidationResponse, err := h.Deps.AuthClient.ValidateKey(ctx, apiKey)

	log.Println("Auth,", authApiValidationResponse.Valid)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		webHookResponse.Error = "DOWNSTREAM_ERROR"
		webHookResponse.Message = "Downstream service failure"
		webHookResponse.ErrorDetails = err.Error()
		json.NewEncoder(w).Encode(webHookResponse)
		return
	}

	if !authApiValidationResponse.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		webHookResponse.Error = "VALIDATION_ERROR"
		webHookResponse.Message = "Invalid API Key"
		json.NewEncoder(w).Encode(webHookResponse)
		return
	}

	idempotencyKey := r.Header.Get("X-Idempotency-Key")
	if idempotencyKey == "" {
		w.WriteHeader(http.StatusBadRequest)
		webHookResponse.Error = "VAIDATION_ERROR"
		webHookResponse.Message = "Invalid request header: No X-Idempotency-Key"
		json.NewEncoder(w).Encode(webHookResponse)
		return

	}

	var redisResponse structs.WebHookResponse
	keyExists, response := db.CheckIdempotency(h.Deps.RDB, idempotencyKey, redisResponse)

	if keyExists {
		log.Println("Duplicate request, returning previous response")
		json.NewEncoder(w).Encode(response)
	}

	var req structs.WebHookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		webHookResponse.Error = "VAIDATION_ERROR"
		webHookResponse.Message = "Invalid request body"
		webHookResponse.ErrorDetails = err.Error()
		json.NewEncoder(w).Encode(webHookResponse)

		return
	}

	appointmentQuery := `SELECT coach_id, calendar_id, start_time, end_time, external_calendar_id FROM coach_appointments WHERE id = $1`

	appointmentCancellationQuery := `UPDATE coach_appointments SET status = 'cancelled', updated_at = NOW(), cancelled_at = NOW(),
webhook_attempts = webhook_attempts + 1, webhook_last_attempt = NOW() WHERE id = $1`

	appointmentConfirmedQuery := `UPDATE coach_appointments SET updated_at = NOW(), confirmed_at = NOW(),
webhook_attempts = webhook_attempts + 1, webhook_last_attempt = NOW() WHERE id = $1`

	slotFreeingQuery := `UPDATE coach_slots SET available = true, updated_at = NOW() WHERE coach_id = $1
AND start_time >= $2 AND start_time < $3`

	insertWebhookEventQuery := `INSERT INTO webhook_events
(id, event_type, event_source, payload, last_attempt, processed_at, created_at)
VALUES ($1, $2, $3, $4, NOW(), NOW(), NOW())`

	b, err := json.Marshal(req)
	if err != nil {
		log.Println("Failed to marshall request")
		w.WriteHeader(http.StatusBadRequest)
		webHookResponse.Error = "VAIDATION_ERROR"
		webHookResponse.Message = "Invalid request body"
		webHookResponse.ErrorDetails = err.Error()
		json.NewEncoder(w).Encode(webHookResponse)

		return
	}

	eventId := uuid.New().String()

	_, err = h.Deps.DB.ExecContext(ctx, insertWebhookEventQuery, eventId, req.EventType, "webhook", b)
	if err != nil {
		log.Println("Failed to insert webhook entry")

	}

	var coachID string
	var calendarID string
	var blockID string
	var startTime time.Time
	var endTime time.Time

	err = h.Deps.DB.QueryRowContext(ctx, appointmentQuery, req.AppointmentID).Scan(&coachID, &calendarID, &startTime, &endTime, &blockID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		webHookResponse.Error = "INTERNAL_ERROR"
		webHookResponse.Message = "Database failure"
		webHookResponse.ErrorDetails = err.Error()
		json.NewEncoder(w).Encode(webHookResponse)

		return

	}

	switch req.EventType {
	case "appointment.cancelled":
		_, err = h.Deps.DB.ExecContext(ctx, appointmentCancellationQuery, req.AppointmentID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			webHookResponse.Error = "INTERNAL_ERROR"
			webHookResponse.Message = "Database failure"
			webHookResponse.ErrorDetails = err.Error()
			json.NewEncoder(w).Encode(webHookResponse)

			return

		}
		_, err = h.Deps.DB.ExecContext(ctx, slotFreeingQuery, coachID, startTime, endTime)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			webHookResponse.Error = "INTERNAL_ERROR"
			webHookResponse.Message = "Database failure"
			webHookResponse.ErrorDetails = err.Error()
			json.NewEncoder(w).Encode(webHookResponse)

			return

		}

		releaseSlotResponse, err := h.Deps.AvailabilityClient.ReleaseSlot(coachID, blockID, startTime, endTime)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			webHookResponse.Error = "GATEWAY_ERROR"
			webHookResponse.Message = "Failed to send a request to block slot to calendar"
			webHookResponse.ErrorDetails = err.Error()
			json.NewEncoder(w).Encode(webHookResponse)
			return
		}

		log.Printf("Response received %s", releaseSlotResponse.BlockId)

	case "appointment.confirmed":
		_, err = h.Deps.DB.ExecContext(ctx, appointmentConfirmedQuery, req.AppointmentID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			webHookResponse.Error = "INTERNAL_ERROR"
			webHookResponse.Message = "Database failure"
			webHookResponse.ErrorDetails = err.Error()
			json.NewEncoder(w).Encode(webHookResponse)

			return

		}

	}


	webHookResponse = &structs.WebHookResponse{
		Received: true,
		EventID:  eventId,
	}

	_ = db.SetIdempotencyKey(h.Deps.RDB, idempotencyKey, webHookResponse)

	json.NewEncoder(w).Encode(webHookResponse)

}


// GetCoachDistribution handles GET /api/coaches/distribution.
// It aggregates appointment counts and utilization metrics for each coach,
// computes the fairness_score across coaches, and returns distribution data.
func (h *SchedulingHandler) GetCoachDistribution(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	query := `
		SELECT id, name, email, score, max_daily_appointments FROM coaches`

	appointmentsCount := `SELECT COUNT(*) as appointments_count FROM coach_appointments
WHERE start_time::date = CURRENT_DATE and coach_id = $1`
	rows, err := h.Deps.DB.QueryContext(ctx, query)
	if err != nil {
	}
	defer rows.Close()

	var coaches []structs.Coach

	for rows.Next() {
		var coach structs.Coach
		err = rows.Scan(
			&coach.ID,
			&coach.Name,
			&coach.Email,
			&coach.Score,
			&coach.MaxDailyAppointments,
		)
		coaches = append(coaches, coach)
	}

	log.Println(coaches)

	var coachDistributionList []structs.CoachDistribution

	var response structs.CoachDistributionResponse

	var utilizationList []float64
	for _, coach := range coaches {
		var appointments int
		err = h.Deps.DB.QueryRowContext(ctx, appointmentsCount, coach.ID).Scan(&appointments)
		utilization := float64(appointments) / float64(coach.MaxDailyAppointments)
		utilizationList = append(utilizationList, utilization)
		coachDistribution := structs.CoachDistribution{
			CoachID:           coach.ID,
			Name:              coach.Name,
			Email:             coach.Email,
			Score:             coach.Score,
			AppointmentsCount: appointments,
			Utilization:       utilization,
		}
		coachDistributionList = append(coachDistributionList, coachDistribution)
	}

	response.Distribution = coachDistributionList
	response.FairnessScore = ComputeFairnessScore(utilizationList)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func ComputeFairnessScore(utilizations []float64) float64 {
	if len(utilizations) == 0 {
		return 0
	}

	var sum float64
	for _, u := range utilizations {
		sum += u
	}
	mean := sum / float64(len(utilizations))

	var variance float64
	for _, u := range utilizations {
		diff := u - mean
		variance += diff * diff
	}
	variance /= float64(len(utilizations))

	stdev := math.Sqrt(variance)

	score := 1 - stdev
	if score < 0 {
		score = 0
	}
	return score
}
