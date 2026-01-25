package protocol_handler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Register registers the agent:// protocol handler for the current platform
func Register() error {
	switch runtime.GOOS {
	case "darwin":
		return registerMacOS()
	case "linux":
		return registerLinux()
	case "windows":
		return registerWindows()
	default:
		return fmt.Errorf("protocol handler registration not supported on %s", runtime.GOOS)
	}
}

// IsRegistered checks if the agent:// protocol handler is already registered
func IsRegistered() bool {
	switch runtime.GOOS {
	case "darwin":
		return isRegisteredMacOS()
	case "linux":
		return isRegisteredLinux()
	case "windows":
		return isRegisteredWindows()
	default:
		return false
	}
}

// registerMacOS registers the protocol handler on macOS
func registerMacOS() error {
	// Get the path to the current executable
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks
	realPath, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		realPath = exePath
	}

	// Create a minimal .app bundle to register as protocol handler
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	appPath := filepath.Join(homeDir, "Library", "Application Support", "ai-pack", "AI-Pack Agent Handler.app")
	contentsDir := filepath.Join(appPath, "Contents")
	macOSDir := filepath.Join(contentsDir, "MacOS")

	// Create app bundle structure
	if err := os.MkdirAll(macOSDir, 0755); err != nil {
		return fmt.Errorf("failed to create app bundle: %w", err)
	}

	// Create Info.plist with protocol handler registration
	infoPlist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleExecutable</key>
	<string>agent-handler</string>
	<key>CFBundleIdentifier</key>
	<string>com.aipack.agent-handler</string>
	<key>CFBundleName</key>
	<string>AI-Pack Agent Handler</string>
	<key>CFBundleVersion</key>
	<string>1.0</string>
	<key>CFBundleURLTypes</key>
	<array>
		<dict>
			<key>CFBundleURLName</key>
			<string>AI-Pack Agent Protocol</string>
			<key>CFBundleURLSchemes</key>
			<array>
				<string>agent</string>
			</array>
		</dict>
	</array>
</dict>
</plist>`)

	if err := os.WriteFile(filepath.Join(contentsDir, "Info.plist"), []byte(infoPlist), 0644); err != nil {
		return fmt.Errorf("failed to create Info.plist: %w", err)
	}

	// Create wrapper script that calls agent-server
	wrapperScript := fmt.Sprintf(`#!/bin/bash
# AI-Pack agent:// protocol handler
exec "%s" "$@"
`, realPath)

	wrapperPath := filepath.Join(macOSDir, "agent-handler")
	if err := os.WriteFile(wrapperPath, []byte(wrapperScript), 0755); err != nil {
		return fmt.Errorf("failed to create wrapper: %w", err)
	}

	// Register the app bundle with Launch Services
	cmd := exec.Command("/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister",
		"-f", appPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to register with Launch Services: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("✅ Protocol handler registered: %s\n", appPath)
	return nil
}

// isRegisteredMacOS checks if the handler is registered on macOS
func isRegisteredMacOS() bool {
	// Check if our app bundle exists
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	appPath := filepath.Join(homeDir, "Library", "Application Support", "ai-pack", "AI-Pack Agent Handler.app")
	if _, err := os.Stat(appPath); err != nil {
		return false
	}

	// Verify it's registered with Launch Services
	cmd := exec.Command("/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister", "-dump")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// Check if our handler is in the output
	return strings.Contains(string(output), "ai-pack-agent-handler") || strings.Contains(string(output), "com.aipack.agent-handler")
}

// registerLinux registers the protocol handler on Linux
func registerLinux() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create .desktop file
	desktopDir := filepath.Join(homeDir, ".local", "share", "applications")
	if err := os.MkdirAll(desktopDir, 0755); err != nil {
		return fmt.Errorf("failed to create applications directory: %w", err)
	}

	desktopFile := filepath.Join(desktopDir, "ai-pack-agent-handler.desktop")
	desktopContent := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=AI-Pack Agent Handler
Exec=%s %%u
MimeType=x-scheme-handler/agent;
NoDisplay=true
`, exePath)

	if err := os.WriteFile(desktopFile, []byte(desktopContent), 0644); err != nil {
		return fmt.Errorf("failed to create desktop file: %w", err)
	}

	// Update desktop database
	if err := exec.Command("update-desktop-database", desktopDir).Run(); err != nil {
		fmt.Printf("⚠️  Could not update desktop database: %v\n", err)
		fmt.Println("   Run manually: update-desktop-database ~/.local/share/applications")
	}

	// Set as default handler
	if err := exec.Command("xdg-mime", "default", "ai-pack-agent-handler.desktop", "x-scheme-handler/agent").Run(); err != nil {
		fmt.Printf("⚠️  Could not set default handler: %v\n", err)
		fmt.Println("   Run manually: xdg-mime default ai-pack-agent-handler.desktop x-scheme-handler/agent")
	}

	fmt.Printf("✅ Protocol handler registered: %s\n", desktopFile)
	return nil
}

// isRegisteredLinux checks if the handler is registered on Linux
func isRegisteredLinux() bool {
	cmd := exec.Command("xdg-mime", "query", "default", "x-scheme-handler/agent")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "ai-pack-agent-handler.desktop")
}

// registerWindows registers the protocol handler on Windows
func registerWindows() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Use reg.exe to add registry keys
	commands := [][]string{
		{"reg", "add", "HKCU\\Software\\Classes\\agent", "/ve", "/d", "URL:AI-Pack Agent Protocol", "/f"},
		{"reg", "add", "HKCU\\Software\\Classes\\agent", "/v", "URL Protocol", "/d", "", "/f"},
		{"reg", "add", "HKCU\\Software\\Classes\\agent\\shell\\open\\command", "/ve", "/d", fmt.Sprintf("\"%s\" \"%%1\"", exePath), "/f"},
	}

	for _, cmdArgs := range commands {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to register protocol: %w", err)
		}
	}

	fmt.Println("✅ Protocol handler registered in Windows registry")
	return nil
}

// isRegisteredWindows checks if the handler is registered on Windows
func isRegisteredWindows() bool {
	cmd := exec.Command("reg", "query", "HKCU\\Software\\Classes\\agent")
	return cmd.Run() == nil
}
