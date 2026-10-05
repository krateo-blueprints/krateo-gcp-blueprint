package main

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"

	"github.com/gobuffalo/flect"
)

// kccGroupSuffix is the API group suffix every Config Connector CRD carries.
const kccGroupSuffix = ".cnrm.cloud.google.com"

// toGolangName mirrors core-provider's internal strutil.ToGolangName so the Kind we compute
// matches the one core-provider derives from the chart name.
func toGolangName(s string) string {
	buf := bytes.NewBuffer([]byte{})
	for i, v := range splitOnAll(s, isNotAGoNameCharacter) {
		if i == 0 && strings.IndexAny(v, "0123456789") == 0 {
			buf.WriteRune('_')
		}
		buf.WriteString(capitaliseFirstLetter(v))
	}
	return buf.String()
}

func capitaliseFirstLetter(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[0:1]) + s[1:]
}

func splitOnAll(s string, shouldSplit func(r rune) bool) []string {
	rv := []string{}
	buf := bytes.NewBuffer([]byte{})
	for _, c := range s {
		if shouldSplit(c) {
			rv = append(rv, buf.String())
			buf.Reset()
		} else {
			buf.WriteRune(c)
		}
	}
	if buf.Len() > 0 {
		rv = append(rv, buf.String())
	}
	return rv
}

func isNotAGoNameCharacter(r rune) bool {
	return !(unicode.IsLetter(r) || unicode.IsDigit(r))
}

// compositionKind reproduces core-provider's gvr.go: flect.Pascalize(toGolangName(chartName)).
func compositionKind(chartName string) string {
	return flect.Pascalize(toGolangName(chartName))
}

// compositionPlural reproduces the plural that crdgen stamps into the generated Composition
// CRD's spec.names.plural, which is what core-provider later reads back from apiserver discovery
// (it does NOT pluralize offline itself).
//
// ORDER MATTERS: crdgen does Pluralize-then-Lower (plumbing transpile.go:660), not
// Lower-then-Pluralize. The two are not equivalent -- "AzureSubscriptionAlias" gives
// "azuresubscriptionaliases" one way and "azuresubscriptionalias" the other, because flect's
// pluralization is case-sensitive. Getting this backwards puts a plural that does not exist into
// the customform's resource path. Mirror crdgen exactly rather than relying on the two orders
// happening to coincide.
func compositionPlural(kind string) string {
	return strings.ToLower(flect.Pluralize(kind))
}

// serviceFromGroup turns a Config Connector API group like "storage.cnrm.cloud.google.com"
// into "storage". It rejects groups that are not Config Connector groups, so pointing the
// generator at a non-KCC CRD fails loudly instead of producing a plausible-looking blueprint.
func serviceFromGroup(group string) (string, error) {
	if !strings.HasSuffix(group, kccGroupSuffix) {
		return "", fmt.Errorf("group %q is not a Config Connector group (expected *%s)", group, kccGroupSuffix)
	}
	service := strings.TrimSuffix(group, kccGroupSuffix)
	if service == "" || strings.Contains(service, ".") {
		return "", fmt.Errorf("cannot derive a service id from group %q", group)
	}
	return service, nil
}

// resourceSlug is the per-resource half of the chart name.
//
// Config Connector Kinds repeat their service ("StorageBucket" lives in the "storage" group),
// which a literal port of the ACK naming would turn into "gcp-storage-storagebucket". We strip
// the redundant service prefix so the name reads like the AWS catalog's (gcp-storage-bucket).
// The prefix is only stripped when a non-empty remainder survives, so Kinds that ARE the
// service (e.g. "Folder" in "resourcemanager") keep their full name.
func resourceSlug(service, kind string) string {
	lower := strings.ToLower(kind)
	if trimmed := strings.TrimPrefix(lower, strings.ToLower(service)); trimmed != "" && trimmed != lower {
		return trimmed
	}
	return lower
}

// chartName builds the blueprint chart name, e.g. gcp-storage-bucket.
func chartName(service, kind string) string {
	return "gcp-" + service + "-" + resourceSlug(service, kind)
}

// compositionVersion maps a chart semver (0.1.0) to the Composition CRD version (v0-1-0).
func compositionVersion(chartVersion string) string {
	return "v" + strings.ReplaceAll(chartVersion, ".", "-")
}

// titleCaseService renders a service id for human-facing text (storage -> Storage).
func titleCaseService(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
