package constants

import (
	"github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/app_specific"
)

// SetAppSpecificConstants sets app specific constants defined in pipe-fittings
func SetAppSpecificConstants() {

	installDir, err := files.Tildefy("~/.flowpipe")
	if err != nil {
		panic(err)
	}

	app_specific.AppName = "flowpipe"
	// TODO unify version logic with steampipe and powerpipe
	//app_specific.AppVersion
	app_specific.AutoVariablesExtension = ".auto.pvars"
	//app_specific.ClientConnectionAppNamePrefix
	//app_specific.ClientSystemConnectionAppNamePrefix
	app_specific.DefaultInstallDir = installDir
	app_specific.DefaultVarsFileName = "flowpipe.pvars"
	//app_specific.DefaultWorkspaceDatabase
	//app_specific.EnvAppPrefix
	app_specific.EnvInputVarPrefix = "P_VAR_"
	//app_specific.InstallDir
	app_specific.ModDataExtension = ".hcl"
	app_specific.ModFileName = "mod.hcl"
	app_specific.VariablesExtension = ".pvars"
	//app_specific.ServiceConnectionAppNamePrefix
	app_specific.WorkspaceIgnoreFile = ".flowpipeignore"
	app_specific.WorkspaceDataDir = ".flowpipe"
}