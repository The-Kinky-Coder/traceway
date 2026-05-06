package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	Id                 uuid.UUID         `json:"id" ch:"id"`
	ProjectId          uuid.UUID         `json:"projectId" ch:"project_id"`
	StartedAt          time.Time         `json:"startedAt" ch:"started_at"`
	EndedAt            *time.Time        `json:"endedAt,omitempty" ch:"ended_at"`
	Duration           int64             `json:"duration" ch:"duration"`
	ClientIP           string            `json:"clientIP" ch:"client_ip"`
	Attributes         map[string]string `json:"attributes" ch:"attributes"`
	AppVersion         string            `json:"appVersion" ch:"app_version"`
	ServerName         string            `json:"serverName" ch:"server_name"`
	DistributedTraceId *uuid.UUID        `json:"distributedTraceId,omitempty" ch:"distributed_trace_id"`
}
