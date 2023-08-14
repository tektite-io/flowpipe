package modconfig

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

type PluginVersionString struct {
	version string
	semver  *semver.Version
}

func NewPluginVersionString(version string) (*PluginVersionString, error) {
	if smv, err := semver.NewVersion(version); err == nil {
		pluginVersion := &PluginVersionString{
			version: version,
			semver:  smv,
		}
		return pluginVersion, nil
	}
	if version == "local" {
		return LocalPluginVersionString(), nil
	}
	return nil, fmt.Errorf("version must be a valid semver or 'local'; got: %s", version)
}

func LocalPluginVersionString() *PluginVersionString {
	return &PluginVersionString{
		version: "local",
	}
}

func (p *PluginVersionString) IsLocal() bool {
	return p.semver == nil
}

func (p *PluginVersionString) IsSemver() bool {
	return p.semver != nil
}

func (p *PluginVersionString) Semver() *semver.Version {
	return p.semver
}

func (p *PluginVersionString) String() string {
	return p.version
}
