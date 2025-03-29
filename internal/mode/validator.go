package mode

import (
	"fmt"
	"regexp"
)

// ValidateMode validates the contents of a Config
func ValidateMode(mode *Config) error {

	if mode.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(mode.GroupsParsed) == 0 {
		return fmt.Errorf("at least one group is required")
	}

	for i, group := range mode.GroupsParsed {
		if group.Options != nil && group.Options.FileRegex != nil {
			if _, err := regexp.Compile(*group.Options.FileRegex); err != nil {
				return fmt.Errorf("invalid fileRegex at index %d: %s", i, err.Error())
			}
		}
	}

	if mode.RoleDefinition == "" {
		return fmt.Errorf("role definition (markdown content) is required")
	}

	return nil
}
