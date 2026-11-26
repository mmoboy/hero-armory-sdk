package sdk

// HeroSnapshot is a sanitized snapshot of one hero inside the Realm.
type HeroSnapshot struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	OwnerAccountID string  `json:"owner_account_id,omitempty"`
	DisplayName    string  `json:"display_name,omitempty"`
	RoomX          int     `json:"room_x"`
	RoomY          int     `json:"room_y"`
	PosX           float32 `json:"pos_x"`
	PosZ           float32 `json:"pos_z"`
	Level          int     `json:"level,omitempty"`    // hero level (API level 2+)
	XP             int64   `json:"xp,omitempty"`       // XP toward the next level (API level 2+)
	Gold           int64   `json:"gold,omitempty"`     // gold balance (API level 2+)
	ClassID        string  `json:"class_id,omitempty"` // starter class picked at creation (API level 2+)
}

// RoomEvent reports a hero stepping into a realm room.
type RoomEvent struct {
	HeroID   string `json:"hero_id"`
	HeroName string `json:"hero_name"`
	RoomX    int    `json:"room_x"`
	RoomY    int    `json:"room_y"`
}

// ChatEvent reports an overworld chat message delivered to the server plugin.
type ChatEvent struct {
	HeroID   string `json:"hero_id"`
	HeroName string `json:"hero_name"`
	Channel  string `json:"channel"`
	Text     string `json:"text"`
}

// UIClickEvent delivers a client HUD button click or input submission to ClientHandler.
type UIClickEvent struct {
	ControlID string            `json:"control_id"`
	Value     string            `json:"value,omitempty"`
	Inputs    map[string]string `json:"inputs,omitempty"`
}

// HostServerPacketCmd instructs the Realm host to send an OpPluginPacket frame.
type HostServerPacketCmd struct {
	TargetHeroID string `json:"target_hero_id,omitempty"` // empty with Broadcast=true sends to room/all
	BroadcastAll bool   `json:"broadcast_all,omitempty"`
	RoomX        int    `json:"room_x,omitempty"`
	RoomY        int    `json:"room_y,omitempty"`
	BroadcastRm  bool   `json:"broadcast_room,omitempty"`
	Channel      uint16 `json:"channel"`
	Payload      []byte `json:"payload"`
}

// HostNoticeCmd instructs the Realm host to show a SystemNotice banner.
type HostNoticeCmd struct {
	TargetHeroID string `json:"target_hero_id,omitempty"`
	BroadcastAll bool   `json:"broadcast_all,omitempty"`
	Severity     string `json:"severity,omitempty"` // "info", "warn", "alert"
	Title        string `json:"title,omitempty"`
	Message      string `json:"message"`
}

// MonsterLoot defines one possible item drop rolled when a custom monster is defeated.
type MonsterLoot struct {
	ItemID string  `json:"item_id"`
	Chance float64 `json:"chance"` // 0.0..1.0
	MinQty int     `json:"min_qty"`
	MaxQty int     `json:"max_qty"`
}

// MonsterSpellSpec defines one active spell castable by a custom monster in tactical battle.
type MonsterSpellSpec struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	APCost        int        `json:"ap_cost,omitempty"`
	MinRange      float32    `json:"min_range,omitempty"`
	MaxRange      float32    `json:"max_range,omitempty"`
	AoERadius     float32    `json:"aoe_radius,omitempty"`
	CooldownTurns int        `json:"cooldown_turns,omitempty"`
	TargetKind    string     `json:"target_kind,omitempty"` // "enemy", "self", "ally"
	EffectKind    string     `json:"effect_kind,omitempty"` // "damage", "heal", "lifesteal", "push", "pull"
	Power         int        `json:"power,omitempty"`
	VFXKind       string     `json:"vfx_kind,omitempty"` // "slash", "bolt", "nova", "ring", "pillar"
	VFXColor      [4]float32 `json:"vfx_color,omitempty"`
}

// MonsterSpec registers a custom monster archetype on the Realm Server.
type MonsterSpec struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Level       int                `json:"level"`
	MaxHP       int                `json:"max_hp"`
	MovePoints  float32            `json:"move_points"`
	AttackPts   int                `json:"attack_pts"`
	Damage      int                `json:"damage"`
	Range       float32            `json:"range"`
	GoldReward  int64              `json:"gold_reward"`
	XPReward    int64              `json:"xp_reward"`
	BodyColor   [3]float32         `json:"body_color"`
	TrimColor   [3]float32         `json:"trim_color"`
	Scale       float32            `json:"scale"`
	ModelPath   string             `json:"model_path,omitempty"` // optional bundled .glb inside client_plugin.gtpack
	ModelID     string             `json:"model_id,omitempty"`   // optional client-registered CustomMeshSpec ID
	AIStyle     string             `json:"ai_style,omitempty"`   // "", "brawler", "caster", "kiter", "support"
	Spells      []MonsterSpellSpec `json:"spells,omitempty"`
	Description string             `json:"description,omitempty"`
	Habitat     string             `json:"habitat,omitempty"`
	LootTable   []MonsterLoot      `json:"loot_table,omitempty"`
}

