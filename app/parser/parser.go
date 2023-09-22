package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Command interface {
	String() string
	Respond() string
}

type Ping struct {
	Literal  string
	Response string
}

func (p Ping) String() string {
	return "PING"
}

func (p Ping) Respond() string {
	return p.Response
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
		if p.current() == '*' {
			p.advance()

			length, err := strconv.Atoi(string(p.current()))
			if err != nil {
				return nil, fmt.Errorf("invalid length of array, received *%s, expeced *INTEGER", string(p.current()))
			}

			p.advance()

			for i := 0; i < length; i++ {
				cmd, err := p.readCommand()
				if err != nil {
					return nil, err
				}

				commands = append(commands, cmd)
			}
		} else if p.current() == '$' {
			cmd, err := p.readCommand()
			if err != nil {
				return nil, err
			}

			commands = append(commands, cmd)
		}
	}

	return commands, nil
}

func (p *Parser) getCommand(str string) (Command, error) {
	switch strings.ToLower(str) {
	case "ping":
		return p.parsePing(str)
	}

	return nil, nil
}

func (p *Parser) parsePing(str string) (Command, error) {
	cmd := Ping{
		Literal:  str,
		Response: "+PONG\r\n",
	}

	if p.current() == '$' {
		p.advance()

		length, err := p.readLength()
		if err != nil {
			return nil, err
		}

		p.advance()

		pingResponse := strings.Builder{}
		for i := 0; i < length; i++ {
			pingResponse.WriteByte(p.current())
			p.advance()
		}

		pingResponse.WriteString("\r\n")

		cmd.Response = pingResponse.String()
	}

	return cmd, nil
}

func (p *Parser) readCommand() (Command, error) {
	if p.current() == '$' {
		p.advance()

		length, err := strconv.Atoi(string(p.current()))
		if err != nil {
			return nil, fmt.Errorf("invalid length of command, received $%s, expeced $INTEGER", string(p.current()))
		}

		p.advance()

		command := strings.Builder{}
		for i := 0; i < length; i++ {
			command.WriteByte(p.current())
			p.advance()
		}

		cmd, err := p.getCommand(command.String())
		if err != nil {
			return nil, err
		}

		return cmd, nil
	}
	return nil, errors.New("invalid command")
}

func (p *Parser) readLength() (int, error) {
	str := strings.Builder{}
	for p.current() != '\r' {
		str.WriteByte(p.current())
		p.advanceWithoutSkippingCRLF()
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

func (p *Parser) advanceWithoutSkippingCRLF() {
	p.currentPosition = p.peekPosition
	p.peekPosition++
}

func (p *Parser) advance() {
	p.currentPosition = p.peekPosition
	p.peekPosition++

	for p.current() == '\r' || p.current() == '\n' {
		p.currentPosition = p.peekPosition
		p.peekPosition++
	}
}

func (p *Parser) peekChar() byte {
	if p.peekPosition >= len(p.input) {
		return 0
	}

	return p.input[p.peekPosition]
}
