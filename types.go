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

// DesignControls are continuous intent signals. Jev chooses these qualities;
// the compiler decides which concrete primitives can express them.
type DesignControls struct {
	SpatialTension     float64 `json:"spatialTension"`
	VisualAbstraction  float64 `json:"visualAbstraction"`
	InformationDensity float64 `json:"informationDensity"`
	NarrativeDepth     float64 `json:"narrativeDepth"`
	TrustPriority      float64 `json:"trustPriority"`
	MotionEnergy       float64 `json:"motionEnergy"`
	Symmetry           float64 `json:"symmetry"`
}

type LayoutSpec struct {
	Axis      string `json:"axis,omitempty"`
	Columns   int    `json:"columns,omitempty"`
	Gap       string `json:"gap,omitempty"`
	Span      int    `json:"span,omitempty"`
	Align     string `json:"align,omitempty"`
	MinHeight string `json:"minHeight,omitempty"`
	Inset     string `json:"inset,omitempty"`
	Reverse   bool   `json:"reverse,omitempty"`
}

type StyleSpec struct {
	Surface  string `json:"surface,omitempty"`
	Scale    string `json:"scale,omitempty"`
	Emphasis string `json:"emphasis,omitempty"`
	Shape    string `json:"shape,omitempty"`
}

type VisualSpec struct {
	Kind      string  `json:"kind,omitempty"`
	Position  string  `json:"position,omitempty"`
	Intensity float64 `json:"intensity,omitempty"`
}

type InteractionSpec struct {
	Trigger  string  `json:"trigger,omitempty"`
	Motion   string  `json:"motion,omitempty"`
	Strength float64 `json:"strength,omitempty"`
}

type NodeContent struct {
	Eyebrow  string          `json:"eyebrow,omitempty"`
	Headline string          `json:"headline,omitempty"`
	Body     string          `json:"body,omitempty"`
	Value    string          `json:"value,omitempty"`
	Items    []BlueprintItem `json:"items,omitempty"`
}

// DesignNode is a safe, recursive design program. Primitive is deliberately
// lower-level than a website section: layout, content, visual and interaction
// can be recombined without introducing arbitrary HTML or JavaScript.
type DesignNode struct {
	ID          string          `json:"id"`
	Primitive   string          `json:"primitive"`
	Role        string          `json:"role,omitempty"`
	Layout      LayoutSpec      `json:"layout,omitempty"`
	Style       StyleSpec       `json:"style,omitempty"`
	Visual      VisualSpec      `json:"visual,omitempty"`
	Interaction InteractionSpec `json:"interaction,omitempty"`
	Content     NodeContent     `json:"content,omitempty"`
	Children    []DesignNode    `json:"children,omitempty"`
}

type PageBlueprint struct {
	Version           int        `json:"version"`
	ID                string     `json:"id"`
	CreativeDirection string     `json:"creativeDirection"`
	Root              DesignNode `json:"root"`
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
	Controls        DesignControls   `json:"controls"`
	Blueprint       PageBlueprint    `json:"blueprint"`
}

type DesignResult struct {
	Mode           string     `json:"mode"`
	LatencyMS      int64      `json:"latencyMs"`
	Prompt         string     `json:"prompt"`
	Generation     int        `json:"generation"`
	Winner         int        `json:"winner"`
	SwarmSize      int        `json:"swarmSize"`
	MutationLog    []string   `json:"mutationLog"`
	Specs          []PageSpec `json:"specs"`
	Generator      string     `json:"generator"`
	ASTSource      string     `json:"astSource"`
	CandidateCount int        `json:"candidateCount"`
	DesignSpace    string     `json:"designSpace"`
}

type AppState struct {
	PreviewURL string       `json:"previewURL"`
	Result     DesignResult `json:"result"`
	HasAPIKey  bool         `json:"hasAPIKey"`
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
