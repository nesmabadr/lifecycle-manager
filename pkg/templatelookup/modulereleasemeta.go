package templatelookup

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"errors"
	"fmt"
	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
)

var (
	errChannelNotFound = errors.New("no versions found for channel")
	errNoChannelsFound = errors.New("no channels found for module")
)

func GetModuleReleaseMeta(ctx context.Context, clnt client.Reader, moduleName string,
	namespace string) (*v1beta2.ModuleReleaseMeta,
	error) {
	mrm := &v1beta2.ModuleReleaseMeta{}
	if err := clnt.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      moduleName,
	}, mrm); err != nil {
		return nil, fmt.Errorf("failed to fetch ModuleReleaseMeta for %s: %w", moduleName, err)
	}

	return mrm, nil
}

func GetChannelVersionForModule(moduleReleaseMeta *v1beta2.ModuleReleaseMeta, desiredChannel string) (string, error) {
	channelAssignments := moduleReleaseMeta.Spec.Channels
	if len(channelAssignments) == 0 {
		return "", fmt.Errorf("%w: %s", errNoChannelsFound, moduleReleaseMeta.Name)
	}

	for _, channelAssignment := range channelAssignments {
		if channelAssignment.Channel == desiredChannel {
			return channelAssignment.Version, nil
		}
	}

	return "", fmt.Errorf("%w: %s in module %s", errChannelNotFound, desiredChannel, moduleReleaseMeta.Name)
}
