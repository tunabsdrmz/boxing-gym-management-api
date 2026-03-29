package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Announcement struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedBy string     `json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type AnnouncementRepository struct {
	db *sql.DB
}

type CreateAnnouncementRequest struct {
	Title     string
	Body      string
	ExpiresAt *time.Time
	CreatedBy string
}

func (a *AnnouncementRepository) Create(ctx context.Context, req CreateAnnouncementRequest) (Announcement, error) {
	var out Announcement
	var exp sql.NullTime
	err := a.db.QueryRowContext(ctx, `
		INSERT INTO announcements (title, body, expires_at, created_by)
		VALUES ($1, $2, $3, $4::uuid)
		RETURNING id::text, title, body, expires_at, created_by::text, created_at, updated_at
	`, req.Title, req.Body, req.ExpiresAt, req.CreatedBy).Scan(
		&out.ID, &out.Title, &out.Body, &exp, &out.CreatedBy, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return Announcement{}, err
	}
	if exp.Valid {
		t := exp.Time
		out.ExpiresAt = &t
	}
	return out, nil
}

func (a *AnnouncementRepository) ListActive(ctx context.Context) ([]Announcement, error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT id::text, title, body, expires_at, created_by::text, created_at, updated_at
		FROM announcements
		WHERE expires_at IS NULL OR expires_at > now()
		ORDER BY created_at DESC
		LIMIT 100
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAnnouncements(rows)
}

func (a *AnnouncementRepository) ListAll(ctx context.Context) ([]Announcement, error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT id::text, title, body, expires_at, created_by::text, created_at, updated_at
		FROM announcements
		ORDER BY created_at DESC
		LIMIT 200
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAnnouncements(rows)
}

func scanAnnouncements(rows *sql.Rows) ([]Announcement, error) {
	var list []Announcement
	for rows.Next() {
		var out Announcement
		var exp sql.NullTime
		if err := rows.Scan(&out.ID, &out.Title, &out.Body, &exp, &out.CreatedBy, &out.CreatedAt, &out.UpdatedAt); err != nil {
			return nil, err
		}
		if exp.Valid {
			t := exp.Time
			out.ExpiresAt = &t
		}
		list = append(list, out)
	}
	return list, rows.Err()
}

type UpdateAnnouncementRequest struct {
	ID            string
	Title         *string
	Body          *string
	ExpiresAt     *time.Time
	ClearExpires  bool
}

func (a *AnnouncementRepository) Update(ctx context.Context, req UpdateAnnouncementRequest) (Announcement, error) {
	cur, err := a.getByID(ctx, req.ID)
	if err != nil {
		return Announcement{}, err
	}
	title := cur.Title
	body := cur.Body
	exp := cur.ExpiresAt
	if req.Title != nil {
		title = *req.Title
	}
	if req.Body != nil {
		body = *req.Body
	}
	if req.ClearExpires {
		exp = nil
	} else if req.ExpiresAt != nil {
		exp = req.ExpiresAt
	}
	var out Announcement
	var expOut sql.NullTime
	err = a.db.QueryRowContext(ctx, `
		UPDATE announcements SET title = $1, body = $2, expires_at = $3, updated_at = now()
		WHERE id = $4::uuid
		RETURNING id::text, title, body, expires_at, created_by::text, created_at, updated_at
	`, title, body, exp, req.ID).Scan(
		&out.ID, &out.Title, &out.Body, &expOut, &out.CreatedBy, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return Announcement{}, err
	}
	if expOut.Valid {
		t := expOut.Time
		out.ExpiresAt = &t
	}
	return out, nil
}

func (a *AnnouncementRepository) getByID(ctx context.Context, id string) (Announcement, error) {
	var out Announcement
	var exp sql.NullTime
	err := a.db.QueryRowContext(ctx, `
		SELECT id::text, title, body, expires_at, created_by::text, created_at, updated_at
		FROM announcements WHERE id = $1::uuid
	`, id).Scan(&out.ID, &out.Title, &out.Body, &exp, &out.CreatedBy, &out.CreatedAt, &out.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Announcement{}, sql.ErrNoRows
	}
	if err != nil {
		return Announcement{}, err
	}
	if exp.Valid {
		t := exp.Time
		out.ExpiresAt = &t
	}
	return out, nil
}

func (a *AnnouncementRepository) Delete(ctx context.Context, id string) error {
	res, err := a.db.ExecContext(ctx, `DELETE FROM announcements WHERE id = $1::uuid`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
