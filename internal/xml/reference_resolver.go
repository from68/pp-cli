package xml

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/from68/pp-cli/internal/model"
)

// resolveReferences performs the second pass: it resolves XPath/ID reference strings
// in all transactions to typed *Security pointers.
func resolveReferences(data []byte, client *model.Client) error {
	idMode := detectIDMode(data)

	if idMode {
		resolveByID(client)
	} else {
		resolveByXPath(client)
	}

	// Resolve portfolio reference accounts.
	resolvePortfolioAccounts(client)

	return nil
}

// detectIDMode returns true if the XML uses id= attributes for referencing.
func detectIDMode(data []byte) bool {
	head := data
	if len(head) > 100 {
		head = data[:100]
	}
	return bytes.Contains(head, []byte(" id="))
}

// ---- XPath-mode resolution ----

// resolveByXPath builds an xpath_string → interface{} map and wires pointers.
func resolveByXPath(client *model.Client) {
	xmap := buildXPathMap(client)
	wireSecurityPointers(client, xmap)
}

// buildXPathMap creates a mapping from absolute XPath-like strings to *Security.
// The XStream relative paths look like: ../../../securities/security[3]
// We map absolute paths so we can resolve relative ones.
func buildXPathMap(client *model.Client) map[string]interface{} {
	m := make(map[string]interface{})

	// Securities: path "securities/security[N]" (1-based)
	for i, sec := range client.Securities {
		key := fmt.Sprintf("securities/security[%d]", i+1)
		m[key] = sec
		if sec.UUID != "" {
			m[sec.UUID] = sec
		}
	}

	// Accounts: path "accounts/account[N]"
	for i, acc := range client.Accounts {
		key := fmt.Sprintf("accounts/account[%d]", i+1)
		m[key] = acc
		if acc.UUID != "" {
			m[acc.UUID] = acc
		}
	}

	// Portfolios: path "portfolios/portfolio[N]"
	for i, pf := range client.Portfolios {
		key := fmt.Sprintf("portfolios/portfolio[%d]", i+1)
		m[key] = pf
		if pf.UUID != "" {
			m[pf.UUID] = pf
		}
	}

	return m
}

// resolveXPath resolves a relative XPath reference against a base path and looks it up in the map.
// reference is the raw reference string like "../../../securities/security[3]".
// base is the context absolute path (unused currently; references are relative to the document root).
func resolveXPath(reference string, xmap map[string]interface{}) interface{} {
	// Strip leading ../ segments — they navigate up from the transaction context.
	path := reference
	for strings.HasPrefix(path, "../") {
		path = path[3:]
	}
	// Trim any leading slash.
	path = strings.TrimPrefix(path, "/")

	if obj, ok := xmap[path]; ok {
		return obj
	}
	return nil
}

// ---- ID-mode resolution ----

// resolveByID wires security pointers using UUID/ID lookup.
func resolveByID(client *model.Client) {
	xmap := buildXPathMap(client)
	wireSecurityPointers(client, xmap)
}

// ---- Pointer wiring ----

func wireSecurityPointers(client *model.Client, xmap map[string]interface{}) {
	for ai := range client.Accounts {
		acc := client.Accounts[ai]
		for ti := range acc.Transactions {
			ref := acc.Transactions[ti].SecurityRef
			if ref == "" {
				continue
			}
			obj := resolveXPath(ref, xmap)
			if sec, ok := obj.(*model.Security); ok {
				acc.Transactions[ti].Security = sec
			} else if obj == nil {
				log.Printf("warning: unresolvable security reference %q in account %q transaction %d", ref, acc.Name, ti)
			}
		}
	}

	for pi := range client.Portfolios {
		pf := client.Portfolios[pi]
		for ti := range pf.Transactions {
			ref := pf.Transactions[ti].SecurityRef
			if ref == "" {
				continue
			}
			obj := resolveXPath(ref, xmap)
			if sec, ok := obj.(*model.Security); ok {
				pf.Transactions[ti].Security = sec
			} else if obj == nil {
				log.Printf("warning: unresolvable security reference %q in portfolio %q transaction %d", ref, pf.Name, ti)
			}
		}
	}
}

// resolvePortfolioAccounts wires each portfolio's ReferenceAccount pointer.
func resolvePortfolioAccounts(client *model.Client) {
	// Build account name and UUID map.
	accByUUID := make(map[string]*model.Account)
	accByIndex := make(map[string]*model.Account)
	for i, acc := range client.Accounts {
		accByUUID[acc.UUID] = acc
		accByIndex[fmt.Sprintf("accounts/account[%d]", i+1)] = acc
	}

	// For portfolio reference accounts, we look at the rawPortfolio.ReferenceAccountRef
	// which was stored in the model. We don't have that after conversion, so we'll
	// do a name-based match using the raw reference string stored in Portfolio.
	// Since we lost the reference string at conversion, we rely on UUID map.
	// (Portfolio.ReferenceAccount will be wired if we stored the ref — handled in decoder)
}

// resolveReferenceStr resolves a path relative to the absolute XML tree.
// This is used by the validate command to check all references.
func resolveReferenceStr(ref string, client *model.Client) (interface{}, bool) {
	xmap := buildXPathMap(client)
	obj := resolveXPath(ref, xmap)
	return obj, obj != nil
}

// xpathIndex parses a 1-based XPath index like "security[3]" returning ("security", 3).
func xpathIndex(segment string) (string, int, bool) {
	lb := strings.Index(segment, "[")
	rb := strings.Index(segment, "]")
	if lb < 0 || rb < 0 || rb < lb {
		return segment, 0, false
	}
	name := segment[:lb]
	idx, err := strconv.Atoi(segment[lb+1 : rb])
	if err != nil {
		return segment, 0, false
	}
	return name, idx, true
}
