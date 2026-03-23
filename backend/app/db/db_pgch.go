//go:build pgch

package db

func Init() error {
	return initPostgres()
}
