package xml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/yourname/pp-cli/internal/model"
)

// Decode reads Portfolio Performance XML from r and returns the decoded Client.
// It performs a two-pass decode: first build the struct tree, then resolve references.
func Decode(r io.Reader) (*model.Client, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading XML: %w", err)
	}

	var raw rawClient
	if err := xml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("decoding XML: %w", err)
	}

	client := raw.toModel()

	if err := resolveReferences(data, client); err != nil {
		return nil, err
	}

	return client, nil
}

// ---- Raw XML types ----

type rawClient struct {
	Version      int           `xml:"version"`
	BaseCurrency string        `xml:"baseCurrency"`
	Securities   []rawSecurity `xml:"securities>security"`
	Accounts     []rawAccount  `xml:"accounts>account"`
	Portfolios   []rawPortfolio `xml:"portfolios>portfolio"`
}

type rawSecurity struct {
	UUID      string     `xml:"uuid"`
	Name      string     `xml:"name"`
	Currency  string     `xml:"currencyCode"`
	ISIN      string     `xml:"isin"`
	Ticker    string     `xml:"tickerSymbol"`
	Feed      string     `xml:"feed"`
	Retired   bool       `xml:"isRetired"`
	UpdatedAt string     `xml:"updatedAt"`
	Prices    []rawPrice `xml:"prices>price"`
	Events    rawEvents  `xml:"events"`
}

type rawPrice struct {
	Date  string `xml:"t,attr"`
	Value int64  `xml:"v,attr"`
}

// rawEvents handles the polymorphic <events> block that may contain
// both <event> and <dividendEvent> child elements.
type rawEvents struct {
	Items []rawEvent
}

type rawEvent struct {
	XMLName xml.Name
	Type    string `xml:"type"`
	Date    string `xml:"date"`
	Amount  int64  `xml:"amount"`
	Details string `xml:"details"`
}

func (e *rawEvents) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for {
		tok, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			if _, ok := tok.(xml.EndElement); ok {
				break
			}
			continue
		}
		if se.Name.Local == "event" || se.Name.Local == "dividendEvent" {
			var ev rawEvent
			if err := d.DecodeElement(&ev, &se); err != nil {
				return err
			}
			e.Items = append(e.Items, ev)
		} else {
			if err := d.Skip(); err != nil {
				return err
			}
		}
	}
	return nil
}

type rawAccount struct {
	UUID         string               `xml:"uuid"`
	Name         string               `xml:"name"`
	Currency     string               `xml:"currencyCode"`
	Retired      bool                 `xml:"isRetired"`
	Transactions []rawAccountTx       `xml:"transactions>account-transaction"`
}

type rawAccountTx struct {
	UUID        string      `xml:"uuid"`
	Type        string      `xml:"type"`
	Date        string      `xml:"date"`
	Amount      int64       `xml:"amount"`
	Currency    string      `xml:"currencyCode"`
	Note        string      `xml:"note"`
	SecurityRef rawSecRef   `xml:"security"`
	Units       []rawTxUnit `xml:"units>unit"`
}

// rawSecRef captures the reference attribute of a <security reference="..."/> element.
type rawSecRef struct {
	Reference string `xml:"reference,attr"`
}

type rawTxUnit struct {
	Type     string `xml:"type,attr"`
	Amount   int64  `xml:"amount"`
	Currency string `xml:"currencyCode"`
	Shares   int64  `xml:"shares"`
}

type rawPortfolio struct {
	UUID               string            `xml:"uuid"`
	Name               string            `xml:"name"`
	ReferenceAccountRef rawSecRef        `xml:"referenceAccount"`
	Transactions       []rawPortfolioTx  `xml:"transactions>portfolio-transaction"`
}

type rawPortfolioTx struct {
	UUID        string      `xml:"uuid"`
	Type        string      `xml:"type"`
	Date        string      `xml:"date"`
	Shares      int64       `xml:"shares"`
	Amount      int64       `xml:"amount"`
	Currency    string      `xml:"currencyCode"`
	Note        string      `xml:"note"`
	SecurityRef rawSecRef   `xml:"security"`
	Units       []rawTxUnit `xml:"units>unit"`
}

// ---- Conversion to model ----

func (r *rawClient) toModel() *model.Client {
	c := &model.Client{
		Version:      r.Version,
		BaseCurrency: r.BaseCurrency,
	}

	for i := range r.Securities {
		c.Securities = append(c.Securities, r.Securities[i].toModel())
	}
	for i := range r.Accounts {
		c.Accounts = append(c.Accounts, r.Accounts[i].toModel())
	}
	for i := range r.Portfolios {
		c.Portfolios = append(c.Portfolios, r.Portfolios[i].toModel())
	}

	return c
}

