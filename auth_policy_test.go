package gw_policies_playground

import (
	"testing"

	authorino "github.com/kuadrant/authorino/api/v1beta1"

	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/apimachinery/pkg/runtime"
)

var (
	testAuthPolicySpec1 = AuthPolicy{
		Patterns: map[string]authorino.JSONPatternExpressions{
			"api-route": []authorino.JSONPatternExpression{
				{
					Selector: "context.request.http.path",
					Operator: "matches",
					Value:    "^/api/.+",
				},
			},
			"api-version": []authorino.JSONPatternExpression{
				{
					Selector: `context.request.http.path.@extract{"sep":"/","pos":2}`,
					Operator: "matches",
					Value:    "^v[0-9]+",
				},
			},
		},
		Conditions: []authorino.JSONPattern{
			{
				JSONPatternRef: authorino.JSONPatternRef{
					JSONPatternName: "api-route",
				},
			},
			{
				JSONPatternRef: authorino.JSONPatternRef{
					JSONPatternName: "api-version",
				},
			},
		},
		Identity: []*authorino.Identity{
			{
				Name:      "friends",
				Anonymous: &authorino.Identity_Anonymous{},
			},
		},
		DenyWith: &authorino.DenyWith{
			Unauthenticated: &authorino.DenyWithSpec{
				Message: &authorino.StaticOrDynamicValue{
					Value: "Please login",
				},
			},
			Unauthorized: &authorino.DenyWithSpec{
				Message: &authorino.StaticOrDynamicValue{
					Value: "Access Denied",
				},
			},
		},
	}

	testAuthPolicySpec2 = AuthPolicy{
		Patterns: map[string]authorino.JSONPatternExpressions{
			"api-version": []authorino.JSONPatternExpression{
				{
					Selector: `context.request.http.path.@extract{"sep":"/","pos":2}`,
					Operator: "eq",
					Value:    "v1",
				},
			},
		},
		Conditions: []authorino.JSONPattern{
			{
				JSONPatternExpression: authorino.JSONPatternExpression{
					Selector: "context.request.http.method",
					Operator: "eq",
					Value:    "GET",
				},
			},
		},
		Identity: []*authorino.Identity{
			{
				Name: "friends",
				APIKey: &authorino.Identity_APIKey{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"authorino.kuadrant.io/identity": "apiKey",
							"my-app.io/group":                "friends",
						},
					},
				},
			},
		},
		Authorization: []*authorino.Authorization{
			{
				Name: "my-policy",
				OPA: &authorino.Authorization_OPA{
					InlineRego: `allow { input.auth.identity.metadata.annotations.my-app\.io/admin == "true" }`,
				},
			},
		},
		DenyWith: &authorino.DenyWith{
			Unauthorized: &authorino.DenyWithSpec{
				Code: 302,
				Headers: []authorino.JsonProperty{
					{
						Name: "Location",
						Value: k8s.RawExtension{
							Raw: []byte("https://my-app.io/login"),
						},
					},
				},
			},
		},
	}
)

func TestMerge_RouteDefaultOverride(t *testing.T) {
	gwc := NewGatewayClass[AuthPolicy]("gwc1")
	gw := gwc.CreateGateway("gw")
	route := gw.CreateRoute("route")

	policy := PolicySpec[AuthPolicy]{
		name:      "auth-policy",
		defaults:  testAuthPolicySpec1,
		overrides: testAuthPolicySpec2,
	}
	route.AddPolicy(policy)

	result := route.MergedPolicies(AuthPolicyMerger)

	assert.Check(t, result.Patterns["api-route"] != nil)
	assert.Check(t, result.Patterns["api-version"] != nil)
	assert.Equal(t, result.Patterns["api-version"][0].Value, "v1")
	assert.Equal(t, len(result.Conditions), 3)
	assert.Equal(t, len(result.Identity), 1)
	assert.Equal(t, result.Identity[0].Name, "friends")
	assert.Check(t, result.Identity[0].Anonymous == nil)
	assert.Check(t, result.Identity[0].APIKey != nil)
	assert.Equal(t, len(result.Authorization), 1)
	assert.Equal(t, result.Authorization[0].Name, "my-policy")
	assert.Check(t, result.DenyWith.Unauthenticated != nil)
	assert.Check(t, result.DenyWith.Unauthorized != nil)
	assert.Equal(t, int(result.DenyWith.Unauthorized.Code), 302)
}

