package main

type ColorValue struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
	Shade string `json:"shade,omitempty"`
}
type IntentArguments struct {
	Channel   string      `json:"channel,omitempty"`
	Color     *ColorValue `json:"color,omitempty"`
	Utility   string      `json:"utility,omitempty"`
	Element   string      `json:"element,omitempty"`
	Placement string      `json:"placement,omitempty"`
	Label     string      `json:"label,omitempty"`
}
type SourceFile struct {
	Path   string `json:"path"`
	Before string `json:"before"`
	After  string `json:"after"`
}
type SourceCheck struct {
	ID        string `json:"id"`
	ClassName string `json:"className,omitempty"`
	ClassKind string `json:"classKind,omitempty"`
	Label     string `json:"label,omitempty"`
	Component string `json:"component,omitempty"`
}
type SourcePatch struct {
	Files    []SourceFile  `json:"files"`
	Checks   []SourceCheck `json:"checks"`
	Executor string        `json:"executor"`
	Error    string        `json:"error,omitempty"`
}
type ComponentEntry struct {
	ID             string   `json:"id"`
	Component      string   `json:"component"`
	Source         string   `json:"source"`
	ImportFrom     string   `json:"importFrom"`
	DefaultContent string   `json:"defaultContent"`
	Variants       []string `json:"variants"`
}
type EditIntent struct {
	Arguments  IntentArguments `json:"arguments"`
	State      string          `json:"state,omitempty"`
	Scope      string          `json:"scope"`
	Family     string          `json:"family"`
	Operation  string          `json:"operation"`
	Direction  string          `json:"direction"`
	Magnitude  string          `json:"magnitude"`
	Breakpoint string          `json:"breakpoint"`
	Value      string          `json:"value,omitempty"`
	Preserve   []string        `json:"preserve"`
	Confidence float64         `json:"confidence"`
}
type Target struct {
	Revision    string            `json:"revision"`
	Fingerprint string            `json:"fingerprint"`
	Start       int               `json:"start"`
	End         int               `json:"end"`
	ClassName   string            `json:"className"`
	ClassKind   string            `json:"classKind"`
	Repeated    bool              `json:"repeated"`
	Empty       bool              `json:"empty"`
	ID          string            `json:"id"`
	Tag         string            `json:"tag"`
	Source      string            `json:"source"`
	Line        int               `json:"line"`
	Parent      string            `json:"parent"`
	Styles      map[string]string `json:"styles"`
	Important   []string          `json:"important"`
	Count       int               `json:"count"`
	Width       int               `json:"width"`
}
type ProjectMeta struct {
	Tailwind   bool              `json:"tailwind"`
	Tokens     map[string]string `json:"tokens"`
	Components []ComponentEntry  `json:"components"`
	Root       string            `json:"root"`
	Targets    []Target          `json:"targets"`
	Version    int               `json:"version"`
}
type Expectation struct {
	ID         string            `json:"id"`
	Breakpoint string            `json:"breakpoint"`
	Properties map[string]string `json:"properties"`
}

type Command struct {
	Revisions    map[string]string `json:"revisions,omitempty"`
	Checks       []SourceCheck     `json:"checks,omitempty"`
	ClientID     string            `json:"clientId"`
	Expectations []Expectation     `json:"expectations,omitempty"`
	ID           string            `json:"id"`
	Kind         string            `json:"kind"`
	Targets      []string          `json:"targets,omitempty"`
	Hash         string            `json:"hash,omitempty"`
}
type BrowserReport struct {
	ClientID  string   `json:"clientId"`
	Project   string   `json:"project"`
	Session   string   `json:"session"`
	Selected  *Target  `json:"selected,omitempty"`
	Targets   []Target `json:"targets,omitempty"`
	CommandID string   `json:"commandId,omitempty"`
	Hash      string   `json:"hash,omitempty"`
	Error     string   `json:"error,omitempty"`
	Selecting bool     `json:"selecting"`
}
type Stage struct {
	Name string `json:"name"`
	MS   int64  `json:"ms"`
}
type EditorState struct {
	Files            []string     `json:"files"`
	Executor         string       `json:"executor"`
	ClientID         string       `json:"-"`
	Project          string       `json:"project"`
	Session          string       `json:"session"`
	PreviewURL       string       `json:"previewURL"`
	Connected        bool         `json:"connected"`
	BrowserConnected bool         `json:"browserConnected"`
	Selected         *Target      `json:"selected"`
	HasAPIKey        bool         `json:"hasAPIKey"`
	Mode             string       `json:"mode"`
	Busy             bool         `json:"busy"`
	Pending          bool         `json:"pending"`
	CanUndo          bool         `json:"canUndo"`
	Selecting        bool         `json:"selecting"`
	Intents          []EditIntent `json:"intents"`
	Diff             string       `json:"diff"`
	Error            string       `json:"error"`
	Stages           []Stage      `json:"stages"`
	ChangeID         string       `json:"changeId"`
}
type Snapshot struct {
	Files    []SourceFile `json:"files,omitempty"`
	Executor string       `json:"executor,omitempty"`
	Project  string       `json:"project"`
	Before   []byte       `json:"before"`
	After    []byte       `json:"after"`
	Pending  bool         `json:"pending"`
	ID       string       `json:"id"`
}
type ApplyRequest struct {
	Prompt   string `json:"prompt"`
	Session  string `json:"session"`
	TargetID string `json:"targetId"`
}
