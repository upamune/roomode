package mode

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
)

// ParseModeFile parses a specified Markdown file and returns a Config
func ParseModeFile(filePath string) (*Config, error) {

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var metadata Metadata
	content, err := frontmatter.Parse(bytes.NewReader(data), &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	base := filepath.Base(absPath)
	slug := base[:len(base)-len(filepath.Ext(base))]

	parsedGroups, err := ParseGroupEntries(metadata.Groups)
	if err != nil {
		return nil, fmt.Errorf("failed to parse group entries: %w", err)
	}

	// Convert content to string and check if it's empty or just whitespace
	contentStr := string(content)
	contentStr = strings.TrimSpace(contentStr)

	// If content is not empty, use it as CustomInstructions
	var customInstructions *string
	if contentStr != "" {
		customInstructions = &contentStr
	}

	config := &Config{
		Slug:               slug,
		Name:               metadata.Name,
		GroupsRaw:          metadata.Groups,
		GroupsParsed:       parsedGroups,
		RoleDefinition:     metadata.RoleDefinition,
		CustomInstructions: customInstructions,
		FilePath:           absPath,
		Source:             metadata.Source,
	}

	return config, nil
}

// ParseGroupEntries parses and validates raw GroupEntry slices
func ParseGroupEntries(rawGroups []GroupEntry) ([]ParsedGroupEntry, error) {
	parsedGroups := make([]ParsedGroupEntry, 0, len(rawGroups))

	for i, entry := range rawGroups {
		switch v := entry.(type) {
		case string:
			// Simple string group
			parsedGroups = append(parsedGroups, ParsedGroupEntry{
				Name:    v,
				Options: nil,
			})
		case []interface{}:
			// Array format [string, options]
			if len(v) != 2 {
				return nil, fmt.Errorf("invalid group entry at index %d: array must have exactly 2 elements", i)
			}

			name, ok := v[0].(string)
			if !ok {
				return nil, fmt.Errorf("invalid group entry at index %d: first element must be a string", i)
			}

			// Handle different types of option maps
			options := &GroupOptions{}

			switch optMap := v[1].(type) {
			case map[interface{}]interface{}:
				// YAML parsing typically produces map[interface{}]interface{}
				if fileRegex, ok := optMap["fileRegex"]; ok {
					if fileRegexStr, ok := fileRegex.(string); ok {
						options.FileRegex = &fileRegexStr
					} else {
						return nil, fmt.Errorf("invalid fileRegex at index %d: must be a string", i)
					}
				}

				if description, ok := optMap["description"]; ok {
					if descriptionStr, ok := description.(string); ok {
						options.Description = &descriptionStr
					} else {
						return nil, fmt.Errorf("invalid description at index %d: must be a string", i)
					}
				}
			case map[string]interface{}:
				// JSON parsing typically produces map[string]interface{}
				if fileRegex, ok := optMap["fileRegex"]; ok {
					if fileRegexStr, ok := fileRegex.(string); ok {
						options.FileRegex = &fileRegexStr
					} else {
						return nil, fmt.Errorf("invalid fileRegex at index %d: must be a string", i)
					}
				}

				if description, ok := optMap["description"]; ok {
					if descriptionStr, ok := description.(string); ok {
						options.Description = &descriptionStr
					} else {
						return nil, fmt.Errorf("invalid description at index %d: must be a string", i)
					}
				}
			case string:
				// Handle case where the second element is a string representation of a map
				// This happens sometimes with YAML parsing of complex structures
				if optMap == "" {
					// Empty options
					options = nil
				} else if strings.HasPrefix(optMap, "map[") {
					// Try to parse map[key:value] format
					optStr := optMap
					// Remove the map[] wrapper
					optStr = strings.TrimPrefix(optStr, "map[")
					optStr = strings.TrimSuffix(optStr, "]")

					// Split by space to get key-value pairs
					parts := strings.Split(optStr, " ")
					for _, part := range parts {
						if part == "" {
							continue
						}

						keyValue := strings.SplitN(part, ":", 2)
						if len(keyValue) != 2 {
							continue
						}

						key := keyValue[0]
						value := keyValue[1]

						switch key {
						case "fileRegex":
							options.FileRegex = &value
						case "description":
							options.Description = &value
						}
					}
				} else {
					return nil, fmt.Errorf("invalid options format at index %d: %s", i, optMap)
				}
			default:
				return nil, fmt.Errorf("invalid options format at index %d: %T", i, v[1])
			}

			parsedGroups = append(parsedGroups, ParsedGroupEntry{
				Name:    name,
				Options: options,
			})
		case map[interface{}]interface{}:
			// Handle YAML object format with key as group name and value as options
			// This format looks like: - groupName: { options }
			if len(v) != 1 {
				return nil, fmt.Errorf("invalid group entry at index %d: map must have exactly 1 key", i)
			}

			// Extract the single key (group name) and value (options)
			var name string
			var optionsMap map[interface{}]interface{}

			for k, val := range v {
				if nameStr, ok := k.(string); ok {
					name = nameStr
					if optMap, ok := val.(map[interface{}]interface{}); ok {
						optionsMap = optMap
					} else {
						return nil, fmt.Errorf("invalid group options at index %d: must be an object", i)
					}
				} else {
					return nil, fmt.Errorf("invalid group name at index %d: must be a string", i)
				}
			}

			// Parse options
			options := &GroupOptions{}

			if fileRegex, ok := optionsMap["fileRegex"]; ok {
				if fileRegexStr, ok := fileRegex.(string); ok {
					options.FileRegex = &fileRegexStr
				} else {
					return nil, fmt.Errorf("invalid fileRegex at index %d: must be a string", i)
				}
			}

			if description, ok := optionsMap["description"]; ok {
				if descriptionStr, ok := description.(string); ok {
					options.Description = &descriptionStr
				} else {
					return nil, fmt.Errorf("invalid description at index %d: must be a string", i)
				}
			}

			parsedGroups = append(parsedGroups, ParsedGroupEntry{
				Name:    name,
				Options: options,
			})
		case map[string]interface{}:
			// Handle JSON object format with key as group name and value as options
			// This format looks like: { "groupName": { options } }
			if len(v) != 1 {
				return nil, fmt.Errorf("invalid group entry at index %d: map must have exactly 1 key", i)
			}

			// Extract the single key (group name) and value (options)
			var name string
			var optionsMap map[string]interface{}

			for k, val := range v {
				name = k
				if optMap, ok := val.(map[string]interface{}); ok {
					optionsMap = optMap
				} else {
					return nil, fmt.Errorf("invalid group options at index %d: must be an object", i)
				}
			}

			// Parse options
			options := &GroupOptions{}

			if fileRegex, ok := optionsMap["fileRegex"]; ok {
				if fileRegexStr, ok := fileRegex.(string); ok {
					options.FileRegex = &fileRegexStr
				} else {
					return nil, fmt.Errorf("invalid fileRegex at index %d: must be a string", i)
				}
			}

			if description, ok := optionsMap["description"]; ok {
				if descriptionStr, ok := description.(string); ok {
					options.Description = &descriptionStr
				} else {
					return nil, fmt.Errorf("invalid description at index %d: must be a string", i)
				}
			}

			parsedGroups = append(parsedGroups, ParsedGroupEntry{
				Name:    name,
				Options: options,
			})
		default:
			return nil, fmt.Errorf("invalid group entry at index %d: must be a string, an array, or a map, got %T", i, entry)
		}
	}

	return parsedGroups, nil
}
