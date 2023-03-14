package c0dec0retwitchbot

/*
import (
	"time"
)

type BasicBot struct {
	channel        string
	MsgRate        time.Duration
	Name           string
	Port           string
	ConfigFilePath string
	Server         string
	Credentials    *OAuthCred
}

type OAuthCred struct {
	Password string `json:"password,omitempty"`
}

type C0deC0reBot interface {
	Connect()
	Disconnect()
	HandelChat() error
	ReadCredentials() (*OAuthCred, error)
	Say(msg string) error
	Start()
}

/*
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
}

// NAME:  Disconnect
// PURPOSE:  To close the connect to the IRC server
// IN: Nothing
// OUT: Nothing
func (ccb *C0deC0reBot) Disconnect() {
	ccb.conn.Close()
	upTime := time.Now().Sub(ccb.startTime).Seconds()
	fmt.Printf("[%s] Closed connection to %s! Live for %fs\n", timeStamp(), ccb.ServerAddr, upTime)

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
						fmt.Printf("Got command: %s", cmd)
						//Channel owner specific commands
						switch cmd {

						case "tbdown":
							if userName == ccb.ChannelName {

								fmt.Printf("[%s} Shutdown command received. Shutting down.\n",
									timeStamp())
								ccb.Disconnect()
								return nil
							}
							break

						case "genprompt":
							fmt.Println("I would be generating a prompt right now")
							err = ccb.Speak("I would be generating a prompt right now")
							if nil != err {
								fmt.Println("err")
							}
							break

						default:
							// Do nothing
						}

					}
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

	fmt.Println("I am HERE")

	comp_msg := fmt.Sprintf("PRIVMSG #%s %s\r\n", ccb.ChannelName, msg)

	fmt.Println(comp_msg)
	// Message was not empty so write the message to the screen
	_, err := ccb.conn.Write([]byte(comp_msg))

	fmt.Println("Passed write")
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
		return
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

*/
import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gempir/go-twitch-irc/v2"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
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

func startServer(ctx context.Context, done chan<- struct{}, data chan<- []byte, out chan<- string) {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	log.Printf("Listening on %s", listener.Addr())

	for {
		select {
		case <-ctx.Done():
			log.Print("Shutting down server...")
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				log.Print(err)
				continue
			}

			go handleConnection(conn, data, out)
		}
	}
}

func handleConnection(conn net.Conn, data chan<- []byte, out chan<- string) {
	defer conn.Close()

	log.Printf("New connection from %s", conn.RemoteAddr())

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Print(err)
			}
			break
		}

		// Send the received data to the `data` channel
		data <- buf[:n]

		// Convert the received data to a string and send it to the `out` channel
		out <- string(buf[:n])
	}

	log.Printf("Connection from %s closed", conn.RemoteAddr())
}

func stateString() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	length := 10 // Length of the random string

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Generate a random string of the given length
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

// /////////////////////////////////STRUCTURES AND INTERFACES///////////////////////////////////////
type C0deC0reBot struct {
	ChannelName    string
	BotName        string
	FilePath       string
	C0deC0reConfig *Config
	Credentials    *OAuthToken
	C0deC0reClient *twitch.Client
	startTime      time.Time
}

type Config struct {
	Secret      string `json:"Secret"`
	ClientID    string `json:"ClientID"`
	TokenURL    string `json:"TokenURL"`
	Permissions string `json:"Permissions"`
	Scope       string `json:"Scope"`
	ListenURL   string `json:"ListenServURL"`
	ListenPort  string `json:"ListenServPort"`
}
type OAuthToken struct {
	AcessToken string `json:"access_token"`
	TokentType string `json:"token_type"`
	ExpiresIn  int    `json:"expires_in"`
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

	// Get the OAuth token
	GetToken() error

	// Validate token scopes
	ValidateToken() (bool, error)
}

func (ccb *C0deC0reBot) Disconnect() {

}

func (ccb *C0deC0reBot) HandleChat() {

}

func (ccb *C0deC0reBot) JoinChannel() {

}

