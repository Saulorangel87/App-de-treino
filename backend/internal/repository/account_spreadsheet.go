package repository

import (
	"context"
	"fmt"

	"github.com/Saulorangel87/App-de-treino/backend/internal/xlsx"
	"github.com/Saulorangel87/App-de-treino/backend/internal/zones"
	"github.com/jackc/pgx/v5"
)

// The spreadsheet carries the essentials an athlete wants to keep: who the
// account is and the training they did. It deliberately leaves out health data
// (limitations, pain, check-ins, body measurements). Dates and times are text in
// the Brazilian format (DD/MM/AAAA) and Brasília time; rows come already ordered.
const (
	brasiliaTime = `'DD/MM/YYYY HH24:MI'`
	brazilDate   = `'DD/MM/YYYY'`
)

const accountSheetQuery = `
	SELECT u.display_name, u.email,
		CASE WHEN u.email_verified_at IS NULL THEN 'Não' ELSE 'Sim' END,
		to_char(u.created_at AT TIME ZONE 'America/Sao_Paulo', ` + brasiliaTime + `),
		COALESCE(CASE ap.experience_level WHEN 'beginner' THEN 'Iniciante'
			WHEN 'intermediate' THEN 'Intermediário' WHEN 'advanced' THEN 'Avançado' END, ''),
		COALESCE(to_char(la.terms_version::date, ` + brazilDate + `), ''),
		COALESCE(to_char(la.accepted_at AT TIME ZONE 'America/Sao_Paulo', ` + brasiliaTime + `), ''),
		to_char(now() AT TIME ZONE 'America/Sao_Paulo', ` + brasiliaTime + `),
		(SELECT count(*) FROM workout_sessions ws
			JOIN athlete_profiles owner ON owner.id = ws.athlete_profile_id
			WHERE owner.user_id = u.id AND ws.status = 'completed')::int
	FROM users u
	LEFT JOIN athlete_profiles ap ON ap.user_id = u.id
	LEFT JOIN LATERAL (
		SELECT terms_version, accepted_at FROM legal_acceptances
		WHERE user_id = u.id ORDER BY accepted_at DESC LIMIT 1
	) la ON true
	WHERE u.id = $1`

const completedWorkoutsSheetQuery = `
	SELECT to_char(w.scheduled_on, ` + brazilDate + `), w.name, w.objective,
		w.duration_minutes, w.target_rpe::float8,
		to_char(ws.started_at AT TIME ZONE 'America/Sao_Paulo', ` + brasiliaTime + `),
		to_char(ws.completed_at AT TIME ZONE 'America/Sao_Paulo', ` + brasiliaTime + `),
		ws.duration_minutes,
		CASE ws.duration_source WHEN 'reported' THEN 'Informada por você' WHEN 'imported' THEN 'Arquivo importado' ELSE 'Cronômetro' END,
		ws.actual_rpe::float8, ws.distance_km::float8, ws.elevation_gain_m,
		ws.average_power_watts, ws.average_heart_rate, ws.average_cadence_rpm
	FROM workout_sessions ws
	JOIN workouts w ON w.id = ws.workout_id
	JOIN athlete_profiles ap ON ap.id = ws.athlete_profile_id
	WHERE ap.user_id = $1 AND ws.status = 'completed'
	ORDER BY w.scheduled_on, ws.completed_at`

const importedActivitiesSheetQuery = `
	SELECT to_char(ia.started_at AT TIME ZONE 'America/Sao_Paulo', ` + brasiliaTime + `), upper(ia.source),
		round(ia.moving_seconds / 60.0)::int, ia.distance_km::float8, ia.elevation_gain_m,
		ia.average_heart_rate, ia.max_heart_rate, ia.average_power_watts, ia.normalized_power_watts,
		ia.average_cadence_rpm, w.name
	FROM imported_activities ia
	LEFT JOIN workouts w ON w.id = ia.workout_id
	WHERE ia.user_id = $1
	ORDER BY ia.started_at`

