package dbgp

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
)

type Response interface {
	String() string
	IsSuccess() bool
}

type dbgpClient struct {
	reader          *bufio.Reader
	writer          io.Writer
	counter         int
	lastSourceBegin int
}

func NewDbgpClient(c net.Conn) *dbgpClient {
	var tmp dbgpClient

	tmp.reader = bufio.NewReader(c)
	tmp.writer = c
	tmp.counter = 1
	tmp.lastSourceBegin = 1

	return &tmp
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

func (dbgp *dbgpClient) ReadResponse() (string, error) {
	/* Read length */
	_, err := dbgp.reader.ReadBytes('\000')

	if err != nil {
		return "", fmt.Errorf("Error reading length: %s", err)
	}

	/* Read data */
	data, err := dbgp.reader.ReadBytes('\000')

	if err != nil {
		return "", fmt.Errorf("Error reading data: %s", err)
	}

	return string(data), nil
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

func (dbgp *dbgpClient) FormatXML(rawXmlData string) (Response, bool) {
	var response Response

	response, err := dbgp.parseResponseXML(rawXmlData)

	if err == nil {
		return response, false
	}

	response, err = dbgp.ParseInitXML(rawXmlData)

	if err == nil {
		return response, false
	}

	response, err = dbgp.parseNotifyXML(rawXmlData)

	if err == nil {
		return response, true
	}

	response, err = dbgp.parseStreamXML(rawXmlData)

	if err == nil {
		return response, true
	}

	response, err = dbgp.parseProxyInitXML(rawXmlData)

	if err == nil {
		return response, false
	}

	response, err = dbgp.parseProxyStopXML(rawXmlData)

	if err == nil {
		return response, false
	}

	return nil, false
}
