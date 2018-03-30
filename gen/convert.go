package main

import (
	"strings"

	"github.com/tehwalris/go-freeipa/thirdparty/snaker"
)

var reservedWords = []string{
	"continue",
	"break",
	"return",
	"type",
	"map",
}

func safeName(s string) string {
	for _, r := range reservedWords {
		if r == s {
			return "_" + s
		}
	}
	return s
}

func toGoType(ipaType string) string {
	switch ipaType {
	case "":
		return "interface{}"
	case "dict":
		return "interface{}"
	case "object":
		return "interface{}"
	case "NoneType":
		return "interface{}"
	case "unicode":
		return "string"
	case "str":
		return "string"
	case "bytes":
		return "string"
	case "datetime":
		return "time.Time"
	case "DN":
		return "string"
	case "Principal":
		return "string"
	case "DNSName":
		return "string"
	case "Decimal":
		return "float64"
	default:
		return ipaType
	}
}

func (t *CommandOutput) GoType(parent *Command) string {
	if t.Type == "dict" {
		cls := strings.Split(parent.ObjClass, "/")[0]
		if cls != "" {
			return upperName(cls)
		}
	}
	return toGoType(t.Type)
}

func lowerName(s string) string {
	return safeName(snaker.SnakeToCamelLower(s))
}

func upperName(s string) string {
	return safeName(snaker.SnakeToCamel(s))
}

func (t *Topic) LowerName() string {
	return lowerName(t.Name)
}

func (t *Topic) UpperName() string {
	return upperName(t.Name)
}

func (t *Class) LowerName() string {
	return lowerName(t.Name)
}

func (t *Class) UpperName() string {
	return upperName(t.Name)
}

func (t *Command) LowerName() string {
	return lowerName(t.Name)
}

func (t *Command) UpperName() string {
	return upperName(t.Name)
}

func (t *Param) LowerName() string {
	return lowerName(t.Name)
}

func (t *Param) UpperName() string {
	return upperName(t.Name)
}

func (t *CommandOutput) LowerName() string {
	return lowerName(t.Name)
}

func (t *CommandOutput) UpperName() string {
	return upperName(t.Name)
}
