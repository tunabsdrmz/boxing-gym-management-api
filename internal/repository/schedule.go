package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"
)

type ScheduleEvent struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	StartAt      time.Time `json:"start_at"`
	EndAt        time.Time `json:"end_at"`
	ResourceType string    `json:"resource_type"`
	TrainerID    *string   `json:"trainer_id,omitempty"`
	Notes        *string   `json:"notes,omitempty"`
	CreatedBy    string    `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ScheduleRepository struct {
	db *sql.DB
}

type CreateScheduleEventRequest struct {
	Title        string
	StartAt      time.Time
	EndAt        time.Time
	ResourceType string
	TrainerID    *string
	Notes        *string
	CreatedBy    string
}

func (s *ScheduleRepository) HasOverlap(ctx context.Context, excludeID *string, resourceType string, start, end time.Time) (bool, error) {
	if !end.After(start) {
		return false, errors.New("end must be after start")
	}
	var q string
	var args []any
	if excludeID != nil && *excludeID != "" {
		q = `
			SELECT EXISTS (
				SELECT 1 FROM schedule_events
				WHERE resource_type = $1
				  AND id <> $2::uuid
				  AND start_at < $4 AND end_at > $3
			)`
		args = []any{resourceType, *excludeID, start, end}
	} else {
		q = `
			SELECT EXISTS (
				SELECT 1 FROM schedule_events
				WHERE resource_type = $1
				  AND start_at < $3 AND end_at > $2
			)`
		args = []any{resourceType, start, end}
	}
	var exists bool
	if err := s.db.QueryRowContext(ctx, q, args...).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (s *ScheduleRepository) Create(ctx context.Context, req CreateScheduleEventRequest) (ScheduleEvent, error) {
	rt := req.ResourceType
	if rt == "" {
		rt = "general"
	}
	trainerStr := ""
	if req.TrainerID != nil {
		trainerStr = *req.TrainerID
	}
	var notes any
	if req.Notes != nil {
		notes = *req.Notes
	}
	var out ScheduleEvent
	var tid, n sql.NullString
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO schedule_events (title, start_at, end_at, resource_type, trainer_id, notes, created_by)
		VALUES ($1, $2, $3, $4,
			CASE WHEN $5 = '' THEN NULL ELSE $5::uuid END,
			$6, $7::uuid)
		RETURNING id::text, title, start_at, end_at, resource_type,
			trainer_id::text, notes, created_by::text, created_at, updated_at
	`, req.Title, req.StartAt, req.EndAt, rt, trainerStr, notes, req.CreatedBy).Scan(
		&out.ID, &out.Title, &out.StartAt, &out.EndAt, &out.ResourceType, &tid, &n, &out.CreatedBy, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return ScheduleEvent{}, err
	}
	if tid.Valid {
		s := tid.String
		out.TrainerID = &s
	}
	if n.Valid {
		s := n.String
		out.Notes = &s
	}
	return out, nil
}

func (s *ScheduleRepository) GetByID(ctx context.Context, id string) (ScheduleEvent, error) {
	var out ScheduleEvent
	var tid, n sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT id::text, title, start_at, end_at, resource_type,
			trainer_id::text, notes, created_by::text, created_at, updated_at
		FROM schedule_events WHERE id = $1::uuid
	`, id).Scan(
		&out.ID, &out.Title, &out.StartAt, &out.EndAt, &out.ResourceType, &tid, &n, &out.CreatedBy, &out.CreatedAt, &out.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return ScheduleEvent{}, sql.ErrNoRows
	}
	if err != nil {
		return ScheduleEvent{}, err
	}
	if tid.Valid {
		s := tid.String
		out.TrainerID = &s
	}
	if n.Valid {
		s := n.String
		out.Notes = &s
	}
	return out, nil
}

type ListScheduleRequest struct {
	From *time.Time
	To   *time.Time
}

func (s *ScheduleRepository) List(ctx context.Context, req ListScheduleRequest) ([]ScheduleEvent, error) {
	q := `
		SELECT id::text, title, start_at, end_at, resource_type,
			trainer_id::text, notes, created_by::text, created_at, updated_at
		FROM schedule_events
		WHERE 1=1`
	var args []any
	n := 1
	if req.From != nil {
		q += ` AND end_at > $` + strconv.Itoa(n)
		args = append(args, *req.From)
		n++
	}
	if req.To != nil {
		q += ` AND start_at < $` + strconv.Itoa(n)
		args = append(args, *req.To)
		n++
	}
	q += ` ORDER BY start_at ASC`
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []ScheduleEvent
	for rows.Next() {
		var out ScheduleEvent
		var tid, notes sql.NullString
		if err := rows.Scan(&out.ID, &out.Title, &out.StartAt, &out.EndAt, &out.ResourceType, &tid, &notes, &out.CreatedBy, &out.CreatedAt, &out.UpdatedAt); err != nil {
			return nil, err
		}
		if tid.Valid {
			s := tid.String
			out.TrainerID = &s
		}
		if notes.Valid {
			s := notes.String
			out.Notes = &s
		}
		list = append(list, out)
	}
	return list, rows.Err()
}

type UpdateScheduleEventRequest struct {
	ID           string
	Title        *string
	StartAt      *time.Time
	EndAt        *time.Time
	ResourceType *string
	TrainerID    *string
	Notes        *string
}

func (s *ScheduleRepository) Update(ctx context.Context, req UpdateScheduleEventRequest) (ScheduleEvent, error) {
	cur, err := s.GetByID(ctx, req.ID)
	if err != nil {
		return ScheduleEvent{}, err
	}
	title := cur.Title
	start := cur.StartAt
	end := cur.EndAt
	resType := cur.ResourceType
	trainerStr := ""
	if cur.TrainerID != nil {
		trainerStr = *cur.TrainerID
	}
	var notesVal any
	if cur.Notes != nil {
		notesVal = *cur.Notes
	}

	if req.Title != nil {
		title = *req.Title
	}
	if req.StartAt != nil {
		start = *req.StartAt
	}
	if req.EndAt != nil {
		end = *req.EndAt
	}
	if req.ResourceType != nil {
		resType = *req.ResourceType
	}
	if req.TrainerID != nil {
		trainerStr = *req.TrainerID
	}
	if req.Notes != nil {
		notesVal = *req.Notes
	}

	var out ScheduleEvent
	var tid, n sql.NullString
	err = s.db.QueryRowContext(ctx, `
		UPDATE schedule_events
		SET title = $1, start_at = $2, end_at = $3, resource_type = $4,
			trainer_id = CASE WHEN $5 = '' THEN NULL ELSE $5::uuid END,
			notes = $6, updated_at = now()
		WHERE id = $7::uuid
		RETURNING id::text, title, start_at, end_at, resource_type,
			trainer_id::text, notes, created_by::text, created_at, updated_at
	`, title, start, end, resType, trainerStr, notesVal, req.ID).Scan(
		&out.ID, &out.Title, &out.StartAt, &out.EndAt, &out.ResourceType, &tid, &n, &out.CreatedBy, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return ScheduleEvent{}, err
	}
	if tid.Valid {
		s := tid.String
		out.TrainerID = &s
	}
	if n.Valid {
		s := n.String
		out.Notes = &s
	}
	return out, nil
}

func (s *ScheduleRepository) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM schedule_events WHERE id = $1::uuid`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
