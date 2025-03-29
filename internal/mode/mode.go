package mode

// GroupOptions defines options for tool groups (like file access restrictions)
type GroupOptions struct {
	FileRegex   *string `yaml:"fileRegex,omitempty" json:"fileRegex,omitempty"`
	Description *string `yaml:"description,omitempty" json:"description,omitempty"`
}

// GroupEntry represents either a simple group name (string) or a [string, GroupOptions] tuple
// Using interface{} for flexibility, requires custom parsing after frontmatter parsing
type GroupEntry interface{}

// Metadata represents data parsed from frontmatter
type Metadata struct {
	Name           string       `yaml:"name" json:"name"`
	Groups         []GroupEntry `yaml:"groups" json:"groups"` // Requires custom parsing after initial parse
	RoleDefinition string       `yaml:"roleDefinition" json:"roleDefinition"`
	Source         string       `yaml:"source,omitempty" json:"source,omitempty"`
}

// Config represents the complete data for a custom mode
type Config struct {
	Slug               string
	Name               string
	GroupsRaw          []GroupEntry       // Raw data from frontmatter
	GroupsParsed       []ParsedGroupEntry // Parsed and validated groups
	RoleDefinition     string             // From frontmatter
	CustomInstructions *string            // Content of the Markdown body (if not empty)
	FilePath           string             // Path to the source Markdown file
	Source             string             // Original source path from frontmatter
}

// ParsedGroupEntry represents a validated group entry
type ParsedGroupEntry struct {
	Name    string
	Options *GroupOptions
}
