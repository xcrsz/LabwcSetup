package main

import (
    "fmt"
    "io"
    "io/fs"
    "os"
    "os/exec"
    "path/filepath"
    "sort"
    "strings"
    "syscall"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type appState int

const (
    menuView appState = iota
    progressView
    actionView
)

const (
    viewWidth     = 60
    menuItemWidth = 32
    logFileName   = "labwcsetup.log"
    defaultEditor = "geany"
)

var (
    titleStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("#00ff00")).
        Padding(1, 2).
        Align(lipgloss.Center).
        Width(viewWidth)

    menuStyle = lipgloss.NewStyle().
        Align(lipgloss.Left).
        Width(viewWidth)

    cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00")).Bold(true)
    disabledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
    logStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("63")).Padding(0, 2).Width(viewWidth)
    actionStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00ff00")).Padding(1, 2).Align(lipgloss.Center).Width(viewWidth)
    infoStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Padding(0, 2).Width(viewWidth)
)

type model struct {
    state        appState
    choices      []string
    cursor       int
    selected     string
    logs         []string
    isProcessing bool
    actionMsg    string
    lastResult   string
}

type statusMsg struct {
    status string
    err    error
}

func initialModel() model {
    clearScreen()
    return model{
        state: menuView,
        choices: []string{
            "Install Labwc",
            "Setup System",
            "Install Default Config",
            "Verify Setup",
            "Save Logs",
            "Exit",
        },
    }
}

func clearScreen() {
    cmd := exec.Command("clear")
    cmd.Stdout = os.Stdout
    _ = cmd.Run()
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch m.state {
        case menuView:
            switch msg.String() {
            case "ctrl+c", "q":
                return m, tea.Quit
            case "up":
                if m.cursor > 0 {
                    m.cursor--
                }
            case "down":
                if m.cursor < len(m.choices)-1 {
                    m.cursor++
                }
            case "enter":
                m.selected = m.choices[m.cursor]
                m.isProcessing = true
                switch m.selected {
                case "Install Labwc":
                    m.state = progressView
                    m.actionMsg = "Installing Labwc and companion packages..."
                    return m, installLabwc()
                case "Setup System":
                    m.state = progressView
                    m.actionMsg = "Configuring system services and runtime environment..."
                    return m, setupSystem()
                case "Install Default Config":
                    m.state = actionView
                    m.actionMsg = "Installing default Labwc configuration..."
                    return m, installLabwcConfig()
                case "Verify Setup":
                    m.state = actionView
                    m.actionMsg = "Verifying Labwc setup..."
                    return m, verifyLabwcSetup()
                case "Save Logs":
                    m.state = actionView
                    m.actionMsg = "Saving logs..."
                    return m, saveLogsToFile(m)
                case "Exit":
                    return m, tea.Quit
                }
            }
        case progressView, actionView:
            return m, nil
        }
    case statusMsg:
        m.logs = append(m.logs, msg.status)
        m.lastResult = msg.status
        m.isProcessing = false
        m.state = menuView
        return m, nil
    }

    return m, nil
}

func (m model) View() string {
    switch m.state {
    case menuView:
        return m.renderMenuView()
    case progressView:
        return m.renderProgressView()
    case actionView:
        return m.renderActionView()
    default:
        return "Unknown state"
    }
}

func (m model) renderMenuView() string {
    title := titleStyle.Render("Labwc Setup Assistant for FreeBSD and GhostBSD")

    menu := strings.Builder{}
    for i, choice := range m.choices {
        if m.cursor == i {
            menu.WriteString(cursorStyle.Render(fmt.Sprintf("> %-*s", menuItemWidth-2, choice)) + "\n")
        } else {
            menu.WriteString(disabledStyle.Render(fmt.Sprintf("  %-*s", menuItemWidth-2, choice)) + "\n")
        }
    }

    footer := "Use arrow keys to move. Press Enter to select. Press q to quit."
    if m.lastResult != "" {
        footer = summarizeStatus(m.lastResult)
    }

    return lipgloss.JoinVertical(
        lipgloss.Left,
        title,
        menuStyle.Render(menu.String()),
        infoStyle.Render(footer),
    )
}

