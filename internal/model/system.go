package model

type SystemStatus struct {
	AppName           string `json:"app_name"`
	Environment       string `json:"environment"`
	DatabaseTime      string `json:"database_time"`
	BootstrappedAt    string `json:"bootstrapped_at"`
	AppliedMigrations int    `json:"applied_migrations"`
}
