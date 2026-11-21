package proto

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
)

// SandboxProbeCheck records one runtime isolation test executed inside the Wasm VM.
type SandboxProbeCheck struct {
	Name    string `json:"name"`
	Target  string `json:"target"`
	Blocked bool   `json:"blocked"`
	Detail  string `json:"detail"`
}

// SandboxProbeReq asks server_plugin.wasm to run sandbox checks for Path inside the Server Wasm VM.
type SandboxProbeReq struct {
	Path string `json:"path"`
	Mode string `json:"mode,omitempty"` // "explore", "read", "write", "audit"
}

// SandboxProbeResp returns every check executed inside the Wasm VM.
type SandboxProbeResp struct {
	Role         string              `json:"role"`
	Path         string              `json:"path"`
	EntriesFound int                 `json:"entries_found"`
	AllBlocked   bool                `json:"all_blocked"`
	Checks       []SandboxProbeCheck `json:"checks"`
}

// RunSandboxProbes executes real Go stdlib filesystem, process exec, network, and env calls
// inside a wasip1 Wasm VM and verifies that every host escape attempt is blocked.
func RunSandboxProbes(role, rawPath, rawMode string) SandboxProbeResp {
	path := strings.TrimSpace(rawPath)
	if path == "" {
		path = "/"
	}
	mode := strings.ToLower(strings.TrimSpace(rawMode))
	if mode == "" {
		mode = "explore"
	}

	checks := make([]SandboxProbeCheck, 0, 10)
	entriesFound := 0

	// 1. User-selected File Explorer target probe.
	switch mode {
	case "read":
		data, err := os.ReadFile(path)
		checks = append(checks, formatErrCheck("Target ReadFile", path, err,
			fmt.Sprintf("UNEXPECTED READ (%d bytes)", len(data))))
	case "write":
		target := path
		if target == "/" || target == "." || target == "" {
			target = "/tmp/wasm_sandbox_escape_test.txt"
		}
		err := os.WriteFile(target, []byte("wasm_escape_probe"), 0644)
		checks = append(checks, formatErrCheck("Target WriteFile", target, err,
			"UNEXPECTED FILE WRITE SUCCEEDED"))
	default:
		ents, err := os.ReadDir(path)
		if err == nil {
			entriesFound = len(ents)
		}
		checks = append(checks, formatErrCheck("Explore ReadDir", path, err,
			fmt.Sprintf("UNEXPECTED LIST (%d entries)", len(ents))))
	}

	// 2. Standard filesystem escape probes.
	if _, err := os.ReadDir("/"); path != "/" || mode != "explore" {
		ents, rErr := os.ReadDir("/")
		_ = err
		checks = append(checks, formatErrCheck("Host Root List", "/", rErr,
			fmt.Sprintf("UNEXPECTED ROOT LIST (%d entries)", len(ents))))
	}
	passwdData, passwdErr := os.ReadFile("/etc/passwd")
	checks = append(checks, formatErrCheck("Read /etc/passwd", "/etc/passwd", passwdErr,
		fmt.Sprintf("UNEXPECTED READ (%d bytes)", len(passwdData))))

	statInfo, statErr := os.Stat("/bin/sh")
	statDetail := "UNEXPECTED STAT"
	if statInfo != nil {
		statDetail = fmt.Sprintf("UNEXPECTED STAT size=%d", statInfo.Size())
	}
	checks = append(checks, formatErrCheck("Stat /bin/sh", "/bin/sh", statErr, statDetail))

	writeErr := os.WriteFile("/tmp/wasm_sandbox_write_probe.txt", []byte("blocked"), 0644)
	checks = append(checks, formatErrCheck("Write Host Temp", "/tmp/wasm_sandbox_write_probe.txt", writeErr,
		"UNEXPECTED WRITE SUCCEEDED"))

	// 3. Process / shell execution probe (os/exec).
	cmdOut, execErr := exec.Command("/bin/sh", "-c", "id").CombinedOutput()
	checks = append(checks, formatErrCheck("Process Exec", "/bin/sh -c id", execErr,
		fmt.Sprintf("UNEXPECTED EXEC OUT=%q", compactDetail(string(cmdOut)))))

	// 4. Raw socket / network probes (local Auth server & external TCP).
	if conn, err := net.Dial("tcp", "127.0.0.1:7700"); err == nil {
		_ = conn.Close()
		checks = append(checks, SandboxProbeCheck{
			Name:    "Localhost Socket",
			Target:  "127.0.0.1:7700",
			Blocked: false,
			Detail:  "UNEXPECTED DIAL SUCCEEDED",
		})
	} else {
		checks = append(checks, formatErrCheck("Localhost Socket", "127.0.0.1:7700", err, ""))
	}

	if conn, err := net.Dial("tcp", "1.1.1.1:80"); err == nil {
		_ = conn.Close()
		checks = append(checks, SandboxProbeCheck{
			Name:    "Internet Socket",
			Target:  "1.1.1.1:80",
			Blocked: false,
			Detail:  "UNEXPECTED DIAL SUCCEEDED",
		})
	} else {
		checks = append(checks, formatErrCheck("Internet Socket", "1.1.1.1:80", err, ""))
	}

	// 5. Working dir & host environment variables probes.
	wd, wdErr := os.Getwd()
	if wdErr != nil {
		checks = append(checks, formatErrCheck("Host Getwd", "os.Getwd()", wdErr, ""))
	} else if wd == "" || wd == "/" || wd == "." {
		checks = append(checks, SandboxProbeCheck{
			Name:    "Host Getwd",
			Target:  "os.Getwd()",
			Blocked: true,
			Detail:  fmt.Sprintf("Isolated: virtual %q stub only (real host path hidden)", wd),
		})
	} else {
		checks = append(checks, SandboxProbeCheck{
			Name:    "Host Getwd",
			Target:  "os.Getwd()",
			Blocked: false,
			Detail:  "UNEXPECTED HOST CWD LEAK: " + compactDetail(wd),
		})
	}

	envVars := os.Environ()
	if len(envVars) == 0 {
		checks = append(checks, SandboxProbeCheck{
			Name:    "Host Env Vars",
			Target:  "os.Environ()",
			Blocked: true,
			Detail:  "Isolated: 0 host environment variables exposed",
		})
	} else {
		checks = append(checks, SandboxProbeCheck{
			Name:    "Host Env Vars",
			Target:  "os.Environ()",
			Blocked: false,
			Detail:  fmt.Sprintf("UNEXPECTED ENV LEAK (%d vars)", len(envVars)),
		})
	}

	allBlocked := true
	for _, c := range checks {
		if !c.Blocked {
			allBlocked = false
			break
		}
	}

	return SandboxProbeResp{
		Role:         role,
		Path:         path,
		EntriesFound: entriesFound,
		AllBlocked:   allBlocked,
		Checks:       checks,
	}
}

func formatErrCheck(name, target string, err error, leakMsg string) SandboxProbeCheck {
	if err == nil {
		return SandboxProbeCheck{
			Name:    name,
			Target:  target,
			Blocked: false,
			Detail:  leakMsg,
		}
	}
	return SandboxProbeCheck{
		Name:    name,
		Target:  target,
		Blocked: true,
		Detail:  "Blocked: " + compactDetail(err.Error()),
	}
}

func compactDetail(raw string) string {
	var b strings.Builder
	for i := 0; i < len(raw) && b.Len() < 110; i++ {
		c := raw[i]
		if c >= 0x20 && c <= 0x7e {
			b.WriteByte(c)
		} else if c == '\n' || c == '\r' || c == '\t' {
			b.WriteByte(' ')
		}
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return "error"
	}
	return out
}
