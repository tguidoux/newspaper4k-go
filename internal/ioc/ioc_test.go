package ioc

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	testify "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseIOC(t *testing.T) {
	tests := []struct {
		ioc  string
		want *IOC
	}{
		{
			"test@test.com",
			&IOC{Type: Email, IOC: "test@test.com"},
		},
		{
			"https://test.com/asdf",
			&IOC{Type: URL, IOC: "https://test.com/asdf"},
		},
	}

	for i, test := range tests {
		if out := ParseIOC(test.ioc); !reflect.DeepEqual(out, test.want) {
			t.Errorf("Error on test %d", i)
		}
	}
}

func TestGetIOCs(t *testing.T) {
	// Test without standardizing defangs
	tests := []struct {
		input string
		want  []*IOC
	}{
		// Bitcoin
		{"1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", []*IOC{{"1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", Bitcoin}}},
		{"1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2\"", []*IOC{{"1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", Bitcoin}}},
		{"1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2:", []*IOC{{"1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", Bitcoin}}},
		{"3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy", []*IOC{{"3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy", Bitcoin}}},
		{"bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq", []*IOC{{"bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq", Bitcoin}}},
		// Hashes
		{"874058e8d8582bf85c115ce319c5b0af", []*IOC{{"874058e8d8582bf85c115ce319c5b0af", MD5}}},
		{"751641b4e4e6cc30f497639eee583b5b392451fb", []*IOC{{"751641b4e4e6cc30f497639eee583b5b392451fb", SHA1}}},
		{"4708a032833b054e4237392c4d75e41b4775dc67845e939487ab39f92de847ce", []*IOC{{"4708a032833b054e4237392c4d75e41b4775dc67845e939487ab39f92de847ce", SHA256}}},
		{"b4ae21eb1e337658368add0d2c177eb366123c8f961325dd1e67492acac84261be29594c1260bb3f249a3dcdf0372e381f2a23c4d026a91b4a7d66c949ddffad", []*IOC{{"b4ae21eb1e337658368add0d2c177eb366123c8f961325dd1e67492acac84261be29594c1260bb3f249a3dcdf0372e381f2a23c4d026a91b4a7d66c949ddffad", SHA512}}},
		{"874058e8d8582bf85c115ce319c5b0a", nil},

		// IPs
		{"8.8.8.8", []*IOC{{"8.8.8.8", IPv4}}},
		{"\"8.8.8.8\"", []*IOC{{"8.8.8.8", IPv4}}},
		{"1.1.1.1", []*IOC{{"1.1.1.1", IPv4}}},
		{"1(.)1.1(.)1", []*IOC{{"1(.)1.1(.)1", IPv4}}},
		{"1(.)1(.)1(.)1", []*IOC{{"1(.)1(.)1(.)1", IPv4}}},
		{"1(.)1[.]1(.)1", []*IOC{{"1(.)1[.]1(.)1", IPv4}}},
		{"10(.)252[.]255(.)255", []*IOC{{"10(.)252[.]255(.)255", IPv4}}},
		{"1.1[.]1[.]1", []*IOC{{"1.1[.]1[.]1", IPv4}}},
		{"1.2[.)3.4", []*IOC{{"1.2[.)3.4", IPv4}}},
		{"1.2[.)3(.)4", []*IOC{{"1.2[.)3(.)4", IPv4}}},
		{"1.2([.])3.4", nil},
		{"2001:0db8:0000:0000:0000:ff00:0042:8329", []*IOC{{"2001:0db8:0000:0000:0000:ff00:0042:8329", IPv6}}},
		{"2001:db8::ff00:42:8329", []*IOC{{"2001:db8::ff00:42:8329", IPv6}}},
		{"::1", []*IOC{{"::1", IPv6}}},
		{"10::1", []*IOC{{"10::1", IPv6}}},
		{"0010::1", []*IOC{{"0010::1", IPv6}}},
		{"300.300.300.300", nil},

		// Emails
		{"test@test.com", []*IOC{{"test.com", Domain}, {"test@test.com", Email}}},
		{"\"test@test.com\"", []*IOC{{"test.com", Domain}, {"test@test.com", Email}}},
		{"test[@]test.com", []*IOC{{"test.com", Domain}, {"test[@]test.com", Email}}},
		{"test(@)test.com", []*IOC{{"test.com", Domain}, {"test(@)test.com", Email}}},

		// Domains
		{"example.com", []*IOC{{"example.com", Domain}}},
		{"www.us-cert.gov", []*IOC{{"www.us-cert.gov", Domain}}},
		{"threat.int.test.blah.blahblah.blahblah.amazon.microsoft.test.com", []*IOC{{"threat.int.test.blah.blahblah.blahblah.amazon.microsoft.test.com", Domain}}},
		{"threat.int.test.blah.blahblah.blahblah.amazon.microsoft.test.com.invalid", []*IOC{{"threat.int.test.blah.blahblah.blahblah.amazon.microsoft.test.com", Domain}}},
		{"test(.)com", []*IOC{{"test(.)com", Domain}}},
		{"test[.]com", []*IOC{{"test[.]com", Domain}}},
		{"test(.)example(.)com", []*IOC{{"test(.)example(.)com", Domain}}},
		{"test(.)example[.]com", []*IOC{{"test(.)example[.]com", Domain}}},
		{"test(.]com", []*IOC{{"test(.]com", Domain}}},
		{"example.pumpkin", nil},

		// Links
		{"\"http://www.example.com/foo/bar?baz=1\"", []*IOC{{"www.example.com", Domain}, {"http://www.example.com/foo/bar?baz=1", URL}}},
		{"http://www.example.com/foo/bar?baz=1", []*IOC{{"www.example.com", Domain}, {"http://www.example.com/foo/bar?baz=1", URL}}},
		{"http://www.example.com", []*IOC{{"www.example.com", Domain}, {"http://www.example.com", URL}}},
		{"http[://]example.com/f", []*IOC{{"example.com", Domain}, {"http[://]example.com/f", URL}}},
		{"http://www.example.com/foo", []*IOC{{"www.example.com", Domain}, {"http://www.example.com/foo", URL}}},
		{"http://www.example.com/foo/", []*IOC{{"www.example.com", Domain}, {"http://www.example.com/foo", URL}}},
		{"https://www.example.com/foo/bar?baz=1", []*IOC{{"www.example.com", Domain}, {"https://www.example.com/foo/bar?baz=1", URL}}},
		{"https://www.example.com", []*IOC{{"www.example.com", Domain}, {"https://www.example.com", URL}}},
		{"https://www.example.com/foo", []*IOC{{"www.example.com", Domain}, {"https://www.example.com/foo", URL}}},
		{"https://www.example.com/foo/", []*IOC{{"www.example.com", Domain}, {"https://www.example.com/foo", URL}}},
		{"https://www[.]example[.]com/foo/", []*IOC{{"www[.]example[.]com", Domain}, {"https://www[.]example[.]com/foo", URL}}},
		{"https://www[.]example[.]com/foo/", []*IOC{{"www[.]example[.]com", Domain}, {"https://www[.]example[.]com/foo", URL}}},
		{"hxxps://185[.]159[.]82[.]15/hollyhole/c644[.]php", []*IOC{{"185[.]159[.]82[.]15", IPv4}, {"hxxps://185[.]159[.]82[.]15/hollyhole/c644[.]php", URL}}},

		// Files
		{"test.doc", []*IOC{{"test.doc", File}}},
		{"test.two.doc", []*IOC{{"test.two.doc", File}}},
		{"test.dll", []*IOC{{"test.dll", File}}},
		{"test.exe", []*IOC{{"test.exe", File}}},
		{"begin.test.test.exe", []*IOC{{"begin.test.test.exe", File}}},
		{"LOGSystem.Agent.Service.exe", []*IOC{{"LOGSystem.Agent.Service.exe", File}}},
		{"test.swf", []*IOC{{"test.swf", File}}},
		{"test.two.swf", []*IOC{{"test.two.swf", File}}},
		{"test.jpg", []*IOC{{"test.jpg", File}}},
		{"LOGSystem.Agent.Service.jpg", []*IOC{{"LOGSystem.Agent.Service.jpg", File}}},
		{"test.plist", []*IOC{{"test.plist", File}}},
		{"test.two.plist", []*IOC{{"test.two.plist", File}}},
		{"test.html", []*IOC{{"test.html", File}}},
		{"test.two.html", []*IOC{{"test.two.html", File}}},
		{"test.zip", []*IOC{{"test.zip", File}}},
		{"test.two.zip", []*IOC{{"test.two.zip", File}}},
		{"test.tar.gz", []*IOC{{"test.tar.gz", File}}},
		{"test.two.tar.gz", []*IOC{{"test.two.tar.gz", File}}},
		{".test.", nil},
		{"test.dl", nil},
		{"..", nil},
		{".", nil},
		{"example.pumpkin", nil},

		// Utility
		{"CVE-1800-0000", []*IOC{{"CVE-1800-0000", CVE}}},
		{"CVE-2016-0000", []*IOC{{"CVE-2016-0000", CVE}}},
		{"CVE-2100-0000", []*IOC{{"CVE-2100-0000", CVE}}},
		{"CVE-2016-00000", []*IOC{{"CVE-2016-00000", CVE}}},
		{"CVE-20100-0000", nil},
		{"CAPEC-13", []*IOC{{"CAPEC-13", CAPEC}}},
		{"CWE-200", []*IOC{{"CWE-200", CWE}}},
		{"cpe:2.3:a:openbsd:openssh:7.5:-:*:*:*:*:*:*", []*IOC{{"cpe:2.3:a:openbsd:openssh:7.5:-:*:*:*:*:*:*", CPE}}},
		{"cpe:/a:openbsd:openssh:7.5:-", []*IOC{{"cpe:/a:openbsd:openssh:7.5:-", CPE}}},
		{"cpe:/a:microsoft:internet_explorer:8.%02:sp%01", []*IOC{{"cpe:/a:microsoft:internet_explorer:8.%02:sp%01", CPE}}},
		{"cpe:/a:hp:insight_diagnostics:7.4.0.1570:-:~~online~win2003~x64~", []*IOC{{"cpe:/a:hp:insight_diagnostics:7.4.0.1570:-:~~online~win2003~x64~", CPE}}},
		{"cpe:2.3:a:microsoft:internet_explorer:8.0.6001:beta:*:*:*:*:*:*", []*IOC{{"cpe:2.3:a:microsoft:internet_explorer:8.0.6001:beta:*:*:*:*:*:*", CPE}}},
		{"cpe:2.3:a:microsoft:internet_explorer:8.*:sp?:*:*:*:*:*:*", []*IOC{{"cpe:2.3:a:microsoft:internet_explorer:8.*:sp?:*:*:*:*:*:*", CPE}}},
		{"cpe:2.3:a:hp:insight:7.4.0.1570:-:*:*:online:win2003:x64:*", []*IOC{{"cpe:2.3:a:hp:insight:7.4.0.1570:-:*:*:online:win2003:x64:*", CPE}}},
		{"cpe:2.3:a:hp:openview_network_manager:7.51:*:*:*:*:linux:*:*", []*IOC{{"cpe:2.3:a:hp:openview_network_manager:7.51:*:*:*:*:linux:*:*", CPE}}},
		{"cpe:2.3:a:foo\\\\bar:big\\$money_2010:*:*:*:*:special:ipod_touch:80gb:*", []*IOC{{"cpe:2.3:a:foo\\\\bar:big\\$money_2010:*:*:*:*:special:ipod_touch:80gb:*", CPE}}},

		// Misc
		{"1.1.1.1 google.com 1.1.1.1", []*IOC{
			{"google.com", Domain},
			{"1.1.1.1", IPv4},
		}},
		{"http://google.com/test/URL 1.3.2.1 Email@test.domain.com sogahgwugh4a49uhgaspd aiweawfa.asdas afw## )#@*)@$*(@ filename.exe", []*IOC{
			{"google.com", Domain},
			{"test.domain.com", Domain},
			{"Email@test.domain.com", Email},
			{"1.3.2.1", IPv4},
			{"http://google.com/test/URL", URL},
			{"filename.exe", File},
		}},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			if iocs := ExtractIOCs(test.input, true); !testify.ElementsMatch(t, iocs, test.want) {
				t.Errorf("IOCType(%q), found %v =/= wanted %v", test.input, iocs, test.want)
			}
		})
	}
}
func TestStandardizeDefangs(t *testing.T) {
	testsStandardizedDefangs := []struct {
		input string
		want  []*IOC
	}{
		// IPs
		{"8.8.8.8", []*IOC{{"8[.]8[.]8[.]8", IPv4}}},
		{"\"8.8.8.8\"", []*IOC{{"8[.]8[.]8[.]8", IPv4}}},
		{"1.1.1.1", []*IOC{{"1[.]1[.]1[.]1", IPv4}}},
		{"1(.)1.1(.)1", []*IOC{{"1[.]1[.]1[.]1", IPv4}}},
		{"1(.)1(.)1(.)1", []*IOC{{"1[.]1[.]1[.]1", IPv4}}},
		{"1(.)1[.]1(.)1", []*IOC{{"1[.]1[.]1[.]1", IPv4}}},
		{"10(.)252[.]255(.)255", []*IOC{{"10[.]252[.]255[.]255", IPv4}}},
		{"1.1[.]1[.]1", []*IOC{{"1[.]1[.]1[.]1", IPv4}}},
		{"1.2[.)3.4", []*IOC{{"1[.]2[.]3[.]4", IPv4}}},
		{"1.2[.)3(.)4", []*IOC{{"1[.]2[.]3[.]4", IPv4}}},
	}

	for _, test := range testsStandardizedDefangs {
		t.Run(test.input, func(t *testing.T) {
			iocs := ExtractIOCs(test.input, true)
			NormalizeDefangs(iocs)
			if !reflect.DeepEqual(iocs, test.want) {
				t.Errorf("[standardizedDefang=true] IOCType(%q), found %v =/= wanted %v", test.input, iocs, test.want)
			}
		})

	}
}
func TestAllFanged(t *testing.T) {
	testsAllFanged := []struct {
		input string
		want  []*IOC
	}{
		// IPs
		{"8.8.8.8", nil},
		{"\"8.8.8.8\"", nil},
		{"1.1.1.1", nil},
		{"1(.)1.1(.)1", []*IOC{{"1[.]1[.]1[.]1", IPv4}}},
		{"1(.)1(.)1(.)1", []*IOC{{"1[.]1[.]1[.]1", IPv4}}},
		{"1(.)1[.]1(.)1", []*IOC{{"1[.]1[.]1[.]1", IPv4}}},
		{"10(.)252[.]255(.)255", []*IOC{{"10[.]252[.]255[.]255", IPv4}}},
		{"1.1[.]1[.]1", []*IOC{{"1[.]1[.]1[.]1", IPv4}}},
		{"1.2[.)3.4", []*IOC{{"1[.]2[.]3[.]4", IPv4}}},
		{"1.2[.)3(.)4", []*IOC{{"1[.]2[.]3[.]4", IPv4}}},
	}

	for _, test := range testsAllFanged {
		t.Run(test.input, func(t *testing.T) {
			iocs := ExtractIOCs(test.input, false)
			NormalizeDefangs(iocs)
			if !reflect.DeepEqual(iocs, test.want) {
				t.Errorf("[allFanged=false] IOCType(%q), found %v =/= wanted %v", test.input, iocs, test.want)
			}
		})
	}
}

