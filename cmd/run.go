package cmd

import (
	"errors"
	"kool-dev/kool/cmd/builder"
	"kool-dev/kool/cmd/parser"
	"kool-dev/kool/environment"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

// KoolRun holds handlers and functions to implement the run command logic
type KoolRun struct {
	DefaultKoolService
	parser     parser.Parser
	envStorage environment.EnvStorage
	commands   []builder.Command
}

// ErrExtraArguments Extra arguments error
var ErrExtraArguments = errors.New("error: you cannot pass in extra arguments to multiple commands scripts")

// ErrKoolScriptNotFound means that the given script was not found
var ErrKoolScriptNotFound = errors.New("script was not found in any kool.yml file")

func init() {
	var (
		run    = NewKoolRun()
		runCmd = NewRunCommand(run)
	)

	rootCmd.AddCommand(runCmd)
}

// NewKoolRun creates a new handler for run logic with default dependencies
func NewKoolRun() *KoolRun {
	return &KoolRun{
		*newDefaultKoolService(),
		parser.NewParser(),
		environment.NewEnvStorage(),
		[]builder.Command{},
	}
}

// Execute runs the run logic with incoming arguments.
func (r *KoolRun) Execute(originalArgs []string) (err error) {
	var (
		script string
		args   []string
	)

	script = originalArgs[0]
	args = originalArgs[1:]

	if r.commands, err = r.parser.Parse(script); err != nil {
		if parser.IsMultipleDefinedScriptError(err) {
			// we should just warn the user about multiple finds for the script
			r.Warning("Attention: the script was found in more than one kool.yml file")
		} else {
			return
		}
	}

	if len(r.commands) == 0 {
		err = ErrKoolScriptNotFound
		return
	}

	if len(args) > 0 && len(r.commands) > 1 {
		err = ErrExtraArguments
		return
	}

	for _, command := range r.commands {
		if len(args) > 0 {
			command.AppendArgs(args...)
		}

		if err = command.Interactive(); err != nil {
			return
		}
	}
	return
}

// NewRunCommand initializes new kool stop command
func NewRunCommand(run *KoolRun) (runCmd *cobra.Command) {
	runCmd = &cobra.Command{
		Use:   "run [SCRIPT]",
		Short: "Runs a custom command defined at kool.yaml in the working directory or in the kool folder of the user's home directory",
		Args:  cobra.MinimumNArgs(1),
		Run:   DefaultCommandRunFunction(run),
	}

	// after a non-flag arg, stop parsing flags
	runCmd.Flags().SetInterspersed(false)

	// look for kool.yml on current working directory
	_ = run.parser.AddLookupPath(run.envStorage.Get("PWD"))
	// look for kool.yml on kool folder within user home directory
	_ = run.parser.AddLookupPath(path.Join(run.envStorage.Get("HOME"), "kool"))

	var (
		usageTempl string
		err        error
	)

	if usageTempl, err = getRunUsageTemplate(run, runCmd); err == nil {
		runCmd.SetUsageTemplate(usageTempl)
	} else if run.envStorage.IsTrue("KOOL_VERBOSE") {
		run.Println("$ got an error trying to add available scripts to command usage template; error:", err.Error())
	}

	return
}

func getRunUsageTemplate(run *KoolRun, cmd *cobra.Command) (templ string, err error) {
	var (
		sb      strings.Builder
		scripts []string
	)

	if scripts, err = run.parser.ParseAvailableScripts(); err != nil {
		return
	}

	sb.WriteString(cmd.UsageTemplate())
	sb.WriteString("\n")
	sb.WriteString("Available Scripts:\n")

	for _, script := range scripts {
		sb.WriteString("  ")
		sb.WriteString(script)
		sb.WriteString("\n")
	}

	templ = sb.String()
	return
}
