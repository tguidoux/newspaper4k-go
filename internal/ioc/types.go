package ioc

import (
	"bytes"
	"fmt"
	"sort"
	"text/tabwriter"
)

// IOC Struct to store an IOC and it's type
type IOC struct {
	IOC  string
	Type Type // hash, url, domain, file
}

// String Takes an IOC and prints in csv form: IOC|Type
func (ioc *IOC) String() string {
	return ioc.IOC + "|" + ioc.Type.String()
}

// Type Type of IOC (bitcoin, sha1, etc)
type Type int

// Types ordered in list of largest to smallest (so an email is > domain since an email contains a domain)
//
//go:generate stringer -type=Type
const (
	Unknown Type = iota
	Bitcoin
	MD5
	SHA1
	SHA256
	SHA512
	Domain
	Email
	IPv4
	IPv6
	URL
	File
	CVE
	CAPEC
	CWE
	CPE
)

// String returns the string representation of the Type.
func (t Type) String() string {
	switch t {
	case Unknown:
		return "Unknown"
	case Bitcoin:
		return "Bitcoin"
	case MD5:
		return "MD5"
	case SHA1:
		return "SHA1"
	case SHA256:
		return "SHA256"
	case SHA512:
		return "SHA512"
	case Domain:
		return "Domain"
	case Email:
		return "Email"
	case IPv4:
		return "IPv4"
	case IPv6:
		return "IPv6"
	case URL:
		return "URL"
	case File:
		return "File"
	case CVE:
		return "CVE"
	case CAPEC:
		return "CAPEC"
	case CWE:
		return "CWE"
	case CPE:
		return "CPE"
	default:
		return "Unknown"
	}
}

// Types of all IOCs
var Types = []Type{
	Bitcoin,
	MD5,
	SHA1,
	SHA256,
	SHA512,
	Domain,
	Email,
	IPv4,
	IPv6,
	URL,
	File,
	CVE,
	CAPEC,
	CWE,
	CPE,
}

// -- []IOC helpers --

// sortByType returns a copy of iocs sorted by Type.
func sortByType(iocs []*IOC) []*IOC {
	if iocs == nil {
		return nil
	}
	copy := make([]*IOC, 0, len(iocs))
	copy = append(copy, iocs...)
	sort.Slice(copy, func(i, j int) bool { return copy[i].Type < copy[j].Type })
	return copy
}

// FormatIOCs formats IOCs according to format ("csv" or "table").
func FormatIOCs(iocs []*IOC, format string) string {
	switch format {
	case "csv":
		return formatIOCsCSV(iocs)
	case "table":
		return formatIOCsTable(iocs)
	default:
		return formatIOCsCSV(iocs)
	}
}

// formatIOCsCSV returns a CSV representation of the provided IOCs.
func formatIOCsCSV(iocs []*IOC) string {
	if len(iocs) == 0 {
		return ""
	}
	buf := &bytes.Buffer{}
	for i, ioc := range iocs {
		buf.WriteString(ioc.String())
		if i < len(iocs)-1 {
			buf.WriteByte('\n')
		}
	}
	return buf.String()
}

// formatIOCsTable returns a table representation of the provided IOCs.
func formatIOCsTable(iocs []*IOC) string {
	w := new(tabwriter.Writer)

	ret := new(bytes.Buffer)
	w.Init(ret, 0, 8, 1, ' ', 0)

	// Loop through and set table
	var lastType Type
	lastType = -1
	for _, ioc := range iocs {
		if ioc.Type != lastType {
			fmt.Fprintln(w, "# "+ioc.Type.String())
			lastType = ioc.Type
		}
		fmt.Fprintln(w, ioc.IOC+"\t"+ioc.Type.String())
	}

	w.Flush()
	return ret.String()
}

// formatIOCsStats returns counts per IOC Type as a string.
func formatIOCsStats(iocs []*IOC) string {
	stats := countsByType(iocs)
	ret := &bytes.Buffer{}
	for iocType, count := range stats {
		fmt.Fprintf(ret, "%s: %d\n", iocType.String(), count)
	}
	return ret.String()
}

// countsByType returns a map with counts per IOC Type.
func countsByType(iocs []*IOC) map[Type]int {
	stats := make(map[Type]int, len(Types))
	for _, i := range iocs {
		stats[i.Type]++
	}
	return stats
}
