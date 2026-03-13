package xml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/from68/pp-cli/internal/model"
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

	// Build client from raw (securities, portfolios only — accounts rebuilt below).
	client := &model.Client{
		Version:      raw.Version,
		BaseCurrency: raw.BaseCurrency,
	}
	for i := range raw.Securities {
		client.Securities = append(client.Securities, raw.Securities[i].toModel())
	}
	// Portfolios may be defined inline anywhere in the document (e.g., inside an
	// account transaction cross-entry) with only a stub reference at the top level.
	// Scan the full document to find all inline portfolio elements.
	for _, raw := range scanAllPortfolios(data) {
		client.Portfolios = append(client.Portfolios, raw.toModel())
	}
	// Scan cross-entry portfolio transactions (camelCase <portfolioTransaction>)
	// that live in buySellEntry / deliveryEntry / portfolioTransferEntry structures.
	scanCrossEntryPortfolioTxs(data, client.Portfolios)

	// Replace the accounts list with a full scan of the entire document.
	// This finds accounts embedded in cross-entries that don't appear at top-level,
	// and collects all accountTransaction (BUY/SELL) elements attributed to each account.
	allAccs := scanAllAccounts(data)

	// Second pass: collect inline transactionTo/transactionFrom from account-transfer
	// crossEntries and attribute them to the correct account via reference resolution.
	collectTransferTransactions(data, allAccs)

	for _, ra := range allAccs {
		client.Accounts = append(client.Accounts, ra.toModel())
	}

	if err := resolveReferences(data, client); err != nil {
		return nil, err
	}

	// Fifth pass: attribute camelCase portfolioTransaction elements that live inside
	// account-transaction crossEntries and carry no <portfolio> child reference.
	// These are resolved via the account → portfolio mapping established above.
	collectAccountCrossEntryPortfolioTxs(data, client)

	return client, nil
}

// ---- Raw XML types ----

type rawClient struct {
	Version      int            `xml:"version"`
	BaseCurrency string         `xml:"baseCurrency"`
	Securities   []rawSecurity  `xml:"securities>security"`
	Accounts     []rawAccount   `xml:"accounts>account"`
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
	Reference    string         // populated from reference attr; if set, this is a stub
	UUID         string         `xml:"uuid"`
	Name         string         `xml:"name"`
	Currency     string         `xml:"currencyCode"`
	Note         string         `xml:"note"`
	Retired      bool           `xml:"isRetired"`
	Transactions []rawAccountTx // populated by custom UnmarshalXML
}

// UnmarshalXML handles <account> elements, including stubs with reference="..." attrs.
// It also decodes only DIRECT <account-transaction> children (not deep nested ones).
func (a *rawAccount) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for _, attr := range start.Attr {
		if attr.Name.Local == "reference" {
			a.Reference = attr.Value
			return d.Skip()
		}
	}
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "uuid":
				if err := d.DecodeElement(&a.UUID, &t); err != nil {
					return err
				}
			case "name":
				if err := d.DecodeElement(&a.Name, &t); err != nil {
					return err
				}
			case "currencyCode":
				if err := d.DecodeElement(&a.Currency, &t); err != nil {
					return err
				}
			case "note":
				if err := d.DecodeElement(&a.Note, &t); err != nil {
					return err
				}
			case "isRetired":
				if err := d.DecodeElement(&a.Retired, &t); err != nil {
					return err
				}
			case "transactions":
				if err := a.decodeTransactions(d); err != nil {
					return err
				}
			default:
				if err := d.Skip(); err != nil {
					return err
				}
			}
		case xml.EndElement:
			return nil
		}
	}
}

