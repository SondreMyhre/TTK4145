package localsingle2

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// Config values that can be loaded from file.
type ElevatorConfig struct {
	DoorOpenDuration float64
	InputPollRateMs  int
}

// DefaultConfig returns the default configuration.
func DefaultConfig() ElevatorConfig {
	return ElevatorConfig{
		DoorOpenDuration: 3.0,
		InputPollRateMs:  25,
	}
}

// LoadConfig loads configuration from a file.
// Lines starting with "--" are parsed as key-value pairs.
// Returns default values if file cannot be opened.
func LoadConfig(filename string) ElevatorConfig {
	cfg := DefaultConfig()

	file, err := os.Open(filename)
	if err != nil {
		return cfg
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "--") {
			continue
		}

		// Remove the "--" prefix and split on whitespace.
		line = strings.TrimPrefix(line, "--")
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := strings.ToLower(fields[0])
		value := fields[1]

		switch key {
		case "dooropenduration_s":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				cfg.DoorOpenDuration = v
			}
		case "inputpollrate_ms":
			if v, err := strconv.Atoi(value); err == nil {
				cfg.InputPollRateMs = v
			}
		}
	}

	return cfg
}
