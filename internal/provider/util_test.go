package provider

import (
	"cmp"
	"regexp"
	"testing"
)

func TestParseSemver(t *testing.T) {
	{
		semver, err := ParseSemver("4.0.0")
		requireNoError(t, err)
		requireEqual(t, semver.Major, 4)
		requireEqual(t, semver.Minor, 0)
		requireEqual(t, semver.Patch, 0)
	}
	{
		semver, err := ParseSemver("1.2.3")
		requireNoError(t, err)
		requireEqual(t, semver.Major, 1)
		requireEqual(t, semver.Minor, 2)
		requireEqual(t, semver.Patch, 3)
	}
	{
		semver, err := ParseSemver("1.2")
		requireError(t, err, "^Found semver with 2 parts, expected 3.$")
		requireNil(t, semver)
	}
	{
		semver, err := ParseSemver("1.2.3.4.5.6")
		requireError(t, err, "^Found semver with 6 parts, expected 3.$")
		requireNil(t, semver)
	}
	{
		semver, err := ParseSemver("a.2.3")
		requireError(t, err, "^Unable to parse semver major component, from: strconv.Atoi: parsing \"a\": invalid syntax$")
		requireNil(t, semver)
	}
	{
		semver, err := ParseSemver("1.0x2.3")
		requireError(t, err, "^Unable to parse semver minor component, from: strconv.Atoi: parsing \"0x2\": invalid syntax$")
		requireNil(t, semver)
	}
	{
		semver, err := ParseSemver("1.2.0b1")
		requireError(t, err, "^Unable to parse semver patch component, from: strconv.Atoi: parsing \"0b1\": invalid syntax$")
		requireNil(t, semver)
	}
	{
		semver, err := ParseSemver("-1.2.3")
		requireError(t, err, "^Unable to validate semver major component, from: -1$")
		requireNil(t, semver)
	}
	{
		semver, err := ParseSemver("1.-2.3")
		requireError(t, err, "^Unable to validate semver minor component, from: -2$")
		requireNil(t, semver)
	}
	{
		semver, err := ParseSemver("1.2.-3")
		requireError(t, err, "^Unable to validate semver patch component, from: -3$")
		requireNil(t, semver)
	}
}

func requireNoError(t *testing.T, actual error) {
	t.Helper()
	if actual != nil {
		t.Errorf("Expected nil, received error: %v", actual)
	}
}

func requireError(t *testing.T, actual error, expectedRegex string) {
	t.Helper()
	if actual == nil {
		t.Errorf("Expected error, received nil")
		return
	}
	match, err := regexp.MatchString(expectedRegex, actual.Error())
	if err != nil {
		t.Errorf("Unable to use %s as a regex pattern.", expectedRegex)
		return
	}
	if !match {
		t.Errorf("Expected error matching %s. Received %s", expectedRegex, actual.Error())
	}
}

func requireNil[T any](t *testing.T, actual *T) {
	t.Helper()
	if actual != nil {
		t.Errorf("Expected nil, received value: %v", actual)
	}
}

func requireEqual[T cmp.Ordered](t *testing.T, actual T, expected T) {
	t.Helper()
	if actual != expected {
		t.Errorf("Expected %v, received %v", expected, actual)
	}
}
