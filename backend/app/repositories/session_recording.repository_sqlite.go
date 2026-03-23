//go:build !pgch

package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/tracewayapp/lit/v2"
	"github.com/tracewayapp/traceway/backend/app/db"
	"github.com/tracewayapp/traceway/backend/app/models"
)

type sessionRecording struct {
	Id          uuid.UUID  `lit:"id"`
	ProjectId   uuid.UUID  `lit:"project_id"`
	ExceptionId uuid.UUID  `lit:"exception_id"`
	FilePath    string     `lit:"file_path"`
	RecordedAt  SQLiteTime `lit:"recorded_at"`
}

func init() {
	models.ExtensionModelRegistrations = append(models.ExtensionModelRegistrations, func(driver lit.Driver) {
		lit.RegisterModel[sessionRecording](driver)
	})
}

type sessionRecordingRepository struct{}

func (r *sessionRecordingRepository) InsertAsync(ctx context.Context, recordings []models.SessionRecording) error {
	if len(recordings) == 0 {
		return nil
	}

	for _, rec := range recordings {
		row := sessionRecording{
			Id:          rec.Id,
			ProjectId:   rec.ProjectId,
			ExceptionId: rec.ExceptionId,
			FilePath:    rec.FilePath,
			RecordedAt:  NewSQLiteTime(rec.RecordedAt),
		}
		if err := lit.InsertExistingUuid(db.TelemetryDB, &row); err != nil {
			return err
		}
	}

	return nil
}

func (r *sessionRecordingRepository) FindByExceptionId(ctx context.Context, projectId uuid.UUID, exceptionId uuid.UUID) (string, error) {
	result, err := lit.SelectSingleNamed[filePathResult](db.TelemetryDB,
		"SELECT file_path FROM session_recordings WHERE project_id = :project_id AND exception_id = :exception_id ORDER BY recorded_at DESC LIMIT 1",
		lit.P{"project_id": projectId, "exception_id": exceptionId})
	if err != nil {
		return "", err
	}
	if result == nil {
		return "", sql.ErrNoRows
	}
	return result.FilePath, nil
}

// Preserve original error contract: returns sql.ErrNoRows when not found
var _ = errors.Is

var SessionRecordingRepository = sessionRecordingRepository{}
