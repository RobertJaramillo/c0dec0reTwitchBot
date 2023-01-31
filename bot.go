package c0dec0retwitchbot

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/textproto"
	"regexp"
	"strings"
	"time"
)

// Constant time format so our bot timestamps messages in the way
const ESTFormat = "Jan 26 22:37:00 EST"

// NAME: timeStamp
// PURPOSE: To return the cuirrent date and time in a formatted string
// IN: Nothing
// OUT: A formatted string containing the current date and time
func timeStamp() string {
	return time.Now().Format(ESTFormat)
}

// /////////////////////////////////STRUCTURES AND INTERFACES///////////////////////////////////////
type C0deC0reBot struct {
	ChannelName string
	MsgRate     time.Duration
	BotName     string
	Port        string
	FilePath    string
	ServerAddr  string
	Credentials *OAuthToken
	conn        net.Conn
	startTime   time.Time
}

type Bot interface {

	// Opens a connection to the Twitch.tv IRC chat server.
	Connect()

	// Closes a connection to the Twitch.tv IRC chat server.
	Disconnect()

	// Listens to chat messages and PING request from the IRC server.
	HandleChat() error

	// Joins a specific chat channel.
	JoinChannel()

	// Parses credentials needed for authentication.
	ReadCredentials() error

	// Sends a message to the connected channel.
	Speak(msg string) error

	// Attempts to keep the bot connected and handling chat.
	Start()
}

type OAuthToken struct {
	Password string `json:"Password,omitempty"`
}

/////////////////////////REGULAR EXPRESSIONS///////////////////////////////////////////////////////

// Regular expression for parsing PRIVMSG. Note this means someone talked in the channel to
// include the bot itself
var msgRegex *regexp.Regexp = regexp.MustCompile(`^:(\w+)!\w+@\w+\.tmi\.twitch\.tv (PRIVMSG) #\w+(?: :(.*))?$`)

// Regular expression for parsing user commands, from already parsed PRIVMSG string
// First matched group is the command name and the second matched group is the argument for the
// command.

var cmdRegex *regexp.Regexp = regexp.MustCompile(`^!(\w+)\s?(\w+)?`)

//////////////////////////////CORE FUNCTIONALITY//////////////////////////////////////////////////

// NAME: Connect
// PURPOSE:  To connect the passed in bot to a twitch IRC sever. It will continously try to connect
//
//	until the but is manaually shutdown.
//
// IN:  b - *Bot, the bot we are trying to connect to the IRC server
// OUT: error, Any error value that is not a timeout error.
func (ccb *C0deC0reBot) Connect() {

	var err error

	//Print intial message to screen
	fmt.Printf("[%s] Connecting to %s....\n", timeStamp(), ccb.ServerAddr)

	//Make connection
	ccb.conn, err = net.Dial("tcp", ccb.ServerAddr+":"+ccb.Port)
	if nil != err {
		fmt.Printf("[%s] Failed to Connect to %s, retrying", timeStamp(), ccb.ServerAddr)
		ccb.Connect()
		return
	}

	fmt.Printf("[%s] Connected to %s\n", timeStamp(), ccb.ServerAddr)
	return
}

// NAME:  Disconnect
// PURPOSE:  To close the connect to the IRC server
// IN: Nothing
// OUT: Nothing
func (ccb *C0deC0reBot) Disconnect() {
	ccb.conn.Close()
	upTime := time.Now().Sub(ccb.startTime).Seconds()
	fmt.Printf("[%s] Closed connection to %s! Live for %fs\n", timeStamp(), ccb.ServerAddr, upTime)
	return
}

