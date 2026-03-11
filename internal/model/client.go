package model

// Client is the root struct representing the Portfolio Performance data file.
type Client struct {
	Version      int
	BaseCurrency string
	Securities   []*Security
	Accounts     []*Account
	Portfolios   []*Portfolio
	Taxonomies   []*Taxonomy
	Settings     *Settings
}
