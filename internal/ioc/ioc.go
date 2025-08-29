package ioc

import (
	"context"
	"io"

	"github.com/tguidoux/newspaper4k-go/internal/helpers"
)

// ParseIOC parses a single IOC and returns the highest-priority IOC found.
// For example, an email contains a domain; the email (higher priority) is returned.
func ParseIOC(ioc string) *IOC {
	iocs := ExtractIOCs(ioc, true)
	ret := &IOC{}
	for _, x := range iocs {
		if x.Type > ret.Type {
			ret = x
		}
	}
	return ret
}

// ExtractIOCs returns the IOCs found in the provided data. When getFangedIOCs
// is true, fanged IOCs are included as well.
func ExtractIOCs(data string, getFangedIOCs bool) []*IOC {
	var iocs []*IOC

	for t, regex := range iocRegexes {
		matches := helpers.UniqueStringsSimple(regex.FindAllString(data, -1))
		for _, m := range matches {
			i := &IOC{IOC: m, Type: t}
			if !i.isFanged() || getFangedIOCs {
				iocs = append(iocs, i)
			}
		}
	}
	return iocs
}

// ExtractIOCsReader reads from reader, extracts IOCs and sends them to matches.
// Note: output order is not deterministic because of map-based deduplication.
func ExtractIOCsReader(ctx context.Context, reader io.Reader, getFangedIOCs bool, matches chan *IOC) error {
	// If context is already done, return immediately.
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	iocs := ExtractIOCs(string(data), getFangedIOCs)
	for _, i := range iocs {
		select {
		case matches <- i:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// NormalizeDefangs runs each IOC through toFanged then toDefanged to normalize
// the defanged style across all IOCs.
func NormalizeDefangs(iocs []*IOC) {
	for idx, i := range iocs {
		iocs[idx] = i.toFanged().toDefanged()
	}
}
