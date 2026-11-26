package vanilla

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/mmoboy/hero-armory-sdk"
)

// HandleChatCommand handles classic Sanctuary chat commands (/help, /where,
// /sanctuary) for ServerPlugins implementing sdk.ChatCommandPlugin. Returns
// true when the message was handled as a command and should be consumed.
func HandleChatCommand(ev sdk.ChatEvent) bool {
	fields := strings.Fields(strings.TrimSpace(ev.Text))
	if len(fields) == 0 {
		return false
	}
	cmd := strings.ToLower(fields[0])
	switch cmd {
	case "/help":
		_ = sdk.SendChatToHero(ev.HeroID, "Sanctuary commands: /sanctuary (return to Courtyard 0,0), /where (location & stats), /difficulty <story|normal|veteran|heroic|mythic>, /help")
		return true
	case "/where":
		hero, found, err := sdk.GetHero(ev.HeroID)
		if err != nil || !found {
			_ = sdk.SendChatToHero(ev.HeroID, "Unable to read your hero location right now.")
			return true
		}
		_ = sdk.SendChatToHero(ev.HeroID, fmt.Sprintf("%s is at room (%d,%d) pos (%.1f, %.1f) - Lv.%d, %dg",
			hero.Name, hero.RoomX, hero.RoomY, hero.PosX, hero.PosZ, hero.Level, hero.Gold))
		return true
	case "/sanctuary":
		err := sdk.TeleportHero(ev.HeroID, 0, 0, [3]float32{0, 0, 2.2}, true)
		if errors.Is(err, sdk.ErrInBattle) {
			_ = sdk.SendChatToHero(ev.HeroID, "You cannot teleport to Sanctuary while in battle.")
			return true
		}
		if err != nil {
			_ = sdk.SendChatToHero(ev.HeroID, "Teleport to Sanctuary failed: "+err.Error())
			return true
		}
		sdk.SendNoticeToHero(ev.HeroID, "info", "Sanctuary Return", "Returned to Sanctuary Courtyard (0,0).")
		return true
	default:
		return false
	}
}
