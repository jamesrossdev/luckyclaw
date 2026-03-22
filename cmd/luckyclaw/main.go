// LuckyClaw - Ultra-lightweight personal AI agent
// Inspired by and based on nanobot: https://github.com/HKUDS/nanobot
// License: MIT
//
// Copyright (c) 2026 LuckyClaw contributors

package main

import (
	"bufio"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata" // Embed timezone database for devices without /usr/share/zoneinfo

	"github.com/chzyer/readline"
	"github.com/jamesrossdev/luckyclaw/pkg/agent"
	"github.com/jamesrossdev/luckyclaw/pkg/auth"
	"github.com/jamesrossdev/luckyclaw/pkg/bus"
	"github.com/jamesrossdev/luckyclaw/pkg/channels"
	"github.com/jamesrossdev/luckyclaw/pkg/config"
	"github.com/jamesrossdev/luckyclaw/pkg/cron"
	"github.com/jamesrossdev/luckyclaw/pkg/devices"
	"github.com/jamesrossdev/luckyclaw/pkg/health"
	"github.com/jamesrossdev/luckyclaw/pkg/heartbeat"
	"github.com/jamesrossdev/luckyclaw/pkg/logger"
	"github.com/jamesrossdev/luckyclaw/pkg/migrate"
	"github.com/jamesrossdev/luckyclaw/pkg/providers"
	"github.com/jamesrossdev/luckyclaw/pkg/skills"
	"github.com/jamesrossdev/luckyclaw/pkg/state"
	"github.com/jamesrossdev/luckyclaw/pkg/tools"
	"github.com/jamesrossdev/luckyclaw/pkg/voice"

	"github.com/jamesrossdev/luckyclaw/firmware"

	// Native WhatsApp
	"github.com/jamesrossdev/luckyclaw/pkg/channels/whatsapp"
)

//go:generate cp -r ../../workspace .
//go:embed workspace
var embeddedFiles embed.FS

var (
	version   = "v0.2.2-alpha-1"
	gitCommit string
	buildTime string
	goVersion string
)

const logo = "🦞"

// formatVersion returns the version string with optional git commit
func formatVersion() string {
	v := version
	if gitCommit != "" {
		v += fmt.Sprintf(" (git: %s)", gitCommit)
	}
	return v
}

// formatBuildInfo returns build time and go version info
func formatBuildInfo() (build string, goVer string) {
	if buildTime != "" {
		build = buildTime
	}
	goVer = goVersion
	if goVer == "" {
		goVer = runtime.Version()
	}
	return
}

func printVersion() {
	fmt.Printf("%s luckyclaw %s\n", logo, formatVersion())
	build, goVer := formatBuildInfo()
	if build != "" {
		fmt.Printf("  Build: %s\n", build)
	}
	if goVer != "" {
		fmt.Printf("  Go: %s\n", goVer)
	}
}

func copyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}

// applyPerformanceDefaults sets memory-optimized Go runtime settings.
// These are baked into the binary so users on memory-constrained devices
// (e.g. Luckfox Pico Plus with 64MB DDR2) don't need to set env vars.
// Env vars GOGC and GOMEMLIMIT can still override these if set.
func applyPerformanceDefaults() {
	if os.Getenv("GOGC") == "" {
		debug.SetGCPercent(20)
	}
	if os.Getenv("GOMEMLIMIT") == "" {
		debug.SetMemoryLimit(8 * 1024 * 1024) // 8MiB
	}
}

func main() {
	applyPerformanceDefaults()

	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "onboard":
		onboard()
	case "agent":
		agentCmd()
	case "gateway":
		gatewayCmd()
	case "stop":
		stopCmd()
	case "restart":
		restartCmd()
	case "install":
		installCmd()
	case "status":
		statusCmd()
	case "migrate":
		migrateCmd()
	case "auth":
		authCmd()
	case "cron":
		cronCmd()
	case "skills":
		if len(os.Args) < 3 {
			skillsHelp()
			return
		}

		subcommand := os.Args[2]

		cfg, err := loadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		workspace := cfg.WorkspacePath()
		installer := skills.NewSkillInstaller(workspace)
		// 获取全局配置目录和内置 skills 目录
		globalDir := filepath.Dir(getConfigPath())
		globalSkillsDir := filepath.Join(globalDir, "skills")
		builtinSkillsDir := filepath.Join(globalDir, "luckyclaw", "skills")
		skillsLoader := skills.NewSkillsLoader(workspace, globalSkillsDir, builtinSkillsDir)

		switch subcommand {
		case "list":
			skillsListCmd(skillsLoader)
		case "install":
			skillsInstallCmd(installer)
		case "remove", "uninstall":
			if len(os.Args) < 4 {
				fmt.Println("Usage: luckyclaw skills remove <skill-name>")
				return
			}
			skillsRemoveCmd(installer, os.Args[3])
		case "install-builtin":
			skillsInstallBuiltinCmd(workspace)
		case "list-builtin":
			skillsListBuiltinCmd()
		case "search":
			skillsSearchCmd(installer)
		case "show":
			if len(os.Args) < 4 {
				fmt.Println("Usage: luckyclaw skills show <skill-name>")
				return
			}
			skillsShowCmd(skillsLoader, os.Args[3])
		default:
			fmt.Printf("Unknown skills command: %s\n", subcommand)
			skillsHelp()
		}
	case "version", "--version", "-v":
		printVersion()
	case "help", "--help", "-h":
		printHelp()
	case "config-reset":
		configResetCmd()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Printf("%s luckyclaw - Personal AI Assistant v%s\n\n", logo, version)
	fmt.Println("Usage: luckyclaw <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  onboard       Initialize luckyclaw configuration and workspace")
	fmt.Println("  agent         Interact with the agent directly")
	fmt.Println("  gateway       Start luckyclaw gateway (-b for background)")
	fmt.Println("  stop          Stop running gateway")
	fmt.Println("  restart       Restart gateway (stop + start in background)")
	fmt.Println("  install       Install init script and SSH banner to system")
	fmt.Println("  status        Show luckyclaw status")
	fmt.Println("  cron          Manage scheduled tasks")
	fmt.Println("  auth          Manage authentication (login, logout, status)")
	fmt.Println("  migrate       Migrate from OpenClaw to LuckyClaw")
	fmt.Println("  skills        Manage skills (install, list, remove)")
	fmt.Println("  config-reset  Safely delete your config.json (API keys/models)")
	fmt.Println("  version       Show version information")
	fmt.Println("  help          Show this help message")
}

