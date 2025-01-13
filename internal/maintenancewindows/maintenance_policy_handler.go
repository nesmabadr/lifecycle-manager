package maintenancewindows

import (
	"fmt"
	"os"

	"github.com/go-logr/logr"

	"github.com/kyma-project/lifecycle-manager/maintenancewindows/resolver"
)

const (
	policyName        = "policy"
	policiesDirectory = "/etc/maintenance-policy"
)

func InitializeMaintenanceWindowsPolicy(log logr.Logger) (*resolver.MaintenanceWindowPolicy, error) {
	if err := os.Setenv(resolver.PolicyPathENV, policiesDirectory); err != nil {
		return nil, fmt.Errorf("failed to set the policy path env variable, %w", err)
	}

	policyFilePath := fmt.Sprintf("%s/%s.json", policiesDirectory, policyName)
	if !maintenancePolicyFileExists(policyFilePath) {
		log.Info("maintenance windows policy file does not exist")
		return nil, nil
	}

	maintenancePolicyPool, err := resolver.GetMaintenancePolicyPool()
	if err != nil {
		return nil, fmt.Errorf("failed to get maintenance policy pool, %w", err)
	}

	maintenancePolicy, err := resolver.GetMaintenancePolicy(maintenancePolicyPool, policyName)
	if err != nil {
		return nil, fmt.Errorf("failed to get maintenance window policy, %w", err)
	}

	return maintenancePolicy, nil
}

func maintenancePolicyFileExists(policyFilePath string) bool {
	if _, err := os.Stat(policyFilePath); os.IsNotExist(err) {
		return false
	}

	return true
}
