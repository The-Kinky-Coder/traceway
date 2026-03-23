//go:build !pgch

package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tracewayapp/lit/v2"
	"github.com/tracewayapp/traceway/backend/app/db"
	"github.com/tracewayapp/traceway/backend/app/models"
)

type FiredNotification struct {
	ProjectId   uuid.UUID
	RuleId      int
	RuleType    string
	RuleName    string
	ChannelType string
	ChannelName string
	Severity    string
	Subject     string
	Body        string
	Status      string
	ErrorMsg    string
	Endpoint    string
	FiredAt     time.Time
}

type firedNotificationRow struct {
	ProjectId   uuid.UUID  `lit:"project_id"`
	RuleId      int        `lit:"rule_id"`
	RuleType    string     `lit:"rule_type"`
	RuleName    string     `lit:"rule_name"`
	ChannelType string     `lit:"channel_type"`
	ChannelName string     `lit:"channel_name"`
	Severity    string     `lit:"severity"`
	Subject     string     `lit:"subject"`
	Body        string     `lit:"body"`
	Status      string     `lit:"status"`
	ErrorMsg    string     `lit:"error_message"`
	Endpoint    string     `lit:"endpoint"`
	FiredAt     SQLiteTime `lit:"fired_at"`
}

func init() {
	models.ExtensionModelRegistrations = append(models.ExtensionModelRegistrations, func(driver lit.Driver) {
		lit.RegisterModel[firedNotificationRow](driver)
	})
}

type firedNotificationRepository struct{}

func (r *firedNotificationRepository) Insert(ctx context.Context, n FiredNotification) error {
	row := firedNotificationRow{
		ProjectId:   n.ProjectId,
		RuleId:      n.RuleId,
		RuleType:    n.RuleType,
		RuleName:    n.RuleName,
		ChannelType: n.ChannelType,
		ChannelName: n.ChannelName,
		Severity:    n.Severity,
		Subject:     n.Subject,
		Body:        n.Body,
		Status:      n.Status,
		ErrorMsg:    n.ErrorMsg,
		Endpoint:    n.Endpoint,
		FiredAt:     NewSQLiteTime(n.FiredAt),
	}

	query, args, err := lit.ParseNamedQuery(db.Driver,
		`INSERT INTO fired_notifications (project_id, rule_id, rule_type, rule_name, channel_type, channel_name, severity, subject, body, status, error_message, endpoint, fired_at)
		VALUES (:project_id, :rule_id, :rule_type, :rule_name, :channel_type, :channel_name, :severity, :subject, :body, :status, :error_message, :endpoint, :fired_at)`,
		lit.P{
			"project_id":    row.ProjectId,
			"rule_id":       row.RuleId,
			"rule_type":     row.RuleType,
			"rule_name":     row.RuleName,
			"channel_type":  row.ChannelType,
			"channel_name":  row.ChannelName,
			"severity":      row.Severity,
			"subject":       row.Subject,
			"body":          row.Body,
			"status":        row.Status,
			"error_message": row.ErrorMsg,
			"endpoint":      row.Endpoint,
			"fired_at":      row.FiredAt,
		})
	if err != nil {
		return err
	}
	_, err = db.TelemetryDB.ExecContext(ctx, query, args...)
	return err
}

var FiredNotificationRepository = firedNotificationRepository{}
