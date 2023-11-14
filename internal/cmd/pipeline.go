package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/flowpipe/internal/cmd/common"
	"github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/printers"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/error_helpers"
)

// pipeline commands
func pipelineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: "Pipeline commands",
	}

	cmd.AddCommand(pipelineListCmd())
	cmd.AddCommand(pipelineShowCmd())
	cmd.AddCommand(pipelineRunCmd())

	return cmd
}

// list
func pipelineListCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:  "list",
		Args: cobra.NoArgs,
		Run:  listPipelineFunc(),
	}

	return cmd
}

func listPipelineFunc() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		limit := int32(25) // int32 | The max number of items to fetch per page of data, subject to a min and max of 1 and 100 respectively. If not specified will default to 25. (optional) (default to 25)
		nextToken := ""    // string | When list results are truncated, next_token will be returned, which is a cursor to fetch the next page of data. Pass next_token to the subsequent list request to fetch the next page of data. (optional)

		apiClient := common.GetApiClient()
		resp, _, err := apiClient.PipelineApi.List(context.Background()).Limit(limit).NextToken(nextToken).Execute()
		if err != nil {
			error_helpers.ShowError(ctx, err)
			return
		}

		if resp != nil {
			printer := printers.GetPrinter(cmd)

			printableResource := types.PrintablePipeline{}
			printableResource.Items, err = printableResource.Transform(resp)
			if err != nil {
				error_helpers.ShowErrorWithMessage(ctx, err, "Error when transforming")
			}

			err := printer.PrintResource(ctx, printableResource, cmd.OutOrStdout())
			if err != nil {
				error_helpers.ShowErrorWithMessage(ctx, err, "Error when printing")
			}
		}
	}
}

// show
func pipelineShowCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:  "show <pipeline-name>",
		Args: cobra.ExactArgs(1),
		Run:  showPipelineFunc(),
	}

	return cmd
}

func showPipelineFunc() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		apiClient := common.GetApiClient()
		resp, _, err := apiClient.PipelineApi.Get(context.Background(), args[0]).Execute()
		if err != nil {
			error_helpers.ShowError(ctx, err)
			return
		}

		if resp != nil {
			output := ""
			if resp.Title != nil {
				output += "Title: " + *resp.Title
			}
			if resp.Title != nil {
				output += "\nName:  " + *resp.Name
			} else {
				output += "Name: " + *resp.Name
			}
			if resp.Tags != nil {
				if resp.Title != nil {
					output += "\nTags:  "
				} else {
					output += "\nTags: "
				}
				isFirstTag := true
				for k, v := range *resp.Tags {
					if isFirstTag {
						output += k + " = " + v
						isFirstTag = false
					} else {
						output += ", " + k + " = " + v
					}
				}
			}
			if resp.Description != nil {
				output += "\n\nDescription:\n" + *resp.Description + "\n"
			}
			if resp.Params != nil {
				output += formatSection("\nParams:", resp.Params)
			}
			if resp.Outputs != nil {
				output += formatSection("\nOutputs:", resp.Outputs)
			}
			output += "\nUsage:" + "\n"
			if resp.Params != nil {
				var pArg string

				// show the minimal required pipeline args
				for _, param := range resp.Params {
					if (param.Default != nil && len(param.Default) > 0) || (param.Optional != nil && *param.Optional) {
						continue
					}
					pArg += " --pipeline-arg " + *param.Name + "=<value>"
				}
				output += "  flowpipe pipeline run " + *resp.Name + pArg
			} else {
				output += "  flowpipe pipeline run " + *resp.Name
			}
			//nolint:forbidigo // CLI console output
			fmt.Println(output)
		}
	}
}

// run
func pipelineRunCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:  "run <pipeline-name>",
		Args: cobra.ExactArgs(1),
		Run:  runPipelineFunc(),
	}

	// Add the pipeline arg flag
	cmdconfig.OnCmd(cmd).
		AddStringArrayFlag(constants.ArgPipelineArg, nil, "Specify the value of a pipeline argument. Multiple --pipeline-arg may be passed.")

	return cmd
}