// promptLine reads a single line from stdin with a prompt.
func promptLine(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

// promptYN asks a yes/no question, returns true for yes.
func promptYN(prompt string) bool {
	resp := strings.ToLower(promptLine(prompt + " (y/n): "))
	return resp == "y" || resp == "yes"
}

// validateOpenRouterKey tests an OpenRouter API key via the /auth/key endpoint.
// This only checks key validity — doesn't need a model, so it can't 404.
func validateOpenRouterKey(apiKey string) error {
	req, _ := http.NewRequest("GET", "https://openrouter.ai/api/v1/auth/key", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return fmt.Errorf("invalid API key")
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error (HTTP %d)", resp.StatusCode)
	}
	return nil
}

// validateTelegramToken checks a Telegram bot token via the getMe API.
func validateTelegramToken(token string) (string, error) {
	resp, err := http.Get("https://api.telegram.org/bot" + token + "/getMe")
	if err != nil {
		return "", fmt.Errorf("connection failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Ok     bool `json:"ok"`
		Result struct {
			Username string `json:"username"`
		} `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	if !result.Ok {
		return "", fmt.Errorf("invalid bot token")
	}
	return result.Result.Username, nil
}

// detectBoardModel reads the device tree model string.
func detectBoardModel() string {
	data, err := os.ReadFile("/proc/device-tree/model")
	if err != nil {
		return ""
	}
	return strings.TrimRight(string(data), "\x00\n")
}

func onboard() {
	configPath := getConfigPath()

	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════╗")
	fmt.Println("  ║  🦞 LuckyClaw Setup Wizard           ║")
	fmt.Printf("  ║  v%-35s║\n", version)
	fmt.Println("  ╚══════════════════════════════════════╝")
	fmt.Println()

	// Show hardware info
	model := detectBoardModel()
	if model != "" {
		fmt.Printf("  Board: %s\n", model)
	}
	// Read memory from /proc/meminfo
	if memData, err := os.ReadFile("/proc/meminfo"); err == nil {
		lines := strings.Split(string(memData), "\n")
		var total, avail int
		for _, line := range lines {
			if strings.HasPrefix(line, "MemTotal:") {
				fmt.Sscanf(line, "MemTotal: %d", &total)
			}
			if strings.HasPrefix(line, "MemAvailable:") {
				fmt.Sscanf(line, "MemAvailable: %d", &avail)
			}
		}
		if total > 0 {
			fmt.Printf("  Memory: %dMB available / %dMB total\n", avail/1024, total/1024)
		}
	}
	fmt.Println()

	// Build config in memory — only save at the very end (atomic)
	cfg := config.DefaultConfig()

	// Step 1: OpenRouter API Key
	fmt.Println("  Step 1: OpenRouter API Key")
	fmt.Println("  ──────────────────────────")
	fmt.Println("  Get your key at: https://openrouter.ai/keys")
	fmt.Println()

	apiKey := promptLine("  API Key: ")
	if apiKey == "" {
		fmt.Println("  Skipped — edit ~/.luckyclaw/config.json later.")
	} else {
		cfg.Providers.OpenRouter.APIKey = apiKey
		cfg.Providers.OpenRouter.APIBase = "https://openrouter.ai/api/v1"

		fmt.Print("  Validating... ")
		if err := validateOpenRouterKey(apiKey); err != nil {
			fmt.Printf("⚠ %v\n", err)
			fmt.Println("  (Key saved anyway — check it later)")
		} else {
			fmt.Println("✓")
		}
	}

	// Step 2: Model
	fmt.Println()
	fmt.Println("  Step 2: Model")
	fmt.Println("  ─────────────")
	fmt.Printf("  Default: %s\n", cfg.Agents.Defaults.Model)
	fmt.Println("  See README.md for how to choose a model.")
	fmt.Println()

	if customModel := promptLine("  Model ID (or Enter for default): "); customModel != "" {
		cfg.Agents.Defaults.Model = customModel
	}
	cfg.Agents.Defaults.Provider = "openrouter"
	cfg.Agents.Defaults.MaxTokens = 4096
	cfg.Agents.Defaults.MaxToolIterations = 10

	// Step 3: Timezone
	fmt.Println()
	fmt.Println("  Step 3: Timezone")
	fmt.Println("  ─────────────────")
	fmt.Println("  Find your Timezone code here:")
	fmt.Println("  https://en.wikipedia.org/wiki/List_of_tz_database_time_zones#List")
	fmt.Println()
	detectedTZ := promptLine("  Enter timezone (e.g. Europe/London): ")
	if detectedTZ != "" {
		// Calculate explicit offset for config
		if loc, err := time.LoadLocation(detectedTZ); err == nil {
			_, offset := time.Now().In(loc).Zone()
			cfg.Gateway.TimezoneName = detectedTZ
			cfg.Gateway.UTCOffset = offset

			// Generate POSIX string (inverted sign)
			posixTZ := getPosixTZ(offset)
			os.Setenv("TZ", posixTZ)

			if f, err := os.OpenFile("/etc/profile.d/timezone.sh", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err == nil {
				f.WriteString(fmt.Sprintf("export TZ='%s'\n", posixTZ))
				f.Close()
				fmt.Printf("  Timezone set: %s ✓\n", detectedTZ)
			} else {
				fmt.Printf("  Timezone set for this session: %s\n", detectedTZ)
			}
		} else {
			fmt.Printf("  Warning: Could not parse timezone %s\n", detectedTZ)
			cfg.Gateway.TimezoneName = "UTC"
			cfg.Gateway.UTCOffset = 0
		}
	}

	// Step 4: Messaging Channels
	fmt.Println()
	fmt.Println("  Step 4: Messaging Channels")
	fmt.Println("  ──────────────────────────")
	fmt.Println("  Set up your chat channels. You can enable multiple platforms.")
	fmt.Println()

	for {
		fmt.Println("  Select a platform to configure:")
		fmt.Println("  1. Telegram")
		fmt.Println("  2. Discord")
		fmt.Println("  3. WhatsApp (Native QR pairing)")
		fmt.Println("  4. Next Step (Skip/Proceed)")
		fmt.Println()

		choice := promptLine("  Choice (1-4): ")
		if choice == "4" || choice == "" {
			break
		}

		switch choice {
		case "1":
			fmt.Println("\n  [Telegram Setup]")
			fmt.Println("  Create a bot via @BotFather on Telegram, then paste the token below.")
			tgToken := promptLine("  Telegram bot token: ")
			if tgToken != "" {
				fmt.Print("  Validating... ")
				username, err := validateTelegramToken(tgToken)
				if err != nil {
					fmt.Printf("⚠ %v\n", err)
					fmt.Println("  (Token saved anyway — check it later)")
				} else {
					fmt.Printf("✓ @%s\n", username)
				}

				cfg.Channels.Telegram.Enabled = true
				cfg.Channels.Telegram.Token = tgToken

				tgUserID := promptLine("  Your Telegram user ID (optional, from @userinfobot): ")
				if tgUserID != "" {
					cfg.Channels.Telegram.AllowFrom = config.FlexibleStringSlice{tgUserID}
				}
				fmt.Println("  ✓ Telegram configured")
			}
		case "2":
			fmt.Println("\n  [Discord Setup]")
			fmt.Println("  Create a bot at https://discord.com/developers/applications")
			fmt.Println("  Enable MESSAGE CONTENT INTENT in Bot settings, then paste the token.")
			dcToken := promptLine("  Discord bot token: ")
			if dcToken != "" {
				cfg.Channels.Discord.Enabled = true
				cfg.Channels.Discord.Token = dcToken

				dcUserID := promptLine("  Your Discord user ID (optional): ")
				if dcUserID != "" {
					cfg.Channels.Discord.AllowFrom = config.FlexibleStringSlice{dcUserID}
				}
				fmt.Println("  ✓ Discord configured")
			}
		case "3":
			fmt.Println("\n  [WhatsApp Setup]")
			fmt.Println("  This will use native WhatsApp Web pairing.")
			if promptYN("  Enable WhatsApp native channel?") {
				cfg.Channels.WhatsApp.Enabled = true
				cfg.Channels.WhatsApp.SessionPath = filepath.Join(cfg.WorkspacePath(), "whatsapp")

				waUserID := promptLine("  Your WhatsApp number (optional, press Enter to allow ALL numbers): ")
				var expectedCode string
				if waUserID != "" {
					codeLetters := []rune("0123456789")
					code := make([]rune, 4)
					for i := range code {
						code[i] = codeLetters[rand.Intn(len(codeLetters))]
					}
					expectedCode = string(code)
				}

				lid, err := whatsapp.PerformSetup(cfg.Channels.WhatsApp.SessionPath, expectedCode)

				if err != nil {
					fmt.Printf("  ⚠ WhatsApp Setup failed: %v\n", err)
					cfg.Channels.WhatsApp.Enabled = false
				} else if expectedCode != "" && lid != "" {
					fmt.Printf("  ✓ Success! Number authorized (Local ID linked).\n")
					cfg.Channels.WhatsApp.AllowFrom = config.FlexibleStringSlice{lid}
				} else {
					cfg.Channels.WhatsApp.AllowFrom = nil
				}

				fmt.Println("  ✓ WhatsApp enabled")
			}
		default:
			fmt.Println("  Invalid choice, please try again.")
		}
		fmt.Println("\n  ──────────────────────────")
	}

	// Atomic save — everything at once
	fmt.Println()
	if err := config.SaveConfig(configPath, cfg); err != nil {
		fmt.Printf("  Error saving config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Config saved: %s ✓\n", configPath)

	// Create workspace
	workspace := cfg.WorkspacePath()
	createWorkspaceTemplates(workspace)

	// Patch USER.md with selected timezone
	if detectedTZ != "" {
		userMD := filepath.Join(workspace, "USER.md")
		if data, err := os.ReadFile(userMD); err == nil {
			patched := strings.Replace(string(data), "(set during onboarding)", detectedTZ, 1)
			os.WriteFile(userMD, []byte(patched), 0644)
		}
	}

	fmt.Printf("  Workspace ready: %s ✓\n", workspace)

	// Start gateway?
	fmt.Println()
	if promptYN("  Start LuckyClaw gateway now?") {
		gatewayStartBackground()
		if cfg.Channels.WhatsApp.Enabled {
			fmt.Println("\n  🦞 WhatsApp pairing is enabled.")
			fmt.Println("  The QR code has been written to: /var/log/luckyclaw.log")
			fmt.Println("  Or run 'luckyclaw stop && luckyclaw gateway' to see it here.")
		}
	}

	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════╗")
	fmt.Println("  ║  🦞 LuckyClaw is ready!              ║")
	fmt.Println("  ╚══════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("  Commands:")
	fmt.Println("    luckyclaw status     — Check system status")
	fmt.Println("    luckyclaw gateway    — Start the gateway")
	fmt.Println("    luckyclaw gateway -b — Start in background")
	fmt.Println("    luckyclaw stop       — Stop the gateway")
	fmt.Println("    luckyclaw restart    — Restart the gateway")
	fmt.Println("    luckyclaw help       — View more commands")
	fmt.Println()
}

// stopCmd stops the running gateway process.
func stopCmd() {
	// Try PID file first
	pidFile := "/var/run/luckyclaw.pid"
	pidData, err := os.ReadFile(pidFile)
	if err == nil {
		pidStr := strings.TrimSpace(string(pidData))
		pid, err := strconv.Atoi(pidStr)
		if err == nil {
			proc, err := os.FindProcess(pid)
			if err == nil {
				if err := proc.Signal(os.Interrupt); err == nil {
					fmt.Printf("🦞 Stopping LuckyClaw (PID %d)...\n", pid)
					time.Sleep(2 * time.Second)
					os.Remove(pidFile)
					fmt.Println("✓ Stopped")
					return
				}
			}
		}
	}

	// Fallback: find by process name
	out, err := exec.Command("pidof", "luckyclaw").Output()
	if err != nil {
		fmt.Println("LuckyClaw is not running.")
		return
	}

	myPid := os.Getpid()
	for _, pidStr := range strings.Fields(string(out)) {
		pid, err := strconv.Atoi(pidStr)
		if err != nil || pid == myPid {
			continue
		}
		proc, err := os.FindProcess(pid)
		if err == nil {
			fmt.Printf("🦞 Stopping LuckyClaw (PID %d)...\n", pid)
			proc.Signal(os.Interrupt)
		}
	}
	time.Sleep(2 * time.Second)
	os.Remove(pidFile)
	fmt.Println("✓ Stopped")
}

// restartCmd stops and restarts the gateway in background mode.
func restartCmd() {
	stopCmd()
	fmt.Println()
	gatewayStartBackground()
}

// gatewayStartBackground starts the gateway as a background process.
func gatewayStartBackground() {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("Error finding executable: %v\n", err)
		os.Exit(1)
	}

	logFile := "/var/log/luckyclaw.log"
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// Fallback to /tmp if no write permission
		logFile = "/tmp/luckyclaw.log"
		f, err = os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Error opening log file: %v\n", err)
			os.Exit(1)
		}
	}

	cmd := exec.Command(exePath, "gateway")
	cmd.Stdout = f
	cmd.Stderr = f
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting gateway: %v\n", err)
		os.Exit(1)
	}

	pid := cmd.Process.Pid

	// Write PID file
	pidFile := "/var/run/luckyclaw.pid"
	if pf, err := os.Create(pidFile); err == nil {
		fmt.Fprintf(pf, "%d", pid)
		pf.Close()
	}

	cmd.Process.Release()
	f.Close()

	fmt.Printf("🦞 LuckyClaw started in background (PID %d)\n", pid)
	fmt.Printf("   Log: %s\n", logFile)
	fmt.Println("   Use 'luckyclaw stop' to stop or 'luckyclaw restart' to restart")
}

func copyEmbeddedToTarget(targetDir string) error {
	// Ensure target directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("Failed to create target directory: %w", err)
	}

	// Walk through all files in embed.FS
	err := fs.WalkDir(embeddedFiles, "workspace", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Read embedded file
		data, err := embeddedFiles.ReadFile(path)
		if err != nil {
			return fmt.Errorf("Failed to read embedded file %s: %w", path, err)
		}

		new_path, err := filepath.Rel("workspace", path)
		if err != nil {
			return fmt.Errorf("Failed to get relative path for %s: %v\n", path, err)
		}

		// Build target file path
		targetPath := filepath.Join(targetDir, new_path)

		// Ensure target file's directory exists
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("Failed to create directory %s: %w", filepath.Dir(targetPath), err)
		}

		// Write file
		if err := os.WriteFile(targetPath, data, 0644); err != nil {
			return fmt.Errorf("Failed to write file %s: %w", targetPath, err)
		}

		return nil
	})

	return err
}

func configResetCmd() {
	configPath := getConfigPath()

	fmt.Println()
	fmt.Println("  ⚠️  WARNING: You are about to erase LuckyClaw's configuration. ⚠️")
	fmt.Println("  This will completely delete your API keys, LLM preferences,")
	fmt.Println("  and custom timezone. (It will NOT delete conversation memory).")
	fmt.Println()

	confirm := promptLine("  Type 'RESET' to confirm: ")
	if confirm != "RESET" {
		fmt.Println("  Config reset cancelled.")
		return
	}

	err := os.Remove(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("\n  No config file found (already reset).")
		} else {
			fmt.Printf("\n  ✗ Failed to delete config file: %v\n", err)
		}
		return
	}

	fmt.Println("\n  ✓ Config successfully erased.")
	fmt.Println("  Run 'luckyclaw onboard' to start fresh.")
}

func createWorkspaceTemplates(workspace string) {
	err := copyEmbeddedToTarget(workspace)
	if err != nil {
		fmt.Printf("Error copying workspace templates: %v\n", err)
	}
}

func migrateCmd() {
	if len(os.Args) > 2 && (os.Args[2] == "--help" || os.Args[2] == "-h") {
		migrateHelp()
		return
	}

	opts := migrate.Options{}

	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--dry-run":
			opts.DryRun = true
		case "--config-only":
			opts.ConfigOnly = true
		case "--workspace-only":
			opts.WorkspaceOnly = true
		case "--force":
			opts.Force = true
		case "--refresh":
			opts.Refresh = true
		case "--openclaw-home":
			if i+1 < len(args) {
				opts.OpenClawHome = args[i+1]
				i++
			}
		case "--luckyclaw-home":
			if i+1 < len(args) {
				opts.LuckyClawHome = args[i+1]
				i++
			}
		default:
			fmt.Printf("Unknown flag: %s\n", args[i])
			migrateHelp()
			os.Exit(1)
		}
	}

	result, err := migrate.Run(opts)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !opts.DryRun {
		migrate.PrintSummary(result)
	}
}

func migrateHelp() {
	fmt.Println("\nMigrate from OpenClaw to LuckyClaw")
	fmt.Println()
	fmt.Println("Usage: luckyclaw migrate [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --dry-run          Show what would be migrated without making changes")
	fmt.Println("  --refresh          Re-sync workspace files from OpenClaw (repeatable)")
	fmt.Println("  --config-only      Only migrate config, skip workspace files")
	fmt.Println("  --workspace-only   Only migrate workspace files, skip config")
	fmt.Println("  --force            Skip confirmation prompts")
	fmt.Println("  --openclaw-home    Override OpenClaw home directory (default: ~/.openclaw)")
	fmt.Println("  --luckyclaw-home    Override LuckyClaw home directory (default: ~/.luckyclaw)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  luckyclaw migrate              Detect and migrate from OpenClaw")
	fmt.Println("  luckyclaw migrate --dry-run    Show what would be migrated")
	fmt.Println("  luckyclaw migrate --refresh    Re-sync workspace files")
	fmt.Println("  luckyclaw migrate --force      Migrate without confirmation")
}

func agentCmd() {
	message := ""
	sessionKey := "cli:default"

	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--debug", "-d":
			logger.SetLevel(logger.DEBUG)
			fmt.Println("🔍 Debug mode enabled")
		case "-m", "--message":
			if i+1 < len(args) {
				message = args[i+1]
				i++
			}
		case "-s", "--session":
			if i+1 < len(args) {
				sessionKey = args[i+1]
				i++
			}
		}
	}

	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}
	applySystemTimezone(cfg)

	provider, err := providers.CreateProvider(cfg)
	if err != nil {
		fmt.Printf("Error creating provider: %v\n", err)
		os.Exit(1)
	}

	msgBus := bus.NewMessageBus()
	agentLoop := agent.NewAgentLoop(cfg, msgBus, provider)

	// Print agent startup info (only for interactive mode)
	startupInfo := agentLoop.GetStartupInfo()
	logger.InfoCF("agent", "Agent initialized",
		map[string]interface{}{
			"tools_count":      startupInfo["tools"].(map[string]interface{})["count"],
			"skills_total":     startupInfo["skills"].(map[string]interface{})["total"],
			"skills_available": startupInfo["skills"].(map[string]interface{})["available"],
		})

	if message != "" {
		ctx := context.Background()
		response, err := agentLoop.ProcessDirect(ctx, message, sessionKey)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\n%s %s\n", logo, response)
	} else {
		fmt.Printf("%s Interactive mode (Ctrl+C to exit)\n\n", logo)
		interactiveMode(agentLoop, sessionKey)
	}
}

func interactiveMode(agentLoop *agent.AgentLoop, sessionKey string) {
	prompt := fmt.Sprintf("%s You: ", logo)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          prompt,
		HistoryFile:     filepath.Join(os.TempDir(), ".luckyclaw_history"),
		HistoryLimit:    100,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})

	if err != nil {
		fmt.Printf("Error initializing readline: %v\n", err)
		fmt.Println("Falling back to simple input mode...")
		simpleInteractiveMode(agentLoop, sessionKey)
		return
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt || err == io.EOF {
				fmt.Println("\nGoodbye!")
				return
			}
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye!")
			return
		}

		ctx := context.Background()
		response, err := agentLoop.ProcessDirect(ctx, input, sessionKey)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("\n%s %s\n\n", logo, response)
	}
}

func simpleInteractiveMode(agentLoop *agent.AgentLoop, sessionKey string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(fmt.Sprintf("%s You: ", logo))
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("\nGoodbye!")
				return
			}
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye!")
			return
		}

		ctx := context.Background()
		response, err := agentLoop.ProcessDirect(ctx, input, sessionKey)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("\n%s %s\n\n", logo, response)
	}
}

func gatewayCmd() {
	// Check for flags
	args := os.Args[2:]
	for _, arg := range args {
		if arg == "--debug" || arg == "-d" {
			logger.SetLevel(logger.DEBUG)
			fmt.Println("🔍 Debug mode enabled")
		}
		if arg == "-b" || arg == "--background" {
			gatewayStartBackground()
			return
		}
	}

	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}
	applySystemTimezone(cfg)

	provider, err := providers.CreateProvider(cfg)
	if err != nil {
		fmt.Printf("Error creating provider: %v\n", err)
		os.Exit(1)
	}

	msgBus := bus.NewMessageBus()
	agentLoop := agent.NewAgentLoop(cfg, msgBus, provider)

	// Print agent startup info
	fmt.Println("\n📦 Agent Status:")
	startupInfo := agentLoop.GetStartupInfo()
	toolsInfo := startupInfo["tools"].(map[string]interface{})
	skillsInfo := startupInfo["skills"].(map[string]interface{})
	fmt.Printf("  • Tools: %d loaded\n", toolsInfo["count"])
	fmt.Printf("  • Skills: %d/%d available\n",
		skillsInfo["available"],
		skillsInfo["total"])

	// Log to file as well
	logger.InfoCF("agent", "Agent initialized",
		map[string]interface{}{
			"tools_count":      toolsInfo["count"],
			"skills_total":     skillsInfo["total"],
			"skills_available": skillsInfo["available"],
		})

	// Setup cron tool and service
	cronService := setupCronTool(agentLoop, msgBus, cfg)

	heartbeatService := heartbeat.NewHeartbeatService(
		cfg.WorkspacePath(),
		cfg.Heartbeat.Interval,
		cfg.Heartbeat.Enabled,
	)
	heartbeatService.SetBus(msgBus)
	heartbeatService.SetHandler(func(prompt, channel, chatID string) *tools.ToolResult {
		// Use cli:direct as fallback if no valid channel
		if channel == "" || chatID == "" {
			channel, chatID = "cli", "direct"
		}
		// Use ProcessHeartbeat - no session history, each heartbeat is independent
		response, err := agentLoop.ProcessHeartbeat(context.Background(), prompt, channel, chatID)
		if err != nil {
			return tools.ErrorResult(fmt.Sprintf("Heartbeat error: %v", err))
		}
		if response == "HEARTBEAT_OK" {
			return tools.SilentResult("Heartbeat OK")
		}
		// For heartbeat, always return silent - the subagent result will be
		// sent to user via processSystemMessage when the async task completes
		return tools.SilentResult(response)
	})

	channelManager, err := channels.NewManager(cfg, msgBus)
	if err != nil {
		fmt.Printf("Error creating channel manager: %v\n", err)
		os.Exit(1)
	}

	// Inject channel manager into agent loop for command handling
	agentLoop.SetChannelManager(channelManager)

	// Register native WhatsApp if enabled
	if cfg.Channels.WhatsApp.Enabled {
		wa, err := whatsapp.NewWhatsAppChannel(cfg.Channels.WhatsApp, msgBus, cfg.Channels.WhatsApp.SessionPath)
		if err != nil {
			logger.ErrorCF("whatsapp", "Failed to initialize native WhatsApp channel", map[string]interface{}{"error": err.Error()})
		} else {
			channelManager.RegisterChannel("whatsapp", wa)
			logger.InfoC("whatsapp", "Native WhatsApp channel registered")
		}
	}

	// Wire Discord moderation tool callbacks from the live DiscordChannel
	if discordCh, ok := channelManager.GetChannel("discord"); ok {
		if dc, ok := discordCh.(*channels.DiscordChannel); ok {
			if tool, ok := agentLoop.GetTool("discord_delete_message"); ok {
				if dt, ok := tool.(*tools.DiscordDeleteMessageTool); ok {
					dt.SetDeleteCallback(dc.DeleteMessage)
					logger.InfoC("discord", "Delete message callback wired to Discord channel")
				}
			}
			if tool, ok := agentLoop.GetTool("discord_timeout_user"); ok {
				if tt, ok := tool.(*tools.DiscordTimeoutUserTool); ok {
					tt.SetTimeoutCallback(dc.TimeoutUser)
					logger.InfoC("discord", "Timeout user callback wired to Discord channel")
				}
			}
		}
	}

	var transcriber *voice.GroqTranscriber
	if cfg.Providers.Groq.APIKey != "" {
		transcriber = voice.NewGroqTranscriber(cfg.Providers.Groq.APIKey)
		logger.InfoC("voice", "Groq voice transcription enabled")
	}

	if transcriber != nil {
		if telegramChannel, ok := channelManager.GetChannel("telegram"); ok {
			if tc, ok := telegramChannel.(*channels.TelegramChannel); ok {
				tc.SetTranscriber(transcriber)
				logger.InfoC("voice", "Groq transcription attached to Telegram channel")
			}
		}
		if discordChannel, ok := channelManager.GetChannel("discord"); ok {
			if dc, ok := discordChannel.(*channels.DiscordChannel); ok {
				dc.SetTranscriber(transcriber)
				logger.InfoC("voice", "Groq transcription attached to Discord channel")
			}
		}
		if slackChannel, ok := channelManager.GetChannel("slack"); ok {
			if sc, ok := slackChannel.(*channels.SlackChannel); ok {
				sc.SetTranscriber(transcriber)
				logger.InfoC("voice", "Groq transcription attached to Slack channel")
			}
		}
	}

	enabledChannels := channelManager.GetEnabledChannels()
	if len(enabledChannels) > 0 {
		fmt.Printf("✓ Channels enabled: %s\n", enabledChannels)
	} else {
		fmt.Println("⚠ Warning: No channels enabled")
	}

	fmt.Printf("✓ Gateway started on %s:%d\n", cfg.Gateway.Host, cfg.Gateway.Port)
	fmt.Println("Press Ctrl+C to stop")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := cronService.Start(); err != nil {
		fmt.Printf("Error starting cron service: %v\n", err)
	}
	fmt.Println("✓ Cron service started")

	if err := heartbeatService.Start(); err != nil {
		fmt.Printf("Error starting heartbeat service: %v\n", err)
	}
	fmt.Println("✓ Heartbeat service started")

	stateManager := state.NewManager(cfg.WorkspacePath())
	deviceService := devices.NewService(devices.Config{
		Enabled:    cfg.Devices.Enabled,
		MonitorUSB: cfg.Devices.MonitorUSB,
	}, stateManager)
	deviceService.SetBus(msgBus)
	if err := deviceService.Start(ctx); err != nil {
		fmt.Printf("Error starting device service: %v\n", err)
	} else if cfg.Devices.Enabled {
		fmt.Println("✓ Device event service started")
	}

	if err := channelManager.StartAll(ctx); err != nil {
		fmt.Printf("Error starting channels: %v\n", err)
	}

	healthServer := health.NewServer(cfg.Gateway.Host, cfg.Gateway.Port)
	go func() {
		if err := healthServer.Start(); err != nil && err != http.ErrServerClosed {
			logger.ErrorCF("health", "Health server error", map[string]interface{}{"error": err.Error()})
		}
	}()
	fmt.Printf("✓ Health endpoints available at http://%s:%d/health and /ready\n", cfg.Gateway.Host, cfg.Gateway.Port)

	go agentLoop.Run(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan

	fmt.Println("\nShutting down...")
	cancel()
	healthServer.Stop(context.Background())
	deviceService.Stop()
	heartbeatService.Stop()
	cronService.Stop()
	agentLoop.Stop()
	channelManager.StopAll(ctx)
	fmt.Println("✓ Gateway stopped")
}

func statusCmd() {
	configPath := getConfigPath()

	fmt.Println()
	fmt.Printf("  %s LuckyClaw Status\n", logo)
	fmt.Printf("  Version: %s\n", formatVersion())
	build, _ := formatBuildInfo()
	if build != "" {
		fmt.Printf("  Build: %s\n", build)
	}

	// Device info
	model := detectBoardModel()
	if model != "" {
		fmt.Printf("  Board: %s\n", model)
	}
	if uptimeOut, err := exec.Command("uptime", "-p").Output(); err == nil {
		fmt.Printf("  Uptime: %s", strings.TrimSpace(string(uptimeOut)))
		fmt.Println()
	}

	// Memory
	if memData, err := os.ReadFile("/proc/meminfo"); err == nil {
		lines := strings.Split(string(memData), "\n")
		var total, avail int
		for _, line := range lines {
			if strings.HasPrefix(line, "MemTotal:") {
				fmt.Sscanf(line, "MemTotal: %d", &total)
			}
			if strings.HasPrefix(line, "MemAvailable:") {
				fmt.Sscanf(line, "MemAvailable: %d", &avail)
			}
		}
		if total > 0 {
			fmt.Printf("  Memory: %dMB available / %dMB total\n", avail/1024, total/1024)
		}
	}

	// Gateway process status
	fmt.Println()
	if pidOut, err := exec.Command("pidof", "luckyclaw").Output(); err == nil {
		pids := strings.Fields(strings.TrimSpace(string(pidOut)))
		// Exclude our own PID
		myPid := fmt.Sprintf("%d", os.Getpid())
		for _, pid := range pids {
			if pid != myPid {
				fmt.Printf("  Gateway: running (PID %s)", pid)
				// Get RSS
				if statData, err := os.ReadFile(fmt.Sprintf("/proc/%s/status", pid)); err == nil {
					for _, line := range strings.Split(string(statData), "\n") {
						if strings.HasPrefix(line, "VmRSS:") {
							var rss int
							fmt.Sscanf(line, "VmRSS: %d", &rss)
							fmt.Printf(" — %dMB RSS", rss/1024)
						}
					}
				}
				fmt.Println()
				break
			}
		}
	} else {
		fmt.Println("  Gateway: stopped")
	}

	// Config status
	fmt.Println()
	cfg, err := loadConfig()
	if err != nil {
		fmt.Println("  Config:", configPath, "✗ (not found — run: luckyclaw onboard)")
		return
	}

	fmt.Println("  Config:", configPath, "✓")
	workspace := cfg.WorkspacePath()
	if _, err := os.Stat(workspace); err == nil {
		fmt.Println("  Workspace:", workspace, "✓")
	} else {
		fmt.Println("  Workspace:", workspace, "✗")
	}

	// Provider status
	fmt.Println()
	fmt.Printf("  Model: %s\n", cfg.Agents.Defaults.Model)
	if cfg.Agents.Defaults.Provider != "" {
		fmt.Printf("  Provider: %s\n", cfg.Agents.Defaults.Provider)
	}

	hasOpenRouter := cfg.Providers.OpenRouter.APIKey != ""
	hasAnthropic := cfg.Providers.Anthropic.APIKey != ""
	hasOpenAI := cfg.Providers.OpenAI.APIKey != ""
	hasGemini := cfg.Providers.Gemini.APIKey != ""
	hasGroq := cfg.Providers.Groq.APIKey != ""

	statusStr := func(enabled bool) string {
		if enabled {
			return "✓"
		}
		return "—"
	}

	activeProviders := []struct {
		name string
		set  bool
	}{
		{"OpenRouter", hasOpenRouter},
		{"OpenAI", hasOpenAI},
		{"Anthropic", hasAnthropic},
		{"Gemini", hasGemini},
		{"Groq", hasGroq},
	}

	for _, p := range activeProviders {
		if p.set {
			fmt.Printf("  %s: %s\n", p.name, statusStr(p.set))
		}
	}

	// Channel status
	fmt.Println()
	hasChannels := false
	if cfg.Channels.Telegram.Enabled {
		fmt.Println("  Telegram: enabled ✓")
		hasChannels = true
	}
	if cfg.Channels.Discord.Enabled {
		fmt.Println("  Discord: enabled ✓")
		hasChannels = true
	}
	if cfg.Channels.WhatsApp.Enabled {
		fmt.Println("  WhatsApp: enabled ✓")
		hasChannels = true
	}
	if cfg.Channels.Slack.Enabled {
		fmt.Println("  Slack: enabled ✓")
		hasChannels = true
	}

	if !hasChannels {
		fmt.Println("  Channels: none enabled")
	}

	store, _ := auth.LoadStore()
	if store != nil && len(store.Credentials) > 0 {
		fmt.Println()
		fmt.Println("  OAuth/Token Auth:")
		for provider, cred := range store.Credentials {
			status := "authenticated"
			if cred.IsExpired() {
				status = "expired"
			} else if cred.NeedsRefresh() {
				status = "needs refresh"
			}
			fmt.Printf("    %s (%s): %s\n", provider, cred.AuthMethod, status)
		}
	}
	fmt.Println()
}

func authCmd() {
	if len(os.Args) < 3 {
		authHelp()
		return
	}

	switch os.Args[2] {
	case "login":
		authLoginCmd()
	case "logout":
		authLogoutCmd()
	case "status":
		authStatusCmd()
	default:
		fmt.Printf("Unknown auth command: %s\n", os.Args[2])
		authHelp()
	}
}

func authHelp() {
	fmt.Println("\nAuth commands:")
	fmt.Println("  login       Login via OAuth or paste token")
	fmt.Println("  logout      Remove stored credentials")
	fmt.Println("  status      Show current auth status")
	fmt.Println()
	fmt.Println("Login options:")
	fmt.Println("  --provider <name>    Provider to login with (openai, anthropic)")
	fmt.Println("  --device-code        Use device code flow (for headless environments)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  luckyclaw auth login --provider openai")
	fmt.Println("  luckyclaw auth login --provider openai --device-code")
	fmt.Println("  luckyclaw auth login --provider anthropic")
	fmt.Println("  luckyclaw auth logout --provider openai")
	fmt.Println("  luckyclaw auth status")
}

func authLoginCmd() {
	provider := ""
	useDeviceCode := false

	args := os.Args[3:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--provider", "-p":
			if i+1 < len(args) {
				provider = args[i+1]
				i++
			}
		case "--device-code":
			useDeviceCode = true
		}
	}

	if provider == "" {
		fmt.Println("Error: --provider is required")
		fmt.Println("Supported providers: openai, anthropic")
		return
	}

	switch provider {
	case "openai":
		authLoginOpenAI(useDeviceCode)
	case "anthropic":
		authLoginPasteToken(provider)
	default:
		fmt.Printf("Unsupported provider: %s\n", provider)
		fmt.Println("Supported providers: openai, anthropic")
	}
}

func authLoginOpenAI(useDeviceCode bool) {
	cfg := auth.OpenAIOAuthConfig()

	var cred *auth.AuthCredential
	var err error

	if useDeviceCode {
		cred, err = auth.LoginDeviceCode(cfg)
	} else {
		cred, err = auth.LoginBrowser(cfg)
	}

	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		os.Exit(1)
	}

	if err := auth.SetCredential("openai", cred); err != nil {
		fmt.Printf("Failed to save credentials: %v\n", err)
		os.Exit(1)
	}

	appCfg, err := loadConfig()
	if err == nil {
		appCfg.Providers.OpenAI.AuthMethod = "oauth"
		if err := config.SaveConfig(getConfigPath(), appCfg); err != nil {
			fmt.Printf("Warning: could not update config: %v\n", err)
		}
	}

	fmt.Println("Login successful!")
	if cred.AccountID != "" {
		fmt.Printf("Account: %s\n", cred.AccountID)
	}
}

func authLoginPasteToken(provider string) {
	cred, err := auth.LoginPasteToken(provider, os.Stdin)
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		os.Exit(1)
	}

	if err := auth.SetCredential(provider, cred); err != nil {
		fmt.Printf("Failed to save credentials: %v\n", err)
		os.Exit(1)
	}

	appCfg, err := loadConfig()
	if err == nil {
		switch provider {
		case "anthropic":
			appCfg.Providers.Anthropic.AuthMethod = "token"
		case "openai":
			appCfg.Providers.OpenAI.AuthMethod = "token"
		}
		if err := config.SaveConfig(getConfigPath(), appCfg); err != nil {
			fmt.Printf("Warning: could not update config: %v\n", err)
		}
	}

	fmt.Printf("Token saved for %s!\n", provider)
}

func authLogoutCmd() {
	provider := ""

	args := os.Args[3:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--provider", "-p":
			if i+1 < len(args) {
				provider = args[i+1]
				i++
			}
		}
	}

	if provider != "" {
		if err := auth.DeleteCredential(provider); err != nil {
			fmt.Printf("Failed to remove credentials: %v\n", err)
			os.Exit(1)
		}

		appCfg, err := loadConfig()
		if err == nil {
			switch provider {
			case "openai":
				appCfg.Providers.OpenAI.AuthMethod = ""
			case "anthropic":
				appCfg.Providers.Anthropic.AuthMethod = ""
			}
			config.SaveConfig(getConfigPath(), appCfg)
		}

		fmt.Printf("Logged out from %s\n", provider)
	} else {
		if err := auth.DeleteAllCredentials(); err != nil {
			fmt.Printf("Failed to remove credentials: %v\n", err)
			os.Exit(1)
		}

		appCfg, err := loadConfig()
		if err == nil {
			appCfg.Providers.OpenAI.AuthMethod = ""
			appCfg.Providers.Anthropic.AuthMethod = ""
			config.SaveConfig(getConfigPath(), appCfg)
		}

		fmt.Println("Logged out from all providers")
	}
}

func authStatusCmd() {
	store, err := auth.LoadStore()
	if err != nil {
		fmt.Printf("Error loading auth store: %v\n", err)
		return
	}

	if len(store.Credentials) == 0 {
		fmt.Println("No authenticated providers.")
		fmt.Println("Run: luckyclaw auth login --provider <name>")
		return
	}

	fmt.Println("\nAuthenticated Providers:")
	fmt.Println("------------------------")
	for provider, cred := range store.Credentials {
		status := "active"
		if cred.IsExpired() {
			status = "expired"
		} else if cred.NeedsRefresh() {
			status = "needs refresh"
		}

		fmt.Printf("  %s:\n", provider)
		fmt.Printf("    Method: %s\n", cred.AuthMethod)
		fmt.Printf("    Status: %s\n", status)
		if cred.AccountID != "" {
			fmt.Printf("    Account: %s\n", cred.AccountID)
		}
		if !cred.ExpiresAt.IsZero() {
			fmt.Printf("    Expires: %s\n", cred.ExpiresAt.Format("2006-01-02 15:04"))
		}
	}
}

func getConfigPath() string {
	if envPath := os.Getenv("LUCKYCLAW_CONFIG"); envPath != "" {
		return envPath
	}
	if _, err := os.Stat("/oem"); err == nil {
		return "/oem/.luckyclaw/config.json"
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".luckyclaw", "config.json")
}

func setupCronTool(agentLoop *agent.AgentLoop, msgBus *bus.MessageBus, cfg *config.Config) *cron.CronService {
	cronStorePath := filepath.Join(cfg.WorkspacePath(), "cron", "jobs.json")

	loc, err := time.LoadLocation(cfg.Gateway.TimezoneName)
	if err != nil {
		loc = time.UTC
	}

	// Create cron service
	cronService := cron.NewCronService(cronStorePath, nil, loc)

	// Create and register CronTool
	cronTool := tools.NewCronTool(cronService, agentLoop, msgBus, cfg.WorkspacePath(), cfg.Agents.Defaults.RestrictToWorkspace)
	agentLoop.RegisterTool(cronTool)

	// Set the onJob handler
	cronService.SetOnJob(func(job *cron.CronJob) (string, error) {
		result := cronTool.ExecuteJob(context.Background(), job)
		return result, nil
	})

	return cronService
}

func loadConfig() (*config.Config, error) {
	cfg, err := config.LoadConfig(getConfigPath())
	if err == nil {
		applyConfigDefaults(cfg)
	}
	return cfg, err
}

func applyConfigDefaults(cfg *config.Config) {
	changed := false
	defaults := config.DefaultConfig()
	if cfg.Agents.Defaults.MaxTokens < defaults.Agents.Defaults.MaxTokens {
		cfg.Agents.Defaults.MaxTokens = defaults.Agents.Defaults.MaxTokens
		changed = true
	}
	if cfg.Agents.Defaults.MaxToolIterations < defaults.Agents.Defaults.MaxToolIterations {
		cfg.Agents.Defaults.MaxToolIterations = defaults.Agents.Defaults.MaxToolIterations
		changed = true
	}
	if changed {
		logger.Info("Auto-upgrading stale memory configuration limits to firmware defaults")
		err := config.SaveConfig(getConfigPath(), cfg)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to save auto-upgraded config: %v", err))
		}
	}
}
func cronCmd() {
	if len(os.Args) < 3 {
		cronHelp()
		return
	}

	subcommand := os.Args[2]

	// Load config to get workspace path
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	cronStorePath := filepath.Join(cfg.WorkspacePath(), "cron", "jobs.json")

	loc, err := time.LoadLocation(cfg.Gateway.TimezoneName)
	if err != nil {
		loc = time.UTC
	}

	switch subcommand {
	case "list":
		cronListCmd(cronStorePath, loc)
	case "add":
		cronAddCmd(cronStorePath, loc)
	case "remove":
		if len(os.Args) < 4 {
			fmt.Println("Usage: luckyclaw cron remove <job_id>")
			return
		}
		cronRemoveCmd(cronStorePath, os.Args[3], loc)
	case "enable":
		cronEnableCmd(cronStorePath, false, loc)
	case "disable":
		cronEnableCmd(cronStorePath, true, loc)
	default:
		fmt.Printf("Unknown cron command: %s\n", subcommand)
		cronHelp()
	}
}

func cronHelp() {
	fmt.Println("\nCron commands:")
	fmt.Println("  list              List all scheduled jobs")
	fmt.Println("  add              Add a new scheduled job")
	fmt.Println("  remove <id>       Remove a job by ID")
	fmt.Println("  enable <id>      Enable a job")
	fmt.Println("  disable <id>     Disable a job")
	fmt.Println()
	fmt.Println("Add options:")
	fmt.Println("  -n, --name       Job name")
	fmt.Println("  -m, --message    Message for agent")
	fmt.Println("  -e, --every      Run every N seconds")
	fmt.Println("  -c, --cron       Cron expression (e.g. '0 9 * * *')")
	fmt.Println("  -d, --deliver     Deliver response to channel")
	fmt.Println("  --to             Recipient for delivery")
	fmt.Println("  --channel        Channel for delivery")
}

func cronListCmd(storePath string, loc *time.Location) {
	cs := cron.NewCronService(storePath, nil, loc)
	jobs := cs.ListJobs(true) // Show all jobs, including disabled

	if len(jobs) == 0 {
		fmt.Println("No scheduled jobs.")
		return
	}

	fmt.Println("\nScheduled Jobs:")
	fmt.Println("----------------")
	for _, job := range jobs {
		var schedule string
		if job.Schedule.Kind == "every" && job.Schedule.EveryMS != nil {
			schedule = fmt.Sprintf("every %ds", *job.Schedule.EveryMS/1000)
		} else if job.Schedule.Kind == "cron" {
			schedule = job.Schedule.Expr
		} else {
			schedule = "one-time"
		}

		nextRun := "scheduled"
		if job.State.NextRunAtMS != nil {
			nextTime := time.UnixMilli(*job.State.NextRunAtMS)
			nextRun = nextTime.Format("2006-01-02 15:04")
		}

		status := "enabled"
		if !job.Enabled {
			status = "disabled"
		}

		fmt.Printf("  %s (%s)\n", job.Name, job.ID)
		fmt.Printf("    Schedule: %s\n", schedule)
		fmt.Printf("    Status: %s\n", status)
		fmt.Printf("    Next run: %s\n", nextRun)
	}
}

func cronAddCmd(storePath string, loc *time.Location) {
	name := ""
	message := ""
	var everySec *int64
	cronExpr := ""
	deliver := false
	channel := ""
	to := ""

	args := os.Args[3:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-n", "--name":
			if i+1 < len(args) {
				name = args[i+1]
				i++
			}
		case "-m", "--message":
			if i+1 < len(args) {
				message = args[i+1]
				i++
			}
		case "-e", "--every":
			if i+1 < len(args) {
				var sec int64
				fmt.Sscanf(args[i+1], "%d", &sec)
				everySec = &sec
				i++
			}
		case "-c", "--cron":
			if i+1 < len(args) {
				cronExpr = args[i+1]
				i++
			}
		case "-d", "--deliver":
			deliver = true
		case "--to":
			if i+1 < len(args) {
				to = args[i+1]
				i++
			}
		case "--channel":
			if i+1 < len(args) {
				channel = args[i+1]
				i++
			}
		}
	}

	if name == "" {
		fmt.Println("Error: --name is required")
		return
	}

	if message == "" {
		fmt.Println("Error: --message is required")
		return
	}

	if everySec == nil && cronExpr == "" {
		fmt.Println("Error: Either --every or --cron must be specified")
		return
	}

	var schedule cron.CronSchedule
	if everySec != nil {
		everyMS := *everySec * 1000
		schedule = cron.CronSchedule{
			Kind:    "every",
			EveryMS: &everyMS,
		}
	} else {
		schedule = cron.CronSchedule{
			Kind: "cron",
			Expr: cronExpr,
		}
	}

	cs := cron.NewCronService(storePath, nil, loc)
	job, err := cs.AddJob(name, schedule, message, deliver, channel, to, 0)
	if err != nil {
		fmt.Printf("Error adding job: %v\n", err)
		return
	}

	fmt.Printf("✓ Added job '%s' (%s)\n", job.Name, job.ID)
}

func cronRemoveCmd(storePath, jobID string, loc *time.Location) {
	cs := cron.NewCronService(storePath, nil, loc)
	if cs.RemoveJob(jobID) {
		fmt.Printf("✓ Removed job %s\n", jobID)
	} else {
		fmt.Printf("✗ Job %s not found\n", jobID)
	}
}

func cronEnableCmd(storePath string, disable bool, loc *time.Location) {
	if len(os.Args) < 4 {
		fmt.Println("Usage: luckyclaw cron enable/disable <job_id>")
		return
	}

	jobID := os.Args[3]
	cs := cron.NewCronService(storePath, nil, loc)
	enabled := !disable

	job := cs.EnableJob(jobID, enabled)
	if job != nil {
		status := "enabled"
		if disable {
			status = "disabled"
		}
		fmt.Printf("✓ Job '%s' %s\n", job.Name, status)
	} else {
		fmt.Printf("✗ Job %s not found\n", jobID)
	}
}

func skillsHelp() {
	fmt.Println("\nSkills commands:")
	fmt.Println("  list                    List installed skills")
	fmt.Println("  install <repo>          Install skill from GitHub")
	fmt.Println("  install-builtin          Install all builtin skills to workspace")
	fmt.Println("  list-builtin             List available builtin skills")
	fmt.Println("  remove <name>           Remove installed skill")
	fmt.Println("  search                  Search available skills")
	fmt.Println("  show <name>             Show skill details")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  luckyclaw skills list")
	fmt.Println("  luckyclaw skills install sipeed/luckyclaw-skills/weather")
	fmt.Println("  luckyclaw skills install-builtin")
	fmt.Println("  luckyclaw skills list-builtin")
	fmt.Println("  luckyclaw skills remove weather")
}

func skillsListCmd(loader *skills.SkillsLoader) {
	allSkills := loader.ListSkills()

	if len(allSkills) == 0 {
		fmt.Println("No skills installed.")
		return
	}

	fmt.Println("\nInstalled Skills:")
	fmt.Println("------------------")
	for _, skill := range allSkills {
		fmt.Printf("  ✓ %s (%s)\n", skill.Name, skill.Source)
		if skill.Description != "" {
			fmt.Printf("    %s\n", skill.Description)
		}
	}
}

func skillsInstallCmd(installer *skills.SkillInstaller) {
	if len(os.Args) < 4 {
		fmt.Println("Usage: luckyclaw skills install <github-repo>")
		fmt.Println("Example: luckyclaw skills install sipeed/luckyclaw-skills/weather")
		return
	}

	repo := os.Args[3]
	fmt.Printf("Installing skill from %s...\n", repo)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := installer.InstallFromGitHub(ctx, repo); err != nil {
		fmt.Printf("✗ Failed to install skill: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Skill '%s' installed successfully!\n", filepath.Base(repo))
}

func skillsRemoveCmd(installer *skills.SkillInstaller, skillName string) {
	fmt.Printf("Removing skill '%s'...\n", skillName)

	if err := installer.Uninstall(skillName); err != nil {
		fmt.Printf("✗ Failed to remove skill: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Skill '%s' removed successfully!\n", skillName)
}

func skillsInstallBuiltinCmd(workspace string) {
	builtinSkillsDir := "./luckyclaw/skills"
	workspaceSkillsDir := filepath.Join(workspace, "skills")

	fmt.Printf("Copying builtin skills to workspace...\n")

	skillsToInstall := []string{
		"weather",
		"news",
		"stock",
		"calculator",
	}

	for _, skillName := range skillsToInstall {
		builtinPath := filepath.Join(builtinSkillsDir, skillName)
		workspacePath := filepath.Join(workspaceSkillsDir, skillName)

		if _, err := os.Stat(builtinPath); err != nil {
			fmt.Printf("⊘ Builtin skill '%s' not found: %v\n", skillName, err)
			continue
		}

		if err := os.MkdirAll(workspacePath, 0755); err != nil {
			fmt.Printf("✗ Failed to create directory for %s: %v\n", skillName, err)
			continue
		}

		if err := copyDirectory(builtinPath, workspacePath); err != nil {
			fmt.Printf("✗ Failed to copy %s: %v\n", skillName, err)
		}
	}

	fmt.Println("\n✓ All builtin skills installed!")
	fmt.Println("Now you can use them in your workspace.")
}

func skillsListBuiltinCmd() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}
	builtinSkillsDir := filepath.Join(filepath.Dir(cfg.WorkspacePath()), "luckyclaw", "skills")

	fmt.Println("\nAvailable Builtin Skills:")
	fmt.Println("-----------------------")

	entries, err := os.ReadDir(builtinSkillsDir)
	if err != nil {
		fmt.Printf("Error reading builtin skills: %v\n", err)
		return
	}

	if len(entries) == 0 {
		fmt.Println("No builtin skills available.")
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			skillName := entry.Name()
			skillFile := filepath.Join(builtinSkillsDir, skillName, "SKILL.md")

			description := "No description"
			if _, err := os.Stat(skillFile); err == nil {
				data, err := os.ReadFile(skillFile)
				if err == nil {
					content := string(data)
					if idx := strings.Index(content, "\n"); idx > 0 {
						firstLine := content[:idx]
						if strings.Contains(firstLine, "description:") {
							descLine := strings.Index(content[idx:], "\n")
							if descLine > 0 {
								description = strings.TrimSpace(content[idx+descLine : idx+descLine])
							}
						}
					}
				}
			}
			status := "✓"
			fmt.Printf("  %s  %s\n", status, entry.Name())
			if description != "" {
				fmt.Printf("     %s\n", description)
			}
		}
	}
}

func skillsSearchCmd(installer *skills.SkillInstaller) {
	fmt.Println("Searching for available skills...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	availableSkills, err := installer.ListAvailableSkills(ctx)
	if err != nil {
		fmt.Printf("✗ Failed to fetch skills list: %v\n", err)
		return
	}

	if len(availableSkills) == 0 {
		fmt.Println("No skills available.")
		return
	}

	fmt.Printf("\nAvailable Skills (%d):\n", len(availableSkills))
	fmt.Println("--------------------")
	for _, skill := range availableSkills {
		fmt.Printf("  📦 %s\n", skill.Name)
		fmt.Printf("     %s\n", skill.Description)
		fmt.Printf("     Repo: %s\n", skill.Repository)
		if skill.Author != "" {
			fmt.Printf("     Author: %s\n", skill.Author)
		}
		if len(skill.Tags) > 0 {
			fmt.Printf("     Tags: %v\n", skill.Tags)
		}
		fmt.Println()
	}
}

func skillsShowCmd(loader *skills.SkillsLoader, skillName string) {
	content, ok := loader.LoadSkill(skillName)
	if !ok {
		fmt.Printf("✗ Skill '%s' not found\n", skillName)
		return
	}

	fmt.Printf("\n📦 Skill: %s\n", skillName)
	fmt.Println("----------------------")
	fmt.Println(content)
}

// OS Timezone Synchronization (POSIX string generation)
func getPosixTZ(offset int) string {
	// offset is in seconds.
	// POSIX TZ format requires the sign to be INVERTED compared to standard ISO offsets.
	// E.g., UTC+9 (Tokyo, offset=32400) -> TZ=LCL-9
	// E.g., UTC-5 (New York, offset=-18000) -> TZ=LCL+5
	// E.g., UTC+5:30 (India, offset=19800) -> TZ=LCL-5:30

	if offset == 0 {
		return "UTC0"
	}

	// Invert the sign for POSIX
	posixOffset := -offset

	sign := ""
	if posixOffset >= 0 {
		sign = "+"
	} else {
		sign = "-"
		posixOffset = -posixOffset
	}

	hours := posixOffset / 3600
	remainder := posixOffset % 3600
	minutes := remainder / 60
	seconds := remainder % 60

	if seconds > 0 {
		return fmt.Sprintf("LCL%s%d:%02d:%02d", sign, hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("LCL%s%d:%02d", sign, hours, minutes)
	}
	return fmt.Sprintf("LCL%s%d", sign, hours)
}

func applySystemTimezone(cfg *config.Config) {
	if cfg.Gateway.TimezoneName == "" {
		return // Legacy config fallback
	}

	// Sync the underlying OS environment timezone with the config
	// so terminal commands executed by the LLM (like `date`) perfectly
	// match the system prompt time without requiring zoneinfo db.
	posixTZ := getPosixTZ(cfg.Gateway.UTCOffset)
	os.Setenv("TZ", posixTZ)
}
func installCmd() {
	if os.Getuid() != 0 {
		fmt.Println("  ⚠ Warning: This command usually requires root privileges to write to /etc.")
	}

	fmt.Println("\n  🦞 LuckyClaw Installation")
	fmt.Println("  ─────────────────────────")

	// 1. Init script
	srcInit := "overlay/etc/init.d/S99luckyclaw"
	dstInit := "/etc/init.d/S99luckyclaw"
	fmt.Printf("  Installing init script to %s... ", dstInit)

	initData, err := firmware.FS.ReadFile(srcInit)
	if err == nil {
		err = os.WriteFile(dstInit, initData, 0755)
	}

	if err != nil {
		fmt.Printf("✗\n    Error: %v\n", err)
	} else {
		fmt.Println("✓")
	}

	// 2. SSH Banner
	srcBanner := "overlay/etc/profile.d/luckyclaw-banner.sh"
	dstBanner := "/etc/profile.d/luckyclaw-banner.sh"
	fmt.Printf("  Installing SSH banner to %s... ", dstBanner)

	bannerData, err := firmware.FS.ReadFile(srcBanner)
	if err == nil {
		err = os.WriteFile(dstBanner, bannerData, 0644)
	}

	if err != nil {
		fmt.Printf("✗\n    Error: %v\n", err)
	} else {
		fmt.Println("✓")
	}

	fmt.Println("\n  ✓ Installation complete.")
	fmt.Println("  You can now start LuckyClaw via: /etc/init.d/S99luckyclaw start")
}

func copyFile(src, dst string, mode os.FileMode) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, mode)
}
