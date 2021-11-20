package rokubtpl

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/go-ps"
	"github.com/picatz/roku"
	"github.com/pkg/errors"
	"github.com/runz0rd/btctl"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type BluetoothPrivateListening struct {
	log         *logrus.Entry
	host        string
	port        int
	btDevAddr   string
	bt          *btctl.BluetoothCtl
	pl          RokuPrivateListening
	ctxCancel   context.CancelFunc
	isPlStarted bool
}

func New(log *logrus.Entry, c *Config, pl RokuPrivateListening) (*BluetoothPrivateListening, error) {
	bt, err := btctl.NewBluetoothCtl()
	if err != nil {
		return nil, err
	}
	return &BluetoothPrivateListening{log: log, host: c.Roku.Host, port: c.Roku.Port, pl: pl, btDevAddr: c.BT.DestinationMacAddr, bt: bt}, nil
}

func (r BluetoothPrivateListening) IsPlStarted() bool {
	return r.isPlStarted
}

func (r BluetoothPrivateListening) IsRokuUp() bool {
	// check if roku is on and specific status is active
	info, err := DeviceInfo(fmt.Sprintf("http://%v:%v/", r.host, r.port), 5*time.Second)
	if err != nil {
		r.log.Debug("roku is down")
		return false
	}
	if info.SupportsPrivateListening == "true" {
		r.log.Debug("roku is up")
		return true
	}
	r.log.Debug("roku is up but doesnt support private listening")
	return false
}

func (r *BluetoothPrivateListening) Start() error {
	ctx, ctxCancel := context.WithCancel(context.Background())
	r.ctxCancel = ctxCancel
	isConnected, err := r.bt.IsConnected(ctx)
	if err != nil {
		return errors.WithMessage(err, "bluetooth connect check failed")
	}
	if !isConnected {
		if err := r.bt.Connect(ctx, r.btDevAddr); err != nil {
			return errors.WithMessage(err, "bluetooth connect failed")
		}
		r.log.Debug("connected bt device")
	}
	go func() {
		for {
			r.log.Debug("starting private listening")
			if err := r.pl.Start(ctx, r.host); err != nil {
				r.log.Debug(errors.WithMessage(err, "private listening failed"))
			}
			time.Sleep(3 * time.Second)
		}
	}()
	r.isPlStarted = true
	return nil
}

func (r *BluetoothPrivateListening) Stop() error {
	if r.ctxCancel != nil {
		r.ctxCancel()
		r.log.Debug("stopped private listening")
		r.isPlStarted = false
	}
	ctx := context.Background()
	isConnected, err := r.bt.IsConnected(ctx)
	if err != nil {
		return errors.WithMessage(err, "bluetooth connect check failed")
	}
	if isConnected {
		if err := r.bt.Disconnect(ctx); err != nil {
			return errors.WithMessage(err, "bluetooth disconnect failed")
		}
		r.log.Debug("disconnected bt device")
	}
	return nil
}

func findPid(parent, query string) (int, error) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("ps ax | grep %q | awk '{print $1}'", query))
	bs, err := cmd.CombinedOutput()
	output := string(bs)
	if err != nil {
		return 0, fmt.Errorf("error: %s, output: %v", err, output)
	}
	pids := strings.Split(output, "\n")
	for _, pid := range pids {
		pidInt, err := strconv.Atoi(pid)
		if err != nil {
			continue
		}
		p, err := ps.FindProcess(pidInt)
		if err != nil {
			return 0, err
		}
		if p != nil && p.Executable() == parent {
			return pidInt, nil
		}
	}
	return 0, fmt.Errorf("couldnt find %q process", query)
}

type Config struct {
	Roku struct {
		Key  string `yaml:"key,omitempty"`
		Host string `yaml:"host,omitempty"`
		Port int    `yaml:"port,omitempty"`
	} `yaml:"roku,omitempty"`
	BT struct {
		DestinationMacAddr string `yaml:"destination_mac_addr,omitempty"`
		SourceAdapterId    string `yaml:"source_adapter_id,omitempty"`
	} `yaml:"bt,omitempty"`
	PrivateListeningBinPath string `yaml:"private_listening_bin_path,omitempty"`
	Debug                   bool   `yaml:"debug,omitempty"`
	CheckDelaySec           int    `yaml:"check_delay_sec,omitempty"`
}

func LoadConfig(path string) (*Config, error) {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	c := Config{}
	if err := yaml.Unmarshal(f, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func DeviceInfo(url string, timeout time.Duration) (*roku.DeviceInfo, error) {
	client := http.Client{Timeout: timeout}
	resp, err := client.Get(url + "/query/device-info")

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	if resp.Body == nil {
		return nil, roku.ErrNoRespBody
	}

	deviceInfo := &roku.DeviceInfo{}
	decoder := xml.NewDecoder(resp.Body)
	err = decoder.Decode(deviceInfo)
	if err != nil {
		return nil, err
	}

	return deviceInfo, nil
}