func runPipelineFunc() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		// API client
		apiClient := common.GetApiClient()
		cmdPipelineRun := flowpipeapiclient.NewCmdPipeline("run")

		// Get the pipeline args from the flag
		pipelineArgs := map[string]string{}
		pipeLineArgValues, err := cmd.Flags().GetStringArray(constants.ArgPipelineArg)
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error getting the value of pipeline-arg flag")
			return
		}

		// validate the pipeline arg input
		err = validatePipelineArgs(pipeLineArgValues)
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Pipeline argument validation failed")
			return
		}

		for _, value := range pipeLineArgValues {
			splitData := strings.SplitN(value, "=", 2)
			pipelineArgs[splitData[0]] = splitData[1]
		}

		// Set the pipeline args
		cmdPipelineRun.ArgsString = &pipelineArgs

		request := apiClient.PipelineApi.Cmd(ctx, args[0]).Request(*cmdPipelineRun)
		resp, _, err := request.Execute()
		if err != nil {
			error_helpers.ShowError(ctx, err)
			return
		}

		if resp != nil && resp["flowpipe"] != nil {
			contents := resp["flowpipe"].(map[string]any)

			err = pollEventLogAndRender(ctx, apiClient, contents, cmd.OutOrStdout())
			if err != nil {
				error_helpers.ShowErrorWithMessage(ctx, err, "Error when polling event log")
				return
			}
		}
	}
}

// Helper function to format a section
func formatSection(sectionName string, items interface{}) string {
	var output string
	if items != nil {
		output += sectionName + "\n"
		switch v := items.(type) {
		case []flowpipeapiclient.FpPipelineParam:
			for _, item := range v {
				output += "  " + paramToString(item) + "\n"
			}
		case []flowpipeapiclient.ModconfigPipelineOutput:
			for _, item := range v {
				output += "  " + outputToString(item) + "\n"
			}
		}
	}
	return output
}

// Helper function to convert Param to string
func paramToString(param flowpipeapiclient.FpPipelineParam) string {
	var strOutput string
	if param.Optional != nil && *param.Optional {
		strOutput = *param.Name + "[" + *param.Type + ",Optional]"
	} else {
		strOutput = *param.Name + "[" + *param.Type + "]"
	}

	if param.Description != nil && len(*param.Description) > 0 {
		strOutput += ": " + *param.Description
	}
	return strOutput
}

// Helper function to convert Output to string
func outputToString(output flowpipeapiclient.ModconfigPipelineOutput) string {
	strOutput := *output.Name
	if output.Description != nil && len(*output.Description) > 0 {
		strOutput += ": " + *output.Description
	}
	return strOutput
}

func validatePipelineArgs(pipelineArgs []string) error {
	validFormat := regexp.MustCompile(`^[\w-]+=[\S\s]+$`)
	for _, arg := range pipelineArgs {
		if !validFormat.MatchString(arg) {
			return fmt.Errorf("invalid format: %s", arg)
		}
	}
	return nil
}

func pollEventLogAndRender(ctx context.Context, client *flowpipeapiclient.APIClient, input map[string]any, w io.Writer) error {
	isComplete := false
	lastIndexRead := -1
	printer := printers.LogLinePrinter{}

	exId := input["execution_id"].(string)
	pId := input["pipeline_execution_id"].(string)
	stale := false
	loadTime := ""
	if input["is_stale"] != nil && input["is_stale"].(bool) {
		stale = true
		loadTime = input["last_loaded"].(string)
	}

	// Print Execution ID / Stale Info
	pi := types.PrintableLogLine{}
	intro := []types.LogLine{
		{Name: "Execution", Message: exId},
	}
	if stale {
		intro = append(intro, types.LogLine{
			Name:    "Execution",
			Message: fmt.Sprintf("Mod is Stale, last loaded: %s", loadTime),
			IsError: true,
		})
	}
	pi.Items = intro
	err := printer.PrintResource(ctx, pi, w)
	if err != nil {
		error_helpers.ShowErrorWithMessage(ctx, err, "Error when printing introduction")
	}

	// Render Processed Log Lines
	for {
		logs, _, err := client.ProcessApi.GetLog(ctx, exId).Execute()
		if err != nil {
			return err
		}

		printableResource := types.PrintableLogLine{}
		printableResource.Items, err = printableResource.Transform(logs)
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error when transforming")
		}

		var render []types.LogLine
		for logIndex, logEntry := range printableResource.Items.([]types.LogLine) {
			if logIndex > lastIndexRead {
				lastIndexRead = logIndex
				render = append(render, logEntry)
			}
		}

		printableResource.Items = render
		err = printer.PrintResource(ctx, printableResource, w)
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error when printing")
		}

		// Check logs received for termination/completion of execution
		for _, logEntry := range logs.Items {
			if *logEntry.EventType == event.HandlerPipelineFinished || *logEntry.EventType == event.HandlerPipelineFailed {
				if logEntry.Payload != nil {
					payload := make(map[string]any)
					if err := json.Unmarshal([]byte(*logEntry.Payload), &payload); err != nil {
						return err
					}

					if payload["pipeline_execution_id"] != nil && payload["pipeline_execution_id"] == pId {
						isComplete = true
						break
					}
				}
			}
		}
		if isComplete {
			break
		}

		time.Sleep(500 * time.Millisecond)
	}
	return nil
}
