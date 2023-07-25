package primitive

import (
	"bufio"
	"context"
	"os/exec"
	"syscall"
	"time"

	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/pipeparser/schema"
)

type Exec struct{}

func (e *Exec) ValidateInput(ctx context.Context, i types.Input) error {
	if i["command"] == nil {
		return fperr.BadRequestWithMessage("Exec input must define a command")
	}
	return nil
}

func (e *Exec) Run(ctx context.Context, input types.Input) (*types.StepOutput, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	// TODO - support arguments per https://www.terraform.io/language/resources/provisioners/local-exec#argument-reference

	//nolint:gosec // TODO G204: Subprocess launched with a potential tainted input or cmd arguments (gosec)
	cmd := exec.Command("sh", "-c", input["command"].(string))

	// Capture stdout in real-time
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stdoutLines := []string{}
	// TODO - by default this has a max line size of 64K, see https://stackoverflow.com/a/16615559
	stdoutScanner := bufio.NewScanner(stdout)
	go func() {
		for stdoutScanner.Scan() {
			t := stdoutScanner.Text()
			// TODO - send to logs immediately
			stdoutLines = append(stdoutLines, t)
		}
	}()

	// Capture stderr in real-time
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	stderrLines := []string{}
	// TODO - by default this has a max line size of 64K, see https://stackoverflow.com/a/16615559
	stderrScanner := bufio.NewScanner(stderr)
	go func() {
		for stderrScanner.Scan() {
			t := stderrScanner.Text()
			// TODO - send to logs immediately
			stderrLines = append(stderrLines, t)
		}
	}()

	start := time.Now().UTC()
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	exitCode := 0

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0

			// This works on both Unix and Windows. Although package
			// syscall is generally platform dependent, WaitStatus is
			// defined for both Unix and Windows and in both cases has
			// an ExitStatus() method with the same signature.
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		} else {
			// Unexpected error type, set exit_code to -1 (because I don't have a better idea)
			// TODO - log a warning
			exitCode = -1
		}
	}
	finish := time.Now().UTC()

	o := types.StepOutput{
		OutputVariables: map[string]interface{}{},
	}

	o.OutputVariables["exit_code"] = exitCode
	o.OutputVariables["stdout_lines"] = stdoutLines
	o.OutputVariables["stderr_lines"] = stderrLines
	o.OutputVariables[schema.AttributeTypeStartedAt] = start
	o.OutputVariables[schema.AttributeTypeFinishedAt] = finish

	return &o, nil
}