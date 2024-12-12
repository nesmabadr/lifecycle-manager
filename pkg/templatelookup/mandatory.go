package templatelookup

import (
	"context"
	"fmt"

	k8slabels "k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Masterminds/semver/v3"
	"github.com/kyma-project/lifecycle-manager/api/shared"
	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
)

// GetMandatory returns ModuleTemplates TOs (Transfer Objects) which are marked are mandatory modules.
func GetMandatory(ctx context.Context, kymaClient client.Reader) (ModuleTemplatesByModuleName,
	error,
) {
	mandatoryModuleTemplateList := &v1beta2.ModuleTemplateList{}
	labelSelector := k8slabels.SelectorFromSet(k8slabels.Set{shared.IsMandatoryModule: shared.EnableLabelValue})
	if err := kymaClient.List(ctx, mandatoryModuleTemplateList,
		&client.ListOptions{LabelSelector: labelSelector}); err != nil {
		return nil, fmt.Errorf("could not list mandatory ModuleTemplates: %w", err)
	}

	mandatoryModules := make(map[string]*ModuleTemplateInfo)
	for _, moduleTemplate := range mandatoryModuleTemplateList.Items {
		if moduleTemplate.DeletionTimestamp.IsZero() {
			currentModuleTemplate := &moduleTemplate
			if mandatoryModules[currentModuleTemplate.Name] != nil {
				var err error
				currentModuleTemplate, err = getDesiredModuleTemplateForMultipleVersions(currentModuleTemplate,
					mandatoryModules[currentModuleTemplate.Name].ModuleTemplate)
				if err != nil {
					mandatoryModules[getModuleName(currentModuleTemplate)] = &ModuleTemplateInfo{
						ModuleTemplate: nil,
						Err:            err,
					}
					continue
				}
			}
			mandatoryModules[getModuleName(currentModuleTemplate)] = &ModuleTemplateInfo{
				ModuleTemplate: currentModuleTemplate,
				Err:            nil,
			}
		}
	}
	return mandatoryModules, nil
}

// TODO: Create an issue to remove this function and only use the spec.ModuleName when mandatory modules use modulectl
func getModuleName(moduleTemplate *v1beta2.ModuleTemplate) string {
	if moduleTemplate.Spec.ModuleName != "" {
		return moduleTemplate.Spec.ModuleName
	}

	return moduleTemplate.Labels[shared.ModuleName]
}

// TODO: Create an issue to remove this function and only use the spec.Version when mandatory modules use modulectl
func getModuleSemverVersion(moduleTemplate *v1beta2.ModuleTemplate) (*semver.Version, error) {
	if moduleTemplate.Spec.Version != "" {
		version, err := semver.NewVersion(moduleTemplate.Spec.Version)
		if err != nil {
			return &semver.Version{}, fmt.Errorf("could not parse version %s: %w", moduleTemplate.Spec.Version, err)
		}
		return version, nil
	}

	version, err := semver.NewVersion(moduleTemplate.Annotations[shared.ModuleVersionAnnotation])
	if err != nil {
		return &semver.Version{}, fmt.Errorf("could not parse version %s: %w",
			moduleTemplate.Annotations[shared.ModuleVersionAnnotation], err)
	}
	return version, nil
}

func getDesiredModuleTemplateForMultipleVersions(firstModuleTemplate, secondModuleTemplate *v1beta2.ModuleTemplate) (*v1beta2.ModuleTemplate,
	error) {
	firstVersion, err := getModuleSemverVersion(firstModuleTemplate)
	if err != nil {
		return nil, err
	}

	secondVersion, err := getModuleSemverVersion(secondModuleTemplate)
	if err != nil {
		return nil, err
	}

	if firstVersion.GreaterThan(secondVersion) {
		return firstModuleTemplate, nil
	}

	return secondModuleTemplate, nil
}
