package api

import (
	"fmt"
)

type Property struct {
	ID    string `mapstructure:"id"`
	Value string `mapstructure:"value"`
	Label string `mapstructure:"label"`
}

type Edge struct {
	ID        string `mapstructure:"id"`
	Label     string `mapstructure:"label"`
	Type      Type   `mapstructure:"type"`
	InVLabel  string `mapstructure:"inVLabel"`
	InV       string `mapstructure:"inV"`
	OutVLabel string `mapstructure:"outVLabel"`
	OutV      string `mapstructure:"outV"`
}

type Vertex struct {
	Type       Type              `mapstructure:"type"`
	ID         string            `mapstructure:"id"`
	Label      string            `mapstructure:"label"`
	Properties VertexPropertyMap `mapstructure:"properties"`
}

type Value struct {
	ID    string      `mapstructure:"id"`
	Value interface{} `mapstructure:"value"`
}

type VertexPropertyMap map[string][]Value

type VertexProperty struct {
	Value
	Label string `mapstructure:"label"`
}

func (p VertexProperty) String() string {
	return fmt.Sprintf("[%s] '%s':'%s'", p.ID, p.Label, p.Value)
}

func (v Vertex) String() string {
	return fmt.Sprintf("%s %s (props %v - type %s", v.ID, v.Label, v.Properties, v.Type)
}

func (e Edge) String() string {
	return fmt.Sprintf("%s (%s)-%s->%s (%s) - type %s", e.InVLabel, e.InV, e.Label, e.OutVLabel, e.OutV, e.Type)
}