// MonsterSpawnSpec places or respawns a monster pack inside a Realm room.
type MonsterSpawnSpec struct {
	ID          string  `json:"id"`
	ArchetypeID string  `json:"archetype_id"`
	RoomX       int     `json:"room_x"`
	RoomY       int     `json:"room_y"`
	PosX        float32 `json:"pos_x"`
	PosZ        float32 `json:"pos_z"`
	PackSize    int     `json:"pack_size"`
	RespawnSec  float32 `json:"respawn_sec,omitempty"` // 0 = default (30s), <0 = one-shot (no auto-respawn)
}

// BattleEndEvent notifies ServerPlugin (implementing BattleObserverPlugin) when a battle ends.
type BattleEndEvent struct {
	BattleID      string   `json:"battle_id"`
	RoomX         int      `json:"room_x"`
	RoomY         int      `json:"room_y"`
	Won           bool     `json:"won"`
	MonsterID     string   `json:"monster_id"` // the MonsterSpawnSpec.ID of the engaged pack
	ArchetypeID   string   `json:"archetype_id"`
	DefeatedCount int      `json:"defeated_count"`
	HeroIDs       []string `json:"hero_ids"`
	HeroNames     []string `json:"hero_names"`
}

// ItemSpec defines an equipment piece, crafting resource, or consumable.
type ItemSpec struct {
	ID           string     `json:"id"` // 2-48 lowercase ASCII [a-z0-9_-]
	Name         string     `json:"name"`
	Kind         string     `json:"kind"`                    // "equipment", "material", "consumable"
	Slot         string     `json:"slot,omitempty"`          // "hat", "amulet", "chest", "belt", "gloves", "ring", "boots", "weapon"
	VisualPreset string     `json:"visual_preset,omitempty"` // optional built-in item ID whose 3D model is reused
	ModelPath    string     `json:"model_path,omitempty"`    // optional bundled .glb inside client_plugin.gtpack
	ModelID      string     `json:"model_id,omitempty"`      // optional client-registered CustomMeshSpec ID
	ModelScale   float32    `json:"model_scale,omitempty"`
	MaxStack     int        `json:"max_stack,omitempty"` // default 1 for equipment, 99 for material/consumable
	Rarity       string     `json:"rarity,omitempty"`    // "Common", "Uncommon", "Rare", "Heroic", "Mythic", "Relic", "Royal"
	Element      string     `json:"element,omitempty"`
	Description  string     `json:"description,omitempty"`
	PrimaryCol   [4]float32 `json:"primary_col,omitempty"`
	AccentCol    [4]float32 `json:"accent_col,omitempty"`
	GlowCol      [4]float32 `json:"glow_col,omitempty"`
	Vitality     int        `json:"vitality,omitempty"`
	Strength     int        `json:"strength,omitempty"`
	Wisdom       int        `json:"wisdom,omitempty"`
	Agility      int        `json:"agility,omitempty"`
	Price        int64      `json:"price,omitempty"`         // merchant buy price in gold
	SellPrice    int64      `json:"sell_price,omitempty"`    // merchant sellback price in gold
	ShopHidden   bool       `json:"shop_hidden,omitempty"`   // exclude from merchant catalog
	ShopCategory string     `json:"shop_category,omitempty"` // "gear", "consumable", "material"
	SetID        string     `json:"set_id,omitempty"`        // optional item set group ID
}

// GatherNodeSpec places a gatherable tree, ore vein, herb, fishing shoal, or bench in a realm room.
type GatherNodeSpec struct {
	ID           string     `json:"id"`
	Job          string     `json:"job"` // "bucheron", "mineur", "herborist", "fisherman", "craftsman"
	Name         string     `json:"name"`
	RoomX        int        `json:"room_x"`
	RoomY        int        `json:"room_y"`
	PosX         float32    `json:"pos_x"`
	PosZ         float32    `json:"pos_z"`
	RewardItemID string     `json:"reward_item_id"`
	MinLevel     int        `json:"min_level,omitempty"`
	GatherSec    float32    `json:"gather_sec,omitempty"`
	XP           int64      `json:"xp,omitempty"`
	RespawnSec   float32    `json:"respawn_sec,omitempty"`
	VisualColor  [4]float32 `json:"visual_color,omitempty"`
	ModelPath    string     `json:"model_path,omitempty"` // optional bundled .glb inside client_plugin.gtpack
	ModelID      string     `json:"model_id,omitempty"`   // optional client-registered CustomMeshSpec ID
	ModelScale   float32    `json:"model_scale,omitempty"`
}

