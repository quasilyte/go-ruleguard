package vendored

import (
	"errors"
	"fmt"
	"strings"

	"github.invalid/globex/logging"
)

func example() {
	err := fmt.Errorf("Failed to configure system. Error: %v", errors.New("test")) // want `\Qnothing special, just testing the Errorf rule`

	logger := logging.GetLogger()
	logger.Errorf("Failed to configure system. Error: %v", err)         // want `\QErrors must be logged as a structured field`
	logger.Errorf("Failed to configure system. Error: %v", err.Error()) // want `\QErrors must be logged as a structured field`

	name := "abc"
	logger.Errorf("Configure system %s", name)
	logger.Errorf("Failed to configure system %s. Error: %v", strings.ToLower(name), err.Error()) // want `\QErrors must be logged as a structured field`
}