// NAME: Connect
// PURPOSE:  To connect the passed in bot to a twitch IRC sever. It will continously try to connect
//
//	until the but is manaually shutdown.
//
// IN:  b - *Bot, the bot we are trying to connect to the IRC server
// OUT: error, Any error value that is not a timeout error.
func (ccb *C0deC0reBot) Connect() {

	//Print intial message to screen
	fmt.Printf("[%s] Connecting to %s....\n", timeStamp(), ccb.ChannelName)

	//Make connection
	ccb.C0deC0reClient = twitch.NewClient(ccb.BotName, ccb.Credentials.AcessToken)
	ccb.C0deC0reClient.Join(ccb.ChannelName)

	fmt.Printf("[%s] Connected to %s\n", timeStamp(), ccb.ChannelName)

	// This is our callback function that listens for messages in the chat
	ccb.C0deC0reClient.OnPrivateMessage(func(message twitch.PrivateMessage) {
		// parse the message and extract the sender and content
		sender := message.User.Name
		content := message.Message

		// respond to a command
		if strings.HasPrefix(content, "!hello") {
			ccb.C0deC0reClient.Say(ccb.ChannelName, fmt.Sprintf("Hello, %s!", sender))
		}
	})

	// connect to Twitch IRC
	err := ccb.C0deC0reClient.Connect()
	if nil != err {
		fmt.Printf("[%s] Got error %s", timeStamp(), err)
	}

	// keep the program running indefinitely
	for {
		time.Sleep(5 * time.Second)
	}
}

func (ccb *C0deC0reBot) GetToken() error {
// Open the file path and get the details for client id and client secret out
credFile, err := ioutil.ReadFile((ccb.FilePath))
if nil != err {
	fmt.Printf("[%s] Failed to read credentials at: %s", timeStamp(), ccb.FilePath)
	return err
}

// Build the config so that we
ccb.C0deC0reConfig = &Config{}

// Dump info from a config file into our structure
err = json.Unmarshal(credFile, &ccb.C0deC0reConfig)
if err != nil {
	fmt.Println("Error parsing JSON:", err)
	return err
}

// Create a buffered channel to signal the completion of server startup
serverStarted := make(chan struct{}, 1)

// Create a channel to receive data from the server
data := make(chan []byte)

// Create a channel to receive string data from the server
out := make(chan string)

// Start the server in a separate goroutine
go func() {
	startServer(context.Background(), serverStarted, data, out)
}()

// Wait for the server to start listening on the specified port
<-serverStarted

// Set up the data to send in the request body
// Create an http post message
state := stateString()

data := url.Values{}
//data.Set("client_id", ccb.C0deC0reConfig.ClientID)
//data.Set("client_secret", ccb.C0deC0reConfig.Secret)
//data.Set("grant_type", ccb.C0deC0reConfig.Permissions)
//data.Set("scope", ccb.C0deC0reConfig.Scope)
data.Set("client_id", ccb.C0deC0reConfig.ClientID)
data.Set("redirect_uri", ccb.C0deC0reConfig.ListenURL)
data.Set("scope", ccb.C0deC0reConfig.Scope)
data.Set("state", state)

client := &http.Client{}
req, _ := http.NewRequest("POST", ccb.C0deC0reConfig.TokenURL, strings.NewReader(data.Encode()))
req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

// Now that our http request is built we send the request
resp, err := client.Do(req)
if nil != err {
	fmt.Printf("[%s] Failed to get a response when getting the token", timeStamp())
	return err
}
defer resp.Body.Close()

// I need to make sure that I have a credentials structure initlized before I decode into it
ccb.Credentials = &OAuthToken{}

// Initialize credentials and Parse the response into my data structure
err = json.NewDecoder(resp.Body).Decode(ccb.Credentials)
if err != nil {
	fmt.Printf("Error decoding response into OAuthToken struct: %s", err)
	return err
}

return nil
}
}

func (ccb *C0deC0reBot) ValidateToken() (bool, error) {
	req, err := http.NewRequest("GET", "https://id.twitch.tv/oauth2/validate", nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "OAuth"+ccb.Credentials.AcessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return false, nil
	}
	fmt.Println(string(body))

	if resp.StatusCode == http.StatusOK {
		return true, nil
	} else {
		return false, nil
	}
}

func Speak(msg string) error {

	return nil
}
