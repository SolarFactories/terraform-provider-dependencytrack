package provider

import (
	"cmp"
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
}

func requireNoError(t *testing.T, actual error) {
	if actual != nil {
		t.Errorf("Expected nil, received %v", actual)
	}
}

func requireEqual[T cmp.Ordered](t *testing.T, actual T, expected T) {
	if actual != expected {
		t.Errorf("Expected %v, received %v", expected, actual)
	}
}