// CustomSubMeshSpec defines one procedural sub-mesh built in Go by a client plugin.
type CustomSubMeshSpec struct {
	Name              string       `json:"name,omitempty"`
	Positions         [][3]float32 `json:"positions"`
	Normals           [][3]float32 `json:"normals,omitempty"`
	Colors            [][4]float32 `json:"colors,omitempty"`
	UVs               [][2]float32 `json:"uvs,omitempty"`
	Indices           []uint32     `json:"indices"`
	Color             [4]float32   `json:"color,omitempty"`
	TexturePath       string       `json:"texture_path,omitempty"`
	Roughness         float32      `json:"roughness,omitempty"`
	Metallic          float32      `json:"metallic,omitempty"`
	EmissiveColor     [4]float32   `json:"emissive_color,omitempty"`
	EmissiveIntensity float32      `json:"emissive_intensity,omitempty"`
	DoubleSided       bool         `json:"double_sided,omitempty"`
	Offset            [3]float32   `json:"offset,omitempty"`
	Rotation          [3]float32   `json:"rotation,omitempty"`
	Scale             [3]float32   `json:"scale,omitempty"`
	AnimMode          string       `json:"anim_mode,omitempty"`   // "", "spin_y", "bob", "orbit"
	ActiveOnly        bool         `json:"active_only,omitempty"` // gatherables: hide when harvesting cooldown active
}

// CustomMeshSpec registers a reusable Go-generated 3D model under ID on the client (RoleClient).
type CustomMeshSpec struct {
	ID    string              `json:"id"`
	Parts []CustomSubMeshSpec `json:"parts"`
}

// RecipeIngredient defines one input item and quantity required by RecipeSpec.
type RecipeIngredient struct {
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
}

// RecipeSpec registers a crafting recipe in a realm's Jobs crafting menu.
type RecipeSpec struct {
	ID           string             `json:"id"`
	Job          string             `json:"job"`
	Name         string             `json:"name"`
	MinLevel     int                `json:"min_level,omitempty"`
	XP           int64              `json:"xp,omitempty"`
	OutputItemID string             `json:"output_item_id"`
	OutputQty    int                `json:"output_qty,omitempty"`
	Ingredients  []RecipeIngredient `json:"ingredients"`
	Description  string             `json:"description,omitempty"`
}

// GiveCurrencyCmd instructs the realm server to change one online hero's gold
// and/or grant XP. Gold may be negative (a charge); XP must be >= 0.
type GiveCurrencyCmd struct {
	TargetHeroID string `json:"target_hero_id"`
	Gold         int64  `json:"gold,omitempty"`
	XP           int64  `json:"xp,omitempty"`
}

// RemoveMonsterSpawnCmd removes a plugin-placed monster pack by MonsterSpawnSpec.ID.
type RemoveMonsterSpawnCmd struct {
	ID string `json:"id"`
}

// HostChatCmd sends a plugin chat line (shown on the "system" channel and
// tagged with the plugin's name) to one hero or to every online hero.
type HostChatCmd struct {
	TargetHeroID string `json:"target_hero_id,omitempty"`
	BroadcastAll bool   `json:"broadcast_all,omitempty"`
	Text         string `json:"text"`
}

// HeroQueryCmd asks the realm server for one online hero's snapshot.
type HeroQueryCmd struct {
	HeroID string `json:"hero_id"`
}

// GiveItemCmd instructs the realm server to grant items into one hero's bag.
type GiveItemCmd struct {
	TargetHeroID string `json:"target_hero_id"`
	ItemID       string `json:"item_id"`
	Quantity     int    `json:"quantity"`
}

// GatherEvent notifies ServerPlugin (implementing WorldObserverPlugin) on harvest completion.
type GatherEvent struct {
	HeroID   string `json:"hero_id"`
	HeroName string `json:"hero_name"`
	Job      string `json:"job"`
	NodeID   string `json:"node_id"`
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
}

// CraftEvent notifies ServerPlugin (implementing WorldObserverPlugin) on craft completion.
type CraftEvent struct {
	HeroID   string `json:"hero_id"`
	HeroName string `json:"hero_name"`
	Job      string `json:"job"`
	RecipeID string `json:"recipe_id"`
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
}

