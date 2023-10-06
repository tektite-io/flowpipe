//nolint:forbidigo // CLI command, expect some fmt.Println
package mod

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/pipeparser/cmdconfig"
	"github.com/turbot/flowpipe/pipeparser/constants"
	"github.com/turbot/flowpipe/pipeparser/error_helpers"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/modinstaller"
	"github.com/turbot/flowpipe/pipeparser/parse"
	"github.com/turbot/flowpipe/pipeparser/utils"
	"github.com/turbot/go-kit/helpers"
)

// mod management commands
func ModCmd(ctx context.Context) (*cobra.Command, error) {
	var cmd = &cobra.Command{
		Use:   "mod [command]",
		Args:  cobra.NoArgs,
		Short: "Flowpipe mod management",
		Long: `Flowpipe mod management.

Mods enable you to run, build, and share dashboards, benchmarks and other resources.

Find pre-built mods in the public registry at https://hub.Flowpipe.io.

Examples:

    # Create a new mod in the current directory
    Flowpipe mod init

    # Install a mod
    Flowpipe mod install github.com/turbot/steampipe-mod-aws-compliance

    # Update a mod
    Flowpipe mod update github.com/turbot/steampipe-mod-aws-compliance

    # List installed mods
    Flowpipe mod list

    # Uninstall a mod
    Flowpipe mod uninstall github.com/turbot/steampipe-mod-aws-compliance
	`,
	}

	cmd.AddCommand(modInstallCmd())
	// cmd.AddCommand(modUninstallCmd())
	// cmd.AddCommand(modUpdateCmd())
	// cmd.AddCommand(modListCmd())
	cmd.AddCommand(modInitCmd())
	cmd.Flags().BoolP(constants.ArgHelp, "h", false, "Help for mod")

	return cmd, nil
}

// install
func modInstallCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "install",
		Run:   runModInstallCmd,
		Short: "Install one or more mods and their dependencies",
		Long:  `Install one or more mods and their dependencies.`,
	}

	cmdconfig.OnCmd(cmd).AddBoolFlag(constants.ArgHelp, false, "Help for init", cmdconfig.FlagOptions.WithShortHand("h"))

	var gitUrlModeEnum = modinstaller.GitUrlModeHTTPS

	cmd.Flags().Var(&gitUrlModeEnum, constants.ArgGitUrlMode, "Git URL mode (https or ssh)")
	err := viper.BindPFlag(constants.ArgGitUrlMode, cmd.Flags().Lookup(constants.ArgGitUrlMode))
	if err != nil {
		log.Fatal(err)
	}

	// cmdconfig.OnCmd(cmd).
	// 	AddBoolFlag(constants.ArgPrune, true, "Remove unused dependencies after installation is complete").
	// 	AddBoolFlag(constants.ArgDryRun, false, "Show which mods would be installed/updated/uninstalled without modifying them").
	// 	AddBoolFlag(constants.ArgForce, false, "Install mods even if plugin/cli version requirements are not met (cannot be used with --dry-run)").

	return cmd
}

func runModInstallCmd(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	utils.LogTime("cmd.runModInstallCmd")
	defer func() {
		utils.LogTime("cmd.runModInstallCmd end")
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
			// exitCode = constants.ExitCodeUnknownErrorPanic
		}
	}()

	// try to load the workspace mod definition
	// - if it does not exist, this will return a nil mod and a nil error
	workspacePath := viper.GetString(constants.ArgModLocation)
	workspaceMod, err := parse.LoadModfile(workspacePath)
	error_helpers.FailOnErrorWithMessage(err, "failed to load mod definition")

	// if no mod was loaded, create a default
	if workspaceMod == nil {
		workspaceMod, err = createWorkspaceMod(ctx, cmd, workspacePath)
		if err != nil {
			error_helpers.FailOnError(err)
		}
	}

	gitUrlMode := viper.GetString(constants.ArgGitUrlMode)

	// if any mod names were passed as args, convert into formed mod names
	opts := modinstaller.NewInstallOpts(workspaceMod, args...)
	opts.ModArgs = utils.TrimGitUrls(opts.ModArgs)
	opts.GitUrlMode = modinstaller.GitUrlMode(gitUrlMode)

	installData, err := modinstaller.InstallWorkspaceDependencies(ctx, opts)
	if err != nil {
		// exitCode = constants.ExitCodeModInstallFailed
		error_helpers.FailOnError(err)
	}

	fmt.Println(modinstaller.BuildInstallSummary(installData))
}

