package sdk

import (
	"fmt"
	"path"
	"strings"
)

const (
	// SDKVersion is the ABI version implemented by both host and plugin modules.
	SDKVersion = 1

	// APILevel is the additive host API revision this SDK targets. Level 2
	// added RemoveMonsterSpawn, GiveGold/GiveXP, GetHero, plugin chat
	// (SendChatToHero/BroadcastChat/ChatCommandPlugin), and PatchOverlay.
	// Level 3 added Hero inventory inspection/consumption (GetHeroInventory,
	// GetHeroItemCount, TakeItemFromHero), TeleportHero, multiple named client
	// overlays (ClearOverlayID + top_center/bottom_center anchors), and dynamic
	// 3D world props (SpawnProp, RemoveProp).
	// Level 4 added bespoke map & atmosphere customization: RoomAtmosphereSpec,
	// SetRoomAtmosphere/ClearRoomAtmosphere, RoomLiquidSpec, RoomLightSpec,
	// RoomEmitterSpec, RoomTriggerZoneSpec + ZoneObserverPlugin, custom-mesh
	// RoomInteractableSpec, RoomSpec MapColor/MapIcon, and dynamic world-map
	// markers (SetMapMarker/RemoveMapMarker).
	// Level 5 added generic combat primitives: dynamic actors (SpawnActor,
	// UpdateActor, RemoveActor), PlayVFX, ShowFloatingText, GroundMarkers,
	// SetHeroHotbar, SetHeroPose, SetHeroControlLock, CombatInputPlugin
	// (OnActorClick, OnGroundClick, OnHotbarPress), and
	// PluginManifest.DisableLegacyCombat.
	// Level 6 added unified action bar slot metadata (Group, KeyHint, Badge,
	// Charges, AccentColor, Selected) and flexible world map customization
	// (SetMapConfig, SetMapRegion, RemoveMapRegion, SetMapRoute,
	// RemoveMapRoute, enriched MapMarkerSpec, and MapInputPlugin).
	// A plugin calling a newer hostcall on an older host fails to load with
	// an error naming the missing host function.
	APILevel = 6

	MaxArchiveBytes   = 16 << 20 // 16 MiB compressed .gtpack max
	MaxExtractedBytes = 32 << 20 // 32 MiB extracted contents max
	MaxEntries        = 256
	MaxPacketBytes    = 64 * 1024
	MaxUISpecBytes    = 128 * 1024
	MaxUINodes        = 128
	MaxUIDepth        = 8
	MaxOverlays       = 4
	MaxHUDActions     = 6
	MaxHUDBadgeLen    = 6
	MaxChatTextLen    = 240
	MaxStringLen      = 512
	ChunkSizeBytes    = 64 * 1024

	MaxDynamicProps         = 256
	MaxDynamicPropsPerRoom  = 32
	MaxRoomLiquids          = 8
	MaxRoomLights           = 8
	MaxRoomEmitters         = 12
	MaxRoomZones            = 16
	MaxMapMarkers           = 64
	MaxMapRegions           = 16
	MaxMapRoutes            = 32
	MaxDynamicActors        = 128
	MaxDynamicActorsPerRoom = 32
	MaxActorStatuses        = 8
	MaxGroundMarkers        = 64
	MaxHotbarSlots          = 10

	MaxKVKeyLen       = 128
	MaxKVValueBytes   = 64 * 1024
	MaxKVKeysPerRealm = 2000            // cloud-realm default; local/dedicated realms are unlimited unless the operator sets a limit
	MaxKVTotalBytes   = 4 * 1024 * 1024 // cloud-realm default; see MaxKVKeysPerRealm
	MaxKVListLimit    = 256

	MaxPrefKeyLen     = 96
	MaxPrefValueBytes = 8 * 1024
	MaxPrefsPerRealm  = 128
	MaxPrefTotalBytes = 128 * 1024
	MaxPrefListLimit  = 128

	MaxCustomMeshes    = 32
	MaxCustomMeshBytes = 1024 * 1024
	MaxCustomMeshParts = 16
	MaxCustomMeshVerts = 16384
	MaxCustomMeshTris  = 32768
	MaxMonsterSpells   = 6
	MaxRoomSpecBytes   = 64 * 1024
	MaxNPCOptions      = 12
)

// PluginManifest describes one realm's client plugin bundle (manifest.json inside .gtpack).
type PluginManifest struct {
	SDKVersion             int      `json:"sdk_version"`
	ID                     string   `json:"id"`
	Name                   string   `json:"name"`
	Version                string   `json:"version"`
	BuiltWithClientVersion string   `json:"built_with_client_version,omitempty"`
	BuiltWithServerVersion string   `json:"built_with_server_version,omitempty"`
	Author                 string   `json:"author,omitempty"`
	Description            string   `json:"description,omitempty"`
	EntryWasm              string   `json:"entry_wasm,omitempty"`
	Icon                   string   `json:"icon,omitempty"`
	Permissions            []string `json:"permissions,omitempty"`
	Required               bool     `json:"required"`
	DisableLegacyCombat    bool     `json:"disable_legacy_combat,omitempty"`
}

