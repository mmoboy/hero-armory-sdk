//go:build wasip1

package sdk

import (
	"encoding/json"
	"runtime"
	"sync"
	"unsafe"
)

//go:wasmimport hero_armory host_log
func hostLog(level, ptr, size uint32)

//go:wasmimport hero_armory host_send_packet
func hostSendPacket(channel, ptr, size uint32) int32

//go:wasmimport hero_armory host_server_send_packet
func hostServerSendPacket(ptr, size uint32) int32

//go:wasmimport hero_armory host_server_send_notice
func hostServerSendNotice(ptr, size uint32) int32

//go:wasmimport hero_armory host_server_register_monster
func hostServerRegisterMonster(ptr, size uint32) int32

//go:wasmimport hero_armory host_server_spawn_monster
func hostServerSpawnMonster(ptr, size uint32) int32

//go:wasmimport hero_armory host_server_register_item
func hostServerRegisterItem(ptr, size uint32) int32

//go:wasmimport hero_armory host_server_spawn_gather_node
func hostServerSpawnGatherNode(ptr, size uint32) int32

//go:wasmimport hero_armory host_server_register_recipe
func hostServerRegisterRecipe(ptr, size uint32) int32

//go:wasmimport hero_armory host_server_give_item
func hostServerGiveItem(ptr, size uint32) int32

//go:wasmimport hero_armory host_server_configure_codex
func hostServerConfigureCodex(ptr, size uint32) int32

//go:wasmimport hero_armory host_server_register_codex_entry
func hostServerRegisterCodexEntry(ptr, size uint32) int32

//go:wasmimport hero_armory host_ui_set_hud_actions
func hostUISetHUDActions(ptr, size uint32) int32

//go:wasmimport hero_armory host_ui_set_overlay
func hostUISetOverlay(ptr, size uint32) int32

//go:wasmimport hero_armory host_ui_open_codex
func hostUIOpenCodex(ptr, size uint32) int32

//go:wasmimport hero_armory host_play_audio
func hostPlayAudio(pathPtr, pathSize, volMilli uint32) int32

//go:wasmimport hero_armory host_server_kv_set
func hostServerKVSet(keyPtr, keyLen, valPtr, valLen uint32) int32

//go:wasmimport hero_armory host_server_kv_get
func hostServerKVGet(keyPtr, keyLen, outPtr, outCap uint32) int32

//go:wasmimport hero_armory host_server_kv_delete
func hostServerKVDelete(keyPtr, keyLen uint32) int32

//go:wasmimport hero_armory host_server_kv_list
func hostServerKVList(prefixPtr, prefixLen, limit, outPtr, outCap uint32) int32

//go:wasmimport hero_armory host_pref_set
func hostPrefSet(keyPtr, keyLen, valPtr, valLen uint32) int32

//go:wasmimport hero_armory host_pref_get
func hostPrefGet(keyPtr, keyLen, outPtr, outCap uint32) int32

//go:wasmimport hero_armory host_pref_delete
func hostPrefDelete(keyPtr, keyLen uint32) int32

//go:wasmimport hero_armory host_pref_list
func hostPrefList(prefixPtr, prefixLen, limit, outPtr, outCap uint32) int32

//go:wasmimport hero_armory host_ui_register_mesh
func hostUIRegisterMesh(ptr, size uint32) int32

//go:wasmimport hero_armory host_server_register_room
func hostServerRegisterRoom(ptr, size uint32) int32

//go:wasmimport hero_armory host_server_remove_room
func hostServerRemoveRoom(ptr, size uint32) int32

//go:wasmimport hero_armory host_server_set_spawn_room
func hostServerSetSpawnRoom(ptr, size uint32) int32

//go:wasmimport hero_armory host_server_register_npc
func hostServerRegisterNPC(ptr, size uint32) int32

//go:wasmimport hero_armory host_server_set_hero_loadout
func hostServerSetHeroLoadout(ptr, size uint32) int32

