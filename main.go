package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"shotgunner/logutil"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/TheTitanrain/w32"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
)

var keyboardHook w32.HHOOK
var wg sync.WaitGroup
var soundInProgress = false
var streamer beep.StreamSeekCloser = nil

const (
	WH_KEYBOARD_LL = 13
	WM_KEYDOWN     = 256

	SHOTGUN_SOUND = "shotgun_1.mp3"
)

func InitializeAudio(path string) {
	f, err := os.Open(path)
	if err != nil {
		logutil.Fatal(fmt.Sprintf("Failed to open audio file. %v", err))
		return
	}
	localStreamer, format, err := mp3.Decode(f)
	if err != nil {
		logutil.Fatal(fmt.Sprintf("Failed to decode audio file. %v", err))
		return
	}

	streamer = localStreamer

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
}

func DeinitializeAudio() {
	if streamer != nil {
		streamer.Close()
	}
}

func PlaySound(path string) {
	if streamer == nil {
		logutil.Warn("Streamer is nil")
		return
	}

	if soundInProgress {
		logutil.Warn("Sound is already playing. Skipping.")
		return
	}

	done := make(chan bool)
	soundInProgress = true
	logutil.Info("Starting sound playback")
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		logutil.Info("Sound playback completed")
		soundInProgress = false
		done <- true
	})))
	<-done

	err := streamer.Seek(0)
	if err != nil {
		logutil.Fatal(fmt.Sprintf("Failed to seek audio file. %v", err))
		return
	}
}

func Start() {
	keyboardHook = w32.SetWindowsHookEx(WH_KEYBOARD_LL, (w32.HOOKPROC)(func(nCode int, wparam w32.WPARAM, lparam w32.LPARAM) w32.LRESULT {
		if nCode == 0 && wparam == WM_KEYDOWN {
			kbdStruct := (*w32.KBDLLHOOKSTRUCT)(unsafe.Pointer(lparam))
			code := byte(kbdStruct.VkCode)
			if code == 'Q' {
				go PlaySound(SHOTGUN_SOUND)
			}
		}
		return w32.CallNextHookEx(keyboardHook, nCode, wparam, lparam)
	}), 0, 0)

	if keyboardHook == 0 {
		logutil.Fatal("Failed to hook SetWindowsHookEx")
		return
	}

	var msg w32.MSG
	for w32.GetMessage(&msg, 0, 0, 0) != 0 {

	}

	w32.UnhookWindowsHookEx(keyboardHook)
	keyboardHook = 0
}

func WaitForKey() {
	fmt.Println("Press any key to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func main() {
	defer WaitForKey()
	wg.Add(1)

	logutil.Info("Program started")
	logutil.Info("Press Ctrl+C to exit.")

	InitializeAudio(SHOTGUN_SOUND)
	defer DeinitializeAudio()

	go Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		logutil.Info("Exiting...")
		wg.Done()
	}()

	wg.Wait()
}
