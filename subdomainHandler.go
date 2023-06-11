package main

import "net/http"

type subdomainHandler struct {
	handlers map[string]http.Handler
}

func NewSubdomainHandler() *subdomainHandler {
	return &subdomainHandler{handlers: make(map[string]http.Handler, 32)}
}

func (s *subdomainHandler) On(domain string, handler http.Handler) *subdomainHandler {
	s.handlers[domain] = handler
	return s
}
func (s *subdomainHandler) Default(handler http.Handler) *subdomainHandler {
	return s.On("", handler)
}

func (s *subdomainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h, ok := s.handlers[r.Host]; ok {
		h.ServeHTTP(w, r)
		return
	}
	if h, ok := s.handlers[""]; ok {
		h.ServeHTTP(w, r)
		return
	}
	http.NotFound(w, r)
}