//go:wasmimport hero_armory host_server_set_starter_config
func hostServerSetStarterConfig(ptr, size uint32) int32

//go:wasmimport hero_armory host_server_register_player_skill
func hostServerRegisterPlayerSkill(ptr, size uint32) int32

//go:wasmimport hero_armory host_server_register_job
func hostServerRegisterJob(ptr, size uint32) int32

//go:wasmimport hero_armory host_server_set_npc_result
func hostServerSetNPCResult(ptr, size uint32) int32

//go:wasmimport hero_armory host_ui_set_create_preview
func hostUISetCreatePreview(ptr, size uint32) int32

//go:wasmimport hero_armory host_ui_submit_char_create
func hostUISubmitCharCreate(ptr, size uint32) int32

//go:wasmimport hero_armory host_server_set_char_create_config
func hostServerSetCharCreateConfig(ptr, size uint32) int32

//go:wasmimport hero_armory host_server_set_char_create_result
func hostServerSetCharCreateResult(ptr, size uint32) int32

var (
	allocMu  sync.Mutex
	allocMap = make(map[uint32][]byte)
)

//go:wasmexport plugin_alloc
func pluginAlloc(size uint32) uint32 {
	if size == 0 || size > MaxUISpecBytes*2 {
		return 0
	}
	buf := make([]byte, size)
	ptr := uint32(uintptr(unsafe.Pointer(&buf[0])))
	allocMu.Lock()
	allocMap[ptr] = buf
	allocMu.Unlock()
	return ptr
}

//go:wasmexport plugin_free
func pluginFree(ptr uint32) {
	if ptr == 0 {
		return
	}
	allocMu.Lock()
	delete(allocMap, ptr)
	allocMu.Unlock()
}

func readGuestSlice(ptr, size uint32) []byte {
	if ptr == 0 || size == 0 {
		return nil
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(uintptr(ptr))), int(size))
}

func bytesPtrLen(b []byte) (uint32, uint32) {
	if len(b) == 0 {
		return 0, 0
	}
	return uint32(uintptr(unsafe.Pointer(&b[0]))), uint32(len(b))
}

func abiLog(level uint32, msg string) {
	b := []byte(msg)
	ptr, sz := bytesPtrLen(b)
	hostLog(level, ptr, sz)
	runtime.KeepAlive(b)
}

func abiSendToServer(channel uint16, payload []byte) int32 {
	ptr, sz := bytesPtrLen(payload)
	res := hostSendPacket(uint32(channel), ptr, sz)
	runtime.KeepAlive(payload)
	return res
}

func abiServerSendPacket(raw []byte) int32 {
	ptr, sz := bytesPtrLen(raw)
	res := hostServerSendPacket(ptr, sz)
	runtime.KeepAlive(raw)
	return res
}

func abiServerSendNotice(raw []byte) int32 {
	ptr, sz := bytesPtrLen(raw)
	res := hostServerSendNotice(ptr, sz)
	runtime.KeepAlive(raw)
	return res
}

func abiServerRegisterMonster(raw []byte) int32 {
	ptr, sz := bytesPtrLen(raw)
	res := hostServerRegisterMonster(ptr, sz)
	runtime.KeepAlive(raw)
	return res
}

func abiServerSpawnMonster(raw []byte) int32 {
	ptr, sz := bytesPtrLen(raw)
	res := hostServerSpawnMonster(ptr, sz)
	runtime.KeepAlive(raw)
	return res
}

func abiServerRegisterItem(raw []byte) int32 {
	ptr, sz := bytesPtrLen(raw)
	res := hostServerRegisterItem(ptr, sz)
	runtime.KeepAlive(raw)
	return res
}

func abiServerSpawnGatherNode(raw []byte) int32 {
	ptr, sz := bytesPtrLen(raw)
	res := hostServerSpawnGatherNode(ptr, sz)
	runtime.KeepAlive(raw)
	return res
}

