package gw_policies_playground

import (
	"fmt"

	authorino "github.com/kuadrant/authorino/api/v1beta1"
)

type AuthPolicy struct {
	// Named sets of JSON patterns that can be referred in `when` conditionals and in JSON-pattern matching policy rules.
	Patterns map[string]authorino.JSONPatternExpressions `json:"patterns,omitempty"`

	// Conditions for the AuthConfig to be enforced.
	// If omitted, the AuthConfig will be enforced for all requests.
	// If present, all conditions must match for the AuthConfig to be enforced; otherwise, Authorino skips the AuthConfig and returns immediately with status OK.
	Conditions []authorino.JSONPattern `json:"when,omitempty"`

	// List of identity sources/authentication modes.
	// At least one config of this list MUST evaluate to a valid identity for a request to be successful in the identity verification phase.
	Identity []*authorino.Identity `json:"identity,omitempty"`

	// List of metadata source configs.
	// Authorino fetches JSON content from sources on this list on every request.
	Metadata []*authorino.Metadata `json:"metadata,omitempty"`

	// Authorization is the list of authorization policies.
	// All policies in this list MUST evaluate to "true" for a request be successful in the authorization phase.
	Authorization []*authorino.Authorization `json:"authorization,omitempty"`

	// List of response configs.
	// Authorino gathers data from the auth pipeline to build custom responses for the client.
	Response []*authorino.Response `json:"response,omitempty"`

	// Custom denial response codes, statuses and headers to override default 40x's.
	DenyWith *authorino.DenyWith `json:"denyWith,omitempty"`
}

func AuthPolicyMerger(p1, p2 AuthPolicy) AuthPolicy {
	result := AuthPolicy{
		Patterns: make(map[string]authorino.JSONPatternExpressions),
	}

	names := make(map[string]interface{})

	// Patterns
	for name, pattern := range p1.Patterns {
		result.Patterns[name] = pattern
		names[fmt.Sprintf("patterns.%s", name)] = nil
	}
	for name, pattern := range p2.Patterns {
		if _, exists := names[fmt.Sprintf("patterns.%s", name)]; exists {
			continue
		}
		result.Patterns[name] = pattern
		names[fmt.Sprintf("patterns.%s", name)] = nil
	}

	// Conditions
	if len(p1.Conditions) > 0 {
		result.Conditions = append(result.Conditions, p1.Conditions...)
	} else if len(p2.Conditions) > 0 {
		result.Conditions = append(result.Conditions, p2.Conditions...)
	}

	// Identity
	result.Identity = append(result.Identity, p1.Identity...)
	for _, identity := range result.Identity {
		names[fmt.Sprintf("identity.%s", identity.Name)] = nil
	}
	for _, identity := range p2.Identity {
		if _, exists := names[fmt.Sprintf("identity.%s", identity.Name)]; exists {
			continue
		}
		result.Identity = append(result.Identity, identity)
		names[fmt.Sprintf("identity.%s", identity.Name)] = nil
	}

	// Metadata
	result.Metadata = append(result.Metadata, p1.Metadata...)
	for _, metadata := range result.Metadata {
		names[fmt.Sprintf("metadata.%s", metadata.Name)] = nil
	}
	for _, metadata := range p2.Metadata {
		if _, exists := names[fmt.Sprintf("metadata.%s", metadata.Name)]; exists {
			continue
		}
		result.Metadata = append(result.Metadata, metadata)
		names[fmt.Sprintf("metadata.%s", metadata.Name)] = nil
	}

	// Authorization
	result.Authorization = append(result.Authorization, p1.Authorization...)
	for _, authorization := range result.Authorization {
		names[fmt.Sprintf("authorization.%s", authorization.Name)] = nil
	}
	for _, authorization := range p2.Authorization {
		if _, exists := names[fmt.Sprintf("authorization.%s", authorization.Name)]; exists {
			continue
		}
		result.Authorization = append(result.Authorization, authorization)
		names[fmt.Sprintf("authorization.%s", authorization.Name)] = nil
	}

	// Response
	result.Response = append(result.Response, p1.Response...)
	for _, response := range result.Response {
		names[fmt.Sprintf("response.%s", response.Name)] = nil
	}
	for _, response := range p2.Response {
		if _, exists := names[fmt.Sprintf("response.%s", response.Name)]; exists {
			continue
		}
		result.Response = append(result.Response, response)
		names[fmt.Sprintf("response.%s", response.Name)] = nil
	}

	// DenyWith
	result.DenyWith = p1.DenyWith

	if denyWith := p2.DenyWith; denyWith != nil {
		if result.DenyWith == nil {
			result.DenyWith = &authorino.DenyWith{}
		}
		if result.DenyWith.Unauthenticated == nil {
			result.DenyWith.Unauthenticated = denyWith.Unauthenticated
		}
		if result.DenyWith.Unauthorized == nil {
			result.DenyWith.Unauthorized = denyWith.Unauthorized
		}
	}

	return result
}