func TestMerge_RouteDefault_RouteDefault(t *testing.T) {
	gwc := NewGatewayClass[AuthPolicy]("gwc1")
	gw := gwc.CreateGateway("gw")
	route := gw.CreateRoute("route")

	p1 := PolicySpec[AuthPolicy]{
		name:     "auth-policy",
		defaults: testAuthPolicySpec1,
	}
	route.AddPolicy(p1)

	p2 := PolicySpec[AuthPolicy]{
		name:     "auth-policy",
		defaults: testAuthPolicySpec2,
	}
	route.AddPolicy(p2)

	result := route.MergedPolicies(AuthPolicyMerger)

	assert.Check(t, result.Patterns["api-route"] != nil)
	assert.Check(t, result.Patterns["api-version"] != nil)
	assert.Equal(t, result.Patterns["api-version"][0].Value, "^v[0-9]+")
	assert.Equal(t, len(result.Conditions), 3)
	assert.Equal(t, len(result.Identity), 1)
	assert.Equal(t, result.Identity[0].Name, "friends")
	assert.Check(t, result.Identity[0].Anonymous != nil)
	assert.Check(t, result.Identity[0].APIKey == nil)
	assert.Equal(t, len(result.Authorization), 1)
	assert.Equal(t, result.Authorization[0].Name, "my-policy")
	assert.Check(t, result.DenyWith.Unauthenticated != nil)
	assert.Check(t, result.DenyWith.Unauthorized != nil)
	assert.Equal(t, int(result.DenyWith.Unauthorized.Code), 0)
}

func TestMerge_RouteDefault_RouteOverride(t *testing.T) {
	gwc := NewGatewayClass[AuthPolicy]("gwc1")
	gw := gwc.CreateGateway("gw")
	route := gw.CreateRoute("route")

	p1 := PolicySpec[AuthPolicy]{
		name:     "auth-policy",
		defaults: testAuthPolicySpec1,
	}
	route.AddPolicy(p1)

	p2 := PolicySpec[AuthPolicy]{
		name:      "auth-policy",
		overrides: testAuthPolicySpec2,
	}
	route.AddPolicy(p2)

	result := route.MergedPolicies(AuthPolicyMerger)

	assert.Check(t, result.Patterns["api-route"] != nil)
	assert.Check(t, result.Patterns["api-version"] != nil)
	assert.Equal(t, result.Patterns["api-version"][0].Value, "v1")
	assert.Equal(t, len(result.Conditions), 3)
	assert.Equal(t, len(result.Identity), 1)
	assert.Equal(t, result.Identity[0].Name, "friends")
	assert.Check(t, result.Identity[0].Anonymous == nil)
	assert.Check(t, result.Identity[0].APIKey != nil)
	assert.Equal(t, len(result.Authorization), 1)
	assert.Equal(t, result.Authorization[0].Name, "my-policy")
	assert.Check(t, result.DenyWith.Unauthenticated != nil)
	assert.Check(t, result.DenyWith.Unauthorized != nil)
	assert.Equal(t, int(result.DenyWith.Unauthorized.Code), 302)
}

func TestMerge_RouteOverride_RouteDefault(t *testing.T) {
	gwc := NewGatewayClass[AuthPolicy]("gwc1")
	gw := gwc.CreateGateway("gw")
	route := gw.CreateRoute("route")

	p1 := PolicySpec[AuthPolicy]{
		name:      "auth-policy",
		overrides: testAuthPolicySpec1,
	}
	route.AddPolicy(p1)

	p2 := PolicySpec[AuthPolicy]{
		name:     "auth-policy",
		defaults: testAuthPolicySpec2,
	}
	route.AddPolicy(p2)

	result := route.MergedPolicies(AuthPolicyMerger)

	assert.Check(t, result.Patterns["api-route"] != nil)
	assert.Check(t, result.Patterns["api-version"] != nil)
	assert.Equal(t, result.Patterns["api-version"][0].Value, "^v[0-9]+")
	assert.Equal(t, len(result.Conditions), 3)
	assert.Equal(t, len(result.Identity), 1)
	assert.Equal(t, result.Identity[0].Name, "friends")
	assert.Check(t, result.Identity[0].Anonymous != nil)
	assert.Check(t, result.Identity[0].APIKey == nil)
	assert.Equal(t, len(result.Authorization), 1)
	assert.Equal(t, result.Authorization[0].Name, "my-policy")
	assert.Check(t, result.DenyWith.Unauthenticated != nil)
	assert.Check(t, result.DenyWith.Unauthorized != nil)
	assert.Equal(t, int(result.DenyWith.Unauthorized.Code), 0)
}