func abiServerRegisterRecipe(raw []byte) int32 {
	ptr, sz := bytesPtrLen(raw)
	res := hostServerRegisterRecipe(ptr, sz)
	runtime.KeepAlive(raw)
	return res
}

func abiServerGiveItem(raw []byte) int32 {
	ptr, sz := bytesPtrLen(raw)
	res := hostServerGiveItem(ptr, sz)
	runtime.KeepAlive(raw)
	return res
}

func abiServerConfigureCodex(raw []byte) int32 {
	ptr, sz := bytesPtrLen(raw)
	res := hostServerConfigureCodex(ptr, sz)
	runtime.KeepAlive(raw)
	return res
}

func abiServerRegisterCodexEntry(raw []byte) int32 {
	ptr, sz := bytesPtrLen(raw)
	res := hostServerRegisterCodexEntry(ptr, sz)
	runtime.KeepAlive(raw)
	return res
}

func abiSetHUDActions(raw []byte) int32 {
	ptr, sz := bytesPtrLen(raw)
	res := hostUISetHUDActions(ptr, sz)
	runtime.KeepAlive(raw)
	return res
}

func abiSetOverlay(raw []byte) int32 {
	ptr, sz := bytesPtrLen(raw)
	res := hostUISetOverlay(ptr, sz)
	runtime.KeepAlive(raw)
	return res
}

func abiOpenCodex(query string) int32 {
	b := []byte(query)
	ptr, sz := bytesPtrLen(b)
	res := hostUIOpenCodex(ptr, sz)
	runtime.KeepAlive(b)
	return res
}

func abiPlayAudio(assetPath string, volMilli uint32) int32 {
	b := []byte(assetPath)
	ptr, sz := bytesPtrLen(b)
	res := hostPlayAudio(ptr, sz, volMilli)
	runtime.KeepAlive(b)
	return res
}

func callSnprintf(fn func(outPtr, outCap uint32) int32, notFoundCode bool, maxSize int) ([]byte, bool, int32) {
	capHint := 4096
	if maxSize > 0 && capHint > maxSize {
		capHint = maxSize
	}
	for attempt := 0; attempt < 3; attempt++ {
		if capHint == 0 {
			capHint = 1
		}
		buf := make([]byte, capHint)
		outPtr, outCap := bytesPtrLen(buf)
		n := fn(outPtr, outCap)
		runtime.KeepAlive(buf)
		if notFoundCode && n == -1 {
			return nil, false, 0
		}
		if n < 0 {
			return nil, false, n
		}
		if n <= int32(outCap) {
			cp := make([]byte, n)
			copy(cp, buf[:n])
			return cp, true, 0
		}
		if maxSize > 0 && int(n) > maxSize {
			return nil, false, -5
		}
		capHint = int(n)
	}
	return nil, false, -5
}

func abiServerKVSet(key string, val []byte) int32 {
	kb := []byte(key)
	kp, kl := bytesPtrLen(kb)
	vp, vl := bytesPtrLen(val)
	res := hostServerKVSet(kp, kl, vp, vl)
	runtime.KeepAlive(kb)
	runtime.KeepAlive(val)
	return res
}

func abiServerKVGet(key string) ([]byte, bool, int32) {
	kb := []byte(key)
	kp, kl := bytesPtrLen(kb)
	defer runtime.KeepAlive(kb)
	return callSnprintf(func(outPtr, outCap uint32) int32 {
		return hostServerKVGet(kp, kl, outPtr, outCap)
	}, true, MaxKVValueBytes)
}

func abiServerKVDelete(key string) int32 {
	kb := []byte(key)
	kp, kl := bytesPtrLen(kb)
	res := hostServerKVDelete(kp, kl)
	runtime.KeepAlive(kb)
	return res
}

func abiServerKVList(prefix string, limit int) ([]byte, int32) {
	pb := []byte(prefix)
	pp, pl := bytesPtrLen(pb)
	defer runtime.KeepAlive(pb)
	raw, _, code := callSnprintf(func(outPtr, outCap uint32) int32 {
		return hostServerKVList(pp, pl, uint32(limit), outPtr, outCap)
	}, false, MaxKVValueBytes)
	return raw, code
}

