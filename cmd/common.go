package cmd

import (
	"time"

	flag "github.com/spf13/pflag"
)

func setFileFlags(flags *flag.FlagSet) {
	// required
	flags.StringP("file", "f", "", "Specify the file path")

	// options
	flags.Bool("force", false, "Force to delete/create/replace hosts")

	flags.Bool("lock", false, "Acquire file lock, useful when multiple process working on the same file")
	flags.StringP("lock-timeout", "", "", "Specify how long to wait to acquire a file lock")
}

func parseDuration(dur string) (time.Duration, error) {
	if dur != "" {
		return time.ParseDuration(dur)
	}

	return 0, nil
}
