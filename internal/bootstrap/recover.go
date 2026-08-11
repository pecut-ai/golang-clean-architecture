package bootstrap

import (
	"fmt"
	"os"
	"runtime/debug"

	"golang-clean-architecture/internal/logging"
)

func RecoverMain(entrypoint string, log **logging.Logger) {
	if recovered := recover(); recovered != nil {
		if log != nil && *log != nil {
			(*log).Errorw("application_panic_recovered", "entrypoint", entrypoint, "panic", recovered, "stacktrace", string(debug.Stack()))
			_ = (*log).Sync()
		} else {
			fmt.Fprintf(os.Stderr, "application panic recovered entrypoint=%s panic=%v\n%s", entrypoint, recovered, debug.Stack())
		}
		os.Exit(1)
	}
}
