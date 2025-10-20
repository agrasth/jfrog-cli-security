package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	commonCommands "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/log"
	"github.com/jfrog/jfrog-cli-security/cli"
	configTests "github.com/jfrog/jfrog-cli-security/tests"
	integrationUtils "github.com/jfrog/jfrog-cli-security/tests/utils/integration"
	clientLog "github.com/jfrog/jfrog-client-go/utils/log"
	clientUtils "github.com/jfrog/jfrog-client-go/utils"
)

func TestMain(m *testing.M) {
	setupIntegrationTests()
	result := m.Run()
	tearDownIntegrationTests()
	os.Exit(result)
}

func setupIntegrationTests() {
	// Disable usage report.
	if err := os.Setenv(coreutils.ReportUsage, "false"); err != nil {
		clientLog.Error(fmt.Sprintf("Couldn't set env: %s. Error: %s", coreutils.ReportUsage, err.Error()))
		os.Exit(1)
	}
	// Disable progress bar and confirmation messages.
	if err := os.Setenv(coreutils.CI, "true"); err != nil {
		clientLog.Error(fmt.Sprintf("Couldn't set env: %s. Error: %s", coreutils.CI, err.Error()))
		os.Exit(1)
	}
	// General
	configTests.InitTestFlags()
	log.SetDefaultLogger()
	// Init
	integrationUtils.InitTestCliDetails(cli.GetJfrogCliSecurityApp())
	integrationUtils.AuthenticateArtifactory()
	integrationUtils.AuthenticateXsc()
	integrationUtils.CreateRequiredRepositories()
	// Create CLI config for binary scan tests (Docker tests create their own isolated configs)
	if *configTests.TestScan {
		createDefaultServerConfig()
	}
}

// createDefaultServerConfig creates the default CLI config for tests without requiring a testing.T context
func createDefaultServerConfig() {
	wd, err := os.Getwd()
	if err != nil {
		clientLog.Error(fmt.Sprintf("Failed to get current dir: %s", err.Error()))
		os.Exit(1)
	}
	if err := os.Setenv(coreutils.HomeDir, filepath.Join(wd, configTests.Out, "jfroghome")); err != nil {
		clientLog.Error(fmt.Sprintf("Failed to set HomeDir: %s", err.Error()))
		os.Exit(1)
	}

	// Delete the default server if exist
	config, err := commonCommands.GetConfig("default", false)
	if err == nil && config.ServerId != "" {
		if err = commonCommands.NewConfigCommand(commonCommands.Delete, "default").Run(); err != nil {
			clientLog.Error(fmt.Sprintf("Failed to delete existing default config: %s", err.Error()))
			os.Exit(1)
		}
	}
	*configTests.JfrogUrl = clientUtils.AddTrailingSlashIfNeeded(*configTests.JfrogUrl)
	if err = commonCommands.NewConfigCommand(commonCommands.AddOrEdit, "default").SetDetails(configTests.XrDetails).SetInteractive(false).SetEncPassword(true).Run(); err != nil {
		clientLog.Error(fmt.Sprintf("Failed to create default config: %s", err.Error()))
		os.Exit(1)
	}
}

func tearDownIntegrationTests() {
	// Important - Virtual repositories must be deleted first
	integrationUtils.DeleteRepos(configTests.CreatedVirtualRepositories)
	integrationUtils.DeleteRepos(configTests.CreatedNonVirtualRepositories)
}
