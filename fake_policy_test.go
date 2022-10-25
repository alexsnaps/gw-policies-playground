package gw_policies_playground

import (
	"testing"

	"gotest.tools/assert"
)

type FakePolicy struct {
	enabled *bool
	value   *int
}

func FakePolicyMerger(overrides FakePolicy, defaults FakePolicy) FakePolicy {
	result := overrides
	if result.enabled == nil {
		result.enabled = defaults.enabled
	}
	if result.value == nil {
		result.value = defaults.value
	}
	return result
}

func TestRouteSimpleMerge(t *testing.T) {
	gwc := NewGatewayClass[FakePolicy]("gwc1")
	gw := gwc.CreateGateway("gw")
	route := gw.CreateRoute("route")

  assert.Equal(t, gw.parent, &gwc)
  assert.Equal(t, gw.name, "gw")
  assert.Equal(t, route.parent, gw)
  assert.Equal(t, route.name, "route")

	value := 42
	enabled := true

	policy := PolicySpec[FakePolicy]{
		name:      "policy",
		defaults:  FakePolicy{value: &value},
		overrides: FakePolicy{enabled: &enabled},
	}
	route.AddPolicy(policy)

  assert.Equal(t, route.policies[0], policy)

	result := route.MergedPolicies(FakePolicyMerger)

	assert.Equal(t, *result.value, 42)
	assert.Check(t, *result.enabled)
}

func TestRouteGwMerge(t *testing.T) {
	gwc := NewGatewayClass[FakePolicy]("gwc1")
	gw := gwc.CreateGateway("gw")
	route := gw.CreateRoute("route")

  assert.Equal(t, gw.parent, &gwc)
  assert.Equal(t, gw.name, "gw")
  assert.Equal(t, route.parent, gw)
  assert.Equal(t, route.name, "route")

	gwDefault := 42
	routeOverride := 420
	enabled := true

	gwPolicy := PolicySpec[FakePolicy]{
		name:      "gw_policy",
		defaults:  FakePolicy{value: &gwDefault},
		overrides: FakePolicy{enabled: &enabled},
	}
	gw.AddPolicy(gwPolicy)

	routePolicy := PolicySpec[FakePolicy]{
		name:      "gw_policy",
		defaults:  FakePolicy{value: &gwDefault},
		overrides: FakePolicy{value: &routeOverride},
	}
	route.AddPolicy(routePolicy)

	result := route.MergedPolicies(FakePolicyMerger)
	assert.Equal(t, *result.value, 420)
	assert.Check(t, *result.enabled)
}

func TestRouteGwMergeDefaults(t *testing.T) {
	gwc := NewGatewayClass[FakePolicy]("gwc1")
	gw := gwc.CreateGateway("gw")
	route := gw.CreateRoute("route")

	gwDefault := 42
	routeDefault := 420
	enabled := true
	disabled := false

	gwPolicy := PolicySpec[FakePolicy]{
		name:      "gw_policy",
		defaults:  FakePolicy{value: &gwDefault},
		overrides: FakePolicy{enabled: &enabled},
	}
	gw.AddPolicy(gwPolicy)

	routePolicy := PolicySpec[FakePolicy]{
		name:      "route_policy",
		defaults:  FakePolicy{enabled: &disabled, value: &routeDefault},
		overrides: FakePolicy{enabled: &disabled, value: nil},
	}
	route.AddPolicy(routePolicy)

	result := route.MergedPolicies(FakePolicyMerger)
	assert.Equal(t, *result.value, 420)
	assert.Check(t, *result.enabled)
}
