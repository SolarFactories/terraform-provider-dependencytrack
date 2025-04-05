package provider

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/google/uuid"
)

type Semver struct {
	Major int
	Minor int
	Patch int
}

const PropertyTypeEncryptedString = "ENCRYPTEDSTRING"

func Filter[T any](items []T, filter func(T) bool) []T {
	filtered := []T{}
	for _, item := range items {
		if !filter(item) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func Find[T any](items []T, filter func(T) bool) (*T, error) {
	filtered := Filter(items, filter)
	if len(filtered) == 0 {
		return nil, errors.New("did not find item")
	} else if len(filtered) > 1 {
		return nil, errors.New("found multiple items")
	}
	item := filtered[0]
	return &item, nil
}

func FilterPaged[T any](
	pageFetchFunc func(dtrack.PageOptions) (dtrack.Page[T], error),
	filter func(T) bool,
) ([]T, error) {
	filtered := []T{}
	err := dtrack.ForEach(pageFetchFunc, func(item T) error {
		if filter(item) {
			filtered = append(filtered, item)
		}
		return nil
	})
	return filtered, err
}

func FindPaged[T any](
	pageFetchFunc func(dtrack.PageOptions) (dtrack.Page[T], error),
	filter func(T) bool,
) (*T, error) {
	filtered, err := FilterPaged(pageFetchFunc, filter)
	if err != nil {
		return nil, err
	}
	if len(filtered) == 0 {
		return nil, errors.New("did not find item")
	} else if len(filtered) > 1 {
		return nil, errors.New("found multiple items")
	}
	item := filtered[0]
	return &item, nil
}

type OIDCMappingInfo struct {
	Team  uuid.UUID
	Group uuid.UUID
}

func FindPagedOidcMapping(
	mappingUUID uuid.UUID,
	teamsFetchFunc func(dtrack.PageOptions) (dtrack.Page[dtrack.Team], error),
) (*OIDCMappingInfo, error) {
	filter := func(team dtrack.Team) bool {
		for _, mappedGroup := range team.MappedOIDCGroups {
			if mappedGroup.UUID == mappingUUID {
				return true
			}
		}
		return false
	}
	team, err := FindPaged(teamsFetchFunc, filter)
	if err != nil {
		return nil, err
	}
	group, err := Find(team.MappedOIDCGroups, func(mapping dtrack.OIDCMapping) bool {
		return mapping.UUID == mappingUUID
	})
	if err != nil {
		return nil, err
	}
	info := OIDCMappingInfo{
		Team:  team.UUID,
		Group: group.Group.UUID,
	}
	return &info, nil
}

func ParseSemver(s string) (*Semver, error) {
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("Found semver with %v parts, expected 3.", len(parts))
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, err
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, err
	}
	semver := Semver{
		Major: major,
		Minor: minor,
		Patch: patch,
	}
	return &semver, nil
}