func (m model) renderProgressView() string {
    body := titleStyle.Render("Working")
    body += "\n" + actionStyle.Render(m.actionMsg)
    body += "\n" + logStyle.Render("Please wait...")
    return body
}

func (m model) renderActionView() string {
    body := titleStyle.Render("Working")
    body += "\n" + actionStyle.Render(m.actionMsg)
    body += "\n" + logStyle.Render("Please wait...")
    return body
}

func summarizeStatus(s string) string {
    lines := strings.Split(strings.TrimSpace(s), "\n")
    if len(lines) == 0 || lines[0] == "" {
        return "Ready."
    }
    first := strings.TrimSpace(lines[0])
    if len(first) > viewWidth-4 {
        first = first[:viewWidth-7] + "..."
    }
    return first
}

func isPackageInstalled(pkg string) bool {
    cmd := exec.Command("pkg", "info", pkg)
    return cmd.Run() == nil
}

func commandExists(name string) bool {
    _, err := exec.LookPath(name)
    return err == nil
}

func findRenderDevice() string {
    entries, err := os.ReadDir("/dev/dri")
    if err != nil {
        return ""
    }

    var renderNodes []string
    for _, e := range entries {
        if strings.HasPrefix(e.Name(), "renderD") {
            renderNodes = append(renderNodes, filepath.Join("/dev/dri", e.Name()))
        }
    }

    if len(renderNodes) == 0 {
        return ""
    }

    sort.Strings(renderNodes)
    return renderNodes[0]
}

func installLabwc() tea.Cmd {
    return func() tea.Msg {
        pkgs := []string{
            "drm-kmod",
            "mesa-libs",
            "mesa-dri",
            "consolekit2",
            "dbus",
            "seatd",
            "pam_xdg",
            "labwc",
            "sfwbar",
            "wofi",
            "foot",
            "pcmanfm",
            defaultEditor,
            "librewolf",
            "grim",
            "slurp",
            "swaybg",
            "swayidle",
            "swaylock",
            "wlopm",
            "mako",
        }

        var logs []string
        var failed []string

        for _, pkg := range pkgs {
            if isPackageInstalled(pkg) {
                logs = append(logs, fmt.Sprintf("Already installed: %s", pkg))
                continue
            }

            cmd := exec.Command("sudo", "pkg", "install", "-y", pkg)
            out, err := cmd.CombinedOutput()
            if err != nil {
                logs = append(logs, fmt.Sprintf("Failed to install %s: %s", pkg, strings.TrimSpace(string(out))))
                failed = append(failed, pkg)
                continue
            }

            logs = append(logs, fmt.Sprintf("Successfully installed %s", pkg))
        }

        logs = append(logs, "")
        logs = append(logs, fmt.Sprintf("Preferred editor package in this build: %s", defaultEditor))
        logs = append(logs, "Edit LabwcSetup.go if you want leafpad instead.")

        if len(failed) > 0 {
            logs = append(logs, "")
            logs = append(logs, fmt.Sprintf("Failed packages (%d): %s", len(failed), strings.Join(failed, ", ")))
            return statusMsg{status: strings.Join(logs, "\n"), err: fmt.Errorf("%d packages failed to install", len(failed))}
        }

        logs = append(logs, "")
        logs = append(logs, "Labwc package installation complete.")
        return statusMsg{status: strings.Join(logs, "\n")}
    }
}

