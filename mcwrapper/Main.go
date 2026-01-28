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
	"syscall"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/sys/windows"
)

// REMEMBER TO CHANGE FOR TOKENS
var PUBLIC_KEY = "7MIyc6g3LVbRU1mvqy+qZKqn3DT7cerlu9jAMJg17/M="

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

func main() {

	fmt.Println("--------------------")
	fmt.Println("MCPL Runtime Wrapper ")
	fmt.Println("--------------------")

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

	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start game: %v\n", err)
		return
	}

	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	cmd.Wait()

}
