package repositories

// SessionAttributeFilter narrows the sessions list by an exact key=value
// match against the JSON `attributes` blob. Defined in a tag-neutral file
// so both the ClickHouse and SQLite repository builds see the same type.
type SessionAttributeFilter struct {
	Key   string
	Value string
}
