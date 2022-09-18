package idgen

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestGenerator_nextID(t *testing.T) {
	type fields struct {
		mutex     *sync.Mutex
		machineID uint16
	}
	tests := []struct {
		name    string
		fields  fields
		want    uint64
		wantErr bool
	}{
		{
			name:    ">0",
			fields:  fields{mutex: new(sync.Mutex), machineID: 0},
			want:    0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Generator{
				mutex:     tt.fields.mutex,
				machineID: tt.fields.machineID,
			}
			got, err := g.ID()
			if (err != nil) != tt.wantErr {
				t.Errorf("nextID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got <= 0 {
				t.Errorf("nextID() firstPrefix got > 0, got: %v", got)
			}
		})
	}
}

func TestNewGenerator(t *testing.T) {
	type args struct {
		opts Options
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "new/ok",
			args: args{Options{}},
		},
		{
			name: "new/with_machine",
			args: args{Options{
				MachineID: func() (uint16, error) {
					return uint16(time.Now().Hour()), nil
				},
				CheckMachineID: nil,
			}},
		},
		{
			name: "new/with_machine/error",
			args: args{Options{
				MachineID: func() (uint16, error) {
					return uint16(time.Now().Hour()), errors.New("simulate machine err")
				},
			}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := NewGenerator(tt.args.opts); got == nil {
				if !tt.wantErr {
					t.Errorf("NewGenerator() = %v", got)
				}
			}
		})
	}
}

func TestGenerator_ID(t *testing.T) {
	type fields struct {
		mutex     *sync.Mutex
		machineID uint16
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "ok/not_empty",
			fields: fields{
				mutex:     new(sync.Mutex),
				machineID: 6,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenerator(Options{})
			got, err := g.ID()
			t.Logf("got ID: %d", got)
			if (err != nil) != tt.wantErr {
				t.Errorf("ID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == 0 {
				t.Errorf("ID() got empty")
			}
		})
	}
}

func TestExtractFromID(t *testing.T) {
	type args struct {
		id uint64
	}
	tests := []struct {
		name        string
		args        args
		firstPrefix uint64
		half2nd     uint64
		machineID   uint64
		eslapeTime  uint64
	}{
		{
			name:        "ok/extract",
			args:        args{22091819345162233},
			firstPrefix: 220918,
			half2nd:     19345162233,
			machineID:   288,
			eslapeTime:  17809401,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got1, got2 := ExtractID(tt.args.id)
			if got1 != tt.machineID {
				t.Errorf("ExtractID() got1 = %v, machineID %v", got1, tt.machineID)
			}
			if got2 != tt.eslapeTime {
				t.Errorf("ExtractID() got2 = %v, eslapeTime %v", got2, tt.eslapeTime)
			}
		})
	}
}
