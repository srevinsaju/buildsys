package config

type Config struct {
	RunStages  []string
	ContextDir string
	CiFile     string
	DryRun     bool
	JobsNumber int
}