// NAME: HandleChat
// PURPOSE: To listen for log messages from the chat and respond to cvommands from the channel
//
//	owner. The bot will continue until its told to quit or is forcefully shut down.
//
// IN: Nothing
// OUT: Nothing
func (ccb *C0deC0reBot) HandleChat() error {

	fmt.Printf("[%s] Watching #%s.... \n", timeStamp(), ccb.ChannelName)

	// Reader that provides generic support for HTTP/NMT/SMTP request/response messages
	tp := textproto.NewReader(bufio.NewReader(ccb.conn))

	// Listen for messages
	for {

		// Grab the message and exit if an error occured
		line, err := tp.ReadLine()
		if nil != err {
			ccb.Disconnect()
			return errors.New("ccb.Bot.HandleChat:  Failed to read the channel.  Disconnected")
		}

		// Log the message from the IRC server
		fmt.Printf("[%s] %s\n", timeStamp(), line)

		// If its a PING message make sure we respond with a Pong message to not get
		// disconnected.
		if "PING:  tmi.twitch.tv" == line {
			ccb.conn.Write([]byte("PONG :tmi.twitch.tv\r\n"))
		} else {

			// Handle PRIVMSG message
			matches := msgRegex.FindStringSubmatch(line)

			// Check if we got a match
			if nil != matches {
				userName := matches[1]
				msgType := matches[2]

				switch msgType {

				case "PRIVMSG":
					msg := matches[3]
					fmt.Printf("[%s] %s %s\n", timeStamp(), userName, msg)

					cmdMatches := cmdRegex.FindStringSubmatch(msg)

					// Parse the commands if we received any that matched our commands
					if nil != cmdMatches {
						// set command and arguments to command
						// TODO: figure out the argument bit a little later
						cmd := cmdMatches[1]
						// arg := cmdMatches[2]

						//Channel owner specific commands
						if userName == ccb.ChannelName {
							switch cmd {
							case "tbdown":
								fmt.Printf("[%s} Shutdown command received. Shutting down.\n",
									timeStamp())
								ccb.Disconnect()
								return nil
							default:
								// Do nothing
							}
						}

					}

				default:
					//Do nothging if it didnt match a command
				}
			}
		}

		// We have to limit how fast we responf to messages on the IRC server
		// or our accounts can get banned
		time.Sleep(ccb.MsgRate)
	}
}

// NAME: JoinChannel
// PURPOSE: To join the channel specified in the bots configuration
// IN:  Nothing
// OUT: Nothing
func (ccb *C0deC0reBot) JoinChannel() {
	fmt.Printf("[%s]  Joining #%s", timeStamp(), ccb.ChannelName)
	ccb.conn.Write([]byte("PASS " + ccb.Credentials.Password + "\r\n"))
	ccb.conn.Write([]byte("NICK " + ccb.BotName + "\r\n"))
	ccb.conn.Write([]byte("JOIN #" + ccb.ChannelName + "\r\n"))

	// TODO:  Look at error checking here to make sure we handle the instance of
	// not connecting to the server

	fmt.Printf("[%s] Joined #%s as @%s!\r\n", timeStamp(), ccb.ChannelName, ccb.BotName)
}

// NAME: ReadCredentials
// PURPOSE: To read login and token information from a json file
//
// IN: Nothing
// OUT: Error if encounter nil if no error
func (ccb *C0deC0reBot) ReadCredentials() error {

	credFile, err := ioutil.ReadFile(ccb.FilePath)
	if nil != err {
		fmt.Printf("[%s] Failed to read credentials file at: %s", timeStamp(), ccb.FilePath)
		return err
	}

	ccb.Credentials = &OAuthToken{}

	// Creates a JSON decoder
	dec := json.NewDecoder(strings.NewReader(string(credFile)))

	// parse the JSON file
	err = dec.Decode(ccb.Credentials)
	if nil != err && io.EOF != err {
		return err
	}

	return nil
}

// NAME: Speak
// PURPOSE: Makes the bot send messages to the channel
// IN: Nothing
// OUT: Error if encountered, nil if no error encountered
func (ccb *C0deC0reBot) Speak(msg string) error {

	// Check for an empty message and return an error if message was empty
	if msg == "" {
		return errors.New("C0deC0reTwitchBot: cant speak, message was empty")

	}

	// Message was not empty so write the message to the screen
	_, err := ccb.conn.Write([]byte(fmt.Sprintf("PRIVMSG #%s %s\r\n", ccb.BotName, msg)))
	if nil != err {
		return err
	}

	return nil
}

// NAME: Start
// PURPOSE: Tells the bot to connect to a specified channel and handle chat messages until it is
//
//	forced to shutdown.
//
// IN: Nothing
// OUT: Nothing
func (ccb *C0deC0reBot) Start() {

	// First grab our credentials
	err := ccb.ReadCredentials()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Aborting...")
	}

	// Credentials are valid no contionouslt try to connect
	for {
		ccb.Connect()
		ccb.JoinChannel()
		err = ccb.HandleChat()
		if nil != err {

			// Attempt to reconnect unpon unexpected error
			fmt.Println(err)
			time.Sleep(1000 * time.Millisecond)
			fmt.Println("Starting C0deC0reBot again...")
		} else {
			ccb.conn.Close()
			return
		}
	}
}
