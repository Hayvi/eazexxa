package server

import (
	"net/http"
)

type Router struct {
	mux    *http.ServeMux
	middleware []func(http.Handler) http.Handler
}

func NewRouter() *Router {
	return &Router{
		mux:        http.NewServeMux(),
		middleware: []func(http.Handler) http.Handler{},
	}
}

func (r *Router) Use(middleware ...func(http.Handler) http.Handler) {
	r.middleware = append(r.middleware, middleware...)
}

func (r *Router) Handle(pattern string, handler http.Handler) {
	r.mux.Handle(pattern, handler)
}

func (r *Router) HandleFunc(pattern string, handler http.HandlerFunc) {
	r.mux.HandleFunc(pattern, handler)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handler := http.Handler(r.mux)
	
	for i := len(r.middleware) - 1; i >= 0; i-- {
		handler = r.middleware[i](handler)
	}
	
	handler.ServeHTTP(w, req)
}

func (r *Router) Group(prefix string, middleware ...func(http.Handler) http.Handler) *Group {
	return &Group{
		router:     r,
		prefix:     prefix,
		middleware: middleware,
	}
}

type Group struct {
	router     *Router
	prefix     string
	middleware []func(http.Handler) http.Handler
}

func (g *Group) Handle(pattern string, handler http.Handler) {
	fullPattern := g.prefix + pattern
	
	wrapped := handler
	for i := len(g.middleware) - 1; i >= 0; i-- {
		wrapped = g.middleware[i](wrapped)
	}
	
	g.router.Handle(fullPattern, wrapped)
}

func (g *Group) HandleFunc(pattern string, handler http.HandlerFunc) {
	g.Handle(pattern, handler)
}
