package main

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"mcbench/benchmark"
	"mcbench/constants"
	"mcbench/game/scenario"
)

// runInfo is a run's run.json.
type runInfo struct {
	Runner    string         `json:"runner"`
	Started   time.Time      `json:"started"`
	Finished  time.Time      `json:"finished"`
	Commit    string         `json:"commit"`
	Dirty     bool           `json:"dirty"`
	Machine   machine        `json:"machine"`
	Settings  settingsInfo   `json:"settings"`
	Scenarios []string       `json:"scenarios"`
	Planned   int            `json:"planned_games"`
	Statuses  map[string]int `json:"statuses"`
	// Resumed lists each time the run was continued after being cut short.
	Resumed []resumeInfo `json:"resumed,omitempty"`
	// Job is the job file the run plays (mcbench jobs), empty for a run by hand.
	Job string `json:"job,omitempty"`
}

type resumeInfo struct {
	Time   time.Time `json:"time"`
	Server string    `json:"server,omitempty"`
	Commit string    `json:"commit"`
}

type settingsInfo struct {
	Player          string             `json:"player"`
	Provider        string             `json:"provider,omitempty"`
	Model           string             `json:"model,omitempty"`
	API             string             `json:"api"`
	Server          string             `json:"server,omitempty"`
	Temperature     float64            `json:"temperature"`
	MaxTokens       int                `json:"max_tokens"`
	Retries         int                `json:"retries"`
	RateLimitWait   time.Duration      `json:"rate_limit_wait,omitempty"`
	ReasoningEffort string             `json:"reasoning_effort,omitempty"`
	ReasoningBudget int                `json:"reasoning_budget,omitempty"`
	Stream          bool               `json:"stream,omitempty"`
	MaxFailures     int                `json:"max_failures"`
	Instructions    []string           `json:"instructions"`
	Seeds           int                `json:"seeds"`
	Samples         int                `json:"samples"`
	Parallel        int                `json:"parallel"`
	Scoring         map[string]float64 `json:"scoring"`
}

// machine fields are best effort: empty when they cannot be read.
type machine struct {
	Hostname string   `json:"hostname"`
	OS       string   `json:"os"`
	Kernel   string   `json:"kernel"`
	Arch     string   `json:"arch"`
	CPU      string   `json:"cpu"`
	Cores    int      `json:"cores"`
	MemoryMB int64    `json:"memory_mb"`
	GPUs     []string `json:"gpus"`
	Go       string   `json:"go"`
}

func startInfo(scs []*scenario.Scenario, o runSettings, scoring map[string]float64, planned int) runInfo {
	info := runInfo{
		Runner:  gitLine("config", "user.name"),
		Started: time.Now().UTC(),
		Commit:  headCommit(),
		Dirty:   gitLine("status", "--porcelain") != "",
		Machine: readMachine(),
		Settings: settingsInfo{
			Player: o.player, Provider: o.provider, Model: o.modelID, API: o.api, Server: o.server, Temperature: o.temp,
			MaxTokens: o.maxTokens, Retries: o.retries, RateLimitWait: o.rateLimitWait, ReasoningEffort: o.effort, ReasoningBudget: o.budget, Stream: o.stream,
			MaxFailures: o.maxFailures, Instructions: o.instr, Seeds: o.seeds, Samples: o.samples,
			Parallel: o.parallel, Scoring: scoring,
		},
		Planned:  planned,
		Statuses: map[string]int{},
		Job:      o.job,
	}
	for _, sc := range scs {
		info.Scenarios = append(info.Scenarios, sc.ID)
	}
	return info
}

func (info *runInfo) finish(dir string, recs []benchmark.Record) error {
	info.Finished = time.Now().UTC()
	for _, r := range recs {
		info.Statuses[r.Status]++
	}
	return info.write(dir)
}

// write saves run.json. A run writes it when it starts too, so that one that is
// killed can still be resumed.
func (info *runInfo) write(dir string) error {
	b, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, constants.RunInfoFile), append(b, '\n'), 0o644)
}

const (
	osLinux  = "linux"
	osDarwin = "darwin"
)

func readMachine() machine {
	m := machine{Arch: runtime.GOARCH, Cores: runtime.NumCPU(), Go: runtime.Version(), OS: runtime.GOOS}
	m.Hostname, _ = os.Hostname()
	switch runtime.GOOS {
	case osLinux:
		if v := fileField("/etc/os-release", "PRETTY_NAME="); v != "" {
			m.OS = strings.Trim(v, `"`)
		}
		m.Kernel = readTrimmed("/proc/sys/kernel/osrelease")
		m.CPU = strings.TrimLeft(fileField("/proc/cpuinfo", "model name"), "\t :")
		kb, _ := strconv.ParseInt(strings.Fields(fileField("/proc/meminfo", "MemTotal:") + " 0")[0], 10, 64)
		m.MemoryMB = kb / 1024
	case osDarwin:
		m.OS = "macOS " + command("sw_vers", "-productVersion")
		m.Kernel = command("sysctl", "-n", "kern.osrelease")
		m.CPU = command("sysctl", "-n", "machdep.cpu.brand_string")
		bytes, _ := strconv.ParseInt(command("sysctl", "-n", "hw.memsize"), 10, 64)
		m.MemoryMB = bytes >> 20
	}
	m.GPUs = gpus()
	return m
}

func gpus() []string {
	if out := command("nvidia-smi", "--query-gpu=name,memory.total", "--format=csv,noheader"); out != "" {
		return strings.Split(out, "\n")
	}
	var out []string
	for _, line := range strings.Split(command("lspci"), "\n") {
		if strings.Contains(line, "VGA") || strings.Contains(line, "3D controller") || strings.Contains(line, "Display controller") {
			if _, name, ok := strings.Cut(line, ": "); ok {
				out = append(out, name)
			}
		}
	}
	return out
}

func fileField(path, prefix string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if v, ok := strings.CutPrefix(sc.Text(), prefix); ok {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func readTrimmed(path string) string {
	b, _ := os.ReadFile(path)
	return strings.TrimSpace(string(b))
}

func command(name string, args ...string) string {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func gitLine(args ...string) string { return command("git", args...) }

// buildCommit is set at build time (-ldflags "-X main.buildCommit=<sha>") for a
// binary that runs without its repository, such as the jobs worker image.
var buildCommit string

func headCommit() string {
	if buildCommit != "" {
		return buildCommit
	}
	return gitLine("rev-parse", "HEAD")
}
