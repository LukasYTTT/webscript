package ast

type Program struct {
	Imports []*ImportStatement
	Servers []*Server // For standard webserver capabilities
}

type ImportStatement struct {
	Path string
}

type Server struct {
	Domain string
	Routes []*Route
}

type Route struct {
	Path   string
	Target *Target
}

type TargetType string

const (
	TargetProxy  TargetType = "proxy"
	TargetStatic TargetType = "static"
)

type Target struct {
	Type  TargetType
	Value string
}
