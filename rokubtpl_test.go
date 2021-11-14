package rokubtpl

import (
	"testing"

	"github.com/muka/go-bluetooth/bluez/profile/device"
)

func Test_connectBtDevice(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"testy", false},
	}

	btDevice, err := device.NewDevice("hci0", "FC:58:FA:44:06:59")
	if err != nil {
		t.Error(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := connectBtDevice(btDevice); (err != nil) != tt.wantErr {
				t.Errorf("connectBtDevice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
