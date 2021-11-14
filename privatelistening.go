package rokubtpl

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type RokuPrivateListening interface {
	Start(ctx context.Context, ip string) error
	Cmd(ctx context.Context, ip string) *exec.Cmd
}

type JarPrivateListening struct {
	binPath string
}

func NewJarPrivateListening(binPath string) *JarPrivateListening {
	return &JarPrivateListening{binPath}
}

func (pl JarPrivateListening) Cmd(ctx context.Context, ip string) *exec.Cmd {
	return exec.CommandContext(ctx, "java", "-jar", pl.binPath, "-i", ip)
}

func (pl JarPrivateListening) Start(ctx context.Context, ip string) error {
	return pl.Cmd(ctx, ip).Start()
}

type PyPrivateListening struct {
	audioSink string
	binPath   string
	syncDelay int
	debug     bool
}

func NewPyAudioReciever(audioSink, binPath string, syncDelay int, debug bool) *PyPrivateListening {
	return &PyPrivateListening{audioSink, binPath, syncDelay, debug}
}

func (_ PyPrivateListening) isRokuAudioReceiverConnected() bool {
	_, err := findPid("python", "roku.py")
	return err == nil
}

func (pl PyPrivateListening) Start(ctx context.Context, ip string) error {
	cmd := exec.CommandContext(ctx, "python", pl.binPath, "run",
		"-roku_ip", ip, "-audio_sink", pl.audioSink, "-audio_video_sync_delay_ms", fmt.Sprint(pl.syncDelay))
	if pl.debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
	}
	return cmd.Start()
}

// attempts rerun if process is borked
func monitorProcess(l *logrus.Entry, cmd *exec.Cmd, delay time.Duration) {
	for {
		if cmd.Process == nil {
			if err := cmd.Start(); err != nil {
				l.Debug(errors.WithMessage(err, "start failed"))
			}
		}
		if err := cmd.Process.Signal(syscall.Signal(0)); err != nil {
			l.Debug(errors.WithMessage(err, "process not up or not responding"))
			if err := cmd.Start(); err != nil {
				l.Debug(errors.WithMessage(err, "start failed"))
			}
		}
		time.Sleep(delay)
	}
}
