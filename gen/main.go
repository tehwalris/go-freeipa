package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"
)

var skipCommands = []string{
	"batch",
	"command_defaults",
	"env",
}

var skipClasses = []string{
	"class",
	"command",
}

func main() {
	if e := actualMain(); e != nil {
		log.Fatalf("%v", e)
	}
}

func actualMain() error {
	schema, e := loadSchema()
	if e != nil {
		return e
	}
	errs, e := loadErrs()
	if e != nil {
		return e
	}
	if e := generateMain(schema, errs); e != nil {
		return e
	}
	return nil
}

func loadSchema() (*Schema, error) {
	input, e := ioutil.ReadFile("../data/schema.json")
	if e != nil {
		return nil, e
	}
	parsed := &SchemaDump{}
	if e = json.Unmarshal(input, parsed); e != nil {
		return nil, e
	}
	schema := parsed.Result.Result

	cmds := make([]*Command, 0, len(schema.Commands))
	for _, c := range schema.Commands {
		var skip bool
		for _, s := range skipCommands {
			if s == c.Name {
				skip = true
			}
		}
		if !skip {
			// HACK FreeIPA does not return the correct type on deletions.
			if c.AttrName == "del" {
				for _, p := range c.Output {
					if p.Name == "result" && p.Type == "dict" {
						p.Type = "interface{}"
					}
				}
			}

			cmds = append(cmds, c)
		}
	}
	schema.Commands = cmds

	classes := make([]*Class, 0, len(schema.Classes))
	for _, c := range schema.Classes {
		var skip bool
		for _, s := range skipCommands {
			if s == c.Name {
				skip = true
			}
		}
		if !skip {
			// HACK FreeIPA admin user has no "givenname" or "cn", even though the schema
			// says these fields are required. This workaround makes it optional.
			if c.Name == "user" {
				for _, p := range c.Params {
					if p.Name == "givenname" || p.Name == "cn" {
						v := false
						p.RequiredRaw = &v
					}
				}
			}

			// HACK FreeIPA sometimes doesn't supply boolean fields which are
			// marked required in schema. This workaround makes them optional.
			for _, p := range c.Params {
				if p.Type == "bool" {
					v := false
					p.RequiredRaw = &v
				}
			}

			classes = append(classes, c)
		}
	}

	// TODO assert "version" on all objects is "1"
	// TODO assert that names are consistent within each object

	return &schema, nil
}

func loadErrs() ([]ErrDesc, error) {
	in, e := ioutil.ReadFile("../data/errors.json")
	if e != nil {
		return nil, e
	}
	var out []ErrDesc
	if e = json.Unmarshal(in, &out); e != nil {
		return nil, e
	}
	return out, nil
}

func generateMain(schema *Schema, errs []ErrDesc) error {
	t, e := template.New("freeipa.gotmpl").Funcs(template.FuncMap{
		"ToGoType":  toGoType,
		"TrimSpace": strings.TrimSpace,
	}).ParseFiles("./freeipa.gotmpl")
	if e != nil {
		return e
	}
	f, e := os.OpenFile("../freeipa/generated.go", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if e != nil {
		return e
	}
	e = t.Execute(f, struct {
		Schema *Schema
		Errs   []ErrDesc
	}{schema, errs})
	if e != nil {
		return e
	}
	return nil
}