// decodeTransactions reads only the DIRECT <account-transaction> children, skipping stubs.
func (a *rawAccount) decodeTransactions(d *xml.Decoder) error {
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "account-transaction" {
				// Check for reference attr (stub) — skip those
				isRef := false
				for _, attr := range t.Attr {
					if attr.Name.Local == "reference" {
						isRef = true
						break
					}
				}
				if isRef {
					if err := d.Skip(); err != nil {
						return err
					}
				} else {
					var tx rawAccountTx
					if err := d.DecodeElement(&tx, &t); err != nil {
						return err
					}
					a.Transactions = append(a.Transactions, tx)
				}
			} else {
				if err := d.Skip(); err != nil {
					return err
				}
			}
		case xml.EndElement:
			return nil
		}
	}
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
	UUID                string           `xml:"uuid"`
	Name                string           `xml:"name"`
	ReferenceAccountRef rawSecRef        `xml:"referenceAccount"`
	Transactions        []rawPortfolioTx `xml:"transactions>portfolio-transaction"`
}

type rawPortfolioTx struct {
	UUID         string      `xml:"uuid"`
	Type         string      `xml:"type"`
	Date         string      `xml:"date"`
	Shares       int64       `xml:"shares"`
	Amount       int64       `xml:"amount"`
	Currency     string      `xml:"currencyCode"`
	Note         string      `xml:"note"`
	SecurityRef  rawSecRef   `xml:"security"`
	PortfolioRef rawSecRef   `xml:"portfolio"`
	Units        []rawTxUnit `xml:"units>unit"`
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
		Note:     r.Note,
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
		UUID:                r.UUID,
		Name:                r.Name,
		ReferenceAccountRef: r.ReferenceAccountRef.Reference,
	}
	for _, tx := range r.Transactions {
		if tx.UUID == "" {
			continue // skip stubs
		}
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

// ---- Full-document account scanner ----

// accScope tracks the state for a single <account>/<accountTo>/<accountFrom> element
// as we walk the token stream. It collects header fields and direct transactions.
type accScope struct {
	depth int      // depth at which the account element opened
	acc   rawAccount

	// Transaction list state (direct <transactions> child of account)
	inTxList    bool
	txListDepth int

	// Current <account-transaction> state (direct child of transactions)
	inTx      bool
	txDepth   int
	currentTx rawAccountTx

	// Which field are we collecting CharData for?
	// Format: "acc.uuid", "acc.name", "acc.currency", "acc.note", "acc.retired",
	//         "tx.uuid", "tx.type", "tx.date", "tx.amount", "tx.currency", "tx.note"
	collectingField string
}

// scanAllAccounts walks the entire XML token stream without consuming subtrees.
// It finds ALL <account>, <accountTo>, and <accountFrom> elements (no reference attr)
// at any nesting depth, collecting their header fields and direct transactions.
// Also collects <accountTransaction> (camelCase) elements, attributing them to the
// innermost account scope (the account that owns the enclosing crossEntry).
// Returns deduplicated accounts ordered by first appearance.
func scanAllAccounts(data []byte) []*rawAccount {
	dec := xml.NewDecoder(bytes.NewReader(data))
	seen := make(map[string]bool)
	var result []*rawAccount
	var scopes []*accScope
	depth := 0

	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++

			// Collect camelCase <accountTransaction> (BUY/SELL from portfolio crossEntries).
			// These belong to the innermost account scope.
			if t.Name.Local == "accountTransaction" {
				isRef := false
				for _, attr := range t.Attr {
					if attr.Name.Local == "reference" {
						isRef = true
						break
					}
				}
				if !isRef && len(scopes) > 0 {
					var tx rawAccountTx
					if err := dec.DecodeElement(&tx, &t); err == nil && tx.UUID != "" {
						scopes[len(scopes)-1].acc.Transactions = append(scopes[len(scopes)-1].acc.Transactions, tx)
					}
					depth-- // compensate for EndElement consumed by DecodeElement
				}
				continue
			}

			// Is this a new (non-reference) account element?
			isAccountElem := t.Name.Local == "account" || t.Name.Local == "accountTo" || t.Name.Local == "accountFrom"
			if isAccountElem {
				isRef := false
				for _, attr := range t.Attr {
					if attr.Name.Local == "reference" {
						isRef = true
						break
					}
				}
				if !isRef {
					scopes = append(scopes, &accScope{depth: depth})
				}
			}

			// Update ALL active scopes based on current depth.
			for _, sc := range scopes {
				// --- Account header fields (direct children of account element) ---
				if depth == sc.depth+1 {
					switch t.Name.Local {
					case "uuid":
						sc.collectingField = "acc.uuid"
					case "name":
						sc.collectingField = "acc.name"
					case "currencyCode":
						sc.collectingField = "acc.currency"
					case "note":
						sc.collectingField = "acc.note"
					case "isRetired":
						sc.collectingField = "acc.retired"
					case "transactions":
						sc.inTxList = true
						sc.txListDepth = depth
					}
					continue
				}

				// --- Transaction list: direct <account-transaction> children ---
				if sc.inTxList && !sc.inTx && depth == sc.txListDepth+1 {
					if t.Name.Local == "account-transaction" {
						isRef := false
						for _, attr := range t.Attr {
							if attr.Name.Local == "reference" {
								isRef = true
								break
							}
						}
						if !isRef {
							sc.inTx = true
							sc.txDepth = depth
							sc.currentTx = rawAccountTx{}
						}
					}
					continue
				}

				// --- Transaction fields (direct children of account-transaction) ---
				if sc.inTx && depth == sc.txDepth+1 {
					switch t.Name.Local {
					case "uuid":
						sc.collectingField = "tx.uuid"
					case "type":
						sc.collectingField = "tx.type"
					case "date":
						sc.collectingField = "tx.date"
					case "amount":
						sc.collectingField = "tx.amount"
					case "currencyCode":
						sc.collectingField = "tx.currency"
					case "note":
						sc.collectingField = "tx.note"
					case "security":
						for _, attr := range t.Attr {
							if attr.Name.Local == "reference" {
								sc.currentTx.SecurityRef.Reference = attr.Value
								break
							}
						}
					}
				}
			}

		case xml.EndElement:
			// Update ALL active scopes based on current depth.
			for _, sc := range scopes {
				// Clear collecting field if closing a field element
				if sc.collectingField != "" {
					// Field elements are one level deeper than their container
					sc.collectingField = ""
				}

				// Closing a direct <account-transaction>
				if sc.inTx && depth == sc.txDepth {
					if sc.currentTx.UUID != "" {
						sc.acc.Transactions = append(sc.acc.Transactions, sc.currentTx)
					}
					sc.inTx = false
					sc.txDepth = 0
				}

				// Closing the <transactions> element
				if sc.inTxList && depth == sc.txListDepth {
					sc.inTxList = false
					sc.txListDepth = 0
				}
			}

			// Finalize innermost scope if it's closing
			if len(scopes) > 0 && depth == scopes[len(scopes)-1].depth {
				sc := scopes[len(scopes)-1]
				scopes = scopes[:len(scopes)-1]
				if sc.acc.UUID != "" && !seen[sc.acc.UUID] {
					seen[sc.acc.UUID] = true
					result = append(result, &sc.acc)
				}
			}
			depth--

		case xml.CharData:
			text := strings.TrimSpace(string(t))
			if text == "" {
				continue
			}
			// Update the field being collected in each scope.
			for _, sc := range scopes {
				switch sc.collectingField {
				case "acc.uuid":
					sc.acc.UUID = text
				case "acc.name":
					sc.acc.Name = text
				case "acc.currency":
					sc.acc.Currency = text
				case "acc.note":
					sc.acc.Note = text
				case "acc.retired":
					sc.acc.Retired = text == "true"
				case "tx.uuid":
					sc.currentTx.UUID = text
				case "tx.type":
					sc.currentTx.Type = text
				case "tx.date":
					sc.currentTx.Date = text
				case "tx.amount":
					if v, err := strconv.ParseInt(text, 10, 64); err == nil {
						sc.currentTx.Amount = v
					}
				case "tx.currency":
					sc.currentTx.Currency = text
				case "tx.note":
					sc.currentTx.Note = text
				}
			}
		}
	}
	return result
}

