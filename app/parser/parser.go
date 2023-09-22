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
	Literal string
}

func (p Ping) String() string {
	return "PING"
}

func (p Ping) Respond() string {
	return "+PONG\r\n"
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
	cmds := make([]Command, 0)
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

				cmds = append(cmds, cmd)
			}
		} else if p.current() == '$' {
			cmd, err := p.readCommand()
			if err != nil {
				return nil, err
			}

			cmds = append(cmds, cmd)
		}
	}

	return cmds, nil
}

func (p *Parser) getCommand(str string) (Command, error) {
	switch strings.ToLower(str) {
	case "ping":
		return Ping{Literal: str}, nil
	}

	return nil, nil
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
		for i := 0; i < int(length); i++ {
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

func (p *Parser) current() byte {
	if p.currentPosition >= len(p.input) {
		return 0
	}

	return p.input[p.currentPosition]
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
