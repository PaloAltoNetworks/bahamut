package {{routes_package_name}}

import (
	"net/http"

	"github.com/aporeto-inc/cid/materia/bahamut"
    "{{base_package}}/{{handlers_package_name}}"
)

// Routes returns the list of all possible routes generated from a Monolithe Specifications Set.
func Routes() []*bahamut.Route {

	var routes []*bahamut.Route

    {% for spec in specifications.values() -%}

    // routes for {{spec.resource_name}}
    {% if spec.allows_create -%}
    routes = append(routes, bahamut.NewRoute("/{{spec.resource_name}}/:id", http.MethodPost, handlers.Create{{spec.entity_name}}))
    {% endif -%}

    {% if spec.allows_get -%}
    {% if spec.is_root -%}
    routes = append(routes, bahamut.NewRoute("/{{spec.resource_name}}", http.MethodGet, handlers.Retrieve{{spec.entity_name}}))
    {% else -%}
    routes = append(routes, bahamut.NewRoute("/{{spec.resource_name}}/:id", http.MethodGet, handlers.Retrieve{{spec.entity_name}}))
    {% endif -%}
    {% endif -%}

    {% if spec.allows_update -%}
    routes = append(routes, bahamut.NewRoute("/{{spec.resource_name}}/:id", http.MethodPut, handlers.Update{{spec.entity_name}}))
    {% endif -%}

    {% if spec.allows_delete -%}
    routes = append(routes, bahamut.NewRoute("/{{spec.resource_name}}/:id", http.MethodDelete, handlers.Delete{{spec.entity_name}}))
    {% endif -%}

    {% for child_api in spec.child_apis -%}
    {% set child_rest_name = child_api.rest_name -%}
    {% set child_spec = specifications[child_rest_name] -%}
    {% set child_resource_name = child_spec.resource_name -%}
    {% set child_entity_name = child_spec.entity_name -%}

    {% if spec.is_root -%}
    {% set child_path = child_resource_name -%}
    {% else -%}
    {% set child_path = "%s/:id/%s" % (spec.resource_name, child_resource_name) -%}
    {% endif -%}

    {% if child_api.allows_create -%}
    routes = append(routes, bahamut.NewRoute("/{{child_path}}", http.MethodPost, handlers.Create{{child_entity_name}}))
    {% endif -%}

    {% if child_api.allows_update and child_api.relationship == "member" -%}
    routes = append(routes, bahamut.NewRoute("/{{child_path}}", http.MethodPatch, handlers.Patch{{child_entity_name}}))
    {% endif -%}

    {% if spec.allows_get -%}
    routes = append(routes, bahamut.NewRoute("/{{child_path}}", http.MethodGet, handlers.RetrieveMany{{child_entity_name}}))
    {% endif -%}

    {% endfor %}
    {% endfor -%}

	return routes
}