var completedWorkoutsHeader = []string{
	"Data planejada", "Treino", "Objetivo do treino", "Duração planejada (min)", "Zona-alvo",
	"Início (horário de Brasília)", "Fim (horário de Brasília)", "Duração real (min)", "Origem da duração", "Zona percebida",
	"Distância (km)", "Altimetria (m)", "Potência média (W)", "FC média (bpm)", "Cadência média (rpm)",
}

var completedWorkoutsKinds = []xlsx.Kind{
	xlsx.KindCenter, xlsx.KindText, xlsx.KindWrap, xlsx.KindInt, xlsx.KindCenter,
	xlsx.KindCenter, xlsx.KindCenter, xlsx.KindInt, xlsx.KindCenter, xlsx.KindCenter,
	xlsx.KindDecimal, xlsx.KindInt, xlsx.KindInt, xlsx.KindInt, xlsx.KindInt,
}

var importedActivitiesKinds = []xlsx.Kind{
	xlsx.KindCenter, xlsx.KindCenter, xlsx.KindInt, xlsx.KindDecimal, xlsx.KindInt,
	xlsx.KindInt, xlsx.KindInt, xlsx.KindInt, xlsx.KindInt, xlsx.KindInt, xlsx.KindText,
}

var importedActivitiesHeader = []string{
	"Início (horário de Brasília)", "Origem", "Tempo em movimento (min)", "Distância (km)", "Altimetria (m)",
	"FC média (bpm)", "FC máxima (bpm)", "Potência média (W)", "Potência normalizada (W)", "Cadência média (rpm)",
	"Treino vinculado",
}

// ExportAccountSpreadsheet returns the sheets of the athlete's data spreadsheet.
// It reads inside one repeatable-read transaction so the sheets agree with each other.
func (s *Store) ExportAccountSpreadsheet(ctx context.Context, userID string) ([]xlsx.Sheet, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var name, email, verified, createdAt, level, termsVersion, termsAcceptedAt, generatedAt string
	var completedWorkouts int
	if err := tx.QueryRow(ctx, accountSheetQuery, userID).
		Scan(&name, &email, &verified, &createdAt, &level, &termsVersion, &termsAcceptedAt, &generatedAt, &completedWorkouts); err != nil {
		return nil, fmt.Errorf("export account sheet: %w", err)
	}
	account := xlsx.Sheet{Name: "Conta", Header: []string{"Campo", "Valor"}, Kinds: []xlsx.Kind{xlsx.KindLabel, xlsx.KindWrap}, Rows: [][]any{
		{"Nome", name},
		{"E-mail", email},
		{"E-mail confirmado", verified},
		{"Conta criada em (horário de Brasília)", createdAt},
		{"Nível de experiência", level},
		{"Treinos realizados", completedWorkouts},
		{"Versão dos termos aceita", termsVersion},
		{"Termos aceitos em (horário de Brasília)", termsAcceptedAt},
		{"Planilha gerada em (horário de Brasília)", generatedAt},
	}}

	workouts, err := queryRows(ctx, tx, completedWorkoutsSheetQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("export completed workouts: %w", err)
	}
	for _, row := range workouts {
		// The engine stores effort as RPE; the athlete reads it as a zone.
		row[4] = zoneLabelOf(row[4])
		row[9] = zoneLabelOf(row[9])
	}
	imported, err := queryRows(ctx, tx, importedActivitiesSheetQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("export imported activities: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return []xlsx.Sheet{
		account,
		{Name: "Treinos realizados", Header: completedWorkoutsHeader, Kinds: completedWorkoutsKinds, Rows: workouts, Filter: true},
		{Name: "Atividades importadas", Header: importedActivitiesHeader, Kinds: importedActivitiesKinds, Rows: imported, Filter: true},
	}, nil
}

func queryRows(ctx context.Context, tx pgx.Tx, query, userID string) ([][]any, error) {
	rows, err := tx.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := [][]any{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		result = append(result, values)
	}
	return result, rows.Err()
}

// zoneLabelOf turns a stored RPE (a float, or nil when not recorded) into "Z2 · Resistência".
func zoneLabelOf(value any) any {
	rpe, ok := value.(float64)
	if !ok {
		return nil
	}
	return zones.Label(rpe)
}