// ValidateManifest validates and normalizes a PluginManifest.
func ValidateManifest(m *PluginManifest) error {
	if m == nil {
		return fmt.Errorf("manifest is required")
	}
	if m.SDKVersion == 0 {
		m.SDKVersion = SDKVersion
	}
	if m.SDKVersion != SDKVersion {
		return fmt.Errorf("unsupported sdk_version %d (want %d)", m.SDKVersion, SDKVersion)
	}
	m.ID = strings.TrimSpace(m.ID)
	if !IsValidPluginID(m.ID) {
		return fmt.Errorf("plugin id must be 3-40 lowercase ASCII chars [a-z0-9_-]")
	}
	m.Name = SanitizeASCII(strings.TrimSpace(m.Name), 48)
	if len(m.Name) < 3 {
		return fmt.Errorf("plugin name must be at least 3 ASCII chars")
	}
	m.Version = SanitizeASCII(strings.TrimSpace(m.Version), 24)
	if m.Version == "" {
		m.Version = "1.0.0"
	}
	m.BuiltWithClientVersion = SanitizeASCII(strings.TrimSpace(m.BuiltWithClientVersion), 24)
	m.BuiltWithServerVersion = SanitizeASCII(strings.TrimSpace(m.BuiltWithServerVersion), 24)
	m.Author = SanitizeASCII(strings.TrimSpace(m.Author), 40)
	m.Description = SanitizeASCII(strings.TrimSpace(m.Description), 240)
	if strings.TrimSpace(m.EntryWasm) == "" {
		m.EntryWasm = "client.wasm"
	}
	cleanEntry, err := CleanBundlePath(m.EntryWasm)
	if err != nil || path.Ext(cleanEntry) != ".wasm" {
		return fmt.Errorf("invalid entry_wasm %q", m.EntryWasm)
	}
	m.EntryWasm = cleanEntry
	if strings.TrimSpace(m.Icon) != "" {
		m.Icon = SanitizeASCII(strings.TrimSpace(m.Icon), 96)
	}
	return nil
}

// IsValidPluginID enforces lowercase ASCII identifier rules.
func IsValidPluginID(id string) bool {
	if len(id) < 3 || len(id) > 40 {
		return false
	}
	for i := 0; i < len(id); i++ {
		c := id[i]
		ok := (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' || c == '-'
		if !ok {
			return false
		}
	}
	return true
}

// IsValidContentID enforces lowercase ASCII identifier rules for items, rooms, NPCs, skills, and assets.
func IsValidContentID(id string) bool {
	if len(id) < 2 || len(id) > 48 {
		return false
	}
	for i := 0; i < len(id); i++ {
		c := id[i]
		ok := (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' || c == '-'
		if !ok {
			return false
		}
	}
	return true
}

// SanitizeASCII strips non-ASCII / control characters and caps byte length.
func SanitizeASCII(raw string, maxLen int) string {
	var b strings.Builder
	for i := 0; i < len(raw); i++ {
		if maxLen > 0 && b.Len() >= maxLen {
			break
		}
		c := raw[i]
		if c == '\r' || c == '\t' {
			b.WriteByte(' ')
			continue
		}
		if c == '\n' {
			b.WriteByte('\n')
			continue
		}
		if c >= 0x20 && c <= 0x7e {
			b.WriteByte(c)
		}
	}
	return strings.TrimSpace(b.String())
}

// CleanBundlePath validates one relative path inside a .gtpack archive.
func CleanBundlePath(raw string) (string, error) {
	trimmed := strings.ReplaceAll(strings.TrimSpace(raw), "\\", "/")
	if trimmed == "" || strings.HasPrefix(trimmed, "/") || strings.Contains(trimmed, "..") || strings.ContainsAny(trimmed, ":\x00") {
		return "", fmt.Errorf("invalid bundle path %q", raw)
	}
	for i := 0; i < len(trimmed); i++ {
		if trimmed[i] < 0x20 || trimmed[i] > 0x7e {
			return "", fmt.Errorf("invalid bundle path %q", raw)
		}
	}
	cleaned := path.Clean(trimmed)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.HasPrefix(cleaned, "/") {
		return "", fmt.Errorf("invalid bundle path %q", raw)
	}
	ext := strings.ToLower(path.Ext(cleaned))
	switch ext {
	case ".wasm", ".json", ".png", ".wav", ".ogg", ".glb":
	default:
		return "", fmt.Errorf("disallowed plugin asset extension %q", ext)
	}
	return cleaned, nil
}

func isAllowedStorageKeyChar(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '.' || c == '_' || c == ':' || c == ',' || c == '/' || c == '-'
}

// IsValidStorageKey reports whether key is safe for plugin KV / pref storage.
func IsValidStorageKey(key string, maxLen int) bool {
	if len(key) == 0 || len(key) > maxLen {
		return false
	}
	if strings.Contains(key, "..") || strings.HasPrefix(key, "/") || strings.HasSuffix(key, "/") {
		return false
	}
	for i := 0; i < len(key); i++ {
		if !isAllowedStorageKeyChar(key[i]) {
			return false
		}
	}
	return true
}

// IsValidStoragePrefix reports whether prefix is safe for plugin key listing.
func IsValidStoragePrefix(prefix string, maxLen int) bool {
	if len(prefix) > maxLen || strings.Contains(prefix, "..") || strings.HasPrefix(prefix, "/") {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		if !isAllowedStorageKeyChar(prefix[i]) {
			return false
		}
	}
	return true
}
