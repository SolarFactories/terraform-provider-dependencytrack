package provider

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	SemverComponentCount        = 3
	PropertyTypeEncryptedString = "ENCRYPTEDSTRING"
	// LifecycleAction.
	LifecycleCreate LifecycleAction = "Create"
	LifecycleRead   LifecycleAction = "Read"
	LifecycleUpdate LifecycleAction = "Update"
	LifecycleDelete LifecycleAction = "Delete"
	LifecycleImport LifecycleAction = "Import"
)

type (
	Semver struct {
		Major int
		Minor int
		Patch int
	}

	OIDCMappingInfo struct {
		Team  uuid.UUID
		Group uuid.UUID
	}

	LifecycleAction string
)

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
	if err != nil {
		return nil, errors.New("Error in FilterPaged: " + err.Error())
	}
	return filtered, nil
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

func FindPagedPolicyCondition(
	conditionUUID uuid.UUID,
	policyFetchFunc func(dtrack.PageOptions) (dtrack.Page[dtrack.Policy], error),
) (*dtrack.PolicyCondition, error) {
	filter := func(policy dtrack.Policy) bool {
		for _, condition := range policy.PolicyConditions {
			if condition.UUID == conditionUUID {
				return true
			}
		}
		return false
	}
	policy, err := FindPaged(policyFetchFunc, filter)
	if err != nil {
		return nil, err
	}
	condition, err := Find(policy.PolicyConditions, func(condition dtrack.PolicyCondition) bool {
		return condition.UUID == conditionUUID
	})
	if err != nil {
		return nil, err
	}
	condition.Policy = policy
	return condition, nil
}

func ParseSemver(s string) (*Semver, error) {
	parts := strings.Split(s, ".")
	if len(parts) != SemverComponentCount {
		return nil, fmt.Errorf("found semver with %v parts, expected 3", len(parts))
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, errors.New("unable to parse semver major component, from: " + err.Error())
	}
	if major < 0 {
		return nil, fmt.Errorf("unable to validate semver major component, from: %d", major)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, errors.New("unable to parse semver minor component, from: " + err.Error())
	}
	if minor < 0 {
		return nil, fmt.Errorf("unable to validate semver minor component, from: %d", minor)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, errors.New("unable to parse semver patch component, from: " + err.Error())
	}
	if patch < 0 {
		return nil, fmt.Errorf("unable to validate semver patch component, from: %d", patch)
	}
	semver := Semver{
		Major: major,
		Minor: minor,
		Patch: patch,
	}
	return &semver, nil
}

func ListDeltas[T cmp.Ordered](current []T, desired []T) (add []T, remove []T) {
	add = []T{}
	remove = []T{}
	for _, v := range current {
		if !slices.Contains(desired, v) {
			remove = append(remove, v)
		}
	}
	for _, v := range desired {
		if !slices.Contains(current, v) {
			add = append(add, v)
		}
	}
	return add, remove
}

func ListDeltasUUID(current []uuid.UUID, desired []uuid.UUID) (add []uuid.UUID, remove []uuid.UUID) {
	currentStr := Map(current, func(cur uuid.UUID) string { return cur.String() })
	desiredStr := Map(desired, func(des uuid.UUID) string { return des.String() })
	addStr, removeStr := ListDeltas(currentStr, desiredStr)
	add = Map(addStr, func(s string) uuid.UUID { return uuid.MustParse(s) })
	remove = Map(removeStr, func(s string) uuid.UUID { return uuid.MustParse(s) })
	return add, remove
}

func Map[T, U any](items []T, actor func(T) U) []U {
	result := make([]U, 0, len(items))
	for _, t := range items {
		u := actor(t)
		result = append(result, u)
	}
	return result
}

func TryMap[T, U any](items []T, actor func(T) (U, error)) ([]U, error) {
	result := make([]U, 0, len(items))
	for _, t := range items {
		u, err := actor(t)
		if err != nil {
			return result, err
		}
		result = append(result, u)
	}
	return result, nil
}

func TryParseUUID(value types.String, action LifecycleAction, tfPath path.Path) (uuid.UUID, diag.Diagnostic) {
	if value.IsUnknown() {
		errDiag := diag.NewAttributeErrorDiagnostic(
			tfPath,
			fmt.Sprintf("Within %s, unable to parse %s into UUID.", action, tfPath.String()),
			fmt.Sprintf("Value for %s is unknown.", tfPath.String()),
		)
		return uuid.Nil, errDiag
	}
	if value.IsNull() {
		errDiag := diag.NewAttributeErrorDiagnostic(
			tfPath,
			fmt.Sprintf("Within %s, unable to parse %s into UUID.", action, tfPath.String()),
			fmt.Sprintf("Value for %s is null.", tfPath.String()),
		)
		return uuid.Nil, errDiag
	}
	ret, err := uuid.Parse(value.ValueString())
	if err != nil {
		errDiag := diag.NewAttributeErrorDiagnostic(
			tfPath,
			fmt.Sprintf("Within %s, unable to parse %s into UUID.", action, tfPath.String()),
			"Error from: "+err.Error(),
		)
		return uuid.Nil, errDiag
	}
	return ret, nil
}

func GetStringList(ctx context.Context, diags *diag.Diagnostics, list types.List) ([]string, error) {
	tagStrings := make([]types.String, 0, len(list.Elements()))
	diags.Append(list.ElementsAs(ctx, &tagStrings, false)...)
	if diags.HasError() {
		return nil, errors.New("type mismatch. Expected []types.String")
	}
	stringList, err := TryMap(tagStrings, func(item types.String) (string, error) {
		if item.IsUnknown() {
			return "", errors.New("received unknown value for tag")
		}
		if item.IsNull() {
			return "", errors.New("received null tag")
		}
		return item.ValueString(), nil
	})
	return stringList, err
}
