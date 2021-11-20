package rokubtpl

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/pkg/errors"
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
	if out, err := pl.Cmd(ctx, ip).CombinedOutput(); err != nil {
		return errors.WithMessage(err, string(out))
	}
	return nil
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
