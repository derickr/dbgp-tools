package protocol

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"github.com/derickr/dbgp-tools/lib/xml"
	"golang.org/x/net/html/charset"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

type Response interface {
	String() string
	IsSuccess() bool
	ExpectMoreResponses() bool
	ShouldCloseConnection() bool
}

type dbgpClient struct {
	connection  net.Conn
	reader      *bufio.Reader
	writer      io.Writer
	counter     int
	smartClient bool

	lastSourceBegin  int
	isInConversation bool
	abortRequested   bool
	commandsToRun    []string
}

func NewDbgpClient(c net.Conn, isSmart bool) *dbgpClient {
	var tmp dbgpClient

	tmp.connection = c
	tmp.reader = bufio.NewReader(c)
	tmp.writer = c
	tmp.counter = 1
	tmp.lastSourceBegin = 1
	tmp.smartClient = isSmart
	tmp.isInConversation = false
	tmp.abortRequested = false

	if isSmart {
		tmp.commandsToRun = append(tmp.commandsToRun, "feature_get -n supports_async")
	}

	return &tmp
}

func (dbgp *dbgpClient) AddCommandToRun(command string) {
	if !dbgp.smartClient {
		return
	}

	dbgp.commandsToRun = append(dbgp.commandsToRun, command)
}

func (dbgp *dbgpClient) GetNextCommand() (string, bool) {
	var item string

	if len(dbgp.commandsToRun) > 0 {
		item, dbgp.commandsToRun = dbgp.commandsToRun[0], dbgp.commandsToRun[1:]

		return item, true
	}

	return "", false
}

func (dbgp *dbgpClient) HasCommandsToRun() bool {
	return len(dbgp.commandsToRun) > 0
}

func (dbgp *dbgpClient) ParseInitXML(rawXmlData string) (dbgpXml.Init, error) {
	init := dbgpXml.Init{}

	reader := strings.NewReader(rawXmlData)

	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel

	err := decoder.Decode(&init)

	if err != nil {
		return init, err
	}

	dbgp.TheConversationIsOn()

	return init, nil
}

func (dbgp *dbgpClient) parseProxyInitXML(rawXmlData string) (dbgpXml.ProxyInit, error) {
	init := dbgpXml.ProxyInit{}

	reader := strings.NewReader(rawXmlData)

	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel

	err := decoder.Decode(&init)

	if err != nil {
		return init, err
	}

	return init, nil
}

func (dbgp *dbgpClient) parseProxyStopXML(rawXmlData string) (dbgpXml.ProxyStop, error) {
	init := dbgpXml.ProxyStop{}

	reader := strings.NewReader(rawXmlData)

	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel

	err := decoder.Decode(&init)

	if err != nil {
		return init, err
	}

	return init, nil
}

func (dbgp *dbgpClient) parseCloudInitXML(rawXmlData string) (dbgpXml.CloudInit, error) {
	init := dbgpXml.CloudInit{}

	reader := strings.NewReader(rawXmlData)

	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel

	err := decoder.Decode(&init)

	if err != nil {
		return init, err
	}

	return init, nil
}

func (dbgp *dbgpClient) parseNotifyXML(rawXmlData string) (dbgpXml.Notify, error) {
	notify := dbgpXml.Notify{}

	reader := strings.NewReader(rawXmlData)

	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel

	err := decoder.Decode(&notify)

	if err != nil {
		return notify, err
	}

	return notify, nil
}

func (dbgp *dbgpClient) TheConversationIsOn() {
	dbgp.isInConversation = true
}
func (dbgp *dbgpClient) IsInConversation() bool {
	return dbgp.isInConversation
}

func (dbgp *dbgpClient) SignalAbort() {
	dbgp.abortRequested = true
}

func (dbgp *dbgpClient) HasAbortBeenSignalled() bool {
	return dbgp.abortRequested
}

func (dbgp *dbgpClient) parseResponseXML(rawXmlData string) (dbgpXml.Response, error) {
	response := dbgpXml.Response{}

	reader := strings.NewReader(rawXmlData)

	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel

	err := decoder.Decode(&response)

	response.LastSourceBegin = dbgp.lastSourceBegin
	if err != nil {
		return response, err
	}

	return response, nil
}

func (dbgp *dbgpClient) parseStreamXML(rawXmlData string) (dbgpXml.Stream, error) {
	stream := dbgpXml.Stream{}

	reader := strings.NewReader(rawXmlData)

	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel

	err := decoder.Decode(&stream)

	if err != nil {
		return stream, err
	}

	return stream, nil
}

func (dbgp *dbgpClient) ReadResponse() (string, error, bool) {
	dbgp.connection.SetReadDeadline(time.Now().Add(1 * time.Second))

	/* Read length */
	_, err := dbgp.reader.ReadBytes('\000')

	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return "", err, true
		}

		return "", fmt.Errorf("Error reading length: %s", err), false
	}

	/* Read data */
	data, err := dbgp.reader.ReadBytes('\000')

	if err != nil {
		return "", fmt.Errorf("Error reading data: %s", err), false
	}

	return string(data), nil, false
}

func (dbgp *dbgpClient) injectIIfNeeded(parts []string) []string {
	for _, item := range parts {
		if item == "-i" {
			return parts
		}
	}

	var newParts []string
	newParts = append(newParts, parts[0])
	newParts = append(newParts, "-i", fmt.Sprintf("%d", dbgp.counter))
	newParts = append(newParts, parts[1:]...)

	dbgp.counter++

	return newParts
}

func (dbgp *dbgpClient) storeSourceBeginIfPresent(parts []string) []string {
	s_found := false

	dbgp.lastSourceBegin = 1

	for _, item := range parts {
		if s_found {
			value, err := strconv.Atoi(item)
			if err == nil && value > 0 {
				dbgp.lastSourceBegin = value
			}
			s_found = false
		}
		if item == "-b" {
			s_found = true
		}
	}

	return parts
}

func (dbgp *dbgpClient) processLine(line string) string {
	parts := strings.Split(strings.TrimSpace(line), " ")

	parts = dbgp.injectIIfNeeded(parts)
	parts = dbgp.storeSourceBeginIfPresent(parts)

	return strings.Join(parts, " ")
}

func (dbgp *dbgpClient) SendCommand(line string) error {
	line = dbgp.processLine(line)

	_, err := dbgp.writer.Write([]byte(line))
	if err != nil {
		fmt.Println("Error writing:", err.Error())
		return err
	}

	_, err = dbgp.writer.Write([]byte("\000"))
	if err != nil {
		fmt.Println("Error writing:", err.Error())
		return err
	}

	return nil
}

func (dbgp *dbgpClient) FormatXML(rawXmlData string) Response {
	var response Response

	response, err := dbgp.parseResponseXML(rawXmlData)

	if err == nil {
		return response
	}

	response, err = dbgp.ParseInitXML(rawXmlData)

	if err == nil {
		return response
	}

	response, err = dbgp.parseNotifyXML(rawXmlData)

	if err == nil {
		return response
	}

	response, err = dbgp.parseStreamXML(rawXmlData)

	if err == nil {
		return response
	}

	response, err = dbgp.parseProxyInitXML(rawXmlData)

	if err == nil {
		return response
	}

	response, err = dbgp.parseProxyStopXML(rawXmlData)

	if err == nil {
		return response
	}

	response, err = dbgp.parseCloudInitXML(rawXmlData)

	if err == nil {
		return response
	}

	return nil
}