func abiPrefSet(key, val string) int32 {
	kb := []byte(key)
	vb := []byte(val)
	kp, kl := bytesPtrLen(kb)
	vp, vl := bytesPtrLen(vb)
	res := hostPrefSet(kp, kl, vp, vl)
	runtime.KeepAlive(kb)
	runtime.KeepAlive(vb)
	return res
}

func abiPrefGet(key string) ([]byte, bool, int32) {
	kb := []byte(key)
	kp, kl := bytesPtrLen(kb)
	defer runtime.KeepAlive(kb)
	return callSnprintf(func(outPtr, outCap uint32) int32 {
		return hostPrefGet(kp, kl, outPtr, outCap)
	}, true, MaxPrefValueBytes)
}

func abiPrefDelete(key string) int32 {
	kb := []byte(key)
	kp, kl := bytesPtrLen(kb)
	res := hostPrefDelete(kp, kl)
	runtime.KeepAlive(kb)
	return res
}

func abiPrefList(prefix string, limit int) ([]byte, int32) {
	pb := []byte(prefix)
	pp, pl := bytesPtrLen(pb)
	defer runtime.KeepAlive(pb)
	raw, _, code := callSnprintf(func(outPtr, outCap uint32) int32 {
		return hostPrefList(pp, pl, uint32(limit), outPtr, outCap)
	}, false, MaxPrefValueBytes*2)
	return raw, code
}

func abiRegisterCustomMesh(raw []byte) int32 {
	ptr, sz := bytesPtrLen(raw)
	res := hostUIRegisterMesh(ptr, sz)
	runtime.KeepAlive(raw)
	return res
}

func abiServerRegisterRoom(raw []byte) int32 {
	ptr, sz := bytesPtrLen(raw)
	res := hostServerRegisterRoom(ptr, sz)
	runtime.KeepAlive(raw)
	return res
}

func abiServerRemoveRoom(raw []byte) int32 {
	ptr, sz := bytesPtrLen(raw)
	res := hostServerRemoveRoom(ptr, sz)
	runtime.KeepAlive(raw)
	return res
}

func abiServerSetSpawnRoom(raw []byte) int32 {
	ptr, sz := bytesPtrLen(raw)
	res := hostServerSetSpawnRoom(ptr, sz)
	runtime.KeepAlive(raw)
	return res
}

func abiServerRegisterNPC(raw []byte) int32 {
	ptr, sz := bytesPtrLen(raw)
	res := hostServerRegisterNPC(ptr, sz)
	runtime.KeepAlive(raw)
	return res
}

func abiServerSetHeroLoadout(raw []byte) int32 {
	ptr, sz := bytesPtrLen(raw)
	res := hostServerSetHeroLoadout(ptr, sz)
	runtime.KeepAlive(raw)
	return res
}

func abiServerSetStarterConfig(raw []byte) int32 {
	ptr, sz := bytesPtrLen(raw)
	res := hostServerSetStarterConfig(ptr, sz)
	runtime.KeepAlive(raw)
	return res
}

func abiServerRegisterPlayerSkill(raw []byte) int32 {
	ptr, sz := bytesPtrLen(raw)
	res := hostServerRegisterPlayerSkill(ptr, sz)
	runtime.KeepAlive(raw)
	return res
}

func abiServerRegisterJob(raw []byte) int32 {
	ptr, sz := bytesPtrLen(raw)
	res := hostServerRegisterJob(ptr, sz)
	runtime.KeepAlive(raw)
	return res
}

func abiServerSetNPCResult(raw []byte) int32 {
	ptr, sz := bytesPtrLen(raw)
	res := hostServerSetNPCResult(ptr, sz)
	runtime.KeepAlive(raw)
	return res
}

