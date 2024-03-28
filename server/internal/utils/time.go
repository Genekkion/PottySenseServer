package utils

import (
	"fmt"
	"time"
)

func GetTimeElapsedPretty(timeRecord time.Time) string {
	elapsedTime := time.Since(timeRecord)
	return fmt.Sprintf("%02d:%02d",
		int(elapsedTime.Hours()),
		int(elapsedTime.Minutes())%60,
	)
}
