package activityimport

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"strings"
	"time"
)

const (
	SourceFIT = "fit"
	SourceGPX = "gpx"
)

// MaxFileSizeBytes bounds an upload, per docs/proxima-fase-dados-reais.md.
// A .fit from a several-hour ride rarely exceeds a few MB; 20 MB is generous.
const MaxFileSizeBytes = 20 << 20

var (
	ErrFileTooLarge       = errors.New("activityimport: arquivo maior que o limite permitido")
	ErrUnrecognizedFormat = errors.New("activityimport: envie um arquivo .fit ou .gpx")
	ErrDuplicate          = errors.New("activityimport: este arquivo já foi importado")
	ErrNotFound           = errors.New("activityimport: atividade importada não encontrada")
)

// Activity is an athlete's imported ride, stored independently of any
// planned workout. Importing never changes a training plan or a workout's
// prescription (see the invariant test in this package): it only becomes
// part of a workout's recorded execution if the athlete reviews the
// suggested match and confirms it through the existing completion/
// correction endpoints (POST /v1/workouts/{id}/complete or /correct).
type Activity struct {
	ID     string
	UserID string
	Parsed
	Source     string
	WorkoutID  *string
	ImportedAt time.Time
}

// WorkoutCandidate is the minimal projection of a planned workout offered as
// a suggested match for an imported activity, based on the same calendar
// date.
type WorkoutCandidate struct {
	ID          string
	ScheduledOn string
	Name        string
	Status      string
}

type Store interface {
	// ActivityExists reports whether this athlete already imported a file
	// with this exact content (by hash), so the same ride is never stored
	// twice.
	ActivityExists(ctx context.Context, userID, fileHash string) (bool, error)
	SaveActivity(ctx context.Context, userID, source, fileHash string, parsed Parsed) (Activity, error)
	ListActivities(ctx context.Context, userID string) ([]Activity, error)
	DeleteActivity(ctx context.Context, userID, activityID string) error
	// WorkoutCandidatesOnDate lists the athlete's planned workouts scheduled
	// for the given calendar date, used to suggest (never force) a match.
	WorkoutCandidatesOnDate(ctx context.Context, userID string, date time.Time) ([]WorkoutCandidate, error)
}

type Service struct{ store Store }

func NewService(store Store) *Service { return &Service{store: store} }

// Import reads, parses and stores an uploaded activity file. It returns the
// stored activity together with any planned workouts on the same date, for
// the frontend to suggest a link — the athlete decides, and any resulting
// change to a workout's recorded execution goes through the existing
// completion/correction endpoints, not through this package.
func (s *Service) Import(ctx context.Context, userID, filename string, r io.Reader) (Activity, []WorkoutCandidate, error) {
	limited := io.LimitReader(r, MaxFileSizeBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return Activity{}, nil, err
	}
	if len(data) > MaxFileSizeBytes {
		return Activity{}, nil, ErrFileTooLarge
	}
	source, parsed, err := parseFile(filename, data)
	if err != nil {
		return Activity{}, nil, err
	}
	sum := sha256.Sum256(data)
	fileHash := hex.EncodeToString(sum[:])
	exists, err := s.store.ActivityExists(ctx, userID, fileHash)
	if err != nil {
		return Activity{}, nil, err
	}
	if exists {
		return Activity{}, nil, ErrDuplicate
	}
	activity, err := s.store.SaveActivity(ctx, userID, source, fileHash, parsed)
	if err != nil {
		return Activity{}, nil, err
	}
	candidates, err := s.store.WorkoutCandidatesOnDate(ctx, userID, activity.StartedAt)
	if err != nil {
		return Activity{}, nil, err
	}
	return activity, candidates, nil
}

func (s *Service) List(ctx context.Context, userID string) ([]Activity, error) {
	return s.store.ListActivities(ctx, userID)
}

func (s *Service) Delete(ctx context.Context, userID, activityID string) error {
	return s.store.DeleteActivity(ctx, userID, activityID)
}

func parseFile(filename string, data []byte) (string, Parsed, error) {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".fit"):
		parsed, err := ParseFIT(bytes.NewReader(data))
		return SourceFIT, parsed, err
	case strings.HasSuffix(lower, ".gpx"):
		parsed, err := ParseGPX(bytes.NewReader(data))
		return SourceGPX, parsed, err
	default:
		return "", Parsed{}, ErrUnrecognizedFormat
	}
}
