package packet

import (
	"bytes"
	"encoding/binary"
	"phoenixbuilder/minecraft/protocol"
)

// SettingsCommand is sent by the client when it changes a setting in the settings that results in the issuing
// of a command to the server, such as when Show Coordinates is enabled.
type SettingsCommand struct {
	// CommandLine is the full command line that was sent to the server as a result of the setting that the
	// client changed.
	CommandLine string
	// SuppressOutput specifies if the client requests the suppressing of the output of the command that was
	// executed. Generally this is set to true, as the client won't need a message to confirm the output of
	// the change.
	SuppressOutput bool
}

// ID ...
func (*SettingsCommand) ID() uint32 {
	return IDSettingsCommand
}

// Marshal ...
func (pk *SettingsCommand) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteString(buf, pk.CommandLine)
	_ = binary.Write(buf, binary.LittleEndian, pk.SuppressOutput)
}

// Unmarshal ...
func (pk *SettingsCommand) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.String(buf, &pk.CommandLine),
		binary.Read(buf, binary.LittleEndian, &pk.SuppressOutput),
	)
}
