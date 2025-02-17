package packet

import (
	"bytes"
	"phoenixbuilder/minecraft/protocol"
)

// RequestChunkRadius is sent by the client to the server to update the server on the chunk view radius that
// it has set in the settings. The server may respond with a ChunkRadiusUpdated packet with either the chunk
// radius requested, or a different chunk radius if the server chooses so.
type RequestChunkRadius struct {
	// ChunkRadius is the requested chunk radius. This value is always the value set in the settings of the
	// player.
	ChunkRadius int32
}

// ID ...
func (*RequestChunkRadius) ID() uint32 {
	return IDRequestChunkRadius
}

// Marshal ...
func (pk *RequestChunkRadius) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVarint32(buf, pk.ChunkRadius)
}

// Unmarshal ...
func (pk *RequestChunkRadius) Unmarshal(buf *bytes.Buffer) error {
	return protocol.Varint32(buf, &pk.ChunkRadius)
}
