package main

type SchemaDump struct {
	Result struct {
		Result Schema `json:"result"`
	} `json:"result"`
}

type Schema struct {
	Topics   []*Topic   `json:"topics"`
	Classes  []*Class   `json:"classes"`
	Commands []*Command `json:"commands"`

	Fingerprint string `json:"fingerprint"`
	TTL         int    `json:"ttl"`
	Version     string `json:"version"`
}

type Topic struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Doc      string `json:"doc"`
	Version  string `json:"version"`

	TopicTopic struct {
		Base64 string `json:"__base64__"`
	} `json:"topic_topic"`
}

type Class struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Version  string `json:"version"`

	Params []*Param `json:"params"`
}

type Command struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Doc      string `json:"doc"`
	Version  string `json:"version"`

	AttrName   string `json:"attr_name"`
	ObjClass   string `json:"obj_class"`
	TopicTopic string `json:"topic_topic"`

	Params []*Param         `json:"params"`
	Output []*CommandOutput `json:"output"`
}

type Param struct {
	Name    string `json:"name"`
	CliName string `json:"cli_name"`
	Label   string `json:"label"`
	Doc     string `json:"doc"`

	Type        string `json:"type"`
	Multivalue  bool   `json:"multivalue"`
	CliMetavar  string `json:"cli_metavar"`
	RequiredRaw *bool  `json:"required"` // use Requried() instead
	Positional  bool   `json:"positional"`

	AlwaysAsk bool     `json:"alwaysask"`
	NoConvert bool     `json:"no_convert"`
	Exclude   []string `json:"exclude"`

	Default          []string `json:"default"`
	DefaultFromParam []string `json:"default_from_param"`
}

type CommandOutput struct {
	Name string `json:"name"`
	Doc  string `json:"doc"`

	Type        string `json:"type"`
	Multivalue  bool   `json:"multivalue"`
	RequiredRaw *bool  `json:"required"` // use Requried() instead
}

func (t *Param) Required() bool {
	if t.RequiredRaw == nil {
		return len(t.Default) == 0 && len(t.DefaultFromParam) == 0
	}
	return *t.RequiredRaw
}

func (t *CommandOutput) Required() bool {
	if t.RequiredRaw == nil {
		return true
	}
	return *t.RequiredRaw
}

func (t *Command) PosParams() []*Param {
	var out []*Param
	for _, p := range t.Params {
		if p.Positional {
			out = append(out, p)
		}
	}
	return out
}

func (t *Command) KwParams() []*Param {
	var out []*Param
	for _, p := range t.Params {
		if !p.Positional {
			out = append(out, p)
		}
	}
	return out
}