// CodexConfigSpec customizes the Bestiary & Resource Codex search UI and visibility rules.
type CodexConfigSpec struct {
	Title           string   `json:"title,omitempty"`
	BannerText      string   `json:"banner_text,omitempty"`
	DefaultCategory string   `json:"default_category,omitempty"`
	ExtraTabLabel   string   `json:"extra_tab_label,omitempty"`
	HideExactSpawns bool     `json:"hide_exact_spawns,omitempty"`
	HideDropRates   bool     `json:"hide_drop_rates,omitempty"`
	HiddenEntryIDs  []string `json:"hidden_entry_ids,omitempty"`
}

// CodexEntrySpec registers a custom searchable entry or realm guide in the Codex database.
type CodexEntrySpec struct {
	ID          string     `json:"id"`
	Category    string     `json:"category,omitempty"` // "bestiary", "resources", "items", "recipes", "custom"
	Title       string     `json:"title"`
	Subtitle    string     `json:"subtitle,omitempty"`
	Description string     `json:"description"`
	KindBadge   string     `json:"kind_badge,omitempty"`
	Location    string     `json:"location,omitempty"`
	RoomX       *int       `json:"room_x,omitempty"`
	RoomY       *int       `json:"room_y,omitempty"`
	IconItemID  string     `json:"icon_item_id,omitempty"`
	ItemIDs     []string   `json:"item_ids,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	AccentColor [4]float32 `json:"accent_color,omitempty"`
	SortOrder   int        `json:"sort_order,omitempty"`
}

// RoomTerrainOpSpec describes a deterministic walkable floor rectangle, ramp slope, or chasm void.
type RoomTerrainOpSpec struct {
	Op        string  `json:"op,omitempty"`      // "rect", "slope", "void"
	Kind      string  `json:"kind,omitempty"`    // alias for Op when Surface is omitted, or surface preset ("courtyard", "bridge", "dais", "ramp", "path", "water", "wing", "alcove")
	Surface   string  `json:"surface,omitempty"` // explicit surface kind ("courtyard", "bridge", "dais", "ramp", "path", "water", "wing", "alcove")
	MinX      float32 `json:"min_x"`
	MaxX      float32 `json:"max_x"`
	MinZ      float32 `json:"min_z"`
	MaxZ      float32 `json:"max_z"`
	RiseDir   string  `json:"rise_dir,omitempty"` // "north", "south", "east", "west"
	Elev      float32 `json:"elev,omitempty"`
	ElevEnd   float32 `json:"elev_end,omitempty"`
	Y         float32 `json:"y,omitempty"`
	MinY      float32 `json:"min_y,omitempty"`
	MaxY      float32 `json:"max_y,omitempty"`
	Tile      string  `json:"tile,omitempty"`       // "grass", "stone", "water", "wood", "sand", "snow", "mud"
	SlopeAxis string  `json:"slope_axis,omitempty"` // "x" or "z"
	Solid     bool    `json:"solid,omitempty"`
}

// RoomObstacleSpec defines a cylindrical collision obstacle inside a realm room.
type RoomObstacleSpec struct {
	Kind   string  `json:"kind,omitempty"`
	X      float32 `json:"x"`
	Z      float32 `json:"z"`
	Radius float32 `json:"radius"`
	Height float32 `json:"height"`
}

// RoomBoxSpec defines an axis-aligned 3D collision box inside a realm room.
type RoomBoxSpec struct {
	MinX float32 `json:"min_x"`
	MaxX float32 `json:"max_x"`
	MinY float32 `json:"min_y"`
	MaxY float32 `json:"max_y"`
	MinZ float32 `json:"min_z"`
	MaxZ float32 `json:"max_z"`
}

// RoomPortalSpec links one room edge (or custom landing pad) to any room coordinate.
type RoomPortalSpec struct {
	Edge         string     `json:"edge"` // "north", "south", "east", "west"
	TargetX      int        `json:"target_x"`
	TargetY      int        `json:"target_y"`
	TargetEdge   string     `json:"target_edge,omitempty"`
	UseTargetPos bool       `json:"use_target_pos,omitempty"`
	TargetPos    [3]float32 `json:"target_pos,omitempty"`
	PadPos       [3]float32 `json:"pad_pos"`
	Label        string     `json:"label,omitempty"`
}

// RoomHouseSpec places a walk-in cottage/building or plot inside a room.
type RoomHouseSpec struct {
	ID        string     `json:"id"`
	Name      string     `json:"name,omitempty"`
	X         float32    `json:"x"`
	Z         float32    `json:"z"`
	W         float32    `json:"w"`
	D         float32    `json:"d"`
	H         float32    `json:"h"`
	DoorDir   string     `json:"door_dir,omitempty"`
	Color     [3]float32 `json:"color,omitempty"`
	RoofColor [3]float32 `json:"roof_color,omitempty"`
}

// RoomInteractableSpec places an interactive prop (portal stone, shrine, chest, anvil) in a room.
type RoomInteractableSpec struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Label       string  `json:"label,omitempty"`
	Description string  `json:"description,omitempty"`
	RewardItem  string  `json:"reward_item,omitempty"`
	X           float32 `json:"x"`
	Y           float32 `json:"y,omitempty"`
	Z           float32 `json:"z"`
	Radius      float32 `json:"radius"`
	Kind        string  `json:"kind"`
	Target      string  `json:"target,omitempty"`
	Notice      string  `json:"notice,omitempty"`
}

// RoomPropSpec places a decorative or solid 3D prop (built-in kind, .glb, or CustomMeshSpec).
type RoomPropSpec struct {
	ID          string     `json:"id,omitempty"`
	Kind        string     `json:"kind,omitempty"`
	X           float32    `json:"x"`
	Y           float32    `json:"y"`
	Z           float32    `json:"z"`
	Scale       float32    `json:"scale,omitempty"`
	Yaw         float32    `json:"yaw,omitempty"`
	Color       [4]float32 `json:"color,omitempty"`
	SolidRadius float32    `json:"solid_radius,omitempty"`
	ModelPath   string     `json:"model_path,omitempty"`
	ModelID     string     `json:"model_id,omitempty"`
}

// RoomSpec completely defines one room chunk in the realm map.
type RoomSpec struct {
	X                int                    `json:"x"`
	Y                int                    `json:"y"`
	Name             string                 `json:"name"`
	Zone             string                 `json:"zone,omitempty"`
	Biome            string                 `json:"biome"`
	LightingPreset   string                 `json:"lighting_preset,omitempty"`
	Ambience         string                 `json:"ambience,omitempty"`
	Description      string                 `json:"description,omitempty"`
	HalfSize         float32                `json:"half_size,omitempty"`
	SafeZone         bool                   `json:"safe_zone,omitempty"`
	HousingZone      bool                   `json:"housing_zone,omitempty"`
	DisableAutoProps bool                   `json:"disable_auto_props,omitempty"`
	FloorColor       [3]float32             `json:"floor_color,omitempty"`
	WallColor        [3]float32             `json:"wall_color,omitempty"`
	AccentColor      [3]float32             `json:"accent_color,omitempty"`
	SpawnPos         [3]float32             `json:"spawn_pos,omitempty"`
	TerrainOps       []RoomTerrainOpSpec    `json:"terrain_ops,omitempty"`
	Obstacles        []RoomObstacleSpec     `json:"obstacles,omitempty"`
	Boxes            []RoomBoxSpec          `json:"boxes,omitempty"`
	Portals          []RoomPortalSpec       `json:"portals,omitempty"`
	Houses           []RoomHouseSpec        `json:"houses,omitempty"`
	Interactables    []RoomInteractableSpec `json:"interactables,omitempty"`
	Props            []RoomPropSpec         `json:"props,omitempty"`
}

// RoomCoordCmd addresses a single room coordinate for removal or targeting.
type RoomCoordCmd struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// SpawnRoomSpec sets the realm's initial spawn room and local spawn position.
type SpawnRoomSpec struct {
	X   int        `json:"x"`
	Y   int        `json:"y"`
	Pos [3]float32 `json:"pos,omitempty"`
}

// NPCDialogOptionSpec defines one selectable choice in an NPC conversation.
type NPCDialogOptionSpec struct {
	ID           string `json:"id,omitempty"`
	Label        string `json:"label"`
	Action       string `json:"action,omitempty"`
	Response     string `json:"response"`
	BlessingKind string `json:"blessing_kind,omitempty"`
	OpenMerchant bool   `json:"open_merchant,omitempty"`
}

// NPCSpec registers an NPC in a realm room with custom equipment, appearance, dialog, and optional shop.
type NPCSpec struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Title       string                `json:"title,omitempty"`
	Role        string                `json:"role,omitempty"`
	RoomX       int                   `json:"room_x"`
	RoomY       int                   `json:"room_y"`
	PosX        float32               `json:"pos_x"`
	PosY        float32               `json:"pos_y,omitempty"`
	PosZ        float32               `json:"pos_z"`
	FacingYaw   float32               `json:"facing_yaw,omitempty"`
	Pose        string                `json:"pose,omitempty"`
	SkinRGB     [3]float32            `json:"skin_rgb,omitempty"`
	HairColor   [3]float32            `json:"hair_color,omitempty"`
	EyeColor    [3]float32            `json:"eye_color,omitempty"`
	AccentRGB   [3]float32            `json:"accent_rgb,omitempty"`
	HairStyle   int                   `json:"hair_style,omitempty"`
	FacialHair  int                   `json:"facial_hair,omitempty"`
	Outfit      map[string]string     `json:"outfit,omitempty"`
	Greeting    string                `json:"greeting"`
	Options     []NPCDialogOptionSpec `json:"options,omitempty"`
	ShopEnabled bool                  `json:"shop_enabled,omitempty"`
}

// NPCInteractEvent delivers an NPC interaction (OptionIndex == -1 for opening dialog) to NPCPlugin.
type NPCInteractEvent struct {
	HeroID        string `json:"hero_id"`
	HeroName      string `json:"hero_name"`
	NPCID         string `json:"npc_id"`
	OptionIndex   int    `json:"option_index"`
	RoomX         int    `json:"room_x"`
	RoomY         int    `json:"room_y"`
	RoomName      string `json:"room_name,omitempty"`
	RoomOccupants int    `json:"room_occupants,omitempty"`
	ActiveBots    int    `json:"active_bots,omitempty"`
	OnlineHeroes  int    `json:"online_heroes,omitempty"`
}

// NPCDialogResult allows NPCPlugin to override an NPC's greeting, options, reply, or open the shop UI.
type NPCDialogResult struct {
	Greeting     string                `json:"greeting,omitempty"`
	ReplyText    string                `json:"reply_text,omitempty"`
	BannerText   string                `json:"banner_text,omitempty"`
	Options      []NPCDialogOptionSpec `json:"options,omitempty"`
	OpenMerchant bool                  `json:"open_merchant,omitempty"`
	ShopEnabled  bool                  `json:"shop_enabled,omitempty"`
}

// SetHeroLoadoutCmd lets the server plugin equip a loadout or update a hero's pose immediately.
type SetHeroLoadoutCmd struct {
	TargetHeroID string            `json:"target_hero_id"`
	Equipped     map[string]string `json:"equipped,omitempty"`
	Pose         string            `json:"pose,omitempty"`
	BannerText   string            `json:"banner_text,omitempty"`
}

// InteractEvent delivers a room interactable click (Option E / click) to InteractPlugin.
type InteractEvent struct {
	HeroID         string `json:"hero_id"`
	HeroName       string `json:"hero_name"`
	InteractableID string `json:"interactable_id"`
	Kind           string `json:"kind,omitempty"`
	Target         string `json:"target,omitempty"`
	RoomX          int    `json:"room_x"`
	RoomY          int    `json:"room_y"`
}

// ItemStackSpec specifies an item quantity pair for starter inventory setup.
type ItemStackSpec struct {
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
}

// JobDefSpec defines or customizes metadata for one gathering/crafting profession.
type JobDefSpec struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ToolLabel   string `json:"tool_label,omitempty"`
	ResourceID  string `json:"resource_id,omitempty"`
	Description string `json:"description,omitempty"`
}

// StarterConfigSpec configures the realm's default spawn, starter gold, equipped loadout, bag items, and base stats.
type StarterConfigSpec struct {
	SpawnRoomX          int               `json:"spawn_room_x,omitempty"`
	SpawnRoomY          int               `json:"spawn_room_y,omitempty"`
	SpawnPos            [3]float32        `json:"spawn_pos,omitempty"`
	StarterGold         int64             `json:"starter_gold,omitempty"`
	StarterEquipped     map[string]string `json:"starter_equipped,omitempty"`
	StarterBag          []ItemStackSpec   `json:"starter_bag,omitempty"`
	StarterTailBag      []ItemStackSpec   `json:"starter_tail_bag,omitempty"`
	IncludeAllEquip     bool              `json:"include_all_equip,omitempty"`
	IncludePortalStyles bool              `json:"include_portal_styles,omitempty"`
	BaseHP              int               `json:"base_hp,omitempty"`
	BaseAP              int               `json:"base_ap,omitempty"`
	BaseMP              float32           `json:"base_mp,omitempty"`
	DisableHousing      bool              `json:"disable_housing,omitempty"`
}

// PlayerSkillSpec registers a tactical combat skill usable by players in turn-based battles.
type PlayerSkillSpec struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	ShortLabel    string     `json:"short_label,omitempty"`
	IconName      string     `json:"icon_name,omitempty"`
	Description   string     `json:"description,omitempty"`
	Pose          string     `json:"pose,omitempty"`
	APCost        int        `json:"ap_cost"`
	CooldownTurns int        `json:"cooldown_turns,omitempty"`
	MinRange      float32    `json:"min_range,omitempty"`
	Range         float32    `json:"range"`
	AoERadius     float32    `json:"aoe_radius,omitempty"`
	BaseDamage    int        `json:"base_damage"`
	ScalingStat   string     `json:"scaling_stat,omitempty"` // "strength", "wisdom", "agility"
	TargetKind    string     `json:"target_kind,omitempty"`  // "enemy", "self", "ally"
	EffectKind    string     `json:"effect_kind,omitempty"`  // "damage", "heal", "lifesteal", "push", "pull"
	VFXKind       string     `json:"vfx_kind,omitempty"`     // "slash", "bolt", "nova", "ring", "pillar"
	VFXColor      [4]float32 `json:"vfx_color,omitempty"`
	SortOrder     int        `json:"sort_order,omitempty"`
}

// Character creation mode identifiers returned by Realm server plugins.
const (
	CharCreateModeVanilla         = "vanilla"
	CharCreateModeVanillaExtended = "vanilla_extended"
	CharCreateModePluginCustom    = "plugin_custom"
)

// HeroAppearanceSpec represents customizable Vanilla body, face, hair, and eye fields.
type HeroAppearanceSpec struct {
	IrisColor        [4]float32 `json:"iris_color,omitempty"`
	IrisColorR       [4]float32 `json:"iris_color_r,omitempty"`
	Heterochromia    bool       `json:"heterochromia,omitempty"`
	HairColor        [4]float32 `json:"hair_color,omitempty"`
	HairHighlight    [4]float32 `json:"hair_highlight,omitempty"`
	HairSheen        float32    `json:"hair_sheen,omitempty"`
	HairStyle        int        `json:"hair_style,omitempty"`
	EarStyle         int        `json:"ear_style,omitempty"`
	EarTint          float32    `json:"ear_tint,omitempty"`
	HeadStyle        int        `json:"head_style,omitempty"`
	HeadScale        float32    `json:"head_scale,omitempty"`
	HeadWidth        float32    `json:"head_width,omitempty"`
	HeadHeight       float32    `json:"head_height,omitempty"`
	HeadDepth        float32    `json:"head_depth,omitempty"`
	JawContour       float32    `json:"jaw_contour,omitempty"`
	SkinColor        [4]float32 `json:"skin_color,omitempty"`
	BrowColor        [4]float32 `json:"brow_color,omitempty"`
	MatchHairBrow    bool       `json:"match_hair_brow,omitempty"`
	BlushColor       [4]float32 `json:"blush_color,omitempty"`
	MarkingColor     [4]float32 `json:"marking_color,omitempty"`
	LipColor         [4]float32 `json:"lip_color,omitempty"`
	LashColor        [4]float32 `json:"lash_color,omitempty"`
	ScleraColor      [4]float32 `json:"sclera_color,omitempty"`
	TunicColor       [4]float32 `json:"tunic_color,omitempty"`
	PantsColor       [4]float32 `json:"pants_color,omitempty"`
	PupilScale       float32    `json:"pupil_scale,omitempty"`
	PupilShape       int        `json:"pupil_shape,omitempty"`
	LashStyle        int        `json:"lash_style,omitempty"`
	BrowStyle        int        `json:"brow_style,omitempty"`
	IrisSize         float32    `json:"iris_size,omitempty"`
	EyeGlow          float32    `json:"eye_glow,omitempty"`
	CatchlightSize   float32    `json:"catchlight_size,omitempty"`
	BrowThickScale   float32    `json:"brow_thick_scale,omitempty"`
	BrowHeightOffset float32    `json:"brow_height_offset,omitempty"`
	MouthSize        float32    `json:"mouth_size,omitempty"`
	MouthStyle       int        `json:"mouth_style,omitempty"`
	BlushLevel       float32    `json:"blush_level,omitempty"`
	NoseTint         float32    `json:"nose_tint,omitempty"`
	Freckles         float32    `json:"freckles,omitempty"`
	FaceMarking      int        `json:"face_marking,omitempty"`
	CustomModelPath  string     `json:"custom_model_path,omitempty"`
	CustomModelID    string     `json:"custom_model_id,omitempty"`
	CustomModelScale float32    `json:"custom_model_scale,omitempty"`
}

// StarterClassSpec defines one realm-offered starter class or origin option.
type StarterClassSpec struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Icon        string            `json:"icon,omitempty"`
	Equipped    map[string]string `json:"equipped,omitempty"`
	ModelPath   string            `json:"model_path,omitempty"`
	ModelID     string            `json:"model_id,omitempty"`
	ModelScale  float32           `json:"model_scale,omitempty"`
	BonusHP     int               `json:"bonus_hp,omitempty"`
	BonusAP     int               `json:"bonus_ap,omitempty"`
	BonusMP     int               `json:"bonus_mp,omitempty"`
}

// CharacterCreateConfig configures Realm-authoritative character creation.
type CharacterCreateConfig struct {
	Mode          string             `json:"mode,omitempty"`
	Title         string             `json:"title,omitempty"`
	Subtitle      string             `json:"subtitle,omitempty"`
	Classes       []StarterClassSpec `json:"classes,omitempty"`
	HideBodyTab   bool               `json:"hide_body_tab,omitempty"`
	HideFaceTab   bool               `json:"hide_face_tab,omitempty"`
	HideEyesTab   bool               `json:"hide_eyes_tab,omitempty"`
	HideNameInput bool               `json:"hide_name_input,omitempty"`
	CustomData    []byte             `json:"custom_data,omitempty"`
}

// BeginCharacterCreateEvent notifies ServerPlugin when a player begins hero creation.
type BeginCharacterCreateEvent struct {
	AccountID   string `json:"account_id"`
	DisplayName string `json:"display_name,omitempty"`
}

// FinalizeCharacterCreateEvent asks ServerPlugin to validate and finalize hero creation.
type FinalizeCharacterCreateEvent struct {
	AccountID   string             `json:"account_id"`
	DisplayName string             `json:"display_name,omitempty"`
	HeroName    string             `json:"hero_name"`
	ClassID     string             `json:"class_id,omitempty"`
	CustomData  []byte             `json:"custom_data,omitempty"`
	Appearance  HeroAppearanceSpec `json:"appearance"`
}

// CharacterCreateResult specifies authoritative starter gear, bag items, gold,
// stats, spawn point, and optional KV writes for a newly created hero.
type CharacterCreateResult struct {
	RejectMessage         string             `json:"reject_message,omitempty"`
	HeroName              string             `json:"hero_name,omitempty"`
	AppearanceOverride    HeroAppearanceSpec `json:"appearance_override,omitempty"`
	UseAppearanceOverride bool               `json:"use_appearance_override,omitempty"`
	Equipped              map[string]string  `json:"equipped,omitempty"`
	OverrideEquipped      bool               `json:"override_equipped,omitempty"`
	StarterBag            []ItemStackSpec    `json:"starter_bag,omitempty"`
	OverrideStarterBag    bool               `json:"override_starter_bag,omitempty"`
	StarterGold           int64              `json:"starter_gold,omitempty"`
	OverrideStarterGold   bool               `json:"override_starter_gold,omitempty"`
	BaseHP                int                `json:"base_hp,omitempty"`
	BaseAP                int                `json:"base_ap,omitempty"`
	BaseMP                int                `json:"base_mp,omitempty"`
	SpawnRoomX            int                `json:"spawn_room_x,omitempty"`
	SpawnRoomY            int                `json:"spawn_room_y,omitempty"`
	SpawnPos              [3]float32         `json:"spawn_pos,omitempty"`
	OverrideSpawn         bool               `json:"override_spawn,omitempty"`
	CustomKV              map[string][]byte  `json:"custom_kv,omitempty"`
	WelcomeNotice         string             `json:"welcome_notice,omitempty"`
}

// CharacterCreateContext is delivered to ClientPlugin.OnEnterCharacterCreate.
type CharacterCreateContext struct {
	ServerID    string                `json:"server_id"`
	ServerName  string                `json:"server_name"`
	Mode        string                `json:"mode"`
	Config      CharacterCreateConfig `json:"config"`
	AccountID   string                `json:"account_id,omitempty"`
	DisplayName string                `json:"display_name,omitempty"`
}

// CharacterCreatePreviewSpec updates the 3D Sanctuary pedestal during character creation.
type CharacterCreatePreviewSpec struct {
	Equipped      map[string]string  `json:"equipped,omitempty"`
	ApplyEquipped bool               `json:"apply_equipped,omitempty"`
	Appearance    HeroAppearanceSpec `json:"appearance,omitempty"`
	UseAppearance bool               `json:"use_appearance,omitempty"`
	ModelPath     string             `json:"model_path,omitempty"`
	ModelID       string             `json:"model_id,omitempty"`
	ModelScale    float32            `json:"model_scale,omitempty"`
	UseModel      bool               `json:"use_model,omitempty"`
	Pose          string             `json:"pose,omitempty"`
	CameraFocus   string             `json:"camera_focus,omitempty"`
}

// CharacterCreateSubmission submits character creation from a custom ClientPlugin screen.
type CharacterCreateSubmission struct {
	Name             string `json:"name"`
	ClassID          string `json:"class_id,omitempty"`
	CustomData       []byte `json:"custom_data,omitempty"`
	EnterImmediately bool   `json:"enter_immediately,omitempty"`
}