func TestMerge_RouteOverride_RouteOverride(t *testing.T) {
	gwc := NewGatewayClass[AuthPolicy]("gwc1")
	gw := gwc.CreateGateway("gw")
	route := gw.CreateRoute("route")

	p1 := PolicySpec[AuthPolicy]{
		name:      "auth-policy",
		overrides: testAuthPolicySpec1,
	}
	route.AddPolicy(p1)

	p2 := PolicySpec[AuthPolicy]{
		name:      "auth-policy",
		overrides: testAuthPolicySpec2,
	}
	route.AddPolicy(p2)

	result := route.MergedPolicies(AuthPolicyMerger)

	assert.Check(t, result.Patterns["api-route"] != nil)
	assert.Check(t, result.Patterns["api-version"] != nil)
	assert.Equal(t, result.Patterns["api-version"][0].Value, "v1")
	assert.Equal(t, len(result.Conditions), 3)
	assert.Equal(t, len(result.Identity), 1)
	assert.Equal(t, result.Identity[0].Name, "friends")
	assert.Check(t, result.Identity[0].Anonymous == nil)
	assert.Check(t, result.Identity[0].APIKey != nil)
	assert.Equal(t, len(result.Authorization), 1)
	assert.Equal(t, result.Authorization[0].Name, "my-policy")
	assert.Check(t, result.DenyWith.Unauthenticated != nil)
	assert.Check(t, result.DenyWith.Unauthorized != nil)
	assert.Equal(t, int(result.DenyWith.Unauthorized.Code), 302)
}

func TestMerge_GatewayDefault_RouteDefault(t *testing.T) {
	gwc := NewGatewayClass[AuthPolicy]("gwc1")
	gw := gwc.CreateGateway("gw")
	route := gw.CreateRoute("route")

	p1 := PolicySpec[AuthPolicy]{
		name:     "auth-policy",
		defaults: testAuthPolicySpec1,
	}
	gw.AddPolicy(p1)

	p2 := PolicySpec[AuthPolicy]{
		name:     "auth-policy",
		defaults: testAuthPolicySpec2,
	}
	route.AddPolicy(p2)

	result := route.MergedPolicies(AuthPolicyMerger)

	assert.Check(t, result.Patterns["api-route"] != nil)
	assert.Check(t, result.Patterns["api-version"] != nil)
	assert.Equal(t, result.Patterns["api-version"][0].Value, "v1")
	assert.Equal(t, len(result.Conditions), 3)
	assert.Equal(t, len(result.Identity), 1)
	assert.Equal(t, result.Identity[0].Name, "friends")
	assert.Check(t, result.Identity[0].Anonymous == nil)
	assert.Check(t, result.Identity[0].APIKey != nil)
	assert.Equal(t, len(result.Authorization), 1)
	assert.Equal(t, result.Authorization[0].Name, "my-policy")
	assert.Check(t, result.DenyWith.Unauthenticated != nil)
	assert.Check(t, result.DenyWith.Unauthorized != nil)
	assert.Equal(t, int(result.DenyWith.Unauthorized.Code), 302)
}

