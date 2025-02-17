package packet

import (
	"bytes"
	"phoenixbuilder/minecraft/protocol"
)

// SetDifficulty is sent by the server to update the client-side difficulty of the client. The actual effect
// of this packet on the client isn't very significant, as the difficulty is handled server-side.
type SetDifficulty struct {
	// Difficulty is the new difficulty that the world has.
	Difficulty uint32
}

// ID ...
func (*SetDifficulty) ID() uint32 {
	return IDSetDifficulty
}

// Marshal ...
func (pk *SetDifficulty) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVaruint32(buf, pk.Difficulty)
}

// Unmarshal ...
func (pk *SetDifficulty) Unmarshal(buf *bytes.Buffer) error {
	return protocol.Varuint32(buf, &pk.Difficulty)
}
