package bahamut

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSOriginMirror instruts to mirror any incoming origin.
// This should not be used in production as this is a development
// feature that is not secure.
const CORSOriginMirror = "_mirror_"

// CORSPolicy allows to configure
// CORS Access Control header of a response.
type CORSPolicy struct {
	additionalOrigins map[string]struct{}
	AllowOrigin       string
	AllowHeaders      []string
	AllowMethods      []string
	ExposeHeaders     []string
	MaxAge            int
	AllowCredentials  bool
}

type corsPolicyController struct {
	policy *CORSPolicy
}

// NewDefaultCORSController returns a CORSPolicyController that always returns a CORSAccessControlPolicy
// with sensible defaults.
func NewDefaultCORSController(origin string, additionalOrigins []string) CORSPolicyController {

	additionalOriginsMap := make(map[string]struct{}, len(additionalOrigins))
	if len(additionalOrigins) > 0 {
		for _, o := range additionalOrigins {
			additionalOriginsMap[o] = struct{}{}
		}
	}

	return &corsPolicyController{
		policy: &CORSPolicy{
			AllowOrigin:       origin,
			additionalOrigins: additionalOriginsMap,
			AllowCredentials:  true,
			MaxAge:            1500,
			AllowHeaders: []string{
				"Authorization",
				"Accept",
				"Content-Type",
				"Cache-Control",
				"Cookie",
				"If-Modified-Since",
				"X-Requested-With",
				"X-Count-Total",
				"X-Namespace",
				"X-External-Tracking-Type",
				"X-External-Tracking-ID",
				"X-TLS-Client-Certificate",
				"Accept-Encoding",
				"X-Fields",
				"X-Read-Consistency",
				"X-Write-Consistency",
				"Idempotency-Key",
			},
			AllowMethods: []string{
				"GET",
				"POST",
				"PUT",
				"DELETE",
				"PATCH",
				"HEAD",
				"OPTIONS",
			},
			ExposeHeaders: []string{
				"X-Requested-With",
				"X-Count-Total",
				"X-Namespace",
				"X-Messages",
				"X-Fields",
				"X-Next",
			},
		},
	}
}

func (c *corsPolicyController) PolicyForRequest(*http.Request) *CORSPolicy {
	return c.policy
}

// Inject injects the CORS header on the given http.Header. It will use
// the given request origin to determine the allow origin policy and the method
// to determine if it should inject pre-flight OPTIONS header.
// If the given http.Header is nil, this function is a no op.
func (a *CORSPolicy) Inject(h http.Header, origin string, preflight bool) {

	if h == nil {
		return
	}

	corsOrigin := a.AllowOrigin

	switch {
	case a.AllowOrigin == "*":
		corsOrigin = "*"

	case a.AllowOrigin == CORSOriginMirror && origin != "":
		corsOrigin = origin

	case a.AllowOrigin == CORSOriginMirror && origin == "":
		corsOrigin = ""

	case func() bool { _, ok := a.additionalOrigins[origin]; return ok }():
		corsOrigin = origin
	}

	if preflight {
		h.Set("Access-Control-Allow-Headers", strings.Join(a.AllowHeaders, ", "))
		h.Set("Access-Control-Allow-Methods", strings.Join(a.AllowMethods, ", "))
		h.Set("Access-Control-Max-Age", strconv.Itoa(a.MaxAge))
	}

	if corsOrigin != "" {
		h.Set("Access-Control-Allow-Origin", corsOrigin)
	}

	h.Set("Access-Control-Expose-Headers", strings.Join(a.ExposeHeaders, ", "))

	if a.AllowCredentials && corsOrigin != "*" && corsOrigin != "" {
		h.Set("Access-Control-Allow-Credentials", "true")
	}
}
