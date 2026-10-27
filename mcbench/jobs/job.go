// Package jobs plays runs queued as job files on a git branch: the job schema,
// the branch layout and the worker that claims, plays and settles each job.
package jobs

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"

	"mcbench/constants"
)

func Parse(b []byte) (Job, error) {
	var j Job
	err := yaml.UnmarshalWithOptions(b, &j, yaml.Strict())
	return j, err
}

// Args are the `run` flags of the fields the job sets.
func (j Job) Args() []string {
	var a []string
	if j.Model.Provider != "" {
		a = append(a, "-"+constants.FlagProvider+"="+j.Model.Provider)
	}
	a = flagArg(a, constants.FlagPlayer, j.Player)
	a = flagArg(a, constants.FlagModel, j.Model.ID)
	a = flagArg(a, constants.FlagTemperature, j.Model.Temperature)
	a = flagArg(a, constants.FlagMaxTokens, j.Model.MaxTokens)
	a = flagArg(a, constants.FlagReasoningEffort, j.Model.ReasoningEffort)
	a = flagArg(a, constants.FlagReasoningBudget, j.Model.ReasoningBudget)
	a = flagArg(a, constants.FlagMaxFailures, j.MaxFailures)
	if j.Instructions != nil {
		a = append(a, "-"+constants.FlagInstructions+"="+strings.Join(*j.Instructions, ","))
	}
	if j.Scenarios != nil {
		a = append(a, "-"+constants.FlagOnly+"="+strings.Join(*j.Scenarios, ","))
	}
	a = flagArg(a, constants.FlagSeeds, j.Seeds)
	if len(j.Pick) > 0 {
		seeds := make([]string, len(j.Pick))
		for i, s := range j.Pick {
			seeds[i] = strconv.Itoa(s)
		}
		a = append(a, "-"+constants.FlagPick+"="+strings.Join(seeds, ","))
	}
	a = flagArg(a, constants.FlagSamples, j.Samples)
	return flagArg(a, constants.FlagParallel, j.Parallel)
}

func flagArg[T any](args []string, name string, v *T) []string {
	if v == nil {
		return args
	}
	return append(args, "-"+name+"="+fmt.Sprint(*v))
}