func TestGetIOCsReader(t *testing.T) {
	// Test without standardizing defangs
	tests := []struct {
		input string
		want  []*IOC
	}{
		// Bitcoin
		{"1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", []*IOC{{"1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", Bitcoin}}},
		{"1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2\"", []*IOC{{"1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", Bitcoin}}},
		{"1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2:", []*IOC{{"1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", Bitcoin}}},
		{"3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy", []*IOC{{"3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy", Bitcoin}}},
		{"bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq", []*IOC{{"bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq", Bitcoin}}},
		// Hashes
		{"874058e8d8582bf85c115ce319c5b0af", []*IOC{{"874058e8d8582bf85c115ce319c5b0af", MD5}}},
		{"751641b4e4e6cc30f497639eee583b5b392451fb", []*IOC{{"751641b4e4e6cc30f497639eee583b5b392451fb", SHA1}}},
		{"4708a032833b054e4237392c4d75e41b4775dc67845e939487ab39f92de847ce", []*IOC{{"4708a032833b054e4237392c4d75e41b4775dc67845e939487ab39f92de847ce", SHA256}}},
		{"b4ae21eb1e337658368add0d2c177eb366123c8f961325dd1e67492acac84261be29594c1260bb3f249a3dcdf0372e381f2a23c4d026a91b4a7d66c949ddffad", []*IOC{{"b4ae21eb1e337658368add0d2c177eb366123c8f961325dd1e67492acac84261be29594c1260bb3f249a3dcdf0372e381f2a23c4d026a91b4a7d66c949ddffad", SHA512}}},
		{"874058e8d8582bf85c115ce319c5b0a", nil},

		// IPs
		{"8.8.8.8", []*IOC{{"8.8.8.8", IPv4}}},
		{"\"8.8.8.8\"", []*IOC{{"8.8.8.8", IPv4}}},
		{"1.1.1.1", []*IOC{{"1.1.1.1", IPv4}}},
		{"1(.)1.1(.)1", []*IOC{{"1(.)1.1(.)1", IPv4}}},
		{"1(.)1(.)1(.)1", []*IOC{{"1(.)1(.)1(.)1", IPv4}}},
		{"1(.)1[.]1(.)1", []*IOC{{"1(.)1[.]1(.)1", IPv4}}},
		{"10(.)252[.]255(.)255", []*IOC{{"10(.)252[.]255(.)255", IPv4}}},
		{"1.1[.]1[.]1", []*IOC{{"1.1[.]1[.]1", IPv4}}},
		{"1.2[.)3.4", []*IOC{{"1.2[.)3.4", IPv4}}},
		{"1.2[.)3(.)4", []*IOC{{"1.2[.)3(.)4", IPv4}}},
		{"1.2([.])3.4", nil},
		{"2001:0db8:0000:0000:0000:ff00:0042:8329", []*IOC{{"2001:0db8:0000:0000:0000:ff00:0042:8329", IPv6}}},
		{"2001:db8::ff00:42:8329", []*IOC{{"2001:db8::ff00:42:8329", IPv6}}},
		{"::1", []*IOC{{"::1", IPv6}}},
		{"10::1", []*IOC{{"10::1", IPv6}}},
		{"0010::1", []*IOC{{"0010::1", IPv6}}},
		{"300.300.300.300", nil},

		// Emails
		{"test@test.com", []*IOC{{"test.com", Domain}, {"test@test.com", Email}}},
		{"\"test@test.com\"", []*IOC{{"test.com", Domain}, {"test@test.com", Email}}},
		{"test[@]test.com", []*IOC{{"test.com", Domain}, {"test[@]test.com", Email}}},
		{"test(@)test.com", []*IOC{{"test.com", Domain}, {"test(@)test.com", Email}}},

		// Domains
		{"example.com", []*IOC{{"example.com", Domain}}},
		{"www.us-cert.gov", []*IOC{{"www.us-cert.gov", Domain}}},
		{"threat.int.test.blah.blahblah.blahblah.amazon.microsoft.test.com", []*IOC{{"threat.int.test.blah.blahblah.blahblah.amazon.microsoft.test.com", Domain}}},
		{"threat.int.test.blah.blahblah.blahblah.amazon.microsoft.test.com.invalid", []*IOC{{"threat.int.test.blah.blahblah.blahblah.amazon.microsoft.test.com", Domain}}},
		{"test(.)com", []*IOC{{"test(.)com", Domain}}},
		{"test[.]com", []*IOC{{"test[.]com", Domain}}},
		{"test(.)example(.)com", []*IOC{{"test(.)example(.)com", Domain}}},
		{"test(.)example[.]com", []*IOC{{"test(.)example[.]com", Domain}}},
		{"test(.]com", []*IOC{{"test(.]com", Domain}}},
		{"example.pumpkin", nil},

		// Links
		{"\"http://www.example.com/foo/bar?baz=1\"", []*IOC{{"www.example.com", Domain}, {"http://www.example.com/foo/bar?baz=1", URL}}},
		{"http://www.example.com/foo/bar?baz=1", []*IOC{{"www.example.com", Domain}, {"http://www.example.com/foo/bar?baz=1", URL}}},
		{"http://www.example.com", []*IOC{{"www.example.com", Domain}, {"http://www.example.com", URL}}},
		{"http[://]example.com/f", []*IOC{{"example.com", Domain}, {"http[://]example.com/f", URL}}},
		{"http://www.example.com/foo", []*IOC{{"www.example.com", Domain}, {"http://www.example.com/foo", URL}}},
		{"http://www.example.com/foo/", []*IOC{{"www.example.com", Domain}, {"http://www.example.com/foo", URL}}},
		{"https://www.example.com/foo/bar?baz=1", []*IOC{{"www.example.com", Domain}, {"https://www.example.com/foo/bar?baz=1", URL}}},
		{"https://www.example.com", []*IOC{{"www.example.com", Domain}, {"https://www.example.com", URL}}},
		{"https://www.example.com/foo", []*IOC{{"www.example.com", Domain}, {"https://www.example.com/foo", URL}}},
		{"https://www.example.com/foo/", []*IOC{{"www.example.com", Domain}, {"https://www.example.com/foo", URL}}},
		{"https://www[.]example[.]com/foo/", []*IOC{{"www[.]example[.]com", Domain}, {"https://www[.]example[.]com/foo", URL}}},
		{"https://www[.]example[.]com/foo/", []*IOC{{"www[.]example[.]com", Domain}, {"https://www[.]example[.]com/foo", URL}}},
		{"hxxps://185[.]159[.]82[.]15/hollyhole/c644[.]php", []*IOC{{"185[.]159[.]82[.]15", IPv4}, {"hxxps://185[.]159[.]82[.]15/hollyhole/c644[.]php", URL}}},

		// Files
		{"test.doc", []*IOC{{"test.doc", File}}},
		{"test.two.doc", []*IOC{{"test.two.doc", File}}},
		{"test.dll", []*IOC{{"test.dll", File}}},
		{"test.exe", []*IOC{{"test.exe", File}}},
		{"begin.test.test.exe", []*IOC{{"begin.test.test.exe", File}}},
		{"LOGSystem.Agent.Service.exe", []*IOC{{"LOGSystem.Agent.Service.exe", File}}},
		{"test.swf", []*IOC{{"test.swf", File}}},
		{"test.two.swf", []*IOC{{"test.two.swf", File}}},
		{"test.jpg", []*IOC{{"test.jpg", File}}},
		{"LOGSystem.Agent.Service.jpg", []*IOC{{"LOGSystem.Agent.Service.jpg", File}}},
		{"test.plist", []*IOC{{"test.plist", File}}},
		{"test.two.plist", []*IOC{{"test.two.plist", File}}},
		{"test.html", []*IOC{{"test.html", File}}},
		{"test.two.html", []*IOC{{"test.two.html", File}}},
		{"test.zip", []*IOC{{"test.zip", File}}},
		{"test.two.zip", []*IOC{{"test.two.zip", File}}},
		{"test.tar.gz", []*IOC{{"test.tar.gz", File}}},
		{"test.two.tar.gz", []*IOC{{"test.two.tar.gz", File}}},
		{".test.", nil},
		{"test.dl", nil},
		{"..", nil},
		{".", nil},
		{"example.pumpkin", nil},

		// Utility
		{"CVE-1800-0000", []*IOC{{"CVE-1800-0000", CVE}}},
		{"CVE-2016-0000", []*IOC{{"CVE-2016-0000", CVE}}},
		{"CVE-2100-0000", []*IOC{{"CVE-2100-0000", CVE}}},
		{"CVE-2016-00000", []*IOC{{"CVE-2016-00000", CVE}}},
		{"CVE-20100-0000", nil},
		{"CAPEC-13", []*IOC{{"CAPEC-13", CAPEC}}},
		{"CWE-200", []*IOC{{"CWE-200", CWE}}},
		{"cpe:2.3:a:openbsd:openssh:7.5:-:*:*:*:*:*:*", []*IOC{{"cpe:2.3:a:openbsd:openssh:7.5:-:*:*:*:*:*:*", CPE}}},
		{"cpe:/a:openbsd:openssh:7.5:-", []*IOC{{"cpe:/a:openbsd:openssh:7.5:-", CPE}}},
		{"cpe:/a:microsoft:internet_explorer:8.%02:sp%01", []*IOC{{"cpe:/a:microsoft:internet_explorer:8.%02:sp%01", CPE}}},
		{"cpe:/a:hp:insight_diagnostics:7.4.0.1570:-:~~online~win2003~x64~", []*IOC{{"cpe:/a:hp:insight_diagnostics:7.4.0.1570:-:~~online~win2003~x64~", CPE}}},
		{"cpe:2.3:a:microsoft:internet_explorer:8.0.6001:beta:*:*:*:*:*:*", []*IOC{{"cpe:2.3:a:microsoft:internet_explorer:8.0.6001:beta:*:*:*:*:*:*", CPE}}},
		{"cpe:2.3:a:microsoft:internet_explorer:8.*:sp?:*:*:*:*:*:*", []*IOC{{"cpe:2.3:a:microsoft:internet_explorer:8.*:sp?:*:*:*:*:*:*", CPE}}},
		{"cpe:2.3:a:hp:insight:7.4.0.1570:-:*:*:online:win2003:x64:*", []*IOC{{"cpe:2.3:a:hp:insight:7.4.0.1570:-:*:*:online:win2003:x64:*", CPE}}},
		{"cpe:2.3:a:hp:openview_network_manager:7.51:*:*:*:*:linux:*:*", []*IOC{{"cpe:2.3:a:hp:openview_network_manager:7.51:*:*:*:*:linux:*:*", CPE}}},
		{"cpe:2.3:a:foo\\\\bar:big\\$money_2010:*:*:*:*:special:ipod_touch:80gb:*", []*IOC{{"cpe:2.3:a:foo\\\\bar:big\\$money_2010:*:*:*:*:special:ipod_touch:80gb:*", CPE}}},

		// Misc
		{"1.1.1.1 google.com 1.1.1.1", []*IOC{
			{"google.com", Domain},
			{"1.1.1.1", IPv4},
		}},
		{"http://google.com/test/URL 1.3.2.1 Email@test.domain.com sogahgwugh4a49uhgaspd aiweawfa.asdas afw## )#@*)@$*(@ filename.exe", []*IOC{
			{"google.com", Domain},
			{"test.domain.com", Domain},
			{"Email@test.domain.com", Email},
			{"1.3.2.1", IPv4},
			{"http://google.com/test/URL", URL},
			{"filename.exe", File},
		}},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			if iocs := ExtractIOCs(test.input, true); !testify.ElementsMatch(t, iocs, test.want) {
				t.Errorf("IOCType(%q), found %v =/= wanted %v", test.input, iocs, test.want)
			}
		})
	}
}

