package item

// schema represents the schema for loading item upgrade data from a JSON file.
type schema struct {
	RenamedIDs    map[string]string            `json:"renamedIds,omitempty"`
	RemappedMetas map[string]map[uint32]string `json:"remappedMetas,omitempty"`
}
