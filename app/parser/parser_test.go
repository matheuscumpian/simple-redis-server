package parser

import (
	"reflect"
	"testing"
)

func TestParser_Parse(t *testing.T) {
	type fields struct {
		input           string
		currentPosition int
		peekPosition    int
	}
	tests := []struct {
		name    string
		fields  fields
		want    []Command
		wantErr bool
	}{
		{
			name: "ping",
			fields: fields{
				input: "*1\r\n$4\r\nping\r\n",
			},
			want: []Command{
				Ping{
					Literal:  "ping",
					Response: "+PONG\r\n",
				},
			},
			wantErr: false,
		},
		{
			name: "ping with argument",
			fields: fields{
				input: "*1\r\n$4\r\nping\r\n$12\r\nhello world\r\n",
			},
			want: []Command{
				Ping{
					Literal:  "ping",
					Response: "hello world\r\n",
				},
			},
			wantErr: false,
		},
		{
			name: "echo",
			fields: fields{
				input: "*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n",
			},
			want: []Command{
				Echo{
					Literal:  "ECHO",
					Response: "hey\r\n",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.fields.input)

			got, err := p.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() got = %v, want %v", got, tt.want)
			}
		})
	}
}