// TestParseIOCEdgeCases tests ParseIOC with various edge cases
func TestParseIOCEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		ioc  string
		want *IOC
	}{
		{
			name: "empty string",
			ioc:  "",
			want: &IOC{},
		},
		{
			name: "whitespace only",
			ioc:  "   \n\t  ",
			want: &IOC{},
		},
		{
			name: "no IOC found",
			ioc:  "this is just plain text with no IOCs",
			want: &IOC{},
		},
		{
			name: "multiple IOCs returns highest priority",
			ioc:  "Visit google.com or email test@google.com for more info",
			want: &IOC{Type: Email, IOC: "test@google.com"},
		},
		{
			name: "bitcoin with surrounding text",
			ioc:  "The wallet address is 1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2 for donations",
			want: &IOC{Type: Bitcoin, IOC: " 1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2"},
		},
		{
			name: "very long domain",
			ioc:  "verylongdomainnameconsistingofmanycharacters.example.com",
			want: &IOC{Type: Domain, IOC: "verylongdomainnameconsistingofmanycharacters.example.com"},
		},
		{
			name: "unicode characters",
			ioc:  "test@пример.испытание",
			want: &IOC{Type: Unknown, IOC: ""},
		},
		{
			name: "special characters in domain",
			ioc:  "test.domain-with-hyphens.com",
			want: &IOC{Type: Domain, IOC: "test.domain-with-hyphens.com"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ParseIOC(test.ioc)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("ParseIOC(%q) = %v, want %v", test.ioc, got, test.want)
			}
		})
	}
}

