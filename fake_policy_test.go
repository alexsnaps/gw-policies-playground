package gw_policies_playground

import (
	"testing"
)

type FakePolicy struct {
	enabled *bool
	value   *int
}

func FakePolicyMerger(policies []FakePolicy) FakePolicy {
	result := policies[0]
	for _, policy := range policies[1:] {
		if result.enabled == nil {
			result.enabled = policy.enabled
		}
		if result.value == nil {
			result.value = policy.value
		}
	}
	return result
}

func TestRouteSimpleMerge(t *testing.T) {
	gwc := NewGatewayClass[FakePolicy]("gwc1")
	gw := gwc.CreateGateway("gw")
	route := gw.CreateRoute("route")

	if gw.parent != &gwc || gw.name != "gw" {
		t.Failed()
	}
	if route.parent != gw || route.name != "route" {
		t.Failed()
	}

	value := 42
	enabled := true

	policy := PolicySpec[FakePolicy]{
		name:      "policy",
		defaults:  FakePolicy{value: &value},
		overrides: FakePolicy{enabled: &enabled},
	}
	route.AddPolicy(policy)

	if route.policies[0] != policy {
		t.Failed()
	}

	result := FakePolicyMerger(route.PoliciesToMerge())
	if *result.value != 42 {
		t.Failed()
	}
	if *result.enabled != true {
		t.Failed()
	}
}

func TestRouteGwMerge(t *testing.T) {
	gwc := NewGatewayClass[FakePolicy]("gwc1")
	gw := gwc.CreateGateway("gw")
	route := gw.CreateRoute("route")

	if gw.parent != &gwc || gw.name != "gw" {
		t.Failed()
	}
	if route.parent != gw || route.name != "route" {
		t.Failed()
	}

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

	result := FakePolicyMerger(route.PoliciesToMerge())
	if *result.value != 420 {
		t.Failed()
	}
	if *result.enabled != true {
		t.Failed()
	}
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

	result := FakePolicyMerger(route.PoliciesToMerge())
	if *result.value != 420 {
		t.Failed()
	}
	if *result.enabled != true {
		t.Failed()
	}
}
