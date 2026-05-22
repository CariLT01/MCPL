package main

import (
	"bufio"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/sys/windows"
)

// REMEMBER TO CHANGE FOR TOKENS
var PUBLIC_KEY = "km9A20ohDKhguLGUYghGaI3Xl3JiJK4v3g2/XYpKtvk="
var logCounter atomic.Int64
var consoleMu sync.Mutex

func shimmerText(text string, tick int) string {
	runes := []rune(text)
	n := len(runes)
	if n == 0 {
		return text
	}

	// Adjust speed: moving 1 character every 2 ticks
	shift := (tick / 2) % n

	var b strings.Builder
	b.Grow(len(text) * 32) // ANSI codes are heavy

	for i := range runes {
		// Calculate circular distance (how far 'i' is ahead of 'shift')
		// This creates a smooth "trailing" effect.
		dist := i - shift
		if dist < 0 {
			dist += n
		}

		// Use a explicit Reset (\x1b[0m) before the peak to ensure
		// "Faint" states from previous characters don't bleed over.
		switch dist {
		case 0:
			// The Peak: Bold Bright White
			b.WriteString("\x1b[0m\x1b[1;97m")
		case 1, n - 1:
			// Soft Edge: Bright White (no bold)
			b.WriteString("\x1b[0m\x1b[97m")
		case 2, n - 2:
			// Fade: Normal White
			b.WriteString("\x1b[0m\x1b[37m")
		default:
			// Background: Gray (Avoiding the '2;' code for better compatibility)
			b.WriteString("\x1b[0m\x1b[90m")
		}
		b.WriteRune(runes[i])
	}

	b.WriteString("\x1b[0m") // Final reset
	return b.String()
}

func formatElapsed(d time.Duration) string {
	totalSeconds := int(d.Seconds())
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func checkToken(issuedToken string) {
	pubBytes, err := base64.StdEncoding.DecodeString(PUBLIC_KEY)
	if err != nil {
		os.Exit(1)
	}
	pubKey := ed25519.PublicKey(pubBytes)
	token, err := jwt.Parse(issuedToken, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodEdDSA.Alg() {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return pubKey, nil
	})
	if err != nil || !token.Valid {
		fmt.Println("Invalid instance authorization token")
		os.Exit(1)
	}
}

func supportsUnicode() bool {
	handle := windows.Handle(os.Stdout.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		return false
	}
	const ENABLE_VIRTUAL_TERMINAL_PROCESSING = 0x0004
	if err := windows.SetConsoleMode(handle, mode|ENABLE_VIRTUAL_TERMINAL_PROCESSING); err != nil {
		return false
	}
	return true
}

func printLog(line string) {
	consoleMu.Lock()
	defer consoleMu.Unlock()

	// Clear the full line (\033[2K) and return to start (\r) to prevent flickering
	fmt.Print("\r\033[2K")
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
	frames    []rune
	delay     time.Duration
	stop      chan struct{}
	wg        sync.WaitGroup
	once      sync.Once
	startedAt time.Time
	text      string
	mu        sync.RWMutex
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
		delay:  60 * time.Millisecond,
		stop:   make(chan struct{}),
		text:   "Launching Minecraft...",
	}
}

func (s *BottomSpinner) Start() {
	s.startedAt = time.Now()
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.delay)
		defer ticker.Stop()

		i := 0
		for {
			select {
			case <-s.stop:
				consoleMu.Lock()
				fmt.Print("\r\033[2K")
				consoleMu.Unlock()
				return
			case <-ticker.C:
				s.mu.RLock()
				text := s.text
				s.mu.RUnlock()

				frame := s.frames[i%len(s.frames)]
				rendered := shimmerText(text, i)
				elapsed := formatElapsed(time.Since(s.startedAt))

				consoleMu.Lock()
				// \r (return) + \033[K (clear to end of line) keeps the update clean
				fmt.Printf("\r\033[K %c  %s  \x1b[1;95m[%s]\x1b[0m", frame, rendered, elapsed)
				consoleMu.Unlock()

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
	if len(os.Args) < 3 {
		fmt.Println("Usage: wrapper.exe <username> <instance authorization token>")
		return
	}

	fmt.Println("--------------------")
	fmt.Println("MCPL Runtime Wrapper ")
	fmt.Println("--------------------")

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
