package repository

import (
	"context"
	"database/sql"
	"time"
)

type AttendanceRecord struct {
	ID          string    `json:"id"`
	GymDate     string    `json:"gym_date"`
	FighterID   string    `json:"fighter_id"`
	Present     bool      `json:"present"`
	Notes       *string   `json:"notes,omitempty"`
	RecordedBy  string    `json:"recorded_by"`
	CreatedAt   time.Time `json:"created_at"`
	FighterName string    `json:"fighter_name,omitempty"`
}

type AttendanceRepository struct {
	db *sql.DB
}

type UpsertAttendanceRequest struct {
	GymDate    string
	FighterID  string
	Present    bool
	Notes      *string
	RecordedBy string
}

func (a *AttendanceRepository) Upsert(ctx context.Context, req UpsertAttendanceRequest) (AttendanceRecord, error) {
	var notes any
	if req.Notes != nil {
		notes = *req.Notes
	}
	var out AttendanceRecord
	var n sql.NullString
	err := a.db.QueryRowContext(ctx, `
		INSERT INTO attendance_records (gym_date, fighter_id, present, notes, recorded_by)
		VALUES ($1::date, $2::uuid, $3, $4, $5::uuid)
		ON CONFLICT (gym_date, fighter_id) DO UPDATE SET
			present = EXCLUDED.present,
			notes = EXCLUDED.notes,
			recorded_by = EXCLUDED.recorded_by,
			created_at = now()
		RETURNING id::text, gym_date::text, fighter_id::text, present, notes, recorded_by::text, created_at
	`, req.GymDate, req.FighterID, req.Present, notes, req.RecordedBy).Scan(
		&out.ID, &out.GymDate, &out.FighterID, &out.Present, &n, &out.RecordedBy, &out.CreatedAt,
	)
	if err != nil {
		return AttendanceRecord{}, err
	}
	if n.Valid {
		s := n.String
		out.Notes = &s
	}
	return out, nil
}

func (a *AttendanceRepository) ListByDate(ctx context.Context, gymDate string) ([]AttendanceRecord, error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT ar.id::text, ar.gym_date::text, ar.fighter_id::text, ar.present, ar.notes,
			ar.recorded_by::text, ar.created_at, f.name
		FROM attendance_records ar
		JOIN fighters f ON f.id = ar.fighter_id
		WHERE ar.gym_date = $1::date
		ORDER BY f.name ASC
	`, gymDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []AttendanceRecord
	for rows.Next() {
		var r AttendanceRecord
		var n sql.NullString
		if err := rows.Scan(&r.ID, &r.GymDate, &r.FighterID, &r.Present, &n, &r.RecordedBy, &r.CreatedAt, &r.FighterName); err != nil {
			return nil, err
		}
		if n.Valid {
			s := n.String
			r.Notes = &s
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

func (a *AttendanceRepository) Delete(ctx context.Context, id string) error {
	res, err := a.db.ExecContext(ctx, `DELETE FROM attendance_records WHERE id = $1::uuid`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