// // uninstall
// func modUninstallCmd() *cobra.Command {
// 	var cmd = &cobra.Command{
// 		Use:   "uninstall",
// 		Run:   runModUninstallCmd,
// 		Short: "Uninstall a mod and its dependencies",
// 		Long:  `Uninstall a mod and its dependencies.`,
// 	}

// 	cmdconfig.OnCmd(cmd).
// 		AddBoolFlag(constants.ArgPrune, true, "Remove unused dependencies after uninstallation is complete").
// 		AddBoolFlag(constants.ArgDryRun, false, "Show which mods would be uninstalled without modifying them").
// 		AddBoolFlag(constants.ArgHelp, false, "Help for uninstall", cmdconfig.FlagOptions.WithShortHand("h"))

// 	return cmd
// }

// func runModUninstallCmd(cmd *cobra.Command, args []string) {
// 	ctx := cmd.Context()
// 	utils.LogTime("cmd.runModInstallCmd")
// 	defer func() {
// 		utils.LogTime("cmd.runModInstallCmd end")
// 		if r := recover(); r != nil {
// 			error_helpers.ShowError(ctx, helpers.ToError(r))
// 			exitCode = constants.ExitCodeUnknownErrorPanic
// 		}
// 	}()

// 	// try to load the workspace mod definition
// 	// - if it does not exist, this will return a nil mod and a nil error
// 	workspaceMod, err := parse.LoadModfile(viper.GetString(constants.ArgModLocation))
// 	error_helpers.FailOnErrorWithMessage(err, "failed to load mod definition")
// 	if workspaceMod == nil {
// 		fmt.Println("No mods installed.")
// 		return
// 	}
// 	opts := modinstaller.NewInstallOpts(workspaceMod, args...)
// 	trimGitUrls(opts)
// 	installData, err := modinstaller.UninstallWorkspaceDependencies(ctx, opts)
// 	error_helpers.FailOnError(err)

// 	fmt.Println(modinstaller.BuildUninstallSummary(installData))
// }

// // update
// func modUpdateCmd() *cobra.Command {
// 	var cmd = &cobra.Command{
// 		Use:   "update",
// 		Run:   runModUpdateCmd,
// 		Short: "Update one or more mods and their dependencies",
// 		Long:  `Update one or more mods and their dependencies.`,
// 	}

// 	cmdconfig.OnCmd(cmd).
// 		AddBoolFlag(constants.ArgPrune, true, "Remove unused dependencies after update is complete").
// 		AddBoolFlag(constants.ArgForce, false, "Update mods even if plugin/cli version requirements are not met (cannot be used with --dry-run)").
// 		AddBoolFlag(constants.ArgDryRun, false, "Show which mods would be updated without modifying them").
// 		AddBoolFlag(constants.ArgHelp, false, "Help for update", cmdconfig.FlagOptions.WithShortHand("h"))

// 	return cmd
// }

// func runModUpdateCmd(cmd *cobra.Command, args []string) {
// 	ctx := cmd.Context()
// 	utils.LogTime("cmd.runModUpdateCmd")
// 	defer func() {
// 		utils.LogTime("cmd.runModUpdateCmd end")
// 		if r := recover(); r != nil {
// 			error_helpers.ShowError(ctx, helpers.ToError(r))
// 			exitCode = constants.ExitCodeUnknownErrorPanic
// 		}
// 	}()

// 	// try to load the workspace mod definition
// 	// - if it does not exist, this will return a nil mod and a nil error
// 	workspaceMod, err := parse.LoadModfile(viper.GetString(constants.ArgModLocation))
// 	error_helpers.FailOnErrorWithMessage(err, "failed to load mod definition")
// 	if workspaceMod == nil {
// 		fmt.Println("No mods installed.")
// 		return
// 	}

// 	opts := modinstaller.NewInstallOpts(workspaceMod, args...)
// 	trimGitUrls(opts)
// 	installData, err := modinstaller.InstallWorkspaceDependencies(ctx, opts)
// 	error_helpers.FailOnError(err)

// 	fmt.Println(modinstaller.BuildInstallSummary(installData))
// }

// // list
// func modListCmd() *cobra.Command {
// 	var cmd = &cobra.Command{
// 		Use:   "list",
// 		Run:   runModListCmd,
// 		Short: "List currently installed mods",
// 		Long:  `List currently installed mods.`,
// 	}

// 	cmdconfig.OnCmd(cmd).AddBoolFlag(constants.ArgHelp, false, "Help for list", cmdconfig.FlagOptions.WithShortHand("h"))
// 	return cmd
// }

