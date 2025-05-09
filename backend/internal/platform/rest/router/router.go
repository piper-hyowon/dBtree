package router

import (
	"context"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"github.com/piper-hyowon/dBtree/internal/platform/rest"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type Router struct {
	routes []*route
	logger *log.Logger
}

type route struct {
	method  string
	regex   *regexp.Regexp
	handler http.HandlerFunc
	params  []string
}

func New(logger *log.Logger) *Router {
	return &Router{
		routes: []*route{},
		logger: logger,
	}
}

func (r *Router) GET(path string, handler http.HandlerFunc) {
	r.addRoute(http.MethodGet, path, handler)
}

func (r *Router) POST(path string, handler http.HandlerFunc) {
	r.addRoute(http.MethodPost, path, handler)
}

func (r *Router) PUT(path string, handler http.HandlerFunc) {
	r.addRoute(http.MethodPut, path, handler)
}

func (r *Router) DELETE(path string, handler http.HandlerFunc) {
	r.addRoute(http.MethodDelete, path, handler)
}

func (r *Router) PATCH(path string, handler http.HandlerFunc) {
	r.addRoute(http.MethodPatch, path, handler)
}

func (r *Router) Handle(method, path string, handler http.HandlerFunc) {
	r.addRoute(method, path, handler)
}

func (r *Router) addRoute(method, path string, handler http.HandlerFunc) {
	// url path 에서 파라미터 추출(/users/:id -> id)
	var params []string
	pattern := "^"

	parts := strings.Split(path, "/")
	for i, part := range parts {
		if i > 0 {
			pattern += "/"
		}

		if strings.HasPrefix(part, ":") {
			params = append(params, part[1:])
			pattern += "([^/]+)"
		} else {
			pattern += part
		}
	}
	pattern += "$"

	regex := regexp.MustCompile(pattern)

	r.routes = append(r.routes, &route{
		method:  method,
		regex:   regex,
		handler: handler,
		params:  params,
	})
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var matchedRoutes []*route
	for _, route := range r.routes {
		if route.regex.MatchString(req.URL.Path) {
			matchedRoutes = append(matchedRoutes, route)
		}
	}

	if len(matchedRoutes) > 0 {
		for _, route := range matchedRoutes {
			if route.method == req.Method {
				matches := route.regex.FindStringSubmatch(req.URL.Path)

				if len(route.params) > 0 && len(matches) > 1 {
					values := make(map[string]string)
					for i, param := range route.params {
						values[param] = matches[i+1]
					}
					req = req.WithContext(withParams(req.Context(), values))
				}

				route.handler(w, req)
				return
			}
		}

		allowedMethods := getAllowedMethods(matchedRoutes)
		w.Header().Set("Allow", allowedMethods)

		rest.HandleError(w, errors.NewMethodNotAllowedError(allowedMethods), r.logger)
		return
	}

	rest.HandleError(w, errors.NewEndpointNotFoundError(req.URL.Path), r.logger)
}

func getAllowedMethods(routes []*route) string {
	methods := make([]string, 0, len(routes))
	for _, route := range routes {
		methods = append(methods, route.method)
	}
	return strings.Join(methods, ", ")
}

type paramsKey struct{}

func withParams(ctx context.Context, params map[string]string) context.Context {
	return context.WithValue(ctx, paramsKey{}, params)
}

func Params(r *http.Request) map[string]string {
	params, _ := r.Context().Value(paramsKey{}).(map[string]string)
	return params
}

func Param(r *http.Request, name string) string {
	params := Params(r)
	if params == nil {
		return ""
	}
	return params[name]
}