func abiSetCreatePreview(raw []byte) int32 {
	ptr, sz := bytesPtrLen(raw)
	res := hostUISetCreatePreview(ptr, sz)
	runtime.KeepAlive(raw)
	return res
}

func abiSubmitCharCreate(raw []byte) int32 {
	ptr, sz := bytesPtrLen(raw)
	res := hostUISubmitCharCreate(ptr, sz)
	runtime.KeepAlive(raw)
	return res
}

func abiServerSetCharCreateConfig(raw []byte) int32 {
	ptr, sz := bytesPtrLen(raw)
	res := hostServerSetCharCreateConfig(ptr, sz)
	runtime.KeepAlive(raw)
	return res
}

func abiServerSetCharCreateResult(raw []byte) int32 {
	ptr, sz := bytesPtrLen(raw)
	res := hostServerSetCharCreateResult(ptr, sz)
	runtime.KeepAlive(raw)
	return res
}

//go:wasmexport plugin_init
func pluginInit() int32 {
	if activeClient != nil {
		activeClient.Init()
	}
	if activeServer != nil {
		activeServer.Init()
	}
	return 0
}

//go:wasmexport plugin_on_tick
func pluginOnTick(dtMs uint32) {
	if activeClient != nil {
		activeClient.OnTick(dtMs)
	}
	if activeServer != nil {
		activeServer.OnTick(dtMs)
	}
}

//go:wasmexport plugin_on_enter_world
func pluginOnEnterWorld(ptr, size uint32) {
	if activeClient == nil {
		return
	}
	var snap HeroSnapshot
	if json.Unmarshal(readGuestSlice(ptr, size), &snap) == nil {
		activeClient.OnEnterWorld(snap)
	}
}

//go:wasmexport plugin_on_client_packet
func pluginOnClientPacket(channel, ptr, size uint32) {
	if activeClient == nil {
		return
	}
	src := readGuestSlice(ptr, size)
	cp := make([]byte, len(src))
	copy(cp, src)
	activeClient.OnPacket(uint16(channel), cp)
}

//go:wasmexport plugin_on_ui_event
func pluginOnUIEvent(ptr, size uint32) {
	if activeClient == nil {
		return
	}
	var ev UIClickEvent
	if json.Unmarshal(readGuestSlice(ptr, size), &ev) == nil {
		activeClient.OnUIEvent(ev)
	}
}

//go:wasmexport plugin_on_player_join
func pluginOnPlayerJoin(ptr, size uint32) {
	if activeServer == nil {
		return
	}
	var snap HeroSnapshot
	if json.Unmarshal(readGuestSlice(ptr, size), &snap) == nil {
		activeServer.OnPlayerJoin(snap)
	}
}

//go:wasmexport plugin_on_player_leave
func pluginOnPlayerLeave(ptr, size uint32) {
	if activeServer == nil {
		return
	}
	activeServer.OnPlayerLeave(string(readGuestSlice(ptr, size)))
}

//go:wasmexport plugin_on_room_enter
func pluginOnRoomEnter(ptr, size uint32) {
	if activeServer == nil {
		return
	}
	var ev RoomEvent
	if json.Unmarshal(readGuestSlice(ptr, size), &ev) == nil {
		activeServer.OnRoomEnter(ev)
	}
}

//go:wasmexport plugin_on_chat
func pluginOnChat(ptr, size uint32) {
	if activeServer == nil {
		return
	}
	var ev ChatEvent
	if json.Unmarshal(readGuestSlice(ptr, size), &ev) == nil {
		activeServer.OnChat(ev)
	}
}

//go:wasmexport plugin_on_server_packet
func pluginOnServerPacket(channel, senderPtr, senderLen, dataPtr, dataLen uint32) {
	if activeServer == nil {
		return
	}
	sender := string(readGuestSlice(senderPtr, senderLen))
	src := readGuestSlice(dataPtr, dataLen)
	cp := make([]byte, len(src))
	copy(cp, src)
	activeServer.OnPacket(sender, uint16(channel), cp)
}

