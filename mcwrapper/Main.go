package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"

	"bufio"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/sys/windows"

	"github.com/fatih/color"
)

// REMEMBER TO CHANGE FOR TOKENS
var PUBLIC_KEY = "7MIyc6g3LVbRU1mvqy+qZKqn3DT7cerlu9jAMJg17/M="

var logCounter atomic.Int64

func checkToken(issuedToken string) {
	pubByes, err := base64.StdEncoding.DecodeString(PUBLIC_KEY)
	if err != nil {
		os.Exit(1)
	}
	pubKey := ed25519.PublicKey(pubByes)
	token, err := jwt.Parse(issuedToken, func(t *jwt.Token) (interface{}, error) {
		// Ensure algorithm is EdDSA
		if t.Method.Alg() != jwt.SigningMethodEdDSA.Alg() {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return pubKey, nil
	})
	if err != nil || !token.Valid {
		fmt.Println("Invalid instance authorization token")
		os.Exit(1) // silently fail
	}
}

func supportsUnicode() bool {
	handle := windows.Handle(os.Stdout.Fd())

	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		return false // not a console (redirected)
	}

	const ENABLE_VIRTUAL_TERMINAL_PROCESSING = 0x0004

	// Try enabling VT mode (safe even if already enabled)
	if err := windows.SetConsoleMode(handle, mode|ENABLE_VIRTUAL_TERMINAL_PROCESSING); err != nil {
		return false
	}

	return true
}

func printLog(line string) {
	// Clear spinner line
	fmt.Print("\r                     \r")

	fmt.Println(line)
}

func streamOutput(r io.Reader, spinner *BottomSpinner) {
	scanner := bufio.NewScanner(r)

	infoColor := color.New(color.FgGreen)
	warnColor := color.New(color.FgYellow)
	errorColor := color.New(color.FgRed)

	spinner.SetText("Java Virtual Machine is compiling the Java bytecode for Minecraft...")

	for scanner.Scan() {
		line := scanner.Text()

		logCounter.Add(1)
		if logCounter.Load() == 1 {
			spinner.SetText("Minecraft is launching...")
		}

		if strings.Contains(line, "Sound engine started") {
			spinner.SetText("Minecraft is running...")
		}

		switch {
		case strings.Contains(line, "/INFO]"):
			printLog(infoColor.Sprint(line))
		case strings.Contains(line, "/WARN]"):
			printLog(warnColor.Sprint(line))
		case strings.Contains(line, "/ERROR]"):
			printLog(errorColor.Sprint(line))
		default:
			printLog(line)
		}
	}
}

type BottomSpinner struct {
	frames []rune
	delay  time.Duration
	stop   chan struct{}
	wg     sync.WaitGroup
	once   sync.Once

	text string
	mu   sync.RWMutex
}

func NewBottomSpinner() *BottomSpinner {
	unicodeFrames := []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
	asciiFrames := []rune{'|', '/', '-', '\\'}

	frames := asciiFrames
	if supportsUnicode() {
		frames = unicodeFrames
	}

	return &BottomSpinner{
		frames: frames,
		delay:  80 * time.Millisecond,
		stop:   make(chan struct{}),
		text:   "Launching Minecraft...",
	}
}

func (s *BottomSpinner) Start() {
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()
		i := 0

		for {
			select {
			case <-s.stop:
				fmt.Print("\r\033[K") // clear line
				return
			default:
				fmt.Printf("\r\033[K %c  %s", s.frames[i%len(s.frames)], s.text)
				time.Sleep(s.delay)
				i++
			}
		}
	}()
}

func (s *BottomSpinner) SetText(t string) {
	s.mu.Lock()
	s.text = t
	s.mu.Unlock()
}

func (s *BottomSpinner) Stop() {
	s.once.Do(func() {
		close(s.stop)
	})
	s.wg.Wait()
}

func main() {

	fmt.Println("--------------------")
	fmt.Println("MCPL Runtime Wrapper ")
	fmt.Println("--------------------")

	logCounter.Store(0)

	if len(os.Args) < 3 {
		fmt.Println("Usage: wrapper.exe <username> <instance authorization token>")
		return
	}

	username := os.Args[1]
	token := os.Args[2]

	checkToken(token)

	cwd, _ := os.Getwd()
	assetsDir := filepath.Join(cwd, "assets")

	cmd := exec.Command(
		".\\java\\bin\\java.exe",

		"-Xmx4G",
		"-Xms1G",
		"--enable-native-access=ALL-UNNAMED",

		"-cp", CLASS_PATH,
		MAIN_CLASS,

		"--accessToken", "0",
		"--version", VERSION,
		"--assetsDir", assetsDir,
		"--assetIndex", strconv.Itoa(ASSET_INDEX),
		"--username", username,
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NO_WINDOW,
	}

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	fmt.Println("Launching Minecraft... Please wait...")
	fmt.Println("> The game will keep running if you decide to close this window.")

	spinner := NewBottomSpinner()

	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start game: %v\n", err)
		return
	}

	spinner.Start()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		streamOutput(stdout, spinner)
	}()

	go func() {
		defer wg.Done()
		streamOutput(stderr, spinner)
	}()

	cmd.Wait()
	wg.Wait()

	spinner.Stop()

}
