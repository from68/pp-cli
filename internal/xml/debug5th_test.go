package xml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"testing"
)

func TestDebug5thPassDirect(t *testing.T) {
	data, _ := os.ReadFile("../../test/data/Portfolio2.portfolio.local.xml")

	// Scan for orphan portfolioTransaction elements and check depth tracking
	dec := xml.NewDecoder(bytes.NewReader(data))
	depth := 0
	depthToAccUUID := make(map[int]string)
	pendingAccDepth := -1
	orphanCount := 0

	for {
		tok, err := dec.Token()
		if err != nil { break }
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			local := t.Name.Local

			if local == "account" || local == "accountTo" || local == "accountFrom" {
				isRef := false
				for _, a := range t.Attr {
					if a.Name.Local == "reference" { isRef = true; break }
				}
				if !isRef {
					depthToAccUUID[depth] = ""
					pendingAccDepth = depth
				}
				continue
			}

			if local == "uuid" && pendingAccDepth >= 0 && depth == pendingAccDepth+1 {
				var s string
				if err2 := dec.DecodeElement(&s, &t); err2 == nil {
					depthToAccUUID[pendingAccDepth] = s
				}
				depth--
				pendingAccDepth = -1
				continue
			}

			if local != "portfolioTransaction" { continue }
			isRef := false
			for _, a := range t.Attr {
				if a.Name.Local == "reference" { isRef = true; break }
			}
			if isRef { continue }

			txDepth := depth
			var tx rawPortfolioTx
			if err2 := dec.DecodeElement(&tx, &t); err2 != nil || tx.UUID == "" {
				depth--; continue
			}
			depth--

			if tx.PortfolioRef.Reference != "" { continue }
			orphanCount++

			// Find best account
			bestDepth := -1
			accUUID := ""
			for d, uuid := range depthToAccUUID {
				if d < txDepth && d > bestDepth && uuid != "" {
					bestDepth = d
					accUUID = uuid
				}
			}
			if orphanCount <= 3 {
				fmt.Printf("Orphan tx[%d]: uuid=%s type=%s accUUID=%s (depthMap=%v)\n",
					orphanCount, tx.UUID[:8], tx.Type, accUUID, depthToAccUUID)
			}

		case xml.EndElement:
			delete(depthToAccUUID, depth)
			depth--
		}
	}
	fmt.Printf("Total orphan portfolioTxs: %d\n", orphanCount)
}
