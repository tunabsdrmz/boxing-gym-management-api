package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Fighter struct {
	ID                      string                   `json:"id"`
	Name                    string                   `json:"name"`
	Age                     int                      `json:"age"`
	Weight                  float64                  `json:"weight"`
	Wins                    int                      `json:"wins"`
	Losses                  int                      `json:"losses"`
	TrainerID               string                   `json:"trainer_id"`
	HealthNotes             *string                  `json:"health_notes,omitempty"`
	ContractEnd             *string                  `json:"contract_end,omitempty"`
	EmergencyContactName    *string                  `json:"emergency_contact_name,omitempty"`
	EmergencyContactPhone   *string                  `json:"emergency_contact_phone,omitempty"`
	WeightClass             *string                  `json:"weight_class,omitempty"`
	FighterStatus           string                   `json:"fighter_status"`
	LicenseNumber           *string                  `json:"license_number,omitempty"`
	AssistantTrainers       []FighterAssistantTrainer `json:"assistant_trainers,omitempty"`
	CreatedAt               time.Time                `json:"created_at"`
	UpdatedAt               time.Time                `json:"updated_at"`
}

type FighterRepository struct {
	db *sql.DB
}

type CreateFighterRequest struct {
	Name                  string  `json:"name"`
	Age                   int     `json:"age"`
	Weight                float64 `json:"weight"`
	Wins                  int     `json:"wins"`
	Losses                int     `json:"losses"`
	TrainerID             string  `json:"trainer_id"`
	HealthNotes           *string `json:"health_notes"`
	ContractEnd           *string `json:"contract_end"`
	EmergencyContactName  *string `json:"emergency_contact_name"`
	EmergencyContactPhone *string `json:"emergency_contact_phone"`
	WeightClass           *string `json:"weight_class"`
	FighterStatus         *string `json:"fighter_status"`
	LicenseNumber         *string `json:"license_number"`
}

func (f *FighterRepository) CreateFighter(ctx context.Context, req CreateFighterRequest) (Fighter, error) {
	status := "amateur"
	if req.FighterStatus != nil && *req.FighterStatus != "" {
		status = *req.FighterStatus
	}
	var contract any
	if req.ContractEnd != nil && *req.ContractEnd != "" {
		contract = *req.ContractEnd
	}
	query := `
		INSERT INTO fighters (
			name, age, weight, wins, losses, trainer_id,
			health_notes, contract_end, emergency_contact_name, emergency_contact_phone,
			weight_class, fighter_status, license_number
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::date, $9, $10, $11, $12, $13)
		RETURNING id, name, age, weight, wins, losses, trainer_id,
			health_notes, contract_end::text, emergency_contact_name, emergency_contact_phone,
			weight_class, fighter_status, license_number, created_at, updated_at
	`
	row := f.db.QueryRowContext(ctx, query,
		req.Name, req.Age, req.Weight, req.Wins, req.Losses, req.TrainerID,
		nullStr(req.HealthNotes), contract, nullStr(req.EmergencyContactName), nullStr(req.EmergencyContactPhone),
		nullStr(req.WeightClass), status, nullStr(req.LicenseNumber),
	)
	return scanFighterRow(row)
}

func nullStr(p *string) any {
	if p == nil {
		return nil
	}
	if *p == "" {
		return nil
	}
	return *p
}

func scanFighterRow(row interface {
	Scan(dest ...any) error
}) (Fighter, error) {
	var fighter Fighter
	var hn, ecn, ecp, wc, lic sql.NullString
	var contractEnd sql.NullString
	err := row.Scan(
		&fighter.ID, &fighter.Name, &fighter.Age, &fighter.Weight, &fighter.Wins, &fighter.Losses, &fighter.TrainerID,
		&hn, &contractEnd, &ecn, &ecp, &wc, &fighter.FighterStatus, &lic, &fighter.CreatedAt, &fighter.UpdatedAt,
	)
	if err != nil {
		return Fighter{}, err
	}
	if hn.Valid {
		s := hn.String
		fighter.HealthNotes = &s
	}
	if contractEnd.Valid {
		s := contractEnd.String
		fighter.ContractEnd = &s
	}
	if ecn.Valid {
		s := ecn.String
		fighter.EmergencyContactName = &s
	}
	if ecp.Valid {
		s := ecp.String
		fighter.EmergencyContactPhone = &s
	}
	if wc.Valid {
		s := wc.String
		fighter.WeightClass = &s
	}
	if lic.Valid {
		s := lic.String
		fighter.LicenseNumber = &s
	}
	return fighter, nil
}

type GetFighterRequest struct {
	ID string `json:"id"`
}

func (f *FighterRepository) GetFighterByID(ctx context.Context, req GetFighterRequest) (Fighter, error) {
	query := `
		SELECT id, name, age, weight, wins, losses, trainer_id,
			health_notes, contract_end::text, emergency_contact_name, emergency_contact_phone,
			weight_class, fighter_status, license_number, created_at, updated_at
		FROM fighters
		WHERE id = $1
	`
	return scanFighterRow(f.db.QueryRowContext(ctx, query, req.ID))
}

type GetAllFightersRequest struct {
	Limit         string `json:"limit"`
	Offset        string `json:"offset"`
	Search        string
	WeightClass   string
	FighterStatus string
	SortField     string
	SortAsc       bool
}

type FighterListResult struct {
	Fighters []Fighter
	Total    int
}

