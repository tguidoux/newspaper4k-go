package ioc

import (
	"regexp"
	"strings"
)

// Defang Structures to defang using our standardized defanging
type defangMap map[Type][]defangPair
type defangPair struct {
	defanged string
	fanged   string
}

// Our standardized defangs use all square brackets []
var defangReplacements = defangMap{
	Email: {
		{"[AT]", "@"},
		{"[.]", "."},
	},
	Domain: {
		{"[.]", "."},
	},
	IPv4: {
		{"[.]", "."},
	},
	IPv6: {
		{"[:]", ":"},
	},
	URL: {
		{"hxxp", "http"},
		{"[://]", "://"},
		{"[.]", "."},
	},
}

// toDefanged converts an IOC to the standardized defanged form.
func (ioc *IOC) toDefanged() *IOC {
	copy := *ioc
	ioc = &copy

	if replacements, ok := defangReplacements[ioc.Type]; ok {
		for _, r := range replacements {
			ioc.IOC = strings.ReplaceAll(ioc.IOC, r.fanged, r.defanged)
		}
	}

	return ioc
}

// Fang Structures to fang using all possible defangs
type regexReplacement struct {
	pattern *regexp.Regexp
	replace string
}

var dotReplace = regexReplacement{regexp.MustCompile(`\ *[([]?\ *((dot)|\.)\ *[])]?\ *`), "."}
var fangReplacements = map[Type][]regexReplacement{
	Email: {
		{regexp.MustCompile(`(\ ?[([]?\ ?(([aA][tT])|@)\ ?[])]?\ ?)`), "@"},
		dotReplace,
	},
	Domain: {
		dotReplace,
	},
	IPv4: {
		dotReplace,
	},
	IPv6: {
		{regexp.MustCompile(`[([]:[])]`), "."},
	},
	URL: {
		{regexp.MustCompile(`hxxp`), "http"},
		{regexp.MustCompile(`[[(]://[)]]`), "://"},
		dotReplace,
	},
}

// toFanged converts a defanged IOC back to its fanged form.
// Example: john[AT]gmail[.]com -> john@gmail.com
func (ioc *IOC) toFanged() *IOC {
	copy := *ioc
	ioc = &copy

	if replacements, ok := defangReplacements[ioc.Type]; ok {
		for _, r := range replacements {
			ioc.IOC = strings.ReplaceAll(ioc.IOC, r.defanged, r.fanged)
		}
	}

	if replacements, ok := fangReplacements[ioc.Type]; ok {
		for _, rr := range replacements {
			offset := 0
			locs := rr.pattern.FindAllStringIndex(ioc.IOC, -1)
			for _, loc := range locs {
				startSize := len(ioc.IOC)
				ioc.IOC = ioc.IOC[0:loc[0]-offset] + rr.replace + ioc.IOC[loc[1]-offset:len(ioc.IOC)]
				offset += startSize - len(ioc.IOC)
			}
		}
	}

	return ioc
}

// isFanged reports whether the IOC is already fanged. Some types are considered
// non-fangable and always return false.
func (ioc *IOC) isFanged() bool {
	switch ioc.Type {
	case Bitcoin, MD5, SHA1, SHA256, SHA512, File, CVE:
		return false
	}

	// If converting to fanged doesn't change the string, it was already fanged.
	if ioc.toFanged().IOC == ioc.IOC {
		return true
	}
	return false
}
