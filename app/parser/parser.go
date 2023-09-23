package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	PingCommand = "PING"
	EchoCommand = "ECHO"
	SetCommand  = "SET"
	GetCommand  = "GET"
)

type Command interface {
	Type() string
	Respond() string
}

// Ping command
type Ping struct {
	Literal  string
	Response string
}

func (p Ping) Type() string {
	return PingCommand
}

func (p Ping) Respond() string {
	str := strings.Builder{}
	str.WriteString("$")
	str.WriteString(strconv.Itoa(len(p.Response)))
	str.WriteString("\r\n")
	str.WriteString(p.Response)
	str.WriteString("\r\n")
	return str.String()
}

// Echo command
type Echo struct {
	Literal  string
	Response string
}

func (e Echo) Type() string {
	return EchoCommand
}

func (e Echo) Respond() string {
	str := strings.Builder{}
	str.WriteString("$")
	str.WriteString(strconv.Itoa(len(e.Response)))
	str.WriteString("\r\n")
	str.WriteString(e.Response)
	str.WriteString("\r\n")
	return str.String()
}

// Set command
type Set struct {
	Literal string

	Key   string
	Value string
}

func (s Set) Type() string {
	return SetCommand
}

func (s Set) Respond() string {
	return "+OK\r\n"
}

// Get command
type Get struct {
	Literal string

	Key   string
	Value string
}

func (g *Get) Type() string {
	return GetCommand
}

func (g *Get) SetValue(value string) {
	g.Value = value
}

func (g *Get) Respond() string {
	if g.Value != "" {
		str := strings.Builder{}
		str.WriteString("$")
		str.WriteString(strconv.Itoa(len(g.Value)))
		str.WriteString("\r\n")
		str.WriteString(g.Value)
		str.WriteString("\r\n")
		return str.String()
	}

	return "$-1\r\n"
}

type Parser struct {
	input           string
	currentPosition int
	peekPosition    int
}

func NewParser(input string) *Parser {
	return &Parser{input: input}
}

func (p *Parser) Parse() ([]Command, error) {
	commands := make([]Command, 0)
	for p.current() != 0 {
		p.advance()

		cmd, err := p.readToken()
		if err != nil {
			return nil, err
		}

		commands = append(commands, cmd)
	}

	return commands, nil
}

func (p *Parser) getCommand(str string) (Command, error) {
	switch strings.ToLower(str) {
	case "ping":
		return p.parsePing(str)
	case "echo":
		return p.parseEcho(str)
	case "set":
		return p.parseSet(str)
	case "get":
		return p.parseGet(str)
	}

	return nil, nil
}

func (p *Parser) parseSet(str string) (Command, error) {
	cmd := Set{
		Literal: str,
	}

	// Get key
	if p.current() == '$' {
		str, err := p.parseBulkString()
		if err != nil {
			return nil, err
		}

		cmd.Key = str
	}

	// Get value
	if p.current() == '$' {
		str, err := p.parseBulkString()
		if err != nil {
			return nil, err
		}

		cmd.Value = str
	}

	return cmd, nil
}

func (p *Parser) parseGet(str string) (Command, error) {
	cmd := &Get{
		Literal: str,
	}

	// Get key
	if p.current() == '$' {
		str, err := p.parseBulkString()
		if err != nil {
			return nil, err
		}

		cmd.Key = str
	}

	return cmd, nil
}

func (p *Parser) parseEcho(str string) (Command, error) {
	cmd := Echo{
		Literal: str,
	}

	// Get argument if it exists
	if p.current() == '$' {
		str, err := p.parseBulkString()
		if err != nil {
			return nil, err
		}

		cmd.Response = str
	}

	return cmd, nil
}

func (p *Parser) parsePing(str string) (Command, error) {
	cmd := Ping{
		Literal:  str,
		Response: "PONG",
	}

	// Get argument if it exists
	if p.current() == '$' {
		str, err := p.parseBulkString()
		if err != nil {
			return nil, err
		}

		cmd.Response = str
	}

	return cmd, nil
}

func (p *Parser) readToken() (Command, error) {
	if p.current() == '$' {
		str, err := p.parseBulkString()
		if err != nil {
			return nil, err
		}

		cmd, err := p.getCommand(str)
		if err != nil {
			return nil, err
		}

		return cmd, nil
	}

	return nil, errors.New("invalid command")
}

func (p *Parser) parseBulkString() (string, error) {
	p.advance()

	length, err := p.readSize()
	if err != nil {
		return "", err
	}

	p.advance()

	str := strings.Builder{}
	for i := 0; i < length; i++ {
		str.WriteByte(p.current())
		p.advance()
	}

	return str.String(), nil
}

func (p *Parser) readSize() (int, error) {
	str := strings.Builder{}
	for p.current() != '\r' {
		str.WriteByte(p.current())
		p.movePointers()
	}

	length, err := strconv.Atoi(str.String())
	if err != nil {
		return 0, fmt.Errorf("invalid length of command, received $%s, expeced $INTEGER", string(p.current()))
	}

	return length, nil
}

func (p *Parser) current() byte {
	if p.currentPosition >= len(p.input) {
		return 0
	}

	return p.input[p.currentPosition]
}

func (p *Parser) movePointers() {
	p.currentPosition = p.peekPosition
	p.peekPosition++
}

func (p *Parser) advance() {
	p.movePointers()

	p.skipArrays()
	p.skipCRLF()
}

// I'm skipping arrays because I don't know how to parse them yet
func (p *Parser) skipArrays() {
	if p.current() == '*' {
		for p.current() != '\r' {
			p.movePointers()
		}
	}
}

func (p *Parser) skipCRLF() {
	for p.current() == '\r' || p.current() == '\n' {
		p.movePointers()
	}
}

func (p *Parser) peekChar() byte {
	if p.peekPosition >= len(p.input) {
		return 0
	}

	return p.input[p.peekPosition]
}
