package sdk

import (
	"encoding/json"
	"fmt"
)

// HUDActionSpec registers one compact action button inside the World "HUD Actions" panel.
type HUDActionSpec struct {
	ID      string `json:"id"`
	Tooltip string `json:"tooltip,omitempty"`
	Icon    string `json:"icon"`            // atlas icon name (e.g. "star") or bundle .png path
	Badge   string `json:"badge,omitempty"` // at most MaxHUDBadgeLen chars; longer badges are cut
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

// UINodePatch updates one existing overlay node (matched by UINodeSpec.ID)
// without rebuilding the panel, so focus, scroll position, and typed input
// survive. Nil fields are left unchanged.
type UINodePatch struct {
	ID       string      `json:"id"`
	Text     *string     `json:"text,omitempty"`     // label/button text
	Value    *string     `json:"value,omitempty"`    // input field value ("" clears it)
	Progress *float32    `json:"progress,omitempty"` // progress bar fill 0..1
	Color    *[4]float32 `json:"color,omitempty"`    // label text / progress fill color
	Hidden   *bool       `json:"hidden,omitempty"`   // hide or show the node
}

// UIOverlayPatch is the payload of PatchOverlay.
type UIOverlayPatch struct {
	PanelID string        `json:"panel_id"`
	Patches []UINodePatch `json:"patches"`
}

// UINodeSpec describes one whitelisted UI node rendered by the native client.
type UINodeSpec struct {
	Kind         string       `json:"kind"` // "panel", "col", "row", "label", "button", "input", "icon", "image", "progress", "scroll"
	ID           string       `json:"id,omitempty"`
	Text         string       `json:"text,omitempty"`
	Placeholder  string       `json:"placeholder,omitempty"`
	Variant      string       `json:"variant,omitempty"` // "default", "active" (alias "primary"), "emerald" (alias "success"), "danger"
	Icon         string       `json:"icon,omitempty"`
	AssetPath    string       `json:"asset_path,omitempty"`
	Progress     float32      `json:"progress,omitempty"` // 0..1 for progress bars (bar height is at least 12 px)
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
	Hidden       bool         `json:"hidden,omitempty"` // rendered hidden; toggle with PatchOverlay (API level 2+)
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

// Pointer helpers for UINodePatch fields.
func PatchText(s string) *string          { return &s }
func PatchProgress(v float32) *float32    { return &v }
func PatchColor(c [4]float32) *[4]float32 { return &c }
func PatchHidden(hidden bool) *bool       { return &hidden }

// ValidateOverlay checks panel against the host overlay limits (MaxUINodes,
// MaxUIDepth, MaxUISpecBytes) before it is sent, so oversized overlays fail
// with a readable error instead of being truncated by the host.
func ValidateOverlay(panel UIPanelSpec) error {
	raw, err := json.Marshal(panel)
	if err != nil {
		return err
	}
	if len(raw) > MaxUISpecBytes {
		return fmt.Errorf("overlay %q is %d bytes (max %d)", panel.ID, len(raw), MaxUISpecBytes)
	}
	if panel.Root == nil {
		return nil
	}
	nodes, depth := countUINodes(panel.Root, 0)
	if nodes > MaxUINodes {
		return fmt.Errorf("overlay %q has %d nodes (max %d)", panel.ID, nodes, MaxUINodes)
	}
	if depth > MaxUIDepth {
		return fmt.Errorf("overlay %q is %d levels deep (max %d)", panel.ID, depth, MaxUIDepth)
	}
	return nil
}

// countUINodes returns the node count and the deepest level below n, counting
// depth the way the client renderer does (root = 0).
func countUINodes(n *UINodeSpec, depth int) (count, maxDepth int) {
	count, maxDepth = 1, depth
	for i := range n.Children {
		c, d := countUINodes(&n.Children[i], depth+1)
		count += c
		if d > maxDepth {
			maxDepth = d
		}
	}
	return count, maxDepth
}