//go:wasmexport plugin_on_battle_end
func pluginOnBattleEnd(ptr, size uint32) {
	obs, ok := activeServer.(BattleObserverPlugin)
	if !ok || obs == nil {
		return
	}
	var ev BattleEndEvent
	if json.Unmarshal(readGuestSlice(ptr, size), &ev) == nil {
		obs.OnBattleEnd(ev)
	}
}

//go:wasmexport plugin_on_gather
func pluginOnGather(ptr, size uint32) {
	obs, ok := activeServer.(WorldObserverPlugin)
	if !ok || obs == nil {
		return
	}
	var ev GatherEvent
	if json.Unmarshal(readGuestSlice(ptr, size), &ev) == nil {
		obs.OnGather(ev)
	}
}

//go:wasmexport plugin_on_craft
func pluginOnCraft(ptr, size uint32) {
	obs, ok := activeServer.(WorldObserverPlugin)
	if !ok || obs == nil {
		return
	}
	var ev CraftEvent
	if json.Unmarshal(readGuestSlice(ptr, size), &ev) == nil {
		obs.OnCraft(ev)
	}
}

//go:wasmexport plugin_on_npc_interact
func pluginOnNPCInteract(ptr, size uint32) int32 {
	h, ok := activeServer.(NPCPlugin)
	if !ok || h == nil {
		return 0
	}
	var ev NPCInteractEvent
	if json.Unmarshal(readGuestSlice(ptr, size), &ev) != nil {
		return 0
	}
	res, handled := h.OnNPCInteract(ev)
	if !handled {
		return 0
	}
	raw, err := json.Marshal(res)
	if err != nil {
		return 0
	}
	if rc := abiServerSetNPCResult(raw); rc < 0 {
		return 0
	}
	return 1
}

//go:wasmexport plugin_on_interact
func pluginOnInteract(ptr, size uint32) int32 {
	h, ok := activeServer.(InteractPlugin)
	if !ok || h == nil {
		return 0
	}
	var ev InteractEvent
	if json.Unmarshal(readGuestSlice(ptr, size), &ev) != nil {
		return 0
	}
	if h.OnInteract(ev) {
		return 1
	}
	return 0
}

//go:wasmexport plugin_on_enter_char_create
func pluginOnEnterCharCreate(ptr, size uint32) {
	h, ok := activeClient.(CharacterCreateClientPlugin)
	if !ok || h == nil {
		return
	}
	var ctx CharacterCreateContext
	if json.Unmarshal(readGuestSlice(ptr, size), &ctx) == nil {
		h.OnEnterCharacterCreate(ctx)
	}
}

//go:wasmexport plugin_on_begin_char_create
func pluginOnBeginCharCreate(ptr, size uint32) int32 {
	h, ok := activeServer.(CharacterCreatePlugin)
	if !ok || h == nil {
		return 0
	}
	var ev BeginCharacterCreateEvent
	if json.Unmarshal(readGuestSlice(ptr, size), &ev) != nil {
		return 0
	}
	cfg, handled := h.OnBeginCharacterCreate(ev)
	if !handled {
		return 0
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return 0
	}
	if rc := abiServerSetCharCreateConfig(raw); rc < 0 {
		return 0
	}
	return 1
}

//go:wasmexport plugin_on_finalize_char_create
func pluginOnFinalizeCharCreate(ptr, size uint32) int32 {
	h, ok := activeServer.(CharacterCreatePlugin)
	if !ok || h == nil {
		return 0
	}
	var ev FinalizeCharacterCreateEvent
	if json.Unmarshal(readGuestSlice(ptr, size), &ev) != nil {
		return 0
	}
	res, err := h.OnFinalizeCharacterCreate(ev)
	if err != nil {
		res = CharacterCreateResult{RejectMessage: err.Error()}
	}
	raw, mErr := json.Marshal(res)
	if mErr != nil {
		return 0
	}
	if rc := abiServerSetCharCreateResult(raw); rc < 0 {
		return 0
	}
	return 1
}
