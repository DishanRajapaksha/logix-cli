package cli

import (
	"fmt"
	"time"

	"github.com/DishanRajapaksha/logix-cli/internal/logixclient"
)

func validateWriteMode(yes, dryRun bool) error {
	if yes && dryRun {
		return fmt.Errorf("%w: --yes and --dry-run cannot be used together", logixclient.ErrValidation)
	}
	return nil
}

func validateWatchOptions(interval time.Duration, count int, duration time.Duration) error {
	if interval <= 0 {
		return fmt.Errorf("%w: interval must be positive", logixclient.ErrValidation)
	}
	if count < 0 {
		return fmt.Errorf("%w: count must be non-negative", logixclient.ErrValidation)
	}
	if duration < 0 {
		return fmt.Errorf("%w: duration must be non-negative", logixclient.ErrValidation)
	}
	return nil
}

func watchShouldStop(started time.Time, completed, count int, duration time.Duration) bool {
	if count > 0 && completed >= count {
		return true
	}
	return duration > 0 && time.Since(started) >= duration
}

func waitForNextWatch(started time.Time, duration, interval time.Duration) bool {
	if duration <= 0 {
		time.Sleep(interval)
		return true
	}
	remaining := duration - time.Since(started)
	if remaining <= 0 {
		return false
	}
	if interval >= remaining {
		time.Sleep(remaining)
		return false
	}
	time.Sleep(interval)
	return true
}