// TestIOCString tests the String method of IOC
func TestIOCString(t *testing.T) {
	tests := []struct {
		name string
		ioc  *IOC
		want string
	}{
		{
			name: "normal IOC",
			ioc:  &IOC{IOC: "test.com", Type: Domain},
			want: "test.com|Domain",
		},
		{
			name: "empty IOC",
			ioc:  &IOC{},
			want: "|Unknown",
		},
		{
			name: "bitcoin address",
			ioc:  &IOC{IOC: "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", Type: Bitcoin},
			want: "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2|Bitcoin",
		},
		{
			name: "email with special chars",
			ioc:  &IOC{IOC: "test+tag@example.com", Type: Email},
			want: "test+tag@example.com|Email",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.ioc.String()
			if got != test.want {
				t.Errorf("IOC.String() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestType_String(t *testing.T) {
	tests := []struct {
		name string
		t    Type
		want string
	}{
		{"Unknown", Unknown, "Unknown"},
		{"Bitcoin", Bitcoin, "Bitcoin"},
		{"MD5", MD5, "MD5"},
		{"SHA1", SHA1, "SHA1"},
		{"SHA256", SHA256, "SHA256"},
		{"SHA512", SHA512, "SHA512"},
		{"Domain", Domain, "Domain"},
		{"Email", Email, "Email"},
		{"IPv4", IPv4, "IPv4"},
		{"IPv6", IPv6, "IPv6"},
		{"URL", URL, "URL"},
		{"File", File, "File"},
		{"CVE", CVE, "CVE"},
		{"CAPEC", CAPEC, "CAPEC"},
		{"CWE", CWE, "CWE"},
		{"CPE", CPE, "CPE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.String(); got != tt.want {
				t.Errorf("Type.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortByType(t *testing.T) {
	iocs := []*IOC{
		{IOC: "test@domain.com", Type: Email},
		{IOC: "example.com", Type: Domain},
		{IOC: "192.168.1.1", Type: IPv4},
	}
	sorted := sortByType(iocs)
	expectedOrder := []Type{Domain, Email, IPv4}
	for i, ioc := range sorted {
		if ioc.Type != expectedOrder[i] {
			t.Errorf("sortByType() at index %d = %v, want %v", i, ioc.Type, expectedOrder[i])
		}
	}
}

func TestFormatIOCs(t *testing.T) {
	iocs := []*IOC{
		{IOC: "example.com", Type: Domain},
		{IOC: "test@domain.com", Type: Email},
	}
	csv := FormatIOCs(iocs, "csv")
	if !strings.Contains(csv, "example.com|Domain") {
		t.Errorf("FormatIOCs csv does not contain expected output")
	}
	table := FormatIOCs(iocs, "table")
	if !strings.Contains(table, "# Domain") {
		t.Errorf("FormatIOCs table does not contain expected output")
	}
}

func TestFormatIOCsCSV(t *testing.T) {
	iocs := []*IOC{
		{IOC: "example.com", Type: Domain},
		{IOC: "test@domain.com", Type: Email},
	}
	csv := formatIOCsCSV(iocs)
	expected := "example.com|Domain\ntest@domain.com|Email"
	if csv != expected {
		t.Errorf("formatIOCsCSV() = %v, want %v", csv, expected)
	}
}

func TestFormatIOCsTable(t *testing.T) {
	iocs := []*IOC{
		{IOC: "example.com", Type: Domain},
		{IOC: "test@domain.com", Type: Email},
	}
	table := formatIOCsTable(iocs)
	if !strings.Contains(table, "# Domain") || !strings.Contains(table, "# Email") {
		t.Errorf("formatIOCsTable() does not contain expected headers")
	}
}

func TestFormatIOCsStats(t *testing.T) {
	iocs := []*IOC{
		{IOC: "example.com", Type: Domain},
		{IOC: "test.com", Type: Domain},
		{IOC: "test@domain.com", Type: Email},
	}
	stats := formatIOCsStats(iocs)
	if !strings.Contains(stats, "Domain: 2") || !strings.Contains(stats, "Email: 1") {
		t.Errorf("formatIOCsStats() = %v, does not contain expected stats", stats)
	}
}

func TestCountsByType(t *testing.T) {
	iocs := []*IOC{
		{IOC: "example.com", Type: Domain},
		{IOC: "test.com", Type: Domain},
		{IOC: "test@domain.com", Type: Email},
	}
	counts := countsByType(iocs)
	if counts[Domain] != 2 || counts[Email] != 1 {
		t.Errorf("countsByType() = %v, want Domain:2, Email:1", counts)
	}
}

// TestDefangEdgeCases tests Defang method with edge cases
func TestDefangEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		ioc  *IOC
		want *IOC
	}{
		{
			name: "already defanged",
			ioc:  &IOC{IOC: "test[AT]example[.]com", Type: Email},
			want: &IOC{IOC: "test[AT]example[[.]]com", Type: Email},
		},
		{
			name: "mixed defanging styles",
			ioc:  &IOC{IOC: "test(@)example[.]com", Type: Email},
			want: &IOC{IOC: "test([AT])example[[.]]com", Type: Email},
		},
		{
			name: "unsupported type",
			ioc:  &IOC{IOC: "test", Type: Bitcoin},
			want: &IOC{IOC: "test", Type: Bitcoin},
		},
		{
			name: "empty string",
			ioc:  &IOC{IOC: "", Type: Domain},
			want: &IOC{IOC: "", Type: Domain},
		},
		{
			name: "IPv6 complex",
			ioc:  &IOC{IOC: "2001:db8::1", Type: IPv6},
			want: &IOC{IOC: "2001[:]db8[:][:]1", Type: IPv6},
		},
		{
			name: "URL with multiple protocols",
			ioc:  &IOC{IOC: "http://test.com", Type: URL},
			want: &IOC{IOC: "hxxp[://]test[.]com", Type: URL},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.ioc.toDefanged()
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Defang() = %v, want %v", got, test.want)
			}
		})
	}
}

// TestFangEdgeCases tests Fang method with edge cases
func TestFangEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		ioc  *IOC
		want *IOC
	}{
		{
			name: "already fanged",
			ioc:  &IOC{IOC: "test@example.com", Type: Email},
			want: &IOC{IOC: "test@example.com", Type: Email},
		},
		{
			name: "multiple defang styles",
			ioc:  &IOC{IOC: "test[AT]example[.]com", Type: Email},
			want: &IOC{IOC: "test@example.com", Type: Email},
		},
		{
			name: "complex defanging",
			ioc:  &IOC{IOC: "test (at) example (dot) com", Type: Email},
			want: &IOC{IOC: "test@example.com", Type: Email},
		},
		{
			name: "unsupported type",
			ioc:  &IOC{IOC: "test", Type: Bitcoin},
			want: &IOC{IOC: "test", Type: Bitcoin},
		},
		{
			name: "empty string",
			ioc:  &IOC{IOC: "", Type: Domain},
			want: &IOC{IOC: "", Type: Domain},
		},
		{
			name: "IPv6 with brackets",
			ioc:  &IOC{IOC: "2001[:]db8[:][:]1", Type: IPv6},
			want: &IOC{IOC: "2001:db8::1", Type: IPv6},
		},
		{
			name: "URL defanged",
			ioc:  &IOC{IOC: "hxxp[://]test[.]com", Type: URL},
			want: &IOC{IOC: "http://test.com", Type: URL},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.ioc.toFanged()
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Fang() = %v, want %v", got, test.want)
			}
		})
	}
}

