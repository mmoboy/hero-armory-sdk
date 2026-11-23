package proto

// ChannelBounty carries JSON state updates and player actions for Realm Bounty Board.
const ChannelBounty uint16 = 1

// ChannelSandbox carries File Explorer & Wasm Sandbox probe requests and results.
const ChannelSandbox uint16 = 2

// ClientAction is sent from client.wasm to server_plugin.wasm.
type ClientAction struct {
	Action string `json:"action"` // "sync", "reroll", "cheer"
	Note   string `json:"note,omitempty"`
}

// BountySnapshot is pushed from server_plugin.wasm to client.wasm for one hero.
type BountySnapshot struct {
	Kind           string `json:"kind"` // "explore", "chat", "cheer"
	Title          string `json:"title"`
	Description    string `json:"description"`
	Progress       int    `json:"progress"`
	Target         int    `json:"target"`
	Completed      bool   `json:"completed"`
	CompletedCount int    `json:"completed_count"`
	RealmCheers    int    `json:"realm_cheers"`
	OriginName     string `json:"origin_name,omitempty"`
	OathID         string `json:"oath_id,omitempty"`
	OathName       string `json:"oath_name,omitempty"`
	RankTitle      string `json:"rank_title"`
	LastEvent      string `json:"last_event"`
}