// collectTransferTransactions scans ALL account-transfer crossEntries in the document,
// finds inline transactionTo/transactionFrom elements, resolves the associated
// accountTo/accountFrom reference (via ../ depth counting) to identify the owning
// account, and appends the transaction. This handles transfers where one side is a
// stub in the account's own transaction list but inline deep in another account's
// cross-entry subtree.
func collectTransferTransactions(data []byte, accs []*rawAccount) {
	byUUID := make(map[string]*rawAccount, len(accs))
	for _, a := range accs {
		if a.UUID != "" {
			byUUID[a.UUID] = a
		}
	}

	// Live map: element depth → account UUID (for currently-open account elements).
	depthToUUID := make(map[int]string)

	// pendingAccDepth: depth of the most recently opened inline account whose UUID
	// we haven't seen yet (UUID is always the first child in PP format).
	pendingAccDepth := -1

	type ceState struct {
		depth          int
		txTo           *rawAccountTx
		txFrom         *rawAccountTx
		accountToRef   string // reference attribute of <accountTo>; empty if inline
		accountFromRef string // reference attribute of <accountFrom>; empty if inline
	}
	var ceStack []*ceState

	// countDots returns the number of ".." path segments that lead the reference string.
	countDots := func(ref string) int {
		n := 0
		for _, seg := range strings.Split(ref, "/") {
			if seg == ".." {
				n++
			} else {
				break
			}
		}
		return n
	}

	dec := xml.NewDecoder(bytes.NewReader(data))
	depth := 0

	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			local := t.Name.Local

			// Track inline account element openings (account/accountTo/accountFrom).
			isAccElem := local == "account" || local == "accountTo" || local == "accountFrom"
			if isAccElem {
				isRef := false
				for _, a := range t.Attr {
					if a.Name.Local == "reference" {
						isRef = true
						break
					}
				}
				if !isRef {
					depthToUUID[depth] = ""
					pendingAccDepth = depth
				}
			}

			// Capture UUID of the pending inline account (UUID is always its first child).
			if local == "uuid" && pendingAccDepth == depth-1 {
				var s string
				if err := dec.DecodeElement(&s, &t); err == nil {
					depthToUUID[pendingAccDepth] = s
				}
				pendingAccDepth = -1
				depth-- // compensate for EndElement consumed by DecodeElement
				continue
			}

			// Track non-reference account-transfer crossEntries.
			if local == "crossEntry" {
				isTransfer, isRef := false, false
				for _, a := range t.Attr {
					switch {
					case a.Name.Local == "class" && a.Value == "account-transfer":
						isTransfer = true
					case a.Name.Local == "reference":
						isRef = true
					}
				}
				if isTransfer && !isRef {
					ceStack = append(ceStack, &ceState{depth: depth})
				}
				continue
			}

			// Handle direct children of the innermost open crossEntry.
			if len(ceStack) > 0 {
				ce := ceStack[len(ceStack)-1]
				if depth == ce.depth+1 {
					isRef := false
					refVal := ""
					for _, a := range t.Attr {
						if a.Name.Local == "reference" {
							isRef = true
							refVal = a.Value
							break
						}
					}
					switch local {
					case "transactionTo":
						if !isRef {
							var tx rawAccountTx
							if err := dec.DecodeElement(&tx, &t); err == nil && tx.UUID != "" {
								cp := tx
								ce.txTo = &cp
							}
							depth-- // compensate
						}
					case "transactionFrom":
						if !isRef {
							var tx rawAccountTx
							if err := dec.DecodeElement(&tx, &t); err == nil && tx.UUID != "" {
								cp := tx
								ce.txFrom = &cp
							}
							depth-- // compensate
						}
					case "accountTo":
						if isRef {
							ce.accountToRef = refVal
						}
					case "accountFrom":
						if isRef {
							ce.accountFromRef = refVal
						}
					}
				}
			}

		case xml.EndElement:
			// Finalize the innermost crossEntry when it closes.
			if len(ceStack) > 0 && depth == ceStack[len(ceStack)-1].depth {
				ce := ceStack[len(ceStack)-1]
				ceStack = ceStack[:len(ceStack)-1]

				addTx := func(tx *rawAccountTx, ref string) {
					if tx == nil || ref == "" {
						return
					}
					n := countDots(ref)
					targetDepth := (ce.depth + 1) - n
					uuid := depthToUUID[targetDepth]
					acc := byUUID[uuid]
					if acc == nil {
						return
					}
					// Dedup by UUID before appending.
					for _, ex := range acc.Transactions {
						if ex.UUID == tx.UUID {
							return
						}
					}
					acc.Transactions = append(acc.Transactions, *tx)
				}

				addTx(ce.txTo, ce.accountToRef)
				addTx(ce.txFrom, ce.accountFromRef)
			}

			// Remove depth entry when the element closes.
			delete(depthToUUID, depth)
			depth--
		}
	}
}

