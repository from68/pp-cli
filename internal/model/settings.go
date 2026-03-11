package model

// Settings holds application settings from the portfolio file.
type Settings struct {
	Bookmarks      []Bookmark
	AttributeTypes []AttributeType
}

// Bookmark represents a saved view/filter configuration.
type Bookmark struct {
	Label string
	URL   string
}

// AttributeType defines a custom attribute type for securities or accounts.
type AttributeType struct {
	ID   string
	Name string
	Type string
}
