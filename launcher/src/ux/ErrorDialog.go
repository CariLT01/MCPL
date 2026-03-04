package ux

import (
	"golang.org/x/sys/windows"
)

func ShowError(title string, message string) {
	titleText, _ := windows.UTF16FromString(title)
	messageText, _ := windows.UTF16FromString(message)

	const MB_OK = 0x00000000
	const MB_ICONERROR = 0x00000010
	const MB_TOPMOST = 0x00040000

	// Combine flags
	flags := MB_OK | MB_ICONERROR | MB_TOPMOST

	// hwnd = 0 means no owner, or use your window handle here
	windows.MessageBox(0, &messageText[0], &titleText[0], uint32(flags))
}

func ShowErrorLog(err error, while string) {
	ShowError("Application Error during an operation: "+while, "Application error that may or may not be fatal depending on the context.\nOperation: "+while+"\nError: "+err.Error())
}