// scanAllPortfolios scans the entire document for inline (non-reference) <portfolio>
// elements at any nesting depth. In real PP XML files the canonical portfolio definition
// is often embedded inside an account transaction cross-entry, with only a stub
// <portfolio reference="..."/> at the top-level <portfolios> section.
// Returns deduplicated portfolios ordered by first appearance.
func scanAllPortfolios(data []byte) []*rawPortfolio {
	dec := xml.NewDecoder(bytes.NewReader(data))
	seen := make(map[string]bool)
	var result []*rawPortfolio

	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if se.Name.Local != "portfolio" {
			continue
		}
		// Skip reference stubs.
		isRef := false
		for _, a := range se.Attr {
			if a.Name.Local == "reference" {
				isRef = true
				break
			}
		}
		if isRef {
			continue
		}

		var pf rawPortfolio
		if err := dec.DecodeElement(&pf, &se); err != nil || pf.UUID == "" {
			continue
		}
		if !seen[pf.UUID] {
			seen[pf.UUID] = true
			cp := pf
			result = append(result, &cp)
		}
	}
	return result
}

// scanCrossEntryPortfolioTxs scans the document for camelCase <portfolioTransaction>,
// <portfolioTransactionFrom>, and <portfolioTransactionTo> elements found inside
// cross-entries (buySellEntry, deliveryEntry, portfolioTransferEntry, etc.).
// Each such element has a <portfolio reference="..."/> child that identifies
// the owning portfolio. Decoded transactions are appended (deduplicated by UUID).
func scanCrossEntryPortfolioTxs(data []byte, portfolios []*model.Portfolio) {
	pfMap := buildPortfolioRefMap(portfolios)

	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		local := se.Name.Local
		if local != "portfolioTransaction" && local != "portfolioTransactionFrom" && local != "portfolioTransactionTo" {
			continue
		}

		// Skip reference stubs.
		isRef := false
		for _, a := range se.Attr {
			if a.Name.Local == "reference" {
				isRef = true
				break
			}
		}
		if isRef {
			continue
		}

		var tx rawPortfolioTx
		if err := dec.DecodeElement(&tx, &se); err != nil || tx.UUID == "" {
			continue
		}

		ref := stripDotDot(tx.PortfolioRef.Reference)
		if ref == "" {
			continue
		}

		pf, ok := pfMap[ref]
		if !ok {
			continue
		}

		// Deduplicate by UUID before appending.
		dup := false
		for _, existing := range pf.Transactions {
			if existing.UUID == tx.UUID {
				dup = true
				break
			}
		}
		if !dup {
			pf.Transactions = append(pf.Transactions, tx.toModel())
		}
	}
}

