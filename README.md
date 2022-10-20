# [Gateway API](https://gateway-api.sigs.k8s.io/) [Policy Attachment](https://gateway-api.sigs.k8s.io/references/policy-attachment/) playground

[Kuadrant](https://github.com/Kuadrant/) is trying to figure how to best implement the default & override hierarchy, as 
proposed by the specification. This is done in two different contexts: 
 - `AuthPolicy`, providing externalized authn/z to your services; and 
 - `RateLimitPolicy`, which lets you rate limit the access to your services from within the gateway.

## The Playground

This simple playground aims at providing a space to programmatically test out different `Policy` "languages" (i.e. how
do you declaratively express these constraints) while supporting the `default` and `override` concepts.

The idea is to use test code, to spec your `Policy` and most importantly your _merging algorithm_ for these and use the 
tests as specification for what the expected outcome to applying these different `Policy`ies would be.

### Some quick examples

- How would a _cluster operator_ express that any traffic to the cluster needs some sort of authentication, 
while leaving it to the _application developer_ to specify how that authentication happens?
- How would a _cluster operator_ require authentication with the company SSO, while the _application developer_
specifies what role a user of the app would need for certain end-points?
- How does a _cluster operator_ rate limits a whole subnet of clients, while the _application developer_ bypasses it 
for authenticated `admin` users from that same subnet?
- … more?

## Ideas behind the playground

 - The array of `Policy`ies is meant to represent "time ordering" of `Policy` CRs, as the "oldest" has precedence. 
 - It builds on the assumption that a `Policy` merge is only required once a `Service` will be hit, i.e. at the 
`HTTPRoute` level.
 - Actual use cases should only require a `_test.go` file, with the `Policy` under test and a `Merger` function 
that knows about the semantic of the `Policy` and the possible "language" used (e.g. the user could submit a `Policy` 
CR with fields different from the actual resulting `Policy` applied, following the "merge").
 - … more?