func (r *rawSecurity) toModel() *model.Security {
	s := &model.Security{
		UUID:     r.UUID,
		Name:     r.Name,
		Currency: r.Currency,
		ISIN:     r.ISIN,
		Ticker:   r.Ticker,
		Feed:     r.Feed,
		Retired:  r.Retired,
	}
	if r.UpdatedAt != "" {
		if t, err := parsePPTime(r.UpdatedAt); err == nil {
			s.UpdatedAt = t
		}
	}
	for _, p := range r.Prices {
		if t, err := parsePPDate(p.Date); err == nil {
			s.Prices = append(s.Prices, model.SecurityPrice{Date: t, Value: p.Value})
		}
	}
	for _, e := range r.Events.Items {
		ev := model.SecurityEvent{
			Type:    e.XMLName.Local,
			Amount:  e.Amount,
			Details: e.Details,
		}
		if e.Date != "" {
			if t, err := parsePPDate(e.Date); err == nil {
				ev.Date = t
			}
		}
		s.Events = append(s.Events, ev)
	}
	return s
}

func (r *rawAccount) toModel() *model.Account {
	a := &model.Account{
		UUID:     r.UUID,
		Name:     r.Name,
		Currency: r.Currency,
		Retired:  r.Retired,
	}
	for _, tx := range r.Transactions {
		a.Transactions = append(a.Transactions, tx.toModel())
	}
	return a
}

func (r *rawAccountTx) toModel() model.AccountTransaction {
	tx := model.AccountTransaction{
		UUID:        r.UUID,
		Type:        r.Type,
		Amount:      r.Amount,
		Currency:    r.Currency,
		Note:        r.Note,
		SecurityRef: r.SecurityRef.Reference,
	}
	if r.Date != "" {
		if t, err := parsePPTime(r.Date); err == nil {
			tx.Date = t
		}
	}
	for _, u := range r.Units {
		tx.Units = append(tx.Units, model.TxUnit{
			Type:     u.Type,
			Amount:   u.Amount,
			Currency: u.Currency,
			Shares:   model.Shares(u.Shares),
		})
	}
	return tx
}

func (r *rawPortfolio) toModel() *model.Portfolio {
	p := &model.Portfolio{
		UUID: r.UUID,
		Name: r.Name,
	}
	for _, tx := range r.Transactions {
		p.Transactions = append(p.Transactions, tx.toModel())
	}
	return p
}

func (r *rawPortfolioTx) toModel() model.PortfolioTransaction {
	tx := model.PortfolioTransaction{
		UUID:        r.UUID,
		Type:        r.Type,
		Shares:      model.Shares(r.Shares),
		Amount:      r.Amount,
		Currency:    r.Currency,
		Note:        r.Note,
		SecurityRef: r.SecurityRef.Reference,
	}
	if r.Date != "" {
		if t, err := parsePPTime(r.Date); err == nil {
			tx.Date = t
		}
	}
	for _, u := range r.Units {
		tx.Units = append(tx.Units, model.TxUnit{
			Type:     u.Type,
			Amount:   u.Amount,
			Currency: u.Currency,
			Shares:   model.Shares(u.Shares),
		})
	}
	return tx
}

// ---- Date/time parsing ----

var ppTimeFmts = []string{
	"2006-01-02T15:04",
	"2006-01-02T15:04:05",
	"2006-01-02",
}

func parsePPTime(s string) (time.Time, error) {
	for _, f := range ppTimeFmts {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	// Try Unix milliseconds (some PP versions use this)
	if ms, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.UnixMilli(ms).UTC(), nil
	}
	return time.Time{}, fmt.Errorf("cannot parse PP time: %q", s)
}

func parsePPDate(s string) (time.Time, error) {
	return parsePPTime(s)
}

// ---- Cross-entry XML handling ----
// crossEntry elements use a class attribute to distinguish BuySell from Transfer.
// They are handled during reference resolution by scanning the raw XML.

type rawCrossEntry struct {
	Class           string          `xml:"class,attr"`
	PortfolioTx     *rawPortfolioTx `xml:"portfolioTransaction"`
	AccountTx       *rawAccountTx   `xml:"accountTransaction"`
}

// parseCrossEntries extracts cross-entries from the raw XML data.
// These are stored under <buysell> or <transfer> elements.
func parseCrossEntries(data []byte) ([]rawCrossEntry, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var entries []rawCrossEntry
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if se.Name.Local == "buysell" || se.Name.Local == "transfer" {
			var ce rawCrossEntry
			ce.Class = se.Name.Local
			if err := dec.DecodeElement(&ce, &se); err != nil {
				return nil, err
			}
			entries = append(entries, ce)
		}
	}
	return entries, nil
}