// buildPortfolioRefMap builds a lookup map from normalised XPath/UUID keys to *model.Portfolio.
func buildPortfolioRefMap(portfolios []*model.Portfolio) map[string]*model.Portfolio {
	m := make(map[string]*model.Portfolio)
	for i, pf := range portfolios {
		m[fmt.Sprintf("portfolios/portfolio[%d]", i+1)] = pf
		if pf.UUID != "" {
			m[pf.UUID] = pf
		}
	}
	return m
}

// stripDotDot removes all leading "../" segments from a reference path.
func stripDotDot(ref string) string {
	path := ref
	for strings.HasPrefix(path, "../") {
		path = path[3:]
	}
	return strings.TrimPrefix(path, "/")
}

// collectAccountCrossEntryPortfolioTxs finds camelCase <portfolioTransaction> elements
// that are nested inside account-transaction crossEntries and carry no <portfolio> child
// reference (so scanCrossEntryPortfolioTxs cannot attribute them). It uses portfolio
// depth tracking: when a non-reference <portfolio> element is open at depth D and a
// portfolioTransaction descendant at depth > D has no PortfolioRef, we assign the
// transaction to the innermost enclosing portfolio.
func collectAccountCrossEntryPortfolioTxs(data []byte, client *model.Client) {
	// Build UUID → portfolio lookup.
	pfByUUID := make(map[string]*model.Portfolio)
	for _, pf := range client.Portfolios {
		if pf.UUID != "" {
			pfByUUID[pf.UUID] = pf
		}
	}
	if len(pfByUUID) == 0 {
		return
	}

	xmap := buildXPathMap(client)

	dec := xml.NewDecoder(bytes.NewReader(data))
	depth := 0
	// pfAtDepth maps the depth at which a non-reference portfolio element opened
	// to that portfolio's UUID (populated once the <uuid> child is seen).
	pfAtDepth := make(map[int]string)
	pendingPFDepth := -1

	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			local := t.Name.Local

			// Track non-reference portfolio elements so we can map depth → UUID.
			if local == "portfolio" {
				isRef := false
				for _, a := range t.Attr {
					if a.Name.Local == "reference" {
						isRef = true
						break
					}
				}
				if !isRef {
					pfAtDepth[depth] = ""
					pendingPFDepth = depth
				}
				continue
			}

			// Capture UUID for the most recently opened portfolio element.
			if local == "uuid" && pendingPFDepth >= 0 && depth == pendingPFDepth+1 {
				var s string
				if err2 := dec.DecodeElement(&s, &t); err2 == nil {
					pfAtDepth[pendingPFDepth] = s
				}
				depth-- // compensate for EndElement consumed by DecodeElement
				pendingPFDepth = -1
				continue
			}

			// Only handle non-reference camelCase portfolio transaction elements.
			if local != "portfolioTransaction" && local != "portfolioTransactionFrom" && local != "portfolioTransactionTo" {
				continue
			}
			isRef := false
			for _, a := range t.Attr {
				if a.Name.Local == "reference" {
					isRef = true
					break
				}
			}
			if isRef {
				continue
			}

			txDepth := depth
			var tx rawPortfolioTx
			if err2 := dec.DecodeElement(&tx, &t); err2 != nil || tx.UUID == "" {
				depth-- // compensate
				continue
			}
			depth-- // compensate for EndElement consumed by DecodeElement

			// Only handle account-centric transactions (no portfolio child reference).
			// scanCrossEntryPortfolioTxs already handles the case where PortfolioRef is set.
			if tx.PortfolioRef.Reference != "" {
				continue
			}

			// Find the innermost containing portfolio (highest depth < txDepth with a UUID).
			bestDepth := -1
			pfUUID := ""
			for d, uuid := range pfAtDepth {
				if d < txDepth && d > bestDepth && uuid != "" {
					bestDepth = d
					pfUUID = uuid
				}
			}
			if pfUUID == "" {
				continue
			}

			pf := pfByUUID[pfUUID]
			if pf == nil {
				continue
			}

			modelTx := tx.toModel()
			if modelTx.SecurityRef != "" {
				if obj := resolveXPath(modelTx.SecurityRef, xmap); obj != nil {
					if sec, ok := obj.(*model.Security); ok {
						modelTx.Security = sec
					}
				}
			}

			// Append, deduplicating by UUID.
			dup := false
			for _, existing := range pf.Transactions {
				if existing.UUID == modelTx.UUID {
					dup = true
					break
				}
			}
			if !dup {
				pf.Transactions = append(pf.Transactions, modelTx)
			}

		case xml.EndElement:
			delete(pfAtDepth, depth)
			depth--
		}
	}
}

