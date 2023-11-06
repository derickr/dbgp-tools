package protocol

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"github.com/derickr/dbgp-tools/lib/dbgpxml"
	"github.com/derickr/dbgp-tools/lib/logger"
	. "github.com/logrusorgru/aurora" // WTFPL
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
	GetErrorMessage() string
}

type dbgpClient struct {
	connection  net.Conn
	logger      logger.Logger
	reader      *bufio.Reader
	writer      io.Writer
	counter     int

	lastSourceBegin  int
	abortRequested   bool
	commandsToRun    []string
}

func NewDbgpClient(c net.Conn, logger logger.Logger) *dbgpClient {
	var tmp dbgpClient

	tmp.connection = c
	tmp.logger = logger
	tmp.reader = bufio.NewReader(c)
	tmp.writer = c
	tmp.counter = 1
	tmp.lastSourceBegin = 1
	tmp.abortRequested = false

	return &tmp
}

func (dbgp *dbgpClient) ParseInitXML(rawXmlData string) (dbgpxml.Init, error) {
	init := dbgpxml.Init{}

	reader := strings.NewReader(rawXmlData)

	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel

	err := decoder.Decode(&init)

	if err != nil {
		return init, err
	}

	return init, nil
}

func (dbgp *dbgpClient) parseProxyInitXML(rawXmlData string) (dbgpxml.ProxyInit, error) {
	init := dbgpxml.ProxyInit{}

	reader := strings.NewReader(rawXmlData)

	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel

	err := decoder.Decode(&init)

	if err != nil {
		return init, err
	}

	return init, nil
}

func (dbgp *dbgpClient) parseProxyStopXML(rawXmlData string) (dbgpxml.ProxyStop, error) {
	init := dbgpxml.ProxyStop{}

	reader := strings.NewReader(rawXmlData)

	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel

	err := decoder.Decode(&init)

	if err != nil {
		return init, err
	}

	return init, nil
}

func (dbgp *dbgpClient) parseCloudInitXML(rawXmlData string) (dbgpxml.CloudInit, error) {
	init := dbgpxml.CloudInit{}

	reader := strings.NewReader(rawXmlData)

	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel

	err := decoder.Decode(&init)

	if err != nil {
		return init, err
	}

	return init, nil
}

func (dbgp *dbgpClient) parseCloudStopXML(rawXmlData string) (dbgpxml.CloudStop, error) {
	init := dbgpxml.CloudStop{}

	reader := strings.NewReader(rawXmlData)

	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel

	err := decoder.Decode(&init)

	if err != nil {
		return init, err
	}

	return init, nil
}

func (dbgp *dbgpClient) parseNotifyXML(rawXmlData string) (dbgpxml.Notify, error) {
	notify := dbgpxml.Notify{}

	reader := strings.NewReader(rawXmlData)

	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel

	err := decoder.Decode(&notify)

	if err != nil {
		return notify, err
	}

	return notify, nil
}

func (dbgp *dbgpClient) SignalAbort() {
	dbgp.abortRequested = true
}

func (dbgp *dbgpClient) HasAbortBeenSignalled() bool {
	return dbgp.abortRequested
}

func (dbgp *dbgpClient) parseResponseXML(rawXmlData string) (dbgpxml.Response, error) {
	response := dbgpxml.Response{}

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

func (dbgp *dbgpClient) parseCtrlResponseXML(rawXmlData string) (dbgpxml.CtrlResponse, error) {
	response := dbgpxml.CtrlResponse{}

	reader := strings.NewReader(rawXmlData)

	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel

	err := decoder.Decode(&response)

	if err != nil {
		return response, err
	}

	return response, nil
}

func (dbgp *dbgpClient) parseStreamXML(rawXmlData string) (dbgpxml.Stream, error) {
	stream := dbgpxml.Stream{}

	reader := strings.NewReader(rawXmlData)

	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel

	err := decoder.Decode(&stream)

	if err != nil {
		return stream, err
	}

	return stream, nil
}

func (dbgp *dbgpClient) readResponse() (string, error, bool) {
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

func (dbgp *dbgpClient) ReadResponse() (string, error) {
	dbgp.connection.SetReadDeadline(time.Time{})

	response, err, _ := dbgp.readResponse()

	return response, err
}

func (dbgp *dbgpClient) ReadResponseWithTimeout(d time.Duration) (string, error, bool) {
	dbgp.connection.SetReadDeadline(time.Now().Add(d))

	return dbgp.readResponse()
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

	_, err := dbgp.writer.Write([]byte(line + "\000"))
	if err != nil {
		dbgp.logger.LogError("dbgp-client", "Error writing data '%s': %s", line, err.Error())
	}

	return err
}

func (dbgp *dbgpClient) FormatXML(rawXmlData string) Response {
	var response Response

	response, err := dbgp.parseResponseXML(rawXmlData)

	if err == nil {
		return response
	}

	response, err = dbgp.parseCtrlResponseXML(rawXmlData)

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

	response, err = dbgp.parseCloudStopXML(rawXmlData)

	if err == nil {
		return response
	}

	return nil
}

func (dbgp *dbgpClient) RunCommand(command string) error {
	err := dbgp.SendCommand(command)

	if err != nil { // writing failed
		return err
	}

	response, err := dbgp.ReadResponse()

	if err != nil { // reading failed
		return err
	}

	if !dbgpxml.IsValidXml(response) {
		return fmt.Errorf("The received XML is not valid, closing connection: %s", response)
	}

	formattedResponse := dbgp.FormatXML(response)

	if formattedResponse.IsSuccess() == false {
		return fmt.Errorf("%s", formattedResponse.GetErrorMessage())
	}

	if formattedResponse == nil {
		return fmt.Errorf("Could not interpret XML, closing connection.")
	}

	return nil
}

/** MERGE WITH ABOVE **/
func RunAndQuit(conn net.Conn, command string, output io.Writer, logOutput logger.Logger, showXML bool) error {
	proto := NewDbgpClient(conn, logOutput)

	err := proto.SendCommand(command)
	if err != nil {
		return fmt.Errorf("Sending %q failed: %s", command, err)
	}

	response, err := proto.ReadResponse()
	if err != nil {
		return fmt.Errorf("%q: %s", command, err)
	}

	if !dbgpxml.IsValidXml(response) {
		fmt.Fprintf(output, "The received XML is not valid: %s", response)
		return nil
	}

	if showXML {
		fmt.Fprintf(output, "%s\n", Faint(response))
	}

	formatted := proto.FormatXML(response)

	if formatted == nil {
		fmt.Fprintf(output, "Could not interpret XML")
		return nil
	}
	fmt.Fprintln(output, formatted)

	if !formatted.IsSuccess() {
		return fmt.Errorf("%q failed", command)
	}

	return nil
}
