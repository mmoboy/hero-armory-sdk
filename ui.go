package sdk

// HUDActionSpec registers one compact action button inside the World "HUD Actions" panel.
type HUDActionSpec struct {
	ID      string `json:"id"`
	Tooltip string `json:"tooltip,omitempty"`
	Icon    string `json:"icon"` // atlas icon name (e.g. "star") or bundle .png path
	Badge   string `json:"badge,omitempty"`
	Active  bool   `json:"active,omitempty"`
}

// UIPanelSpec describes one floating declarative UI overlay panel in World mode.
type UIPanelSpec struct {
	ID          string      `json:"id"`
	Visible     bool        `json:"visible"`
	Anchor      string      `json:"anchor,omitempty"` // "top_left", "top_right", "bottom_left", "bottom_right", "center"
	OffsetX     float32     `json:"offset_x,omitempty"`
	OffsetY     float32     `json:"offset_y,omitempty"`
	Width       float32     `json:"width,omitempty"`
	Height      float32     `json:"height,omitempty"`
	MaxHeight   float32     `json:"max_height,omitempty"`
	Padding     float32     `json:"padding,omitempty"`
	Gap         float32     `json:"gap,omitempty"`
	BgColor     [4]float32  `json:"bg_color,omitempty"`
	BorderColor [4]float32  `json:"border_color,omitempty"`
	BorderWidth float32     `json:"border_width,omitempty"`
	Radius      float32     `json:"radius,omitempty"`
	Root        *UINodeSpec `json:"root,omitempty"`
}

// UINodeSpec describes one whitelisted UI node rendered by the native client.
type UINodeSpec struct {
	Kind         string       `json:"kind"` // "panel", "col", "row", "label", "button", "input", "icon", "image", "progress", "scroll"
	ID           string       `json:"id,omitempty"`
	Text         string       `json:"text,omitempty"`
	Placeholder  string       `json:"placeholder,omitempty"`
	Variant      string       `json:"variant,omitempty"` // "default", "active", "emerald", "danger"
	Icon         string       `json:"icon,omitempty"`
	AssetPath    string       `json:"asset_path,omitempty"`
	Progress     float32      `json:"progress,omitempty"` // 0..1 for progress bars
	Width        float32      `json:"width,omitempty"`
	Height       float32      `json:"height,omitempty"`
	PercentWidth float32      `json:"percent_width,omitempty"`
	FlexGrow     float32      `json:"flex_grow,omitempty"`
	Padding      float32      `json:"padding,omitempty"`
	Gap          float32      `json:"gap,omitempty"`
	TextSize     float32      `json:"text_size,omitempty"`
	Wrap         bool         `json:"wrap,omitempty"`
	Color        [4]float32   `json:"color,omitempty"`
	BorderColor  [4]float32   `json:"border_color,omitempty"`
	Radius       float32      `json:"radius,omitempty"`
	Children     []UINodeSpec `json:"children,omitempty"`
}

// ColNode creates a vertical flex layout container node.
func ColNode(id string, gap float32, children ...UINodeSpec) UINodeSpec {
	return UINodeSpec{Kind: "col", ID: id, Gap: gap, PercentWidth: 100, Children: children}
}

// RowNode creates a horizontal flex layout container node.
func RowNode(id string, gap float32, children ...UINodeSpec) UINodeSpec {
	return UINodeSpec{Kind: "row", ID: id, Gap: gap, PercentWidth: 100, Children: children}
}

// CardNode creates a styled Panel container node.
func CardNode(id string, padding, gap float32, bg, border [4]float32, children ...UINodeSpec) UINodeSpec {
	return UINodeSpec{
		Kind: "panel", ID: id, Padding: padding, Gap: gap, PercentWidth: 100,
		Color: bg, BorderColor: border, Radius: 7, Children: children,
	}
}

// LabelNode creates a text label node.
func LabelNode(id, text string, size float32, color [4]float32, wrap bool) UINodeSpec {
	return UINodeSpec{Kind: "label", ID: id, Text: text, TextSize: size, Color: color, Wrap: wrap}
}

// ButtonNode creates a clickable HUD button node triggering OnUIEvent(id, "").
func ButtonNode(id, label, variant string, width, height float32) UINodeSpec {
	return UINodeSpec{Kind: "button", ID: id, Text: label, Variant: variant, Width: width, Height: height}
}

// InputNode creates a text input box triggering OnUIEvent(id, value) on submit.
func InputNode(id, placeholder, value string, height float32) UINodeSpec {
	return UINodeSpec{Kind: "input", ID: id, Placeholder: placeholder, Text: value, Height: height, PercentWidth: 100}
}

// ProgressNode creates a bounded horizontal progress meter (0.0 .. 1.0).
func ProgressNode(id string, progress float32, height float32, fillColor [4]float32) UINodeSpec {
	return UINodeSpec{Kind: "progress", ID: id, Progress: progress, Height: height, Color: fillColor, PercentWidth: 100}
}

// ImageNode creates a sandboxed PNG texture view loaded from inside the .gtpack archive.
func ImageNode(id, assetPath string, width, height float32) UINodeSpec {
	return UINodeSpec{Kind: "image", ID: id, AssetPath: assetPath, Width: width, Height: height}
}