func setupSystem() tea.Cmd {
    return func() tea.Msg {
        var logs []string

        steps := []struct {
            desc string
            cmd  []string
        }{
            {"Enabling dbus service", []string{"sudo", "sysrc", "dbus_enable=YES"}},
            {"Starting dbus service", []string{"sudo", "service", "dbus", "start"}},
            {"Enabling seatd service", []string{"sudo", "sysrc", "seatd_enable=YES"}},
            {"Starting seatd service", []string{"sudo", "service", "seatd", "start"}},
        }

        for _, step := range steps {
            cmd := exec.Command(step.cmd[0], step.cmd[1:]...)
            out, err := cmd.CombinedOutput()
            if err != nil {
                outStr := string(out)
                if strings.Contains(outStr, "already running") {
                    logs = append(logs, fmt.Sprintf("%s: already running", step.desc))
                } else {
                    logs = append(logs, fmt.Sprintf("Warning: %s: %s", step.desc, strings.TrimSpace(outStr)))
                }
            } else {
                logs = append(logs, fmt.Sprintf("%s: OK", step.desc))
            }
        }

        currentUser := os.Getenv("USER")
        if currentUser == "" {
            currentUser = os.Getenv("LOGNAME")
        }
        if currentUser != "" {
            cmd := exec.Command("sudo", "pw", "groupmod", "video", "-m", currentUser)
            out, err := cmd.CombinedOutput()
            if err != nil {
                logs = append(logs, fmt.Sprintf("Warning: Adding user to video group: %s", strings.TrimSpace(string(out))))
            } else {
                logs = append(logs, fmt.Sprintf("Added user '%s' to video group: OK", currentUser))
            }
        } else {
            logs = append(logs, "Warning: Could not determine current user for group setup")
        }

        cmd := exec.Command("sudo", "kldload", "drm")
        out, err := cmd.CombinedOutput()
        if err != nil {
            outStr := string(out)
            if strings.Contains(outStr, "already loaded") || strings.Contains(outStr, "module already loaded") {
                logs = append(logs, "Loading DRM kernel module: already loaded")
            } else {
                logs = append(logs, fmt.Sprintf("Warning: Loading DRM kernel module: %s", strings.TrimSpace(outStr)))
            }
        } else {
            logs = append(logs, "Loading DRM kernel module: OK")
        }

        cmd = exec.Command("sudo", "sysrc", "kld_list+=drm")
        out, err = cmd.CombinedOutput()
        if err != nil {
            logs = append(logs, fmt.Sprintf("Warning: Persisting DRM module to boot: %s", strings.TrimSpace(string(out))))
        } else {
            logs = append(logs, "Persisting DRM module to boot: OK")
        }

        homeDir, homeErr := os.UserHomeDir()
        if homeErr != nil {
            logs = append(logs, fmt.Sprintf("Warning: Could not determine home directory: %v", homeErr))
        } else {
            profilePath := filepath.Join(homeDir, ".profile")
            xdgLine := fmt.Sprintf("export XDG_RUNTIME_DIR=/tmp/%d-runtime-dir", os.Geteuid())
            envLines := []string{
                "# Wayland runtime for Labwc",
                xdgLine,
                "export LIBSEAT_BACKEND=consolekit2",
            }

            profileContent, _ := os.ReadFile(profilePath)
            profileStr := string(profileContent)

            needsWrite := false
            builder := strings.Builder{}
            for _, line := range envLines {
                if !strings.Contains(profileStr, line) {
                    builder.WriteString(line + "\n")
                    needsWrite = true
                }
            }

            if needsWrite {
                f, err := os.OpenFile(profilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
                if err != nil {
                    logs = append(logs, fmt.Sprintf("Warning: Could not update %s: %v", profilePath, err))
                } else {
                    _, _ = f.WriteString("\n" + builder.String())
                    _ = f.Close()
                    logs = append(logs, fmt.Sprintf("Updated %s with Labwc session environment: OK", profilePath))
                }
            } else {
                logs = append(logs, ".profile already contains Wayland session environment: OK")
            }
        }

        renderDev := findRenderDevice()
        if renderDev != "" {
            logs = append(logs, fmt.Sprintf("Found DRM render device: %s", renderDev))
            f, err := os.Open(renderDev)
            if err != nil {
                logs = append(logs, fmt.Sprintf("Warning: Cannot access %s: %v", renderDev, err))
            } else {
                _ = f.Close()
                logs = append(logs, fmt.Sprintf("DRM render device %s is accessible: OK", renderDev))
            }
        } else {
            logs = append(logs, "Warning: No DRM render device found in /dev/dri")
        }

        logs = append(logs, "")
        logs = append(logs, "System setup complete. Log out and back in if group changes do not apply immediately.")
        logs = append(logs, "Start Labwc from a TTY with:")
        logs = append(logs, "  LIBSEAT_BACKEND=consolekit2 ck-launch-session dbus-launch labwc")

        return statusMsg{status: strings.Join(logs, "\n")}
    }
}

func installLabwcConfig() tea.Cmd {
    return func() tea.Msg {
        homeDir, err := os.UserHomeDir()
        if err != nil {
            return statusMsg{status: "Failed to determine home directory", err: err}
        }

        destDir := filepath.Join(homeDir, ".config", "labwc")
        if err := os.MkdirAll(destDir, 0755); err != nil {
            return statusMsg{status: fmt.Sprintf("Failed to create %s: %v", destDir, err), err: err}
        }

        srcDir, err := findBundledConfigDir()
        if err != nil {
            return statusMsg{status: err.Error(), err: err}
        }

        copied, copyErr := copyDirRecursive(srcDir, destDir)
        if copyErr != nil {
            return statusMsg{status: fmt.Sprintf("Failed to install Labwc configuration: %v", copyErr), err: copyErr}
        }

        msg := []string{
            fmt.Sprintf("Installed Labwc configuration into %s", destDir),
            fmt.Sprintf("Copied files: %s", strings.Join(copied, ", ")),
            "",
            "You can reload the compositor after editing rc.xml or menu.xml with:",
            "  labwc --reconfigure",
        }
        return statusMsg{status: strings.Join(msg, "\n")}
    }
}

func verifyLabwcSetup() tea.Cmd {
    return func() tea.Msg {
        homeDir, err := os.UserHomeDir()
        if err != nil {
            return statusMsg{status: "Failed to determine home directory", err: err}
        }

        requiredFiles := []string{
            filepath.Join(homeDir, ".config", "labwc", "rc.xml"),
            filepath.Join(homeDir, ".config", "labwc", "menu.xml"),
            filepath.Join(homeDir, ".config", "labwc", "autostart"),
            filepath.Join(homeDir, ".config", "labwc", "environment"),
            filepath.Join(homeDir, ".config", "labwc", "backgrounds", "dark_blue_bg.png"),
        }

        requiredCommands := []string{
            "labwc",
            "sfwbar",
            "wofi",
            "foot",
            "pcmanfm",
            "librewolf",
            defaultEditor,
            "grim",
            "slurp",
            "swaybg",
            "swayidle",
            "swaylock",
            "wlopm",
            "mako",
        }

        var lines []string
        var missing []string

        for _, f := range requiredFiles {
            if _, err := os.Stat(f); err != nil {
                lines = append(lines, fmt.Sprintf("Missing file: %s", f))
                missing = append(missing, f)
            } else {
                lines = append(lines, fmt.Sprintf("OK file: %s", f))
            }
        }

        for _, cmdName := range requiredCommands {
            if commandExists(cmdName) {
                lines = append(lines, fmt.Sprintf("OK command: %s", cmdName))
            } else {
                lines = append(lines, fmt.Sprintf("Missing command in PATH: %s", cmdName))
                missing = append(missing, cmdName)
            }
        }

        renderDev := findRenderDevice()
        if renderDev != "" {
            lines = append(lines, fmt.Sprintf("Detected DRM render node: %s", renderDev))
        } else {
            lines = append(lines, "Warning: No DRM render node detected")
        }

        lines = append(lines, "")
        if len(missing) > 0 {
            lines = append(lines, "Verification completed with missing items.")
            return statusMsg{status: strings.Join(lines, "\n"), err: fmt.Errorf("verification found %d missing items", len(missing))}
        }

        lines = append(lines, "Verification passed. Labwc looks ready.")
        return statusMsg{status: strings.Join(lines, "\n")}
    }
}

func findBundledConfigDir() (string, error) {
    candidates := []string{}

    if exePath, err := os.Executable(); err == nil {
        exeDir := filepath.Dir(exePath)
        candidates = append(candidates,
            filepath.Join(exeDir, "configs", "labwc"),
            filepath.Join(exeDir, "..", "configs", "labwc"),
        )
    }

    if cwd, err := os.Getwd(); err == nil {
        candidates = append(candidates,
            filepath.Join(cwd, "configs", "labwc"),
            filepath.Join(cwd, "LabwcSetup", "configs", "labwc"),
        )
    }

    for _, candidate := range candidates {
        if info, err := os.Stat(candidate); err == nil && info.IsDir() {
            return candidate, nil
        }
    }

    return "", fmt.Errorf("could not locate configs/labwc next to executable or in current directory")
}

func copyDirRecursive(srcDir, destDir string) ([]string, error) {
    var copied []string

    err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }

        rel, err := filepath.Rel(srcDir, path)
        if err != nil {
            return err
        }
        if rel == "." {
            return nil
        }

        target := filepath.Join(destDir, rel)
        if d.IsDir() {
            return os.MkdirAll(target, 0755)
        }

        if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
            return err
        }

        if err := copyFile(path, target); err != nil {
            return err
        }
        copied = append(copied, rel)
        return nil
    })
    if err != nil {
        return copied, err
    }

    sort.Strings(copied)
    return copied, nil
}

