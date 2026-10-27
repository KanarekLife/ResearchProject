package jobs

import (
	"time"

	"mcbench/online"
)

// Job is one job: the settings of one run. A field left out keeps the worker's
// default (config.yaml, then the built-in one), as a flag left out of `run`
// does; an empty list is set (instructions: [] is no documents).
type Job struct {
	Name         string    `yaml:"name"` // shown in the log; the job is its file name
	Player       *string   `yaml:"player"`
	Model        Model     `yaml:"model"`
	MaxFailures  *int      `yaml:"max_failures"`
	Instructions *[]string `yaml:"instructions"`
	Scenarios    *[]string `yaml:"scenarios"` // []: all
	Seeds        *int      `yaml:"seeds"`
	Pick         []int     `yaml:"pick"`
	Samples      *int      `yaml:"samples"`
	Parallel     *int      `yaml:"parallel"`
}

// Model is what is measured. How the model is reached (api, server, key,
// headers, stream, retries) is the worker's provider of that name.
type Model struct {
	Provider        string   `yaml:"provider"`
	ID              *string  `yaml:"id"`
	Temperature     *float64 `yaml:"temperature"`
	MaxTokens       *int     `yaml:"max_tokens"`
	ReasoningEffort *string  `yaml:"reasoning_effort"`
	ReasoningBudget *int     `yaml:"reasoning_budget"`
}

// Branch is the jobs branch and its local checkout.
type Branch struct {
	Store online.Store
}

// Claimed is a job moved to running/ by this worker.
type Claimed struct {
	Name string
	Data []byte
}

// Run is a local run directory of a job, as the worker reads it back.
type Run struct {
	Dir       string
	Statuses  map[string]int
	Planned   int
	Published bool
}

// Worker plays the jobs of the jobs branch one at a time. A job is claimed by
// pushing its move to running/, so two workers never play the same one. A job
// that fails is never resumed automatically: its error.txt says how.
type Worker struct {
	Branch Branch
	Host   string
	// Args start the `run` flags of every job, e.g. its config file.
	Args []string
	// Play runs `run` with these flags.
	Play func(args []string) error
	// LatestRun is the newest local run of a job started at since or later.
	LatestRun func(job string, since time.Time) *Run
}

// failure is a job's error.txt: why it failed and, when it has a run, how to
// resume that run on the worker.
type failure struct {
	job, worker, reason string
	time                time.Time
	run                 *Run
	args                []string
}