// TestGetIOCsEdgeCases tests GetIOCs with edge cases
func TestGetIOCsEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		getFangedIOCs bool
		want          []*IOC
	}{
		{
			name:          "empty string",
			input:         "",
			getFangedIOCs: true,
			want:          []*IOC{},
		},
		{
			name:          "only whitespace",
			input:         "   \n\t\r  ",
			getFangedIOCs: true,
			want:          []*IOC{},
		},
		{
			name:          "very long string",
			input:         strings.Repeat("a", 10000) + " test.com " + strings.Repeat("b", 10000),
			getFangedIOCs: true,
			want:          []*IOC{{IOC: "test.com", Type: Domain}},
		},
		{
			name:          "duplicate IOCs",
			input:         "test.com test.com test.com",
			getFangedIOCs: true,
			want:          []*IOC{{IOC: "test.com", Type: Domain}},
		},
		{
			name:          "overlapping matches",
			input:         "test@test.com and test.com",
			getFangedIOCs: true,
			want: []*IOC{
				{IOC: "test.com", Type: Domain},
				{IOC: "test@test.com", Type: Email},
			},
		},
		{
			name:          "special characters",
			input:         "test@domain.com!@#$%^&*()",
			getFangedIOCs: true,
			want: []*IOC{
				{IOC: "domain.com", Type: Domain},
				{IOC: "test@domain.com", Type: Email},
			},
		},
		{
			name:          "unicode characters",
			input:         "test@пример.испытание",
			getFangedIOCs: true,
			want:          []*IOC{},
		},
		{
			name:          "binary data",
			input:         string([]byte{0x00, 0x01, 0x02, 0x03}),
			getFangedIOCs: true,
			want:          []*IOC{},
		},
		{
			name:          "getFangedIOCs false with fanged content",
			input:         "test[.]com",
			getFangedIOCs: false,
			want:          []*IOC{{IOC: "test[.]com", Type: Domain}},
		},
		{
			name:          "getFangedIOCs true with fanged content",
			input:         "test[.]com",
			getFangedIOCs: true,
			want:          []*IOC{{IOC: "test[.]com", Type: Domain}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ExtractIOCs(test.input, test.getFangedIOCs)
			if !testify.ElementsMatch(t, got, test.want) {
				t.Errorf("ExtractIOCs(%q, %v) = %v, want %v", test.input, test.getFangedIOCs, got, test.want)
			}
		})
	}
}

