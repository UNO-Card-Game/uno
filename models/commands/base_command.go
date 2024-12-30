package commands

import (
	"encoding/json"
	"fmt"
)

type BaseCommand struct {
	Type   string          `json:"type"`
	Object json.RawMessage `json:"obj"`
}

func DeserializeCommand(data []byte) (interface{}, error) {
	// Unmarshal the base command to determine the type
	var baseCmd BaseCommand
	if err := json.Unmarshal(data, &baseCmd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal command object: %w", err)
	}

	// Get the factory for the command type
	factory, err := GetCommandFactory(baseCmd.Type)
	if err != nil {
		return nil, err
	}

	// Use the factory to create a new instance of the command
	cmdInstance := factory()
	if err := json.Unmarshal(baseCmd.Object, cmdInstance); err != nil {
		return nil, fmt.Errorf("failed to unmarshal object for command type %s: %w", baseCmd.Type, err)
	}

	return cmdInstance, nil
}
