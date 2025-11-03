package dictionary

// SourceMetadata defines the rich metadata for an attribute's source.
type SourceMetadata struct {
	Primary   string `json:"primary"`
	Secondary string `json:"secondary,omitempty"`
	Tertiary  string `json:"tertiary,omitempty"`
}

// SinkMetadata defines the rich metadata for an attribute's sink.
type SinkMetadata struct {
	Primary   string `json:"primary"`
	Secondary string `json:"secondary,omitempty"`
	Tertiary  string `json:"tertiary,omitempty"`
}

// Attribute is the central "pillar" of the data dictionary.
// This rich struct maps directly to the 'dictionary' table's JSONB columns.
type Attribute struct {
	AttributeID     string         `json:"attribute_id"`
	Name            string         `json:"name"`
	LongDescription string         `json:"long_description"`
	GroupID         string         `json:"group_id"`
	Mask            string         `json:"mask"`
	Domain          string         `json:"domain"`
	Vector          string         `json:"vector,omitempty"`
	Source          SourceMetadata `json:"source"`
	Sink            SinkMetadata   `json:"sink"`
}
