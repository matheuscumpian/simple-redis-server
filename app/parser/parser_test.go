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
					Literal: "ping",
				},
			},
			wantErr: false,
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
