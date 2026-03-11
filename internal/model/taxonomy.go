package model

// Taxonomy represents a classification taxonomy (e.g., asset allocation, industry).
type Taxonomy struct {
	ID              string
	Name            string
	Root            Classification
}

// Classification is a node in the taxonomy tree.
type Classification struct {
	ID              string
	Name            string
	Color           string
	Weight          int
	Children        []Classification
	Assignments     []Assignment
}

// Assignment links a security to a classification with a weight.
type Assignment struct {
	SecurityRef string
	Security    *Security
	Weight      int
}