func copyFile(srcPath, destPath string) error {
    in, err := os.Open(srcPath)
    if err != nil {
        return err
    }
    defer in.Close()

    info, err := in.Stat()
    if err != nil {
        return err
    }

    mode := info.Mode()
    out, err := os.OpenFile(destPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode.Perm())
    if err != nil {
        return err
    }
    defer out.Close()

    if _, err := io.Copy(out, in); err != nil {
        return err
    }
    return nil
}

func saveLogsToFile(m model) tea.Cmd {
    return func() tea.Msg {
        if len(m.logs) == 0 {
            return statusMsg{status: "No logs available yet. Run an action first."}
        }

        logFile := filepath.Join(os.TempDir(), logFileName)
        file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if err != nil {
            return statusMsg{status: "Failed to open log file for writing", err: err}
        }
        defer file.Close()

        for _, entry := range m.logs {
            if _, err := file.WriteString(entry + "\n\n"); err != nil {
                return statusMsg{status: "Failed to write to log file", err: err}
            }
        }

        return statusMsg{status: fmt.Sprintf("Logs saved to %s", logFile)}
    }
}

func setupEnvironment() error {
    userID := os.Geteuid()
    runtimeDir := fmt.Sprintf("/tmp/%d-runtime-dir", userID)
    _ = os.Setenv("XDG_RUNTIME_DIR", runtimeDir)

    if _, err := os.Stat(runtimeDir); os.IsNotExist(err) {
        return os.Mkdir(runtimeDir, 0700)
    }

    info, err := os.Stat(runtimeDir)
    if err != nil {
        return err
    }

    stat, ok := info.Sys().(*syscall.Stat_t)
    if !ok {
        return fmt.Errorf("failed to get ownership information for %s", runtimeDir)
    }

    if stat.Uid != uint32(userID) {
        return fmt.Errorf("XDG_RUNTIME_DIR %q is owned by UID %d, not our UID %d", runtimeDir, stat.Uid, userID)
    }

    return nil
}

func main() {
    if err := setupEnvironment(); err != nil {
        fmt.Fprintf(os.Stderr, "LabwcSetup environment error: %v\n", err)
        os.Exit(1)
    }

    p := tea.NewProgram(initialModel())
    if err := p.Start(); err != nil {
        fmt.Fprintf(os.Stderr, "LabwcSetup error: %v\n", err)
        os.Exit(1)
    }
}