func TestMerge_GatewayDefault_RouteOverride(t *testing.T) {
	gwc := NewGatewayClass[AuthPolicy]("gwc1")
	gw := gwc.CreateGateway("gw")
	route := gw.CreateRoute("route")

	p1 := PolicySpec[AuthPolicy]{
		name:     "auth-policy",
		defaults: testAuthPolicySpec1,
	}
	gw.AddPolicy(p1)

	p2 := PolicySpec[AuthPolicy]{
		name:      "auth-policy",
		overrides: testAuthPolicySpec2,
	}
	route.AddPolicy(p2)

	result := route.MergedPolicies(AuthPolicyMerger)

	assert.Check(t, result.Patterns["api-route"] != nil)
	assert.Check(t, result.Patterns["api-version"] != nil)
	assert.Equal(t, result.Patterns["api-version"][0].Value, "v1")
	assert.Equal(t, len(result.Conditions), 3)
	assert.Equal(t, len(result.Identity), 1)
	assert.Equal(t, result.Identity[0].Name, "friends")
	assert.Check(t, result.Identity[0].Anonymous == nil)
	assert.Check(t, result.Identity[0].APIKey != nil)
	assert.Equal(t, len(result.Authorization), 1)
	assert.Equal(t, result.Authorization[0].Name, "my-policy")
	assert.Check(t, result.DenyWith.Unauthenticated != nil)
	assert.Check(t, result.DenyWith.Unauthorized != nil)
	assert.Equal(t, int(result.DenyWith.Unauthorized.Code), 302)
}

func TestMerge_GatewayOverride_RouteDefault(t *testing.T) {
	gwc := NewGatewayClass[AuthPolicy]("gwc1")
	gw := gwc.CreateGateway("gw")
	route := gw.CreateRoute("route")

	p1 := PolicySpec[AuthPolicy]{
		name:      "auth-policy",
		overrides: testAuthPolicySpec1,
	}
	gw.AddPolicy(p1)

	p2 := PolicySpec[AuthPolicy]{
		name:     "auth-policy",
		defaults: testAuthPolicySpec2,
	}
	route.AddPolicy(p2)

	result := route.MergedPolicies(AuthPolicyMerger)

	assert.Check(t, result.Patterns["api-route"] != nil)
	assert.Check(t, result.Patterns["api-version"] != nil)
	assert.Equal(t, result.Patterns["api-version"][0].Value, "^v[0-9]+")
	assert.Equal(t, len(result.Conditions), 3)
	assert.Equal(t, len(result.Identity), 1)
	assert.Equal(t, result.Identity[0].Name, "friends")
	assert.Check(t, result.Identity[0].Anonymous != nil)
	assert.Check(t, result.Identity[0].APIKey == nil)
	assert.Equal(t, len(result.Authorization), 1)
	assert.Equal(t, result.Authorization[0].Name, "my-policy")
	assert.Check(t, result.DenyWith.Unauthenticated != nil)
	assert.Check(t, result.DenyWith.Unauthorized != nil)
	assert.Equal(t, int(result.DenyWith.Unauthorized.Code), 0)
}

func TestMerge_GatewayOverride_RouteOverride(t *testing.T) {
	gwc := NewGatewayClass[AuthPolicy]("gwc1")
	gw := gwc.CreateGateway("gw")
	route := gw.CreateRoute("route")

	p1 := PolicySpec[AuthPolicy]{
		name:      "auth-policy",
		overrides: testAuthPolicySpec1,
	}
	gw.AddPolicy(p1)

	p2 := PolicySpec[AuthPolicy]{
		name:      "auth-policy",
		overrides: testAuthPolicySpec2,
	}
	route.AddPolicy(p2)

	result := route.MergedPolicies(AuthPolicyMerger)

	assert.Check(t, result.Patterns["api-route"] != nil)
	assert.Check(t, result.Patterns["api-version"] != nil)
	assert.Equal(t, result.Patterns["api-version"][0].Value, "^v[0-9]+")
	assert.Equal(t, len(result.Conditions), 3)
	assert.Equal(t, len(result.Identity), 1)
	assert.Equal(t, result.Identity[0].Name, "friends")
	assert.Check(t, result.Identity[0].Anonymous != nil)
	assert.Check(t, result.Identity[0].APIKey == nil)
	assert.Equal(t, len(result.Authorization), 1)
	assert.Equal(t, result.Authorization[0].Name, "my-policy")
	assert.Check(t, result.DenyWith.Unauthenticated != nil)
	assert.Check(t, result.DenyWith.Unauthorized != nil)
	assert.Equal(t, int(result.DenyWith.Unauthorized.Code), 0)
}
