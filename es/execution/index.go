package execution

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"path"

	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/types"
)

func LoadEventLogEntries(executionID string) ([]types.EventLogEntry, error) {

	// Open the JSONL file
	fileName := path.Join(viper.GetString("log.dir"), executionID+".jsonl")
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Create a slice to hold the parsed eventLogEntries
	var eventLogEntries []types.EventLogEntry

	// Read the file line by line
	for scanner.Scan() {
		line := scanner.Bytes()

		// Create a new Event struct to hold the parsed event
		var event types.EventLogEntry

		// Parse the line into the Event struct
		err := json.Unmarshal(line, &event)
		if err != nil {
			log.Println("Error parsing line:", err)
			continue
		}

		// Append the parsed event to the events slice
		eventLogEntries = append(eventLogEntries, event)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return eventLogEntries, nil
}