// TestGetIOCsReaderEdgeCases tests GetIOCsReader with edge cases
func TestGetIOCsReaderEdgeCases(t *testing.T) {
	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		iocs := make(chan *IOC)

		go func() {
			defer close(iocs)
			longString := strings.Repeat("test.com ", 1000)
			err := ExtractIOCsReader(ctx, strings.NewReader(longString), true, iocs)
			require.Error(t, err)
			require.Equal(t, context.Canceled, err)
		}()

		// Cancel context immediately
		cancel()

		// Should not block
		select {
		case <-iocs:
		case <-time.After(100 * time.Millisecond):
			t.Error("GetIOCsReader should have returned quickly after context cancellation")
		}
	})

	t.Run("empty reader", func(t *testing.T) {
		iocs := make(chan *IOC)
		go func() {
			defer close(iocs)
			err := ExtractIOCsReader(context.Background(), strings.NewReader(""), true, iocs)
			require.NoError(t, err)
		}()

		count := 0
		for range iocs {
			count++
		}
		if count != 0 {
			t.Errorf("Expected 0 IOCs from empty reader, got %d", count)
		}
	})

	t.Run("large reader", func(t *testing.T) {
		iocs := make(chan *IOC)
		largeInput := strings.Repeat("test.com hash12345678901234567890123456789012 ", 1000)

		go func() {
			defer close(iocs)
			err := ExtractIOCsReader(context.Background(), strings.NewReader(largeInput), true, iocs)
			require.NoError(t, err)
		}()

		count := 0
		for range iocs {
			count++
		}
		// Should find many IOCs
		if count == 0 {
			t.Error("Expected to find IOCs in large input")
		}
	})
}

// TestStandardizeDefangsEdgeCases tests StandardizeDefangs with edge cases
func TestStandardizeDefangsEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		iocs []*IOC
		want []*IOC
	}{
		{
			name: "empty slice",
			iocs: []*IOC{},
			want: []*IOC{},
		},
		{
			name: "nil slice",
			iocs: nil,
			want: nil,
		},
		{
			name: "mixed defanging styles",
			iocs: []*IOC{
				{IOC: "test(@)example[.]com", Type: Email},
				{IOC: "1(.)2(.)3(.)4", Type: IPv4},
			},
			want: []*IOC{
				{IOC: "test[AT]example[.]com", Type: Email},
				{IOC: "1[.]2[.]3[.]4", Type: IPv4},
			},
		},
		{
			name: "already standardized",
			iocs: []*IOC{
				{IOC: "test[AT]example[.]com", Type: Email},
			},
			want: []*IOC{
				{IOC: "test[AT]example[.]com", Type: Email},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			NormalizeDefangs(test.iocs)
			if !reflect.DeepEqual(test.iocs, test.want) {
				t.Errorf("NormalizeDefangs() = %v, want %v", test.iocs, test.want)
			}
		})
	}
}
