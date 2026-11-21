//go:build !wasip1

package sdk

func abiLog(uint32, string)                       {}
func abiSendToServer(uint16, []byte) int32        { return 0 }
func abiServerSendPacket([]byte) int32            { return 0 }
func abiServerSendNotice([]byte) int32            { return 0 }
func abiServerRegisterMonster([]byte) int32       { return 0 }
func abiServerSpawnMonster([]byte) int32          { return 0 }
func abiServerRegisterItem([]byte) int32          { return 0 }
func abiServerSpawnGatherNode([]byte) int32       { return 0 }
func abiServerRegisterRecipe([]byte) int32        { return 0 }
func abiServerGiveItem([]byte) int32              { return 0 }
func abiServerConfigureCodex([]byte) int32        { return 0 }
func abiServerRegisterCodexEntry([]byte) int32    { return 0 }
func abiSetHUDActions([]byte) int32               { return 0 }
func abiSetOverlay([]byte) int32                  { return 0 }
func abiOpenCodex(string) int32                   { return 0 }
func abiPlayAudio(string, uint32) int32           { return 0 }
func abiServerKVSet(string, []byte) int32         { return 0 }
func abiServerKVGet(string) ([]byte, bool, int32) { return nil, false, 0 }
func abiServerKVDelete(string) int32              { return 0 }
func abiServerKVList(string, int) ([]byte, int32) { return []byte("[]"), 0 }
func abiPrefSet(string, string) int32             { return 0 }
func abiPrefGet(string) ([]byte, bool, int32)     { return nil, false, 0 }
func abiPrefDelete(string) int32                    { return 0 }
func abiPrefList(string, int) ([]byte, int32)       { return []byte("[]"), 0 }
func abiRegisterCustomMesh([]byte) int32            { return 0 }
func abiServerRegisterRoom([]byte) int32            { return 0 }
func abiServerRemoveRoom([]byte) int32              { return 0 }
func abiServerSetSpawnRoom([]byte) int32            { return 0 }
func abiServerRegisterNPC([]byte) int32             { return 0 }
func abiServerSetHeroLoadout([]byte) int32          { return 0 }
func abiServerSetStarterConfig([]byte) int32        { return 0 }
func abiServerRegisterPlayerSkill([]byte) int32     { return 0 }
func abiServerRegisterJob([]byte) int32             { return 0 }
func abiServerSetNPCResult([]byte) int32            { return 0 }
