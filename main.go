package main

import (
	"C"
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/pterm/pterm"
	"golang.org/x/term"
	"io/ioutil"
	"os"
	"path/filepath"
	fbauth "phoenixbuilder/cv4/auth"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/command"
	"phoenixbuilder/minecraft/mctype"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/minecraft/utils"
	"phoenixbuilder/minecraft/function"
	"phoenixbuilder/minecraft/configuration"
	//"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/hotbarmanager"
	//"phoenixbuilder/minecraft/enchant"
	"strings"
	"syscall"
	"runtime"
	"runtime/debug"
	"phoenixbuilder/minecraft/fbtask"
	"phoenixbuilder/minecraft/plugin"
	"phoenixbuilder/minecraft/move"
)

type FBPlainToken struct {
	EncryptToken bool   `json:"encrypt_token"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

func main() {
	pterm.Error.Prefix = pterm.Prefix{
		Text:  "ERROR",
		Style: pterm.NewStyle(pterm.BgBlack, pterm.FgRed),
	}
	//Version num should seperate from fellow strings
	//for implenting print version feature later
	const FBVersion = "0.3.9"
	const FBCodeName = "Phoenix"

	pterm.DefaultBox.Println(pterm.LightCyan("Copyright notice: \n" +
		"FastBuilder Phoenix used codes\n" +
		"from Sandertv's Gophertunnel that\n" +
		"licensed under MIT license, at:\n" +
		"https://github.com/Sandertv/gophertunnel"))
	pterm.Println(pterm.Yellow("ファスト　ビルダー"))
	pterm.Println(pterm.Yellow("F A S T  B U I L D E R"))
	pterm.Println(pterm.Yellow("Contributors: Ruphane, CAIMEO"))
	pterm.Println(pterm.Yellow("Copyright (c) FastBuilder DevGroup, Bouldev 2021"))
	pterm.Println(pterm.Yellow("FastBuilder Phoenix Alpha " + FBVersion))
	//if runtime.GOOS == "windows" {}
	defer func() {
		if err:=recover(); err!=nil {
			debug.PrintStack()
			pterm.Error.Println("Oh no! FastBuilder Phoenix crashed! ")
			pterm.Error.Println("Stack dump was shown above, error:")
			pterm.Error.Println(err)
			if runtime.GOOS == "windows" {
				pterm.Error.Println("Press ENTER to exit.")
				_, _=bufio.NewReader(os.Stdin).ReadString('\n')
			}
			os.Exit(1)
		}
		os.Exit(0)
		//os.Exit(rand.Int())
	}()
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	token := loadTokenPath()
	version, err := utils.GetHash(ex)
	if err != nil {
		panic(err)
	}
	if _, err := os.Stat(token); os.IsNotExist(err) {
		fbusername, err := getInputUserName()
		if err != nil {
			panic(err)
		}
		fbuntrim := fmt.Sprintf("%s", strings.TrimSuffix(fbusername, "\n"))
		fbun := strings.TrimRight(fbuntrim, "\r\n")
		fmt.Printf("Enter your FastBuilder User Center password: ")
		fbpassword, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Printf("\n")
		tokenstruct := &FBPlainToken{
			EncryptToken: true,
			Username:     fbun,
			Password:     string(fbpassword),
		}
		token, err := json.Marshal(tokenstruct)
		if err != nil {
			fmt.Println("Failed to generate temp token")
			fmt.Println(err)
			return
		}
		runShellClient(string(token), version)

	} else {
		token, err := readToken(token)
		if err != nil {
			fmt.Println(err)
			return
		}
		runShellClient(token, version)
	}
}

//export iOSAppStart
func iOSAppStart(token string, version string, serverCode string, serverPasswd string, onError func()) {
	defer func() {
		onError()
	}()
	runClient(token, version, serverCode, serverPasswd)
}

func runShellClient(token string, version string) {
	code, serverPasswd, err := getRentalServerCode()
	if err != nil {
		fmt.Println(err)
		return
	}
	runClient(token, version, code, serverPasswd)
}

func runClient(token string, version string, code string, serverPasswd string) {
	worldchatchannel:=make(chan []string)
	client := fbauth.CreateClient(worldchatchannel)
	if token[0] == '{' {
		token = client.GetToken("", token)
		if token == "" {
			fmt.Println("Incorrect username or password")
			return
		}
		tokenPath := loadTokenPath()
		if fi, err := os.Create(tokenPath); err != nil {
			fmt.Println("Error creating token file: ", err)
			fmt.Println("Error ignored.")
		} else {
			configuration.UserToken=token
			_, err = fi.WriteString(token)
			if err != nil {
				fmt.Println("Error saving token: ", err)
				fmt.Println("Error ignored.")
			}
			fi.Close()
			fi = nil
		}
	}else{
		configuration.UserToken=token
	}
	serverCode := fmt.Sprintf("%s", strings.TrimSuffix(code, "\n"))
	pterm.Println(pterm.Yellow(fmt.Sprintf("Server: %s", serverCode)))
	dialer := minecraft.Dialer{
		ServerCode: strings.TrimRight(serverCode, "\r\n"),
		Password:   serverPasswd,
		Version:    version,
		Token:      token,
		Client:     client,
	}
	conn, err := dialer.Dial("raknet", "")

	if err != nil {
		pterm.Error.Println(err)
		panic(err)
	}
	defer conn.Close()
	pterm.Println(pterm.Yellow("Successfully created minecraft dialer."))
	user := client.ShouldRespondUser()
	configuration.RespondUser=user
	// delay := 1000 //BP MMS
	// Make the client spawn in the world: This is a blocking operation that will return an error if the
	// client times out while spawning.
	
	conn.WritePacket(&packet.PlayerAction {
		EntityRuntimeID: conn.GameData().EntityRuntimeID,
		ActionType: packet.PlayerActionRespawn,
	})
	go func() {
		for {
			csmsg:=<-worldchatchannel
			command.WorldChatTellraw(conn, csmsg[0], csmsg[1])
		}
	} ()
	
	plugin.StartPluginSystem(conn)
	
	/*if err := conn.DoSpawn(); err != nil {
		pterm.Error.Println("Failed to spawn")
		panic(err)
	}*/
	//pterm.Println(pterm.Yellow("Player spawned successfully."))

	function.InitInternalFunctions()
	fbtask.InitTaskStatusDisplay(conn)
	hotbarmanager.Init()

	zeroId, _ := uuid.NewUUID()
	oneId, _ := uuid.NewUUID()
	configuration.ZeroId=zeroId
	configuration.OneId=oneId
	mctype.ForwardedBrokSender=fbtask.BrokSender
	tellraw(conn, "Welcome to FastBuilder!")
	tellraw(conn, fmt.Sprintf("Operator: %s", user))
	sendCommand("testforblock ~ ~ ~ air", zeroId, conn)
	go func() {
		for {
			cmd, _:=getInput()
			if len(cmd) == 0 {
				continue
			}
			if cmd[0] == '.' {
				ud,_:=uuid.NewUUID()
				chann:=make(chan *packet.CommandOutput)
				command.UUIDMap.Store(ud.String(), chann)
				command.SendCommand(cmd[1:], ud, conn)
				resp:=<-chann
				fmt.Printf("%+v\n", resp)
			}else if cmd[0] == '!' {
				ud,_:=uuid.NewUUID()
				chann:=make(chan *packet.CommandOutput)
				command.UUIDMap.Store(ud.String(), chann)
				command.SendWSCommand(cmd[1:], ud, conn)
				resp:=<-chann
				fmt.Printf("%+v\n", resp)
			}
			if cmd=="menu" {
				move.OpenMenu(conn)
				fmt.Printf("OK\n")
				continue
			}
			if cmd[0] == '>'&&len(cmd)>1 {
				umsg:=cmd[1:]
				if(!client.CanSendMessage()) {
					command.WorldChatTellraw(conn, "FastBuildeｒ", "Lost connection to the authentication server.")
					break
				}
				client.WorldChat(umsg)
			}
			function.Process(conn, cmd)
		}
	} ()
	// A loop that reads packets from the connection until it is closed.
	for {
		// Read a packet from the connection: ReadPacket returns an error if the connection is closed or if
		// a read timeout is set. You will generally want to return or break if this happens.
		pk, err := conn.ReadPacket()
		if err != nil {
			panic(err)
		}

		switch p := pk.(type) {
		case *packet.StructureTemplateDataResponse:
			//fmt.Printf("RESPONSE %+v\n",p.StructureTemplate)
			fbtask.ExportWaiter<-p.StructureTemplate
			break
		case *packet.MovePlayer:
			if(p.EntityRuntimeID==move.OPRuntimeId) {
				move.LastOPPitch=p.Pitch
			}
		case *packet.SetActorData:
			if(p.EntityRuntimeID==move.OPRuntimeId) {
				if len(p.EntityMetadata)!=1 {
					break
				}
				_,isSneak:=p.EntityMetadata[91]
				if isSneak {
					move.LastOPSneak=!move.LastOPSneak
					if(move.LastOPMouSneaked) {
						move.LastOPMouSneaked=false
					}
				}
			}
		/*case *packet.InventoryContent:
			for _, item := range p.Content {
				fmt.Printf("InventorySlot %+v\n",item.Stack.NBTData["dataField"])
			}
			break*/
		/*case *packet.InventorySlot:
			fmt.Printf("Slot %d:%+v",p.Slot,p.NewItem.Stack)*/
		case *packet.Text:
			if p.TextType == packet.TextTypeChat {
				if user == p.SourceName {
					//if (strings.Contains(p.Message,"a")||strings.Contains(p.Message,"A")||strings.Contains(p.Message,"An")||strings.Contains(p.Message,"an")) {
					//	move.OpenMenu(conn)
					//}
					if p.Message[0] == '>'&&len(p.Message)>1 {
						umsg:=p.Message[1:]
						if(!client.CanSendMessage()) {
							command.WorldChatTellraw(conn, "FasｔBuildeｒ", "Lose connection to the authentication server.")
							break
						}
						client.WorldChat(umsg)
					}
					break
					pterm.Println(pterm.Yellow(fmt.Sprintf("<%s>", user)), pterm.LightCyan(p.Message))
					if p.Message[0] == '>' {
						//umsg:=p.Message[1:]
						//
					}
					function.Process(conn, p.Message)
					break
				}
			}
		case *packet.AddPlayer:
			if (p.Username == user) {
				move.OPRuntimeId=p.EntityRuntimeID;
				if(move.OPRuntimeIdReceivedChannel!=nil) {
					move.OPRuntimeIdReceivedChannel<-true
				}
			}
			//if (p.Username == user && enchant.AddPlayerItemChannel != nil) {
			//	enchant.AddPlayerItemChannel<-&p.HeldItem
			//	pterm.Println(pterm.Yellow(fmt.Sprintf("[%s] Operator joined Game", user)))
			//}
			//fmt.Printf("%+v\n",p.EntityMetadata)
		case *packet.CommandOutput:
			//if p.SuccessCount > 0 {
				if p.CommandOrigin.UUID.String() == configuration.ZeroId.String() {
					pos, _ := utils.SliceAtoi(p.OutputMessages[0].Parameters)
					if len(pos) == 0 {
						tellraw(conn, "Invalid position")
						break
					}
					configuration.GlobalFullConfig().Main().Position = mctype.Position{
						X: pos[0],
						Y: pos[1],
						Z: pos[2],
					}
					tellraw(conn, fmt.Sprintf("Position got: %v", pos))
					break
				}else if p.CommandOrigin.UUID.String() == configuration.OneId.String() {
					pos, _ := utils.SliceAtoi(p.OutputMessages[0].Parameters)
					if len(pos) == 0 {
						tellraw(conn, "Invalid position")
						break
					}
					configuration.GlobalFullConfig().Main().End = mctype.Position{
						X: pos[0],
						Y: pos[1],
						Z: pos[2],
					}
					tellraw(conn, fmt.Sprintf("End Position got: %v", pos))
					break
				}
			//}
			pr, ok := command.UUIDMap.LoadAndDelete(p.CommandOrigin.UUID.String())
			if ok {
				pu:=pr.(chan *packet.CommandOutput)
				pu<-p
			}
		}

	}
}

func getInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	inp, err := reader.ReadString('\n')
	inpl:=strings.TrimRight(inp, "\r\n")
	return inpl, err
}

func getInputUserName() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	pterm.Printf("Enter your FastBuilder User Center username: ")
	fbusername, err := reader.ReadString('\n')
	return fbusername, err
}

func getRentalServerCode() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Please enter the rental server number: ")
	code, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}
	fmt.Printf("Enter Password (Press [Enter] if not set, input won't be echoed): ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Printf("\n")
	return code, string(bytePassword), err
}

func readToken(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func sendCommand(commands string, UUID uuid.UUID, conn *minecraft.Conn) error {
	/*requestId, _ := uuid.Parse("96045347-a6a3-4114-94c0-1bc4cc561694")
	origin := protocol.CommandOrigin{
		Origin:         protocol.CommandOriginPlayer,
		UUID:           UUID,
		RequestID:      requestId.String(),
		PlayerUniqueID: 0,
	}
	commandRequest := &packet.CommandRequest{
		CommandLine:   command,
		CommandOrigin: origin,
		Internal:      false,
		UnLimited:     false,
	}
	return conn.WritePacket(commandRequest)*/
	return command.SendCommand(commands,UUID,conn)
}

func tellraw(conn *minecraft.Conn, lines ...string) error {
	//fmt.Printf("%s\n",lines[0])
	//return nil
	uuid1, _ := uuid.NewUUID()
	return sendCommand(command.TellRawRequest(mctype.AllPlayers, lines...), uuid1, conn)
}

func decideDelay(delaytype byte) int64 {
	// Will add system check later,so don't merge into other functions.
	if delaytype==mctype.DelayModeContinuous {
		return 1000
	}else if delaytype==mctype.DelayModeDiscrete {
		return 15
	}else{
		return 0
	}
}

func decideDelayThreshold() int {
	// Will add system check later,so don't merge into other functions.
	return 20000
}

func loadTokenPath() string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("WARNING - Failed to obtain the user's home directory. made homedir=\".\";")
		homedir="."
	}
	fbconfigdir := filepath.Join(homedir, ".config/fastbuilder")
	os.MkdirAll(fbconfigdir, 0755)
	token := filepath.Join(fbconfigdir,"fbtoken")
	return token
}
