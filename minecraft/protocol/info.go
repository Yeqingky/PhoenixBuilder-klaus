package protocol

const (
	// CurrentProtocol is the current protocol version for the version below.
	CurrentProtocol = 422
	// CurrentVersion is the current version of Minecraft as supported by the `packet` package.
	CurrentVersion = "1.16.201"
	// CurrentBlockVersion is the current version of blocks (states) of the game. This version is composed
	// of 4 bytes indicating a version, interpreted as a big endian int. The current version represents
	// 1.16.0.14 {1, 16, 0, 14}.
	CurrentBlockVersion int32 = 17825806
)
