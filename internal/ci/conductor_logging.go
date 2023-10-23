package ci

import (
	"github.com/sirupsen/logrus"
	"os"
)

// NewLogger creates a new logger with the given configuration
// The logger is configured to write to stdout
// The logger is configured to write in text format if the child behavior is enabled.
// The logger is configured to write in JSON format if the JSON logging is enabled.
func NewLogger(cfg ConductorConfig) *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:    false,
		DisableTimestamp: cfg.Behavior.Child.Enabled,
	})
	switch cfg.Interface.Verbosity {
	case -1:
	case 0:
		logger.SetLevel(logrus.InfoLevel)
		break
	case 1:
		logger.SetLevel(logrus.DebugLevel)
		break
	default:
		logger.SetLevel(logrus.TraceLevel)
		break
	}
	if cfg.Behavior.Ci {
		logger.SetFormatter(&logrus.TextFormatter{
			DisableColors:             false,
			EnvironmentOverrideColors: false,
			ForceColors:               true,
			ForceQuote:                false,
		})
	}
	if cfg.Behavior.Child.Enabled {
		logger.SetFormatter(&logrus.TextFormatter{
			DisableTimestamp:          true,
			DisableColors:             false,
			EnvironmentOverrideColors: false,
			ForceColors:               true,
			ForceQuote:                false,
		})
	}
	if cfg.Interface.JSONLogging {
		logger.SetFormatter(&logrus.JSONFormatter{})
	}
	return logger
}
