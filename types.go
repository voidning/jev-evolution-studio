package main

type Decision struct {
	Label        string             `json:"label"`
	Value        string             `json:"value"`
	Confidence   float64            `json:"confidence"`
	Group        string             `json:"group"`
	Distribution map[string]float64 `json:"distribution,omitempty"`
}

type Scorecard struct {
	Originality int `json:"originality"`
	Clarity     int `json:"clarity"`
	Trust       int `json:"trust"`
	Conversion  int `json:"conversion"`
	Composite   int `json:"composite"`
}

type FeatureContent struct {
	Kicker string `json:"kicker"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

type MetricContent struct {
	Value string `json:"value"`
	Unit  string `json:"unit"`
	Label string `json:"label"`
}

// BlueprintItem is a small semantic unit inside a generated section. The
// renderer decides how to present it from the section's layout and visual gene.
type BlueprintItem struct {
	Label string `json:"label"`
	Title string `json:"title"`
	Body  string `json:"body"`
	Value string `json:"value"`
}

// SectionNode is the recursive page grammar. It deliberately describes intent
// instead of HTML or Tailwind classes, keeping generated output safe and valid.
type SectionNode struct {
	ID       string          `json:"id"`
	Kind     string          `json:"kind"`
	Layout   string          `json:"layout"`
	Visual   string          `json:"visual"`
	Eyebrow  string          `json:"eyebrow"`
	Headline string          `json:"headline"`
	Body     string          `json:"body"`
	Items    []BlueprintItem `json:"items"`
	Children []SectionNode   `json:"children"`
}

type PageBlueprint struct {
	Version           int           `json:"version"`
	ID                string        `json:"id"`
	CreativeDirection string        `json:"creativeDirection"`
	Sections          []SectionNode `json:"sections"`
}

type PageSpec struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Descriptor      string           `json:"descriptor"`
	Strategy        string           `json:"strategy"`
	Generation      int              `json:"generation"`
	Mutation        string           `json:"mutation"`
	Theme           string           `json:"theme"`
	Hero            string           `json:"hero"`
	Visual          string           `json:"visual"`
	Features        string           `json:"features"`
	Density         string           `json:"density"`
	Navigation      string           `json:"navigation"`
	Motion          string           `json:"motion"`
	Story           string           `json:"story"`
	CTA             string           `json:"cta"`
	World           string           `json:"world"`
	Brand           string           `json:"brand"`
	Eyebrow         string           `json:"eyebrow"`
	SectionLabel    string           `json:"sectionLabel"`
	SectionTitle    string           `json:"sectionTitle"`
	FeaturesContent []FeatureContent `json:"featuresContent"`
	Metrics         []MetricContent  `json:"metrics"`
	ShowLogos       bool             `json:"showLogos"`
	ShowPricing     bool             `json:"showPricing"`
	ShowStats       bool             `json:"showStats"`
	Title           string           `json:"title"`
	Description     string           `json:"description"`
	Decisions       []Decision       `json:"decisions"`
	Scores          Scorecard        `json:"scores"`
	Blueprint       PageBlueprint    `json:"blueprint"`
}

type DesignResult struct {
	Mode        string     `json:"mode"`
	LatencyMS   int64      `json:"latencyMs"`
	Prompt      string     `json:"prompt"`
	Generation  int        `json:"generation"`
	Winner      int        `json:"winner"`
	SwarmSize   int        `json:"swarmSize"`
	MutationLog []string   `json:"mutationLog"`
	Specs       []PageSpec `json:"specs"`
	Generator   string     `json:"generator"`
	ASTSource   string     `json:"astSource"`
}

type AppState struct {
	PreviewURL      string       `json:"previewURL"`
	Result          DesignResult `json:"result"`
	HasAPIKey       bool         `json:"hasAPIKey"`
	HasGeneratorKey bool         `json:"hasGeneratorKey"`
}

type Answer struct {
	Type          string             `json:"type,omitempty"`
	Choice        string             `json:"choice,omitempty"`
	Confidence    float64            `json:"confidence,omitempty"`
	Noul          float64            `json:"noul,omitempty"`
	Score         float64            `json:"score,omitempty"`
	Probabilities map[string]float64 `json:"probabilities,omitempty"`
}

type Answers map[string]Answer

const defaultPrompt = "为一款实时 AI 数据分析产品做主页。面向开发者，深色、克制、有速度感，重点突出实时分析。"
