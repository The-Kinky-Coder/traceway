// Package tracewaybackend provides an embeddable Traceway backend that can be
// run inside your own Go application.
//
// !! WARNING: Linux and macOS only.
// !! The embedded backend uses chdb (embedded ClickHouse), which does NOT
// !! support Windows. libchdb must be installed on the system before use:
// !!
// !!   curl -sL https://lib.chdb.io | bash
package tracewaybackend

import "github.com/tracewayapp/traceway/backend/cmd"

type Option = cmd.Option

var (
	Run                = cmd.Run
	WithPort           = cmd.WithPort
	WithServerURL      = cmd.WithServerURL
	WithSQLitePath     = cmd.WithSQLitePath
	WithClickhousePath = cmd.WithClickhousePath
	WithDefaultUser    = cmd.WithDefaultUser
	WithDefaultProject = cmd.WithDefaultProject
	DisableLogging     = cmd.DisableLogging
)
