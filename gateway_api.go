package gw_policies_playground

type void struct{}

var sentinel void

type GatewayClass[T Policy] struct {
	name     string
	gateways map[*Gateway[T]]void
	policies []PolicySpec[T]
}

func (gwc *GatewayClass[T]) CreateGateway(name string) *Gateway[T] {
	gw := &Gateway[T]{
		parent: gwc,
		name:   name,
		routes: make(map[*HttpRoute[T]]void),
	}
	gwc.gateways[gw] = sentinel
	return gw
}

type Gateway[T Policy] struct {
	parent   *GatewayClass[T]
	name     string
	routes   map[*HttpRoute[T]]void
	policies []PolicySpec[T]
}

func (gw *Gateway[T]) CreateRoute(name string) *HttpRoute[T] {
	r := &HttpRoute[T]{
		parent: gw,
		name:   name,
	}
	gw.routes[r] = sentinel
	return r
}

func (gw *Gateway[T]) AddPolicy(policy PolicySpec[T]) {
	gw.policies = append(gw.policies, policy)
}

type HttpRoute[T Policy] struct {
	parent   *Gateway[T]
	name     string
	policies []PolicySpec[T]
}

func (r *HttpRoute[T]) AddPolicy(policy PolicySpec[T]) {
	r.policies = append(r.policies, policy)
}

func (r *HttpRoute[T]) MergedPolicies(merger func(T, T) T) T {
	var policies []T
	for _, policy := range r.policies {
		policies = append(policies, policy.defaults)
		policies = append([]T{policy.overrides}, policies...)
	}
	for _, policy := range r.parent.policies {
		policies = append(policies, policy.defaults)
		policies = append([]T{policy.overrides}, policies...)
	}
	for _, policy := range r.parent.parent.policies {
		policies = append(policies, policy.defaults)
		policies = append([]T{policy.overrides}, policies...)
	}
	result := policies[0]
	for _, policy := range policies[1:] {
		result = merger(result, policy)
	}
	return result
}

type PolicySpec[T Policy] struct {
	name      string
	defaults  T
	overrides T
}

type Policy interface {
}

func NewGatewayClass[T Policy](name string) GatewayClass[T] {
	return GatewayClass[T]{
		name:     name,
		gateways: make(map[*Gateway[T]]void),
	}
}