func fighterFilterSQL(req GetAllFightersRequest) (where string, args []any) {
	parts := []string{"1=1"}
	args = []any{}
	i := 1
	if req.Search != "" {
		parts = append(parts, fmt.Sprintf("name ILIKE $%d", i))
		args = append(args, "%"+req.Search+"%")
		i++
	}
	if req.WeightClass != "" {
		parts = append(parts, fmt.Sprintf("weight_class ILIKE $%d", i))
		args = append(args, req.WeightClass)
		i++
	}
	if req.FighterStatus != "" {
		parts = append(parts, fmt.Sprintf("fighter_status = $%d", i))
		args = append(args, req.FighterStatus)
		i++
	}
	return strings.Join(parts, " AND "), args
}

func fighterOrderSQL(sortField string, asc bool) string {
	col := "created_at"
	switch sortField {
	case "name":
		col = "name"
	case "weight":
		col = "weight"
	case "created_at":
		col = "created_at"
	}
	dir := "DESC"
	if asc {
		dir = "ASC"
	}
	return col + " " + dir
}

func (f *FighterRepository) GetAllFighters(ctx context.Context, req GetAllFightersRequest) (FighterListResult, error) {
	whereSQL, whereArgs := fighterFilterSQL(req)
	countQuery := "SELECT COUNT(*) FROM fighters WHERE " + whereSQL
	var total int
	if err := f.db.QueryRowContext(ctx, countQuery, whereArgs...).Scan(&total); err != nil {
		return FighterListResult{}, err
	}

	li := len(whereArgs) + 1
	lo := len(whereArgs) + 2
	dataQuery := `
		SELECT id, name, age, weight, wins, losses, trainer_id,
			health_notes, contract_end::text, emergency_contact_name, emergency_contact_phone,
			weight_class, fighter_status, license_number, created_at, updated_at
		FROM fighters
		WHERE ` + whereSQL + `
		ORDER BY ` + fighterOrderSQL(req.SortField, req.SortAsc) + `
		LIMIT $` + strconv.Itoa(li) + ` OFFSET $` + strconv.Itoa(lo)

	args := append(append([]any{}, whereArgs...), req.Limit, req.Offset)
	rows, err := f.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return FighterListResult{}, err
	}
	defer rows.Close()
	var fighters []Fighter
	for rows.Next() {
		fighter, err := scanFighterRow(rows)
		if err != nil {
			return FighterListResult{}, err
		}
		fighters = append(fighters, fighter)
	}
	if err := rows.Err(); err != nil {
		return FighterListResult{}, err
	}
	return FighterListResult{Fighters: fighters, Total: total}, nil
}

type UpdateFighterRequest struct {
	ID                      string   `json:"id"`
	Name                    *string  `json:"name"`
	Age                     *int     `json:"age"`
	Weight                  *float64 `json:"weight"`
	Wins                    *int     `json:"wins"`
	Losses                  *int     `json:"losses"`
	TrainerID               *string  `json:"trainer_id"`
	HealthNotes             *string  `json:"health_notes"`
	ContractEnd             *string  `json:"contract_end"`
	EmergencyContactName    *string  `json:"emergency_contact_name"`
	EmergencyContactPhone   *string  `json:"emergency_contact_phone"`
	WeightClass             *string  `json:"weight_class"`
	FighterStatus           *string  `json:"fighter_status"`
	LicenseNumber           *string  `json:"license_number"`
}

func (f *FighterRepository) UpdateFighter(ctx context.Context, req UpdateFighterRequest) (Fighter, error) {
	var setParts []string
	var args []any
	n := 1

	add := func(col string, val any) {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", col, n))
		args = append(args, val)
		n++
	}

	if req.Name != nil {
		add("name", *req.Name)
	}
	if req.Age != nil {
		add("age", *req.Age)
	}
	if req.Weight != nil {
		add("weight", *req.Weight)
	}
	if req.Wins != nil {
		add("wins", *req.Wins)
	}
	if req.Losses != nil {
		add("losses", *req.Losses)
	}
	if req.TrainerID != nil {
		add("trainer_id", *req.TrainerID)
	}
	if req.HealthNotes != nil {
		add("health_notes", nullStr(req.HealthNotes))
	}
	if req.ContractEnd != nil {
		if *req.ContractEnd == "" {
			setParts = append(setParts, fmt.Sprintf("contract_end = $%d", n))
			args = append(args, nil)
			n++
		} else {
			setParts = append(setParts, fmt.Sprintf("contract_end = $%d::date", n))
			args = append(args, *req.ContractEnd)
			n++
		}
	}
	if req.EmergencyContactName != nil {
		add("emergency_contact_name", nullStr(req.EmergencyContactName))
	}
	if req.EmergencyContactPhone != nil {
		add("emergency_contact_phone", nullStr(req.EmergencyContactPhone))
	}
	if req.WeightClass != nil {
		add("weight_class", nullStr(req.WeightClass))
	}
	if req.FighterStatus != nil {
		add("fighter_status", *req.FighterStatus)
	}
	if req.LicenseNumber != nil {
		add("license_number", nullStr(req.LicenseNumber))
	}

	if len(setParts) == 0 {
		return f.GetFighterByID(ctx, GetFighterRequest{ID: req.ID})
	}

	setParts = append(setParts, "updated_at = now()")
	query := fmt.Sprintf(`
		UPDATE fighters
		SET %s
		WHERE id = $%d
		RETURNING id, name, age, weight, wins, losses, trainer_id,
			health_notes, contract_end::text, emergency_contact_name, emergency_contact_phone,
			weight_class, fighter_status, license_number, created_at, updated_at
	`, strings.Join(setParts, ", "), n)
	args = append(args, req.ID)
	return scanFighterRow(f.db.QueryRowContext(ctx, query, args...))
}

type DeleteFighterRequest struct {
	ID string `json:"id"`
}

func (f *FighterRepository) DeleteFighter(ctx context.Context, req DeleteFighterRequest) error {
	query := `DELETE FROM fighters WHERE id = $1`
	_, err := f.db.ExecContext(ctx, query, req.ID)
	return err
}