// func runModListCmd(cmd *cobra.Command, _ []string) {
// 	ctx := cmd.Context()
// 	utils.LogTime("cmd.runModListCmd")
// 	defer func() {
// 		utils.LogTime("cmd.runModListCmd end")
// 		if r := recover(); r != nil {
// 			error_helpers.ShowError(ctx, helpers.ToError(r))
// 			exitCode = constants.ExitCodeUnknownErrorPanic
// 		}
// 	}()

// 	// try to load the workspace mod definition
// 	// - if it does not exist, this will return a nil mod and a nil error
// 	workspaceMod, err := parse.LoadModfile(viper.GetString(constants.ArgModLocation))
// 	error_helpers.FailOnErrorWithMessage(err, "failed to load mod definition")
// 	if workspaceMod == nil {
// 		fmt.Println("No mods installed.")
// 		return
// 	}

// 	opts := modinstaller.NewInstallOpts(workspaceMod)
// 	installer, err := modinstaller.NewModInstaller(opts)
// 	error_helpers.FailOnError(err)

// 	treeString := installer.GetModList()
// 	if len(strings.Split(treeString, "\n")) > 1 {
// 		fmt.Println()
// 	}
// 	fmt.Println(treeString)
// }

// // init
func modInitCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "init",
		Run:   runModInitCmd,
		Short: "Initialize the current directory with a mod.sp file",
		Long:  `Initialize the current directory with a mod.sp file.`,
	}

	cmdconfig.OnCmd(cmd).AddBoolFlag(constants.ArgHelp, false, "Help for init", cmdconfig.FlagOptions.WithShortHand("h"))

	return cmd
}

func createWorkspaceMod(ctx context.Context, cmd *cobra.Command, workspacePath string) (*modconfig.Mod, error) {
	if !modinstaller.ValidateModLocation(ctx, workspacePath) {
		return nil, fmt.Errorf("mod %s cancelled", cmd.Name())
	}

	if parse.ModfileExists(workspacePath) {
		fmt.Println("Working folder already contains a mod definition file")
		return nil, nil
	}
	mod := modconfig.CreateDefaultMod(workspacePath)
	if err := mod.Save(); err != nil {
		return nil, err
	}

	// TODO: this needs the HCL stuff
	// load up the written mod file so that we get the updated
	// block ranges
	// mod, err := parse.LoadModfile(workspacePath)
	// if err != nil {
	// 	return nil, err
	// }

	return mod, nil
}

func runModInitCmd(cmd *cobra.Command, args []string) {
	workspacePath := viper.GetString(constants.ArgModLocation)
	_, err := createWorkspaceMod(cmd.Context(), cmd, workspacePath)
	if err != nil {
		log.Fatal(err)
	}

}

// 	utils.LogTime("cmd.runModInitCmd")
// 	ctx := cmd.Context()

// 	defer func() {
// 		utils.LogTime("cmd.runModInitCmd end")
// 		if r := recover(); r != nil {
// 			error_helpers.ShowError(ctx, helpers.ToError(r))
// 			exitCode = constants.ExitCodeUnknownErrorPanic
// 		}
// 	}()
// 	workspacePath := viper.GetString(constants.ArgModLocation)
// 	if _, err := createWorkspaceMod(ctx, cmd, workspacePath); err != nil {
// 		exitCode = constants.ExitCodeModInitFailed
// 		error_helpers.FailOnError(err)
// 	}
// 	fmt.Printf("Created mod definition file '%s'\n", filepaths.ModFilePath(workspacePath))
// }

// // helpers
// func createWorkspaceMod(ctx context.Context, cmd *cobra.Command, workspacePath string) (*modconfig.Mod, error) {
// 	if !modinstaller.ValidateModLocation(ctx, workspacePath) {
// 		return nil, fmt.Errorf("mod %s cancelled", cmd.Name())
// 	}

// 	if parse.ModfileExists(workspacePath) {
// 		fmt.Println("Working folder already contains a mod definition file")
// 		return nil, nil
// 	}
// 	mod := modconfig.CreateDefaultMod(workspacePath)
// 	if err := mod.Save(); err != nil {
// 		return nil, err
// 	}

// 	// load up the written mod file so that we get the updated
// 	// block ranges
// 	mod, err := parse.LoadModfile(workspacePath)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return mod, nil
// }

// // Modifies(trims) the URL if contains http ot https in arguments
// func trimGitUrls(opts *modinstaller.InstallOpts) {
// 	for i, url := range opts.ModArgs {
// 		opts.ModArgs[i] = strings.TrimPrefix(url, "http://")
// 		opts.ModArgs[i] = strings.TrimPrefix(url, "https://")
// 	}
// }
