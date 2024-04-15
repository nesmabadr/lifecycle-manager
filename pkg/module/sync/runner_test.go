package sync_test

import (
	"context"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	machineryruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/kyma-project/lifecycle-manager/api/shared"
	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
	"github.com/kyma-project/lifecycle-manager/pkg/module/sync"
	"github.com/kyma-project/lifecycle-manager/pkg/testutils"
)

const (
	InvalidModulePrefix = "invalid_"
	ModuleShouldKeep    = "ModuleShouldKeep"
	ModuleToBeRemoved   = "ModuleToBeRemoved"
)

func moduleDeletedSuccessfullyMock(_ context.Context, _ client.Object) error {
	return apierrors.NewNotFound(schema.GroupResource{}, "module-no-longer-exists")
}

func moduleStillExistsInClusterMock(_ context.Context, _ client.Object) error {
	return apierrors.NewAlreadyExists(schema.GroupResource{}, "module-still-exists")
}

type ModuleMockMetrics struct {
	mock.Mock
}

func (m *ModuleMockMetrics) RemoveModuleStateMetrics(kymaName, moduleName string) {
	m.Called(kymaName, moduleName)
}

func TestMetricsOnDeleteNoLongerExistingModuleStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                         string
		ModuleInStatus               string
		getModule                    sync.GetModuleFunc
		expectModuleMetricsGetCalled bool
	}{
		{
			"When status.modules contains Manifest not found in cluster, expect RemoveModuleStateMetrics get called",
			ModuleToBeRemoved,
			moduleDeletedSuccessfullyMock,
			true,
		},
		{
			"When status.modules contains Manifest still exits in cluster, expect RemoveModuleStateMetrics not called",
			ModuleToBeRemoved,
			moduleStillExistsInClusterMock,
			false,
		},
		{
			"When status.modules contains not valid Manifest, expect RemoveModuleStateMetrics get called",
			InvalidModulePrefix + ModuleToBeRemoved,
			moduleStillExistsInClusterMock,
			true,
		},
		{
			"When status.modules contains module in spec.module, expect RemoveModuleStateMetrics not called",
			ModuleShouldKeep,
			moduleStillExistsInClusterMock,
			false,
		},
	}

	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			kyma := testutils.NewTestKyma("test-kyma")
			configureModuleInKyma(kyma, []string{ModuleShouldKeep}, []string{testCase.ModuleInStatus})
			mockMetrics := &ModuleMockMetrics{}
			const methodToBeCalled = "RemoveModuleStateMetrics"
			mockMetrics.On(methodToBeCalled, kyma.Name, testCase.ModuleInStatus).Return()
			sync.DeleteNoLongerExistingModuleStatus(context.TODO(), kyma, testCase.getModule, mockMetrics)
			if testCase.expectModuleMetricsGetCalled {
				mockMetrics.AssertCalled(t, methodToBeCalled, kyma.Name, testCase.ModuleInStatus)
			} else {
				mockMetrics.AssertNotCalled(t, methodToBeCalled, kyma.Name, testCase.ModuleInStatus)
			}
		})
	}
}

func TestDeleteNoLongerExistingModuleStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                        string
		ModulesInKymaSpec           []string
		ModulesInKymaStatus         []string
		ExpectedModulesInKymaStatus []string
		getModule                   sync.GetModuleFunc
	}{
		{
			"When status.modules contains valid modules not in spec.module, expect removed and spec.module keep",
			[]string{ModuleShouldKeep},
			[]string{ModuleShouldKeep, ModuleToBeRemoved},
			[]string{ModuleShouldKeep},
			moduleDeletedSuccessfullyMock,
		},
		{
			"When status.modules contains invalid modules not in spec.module, expect removed and spec.module keep",
			[]string{ModuleShouldKeep},
			[]string{ModuleShouldKeep, InvalidModulePrefix + ModuleToBeRemoved},
			[]string{ModuleShouldKeep},
			moduleDeletedSuccessfullyMock,
		},
		{
			"When status.modules contains invalid modules in spec.module, expect keep",
			[]string{InvalidModulePrefix + ModuleShouldKeep},
			[]string{InvalidModulePrefix + ModuleShouldKeep, ModuleToBeRemoved},
			[]string{InvalidModulePrefix + ModuleShouldKeep},
			moduleDeletedSuccessfullyMock,
		},
		{
			"When status.modules contains valid modules not in spec.module, " +
				"expect keep if module still in cluster",
			[]string{ModuleShouldKeep},
			[]string{ModuleShouldKeep, ModuleToBeRemoved},
			[]string{ModuleShouldKeep, ModuleToBeRemoved},
			moduleStillExistsInClusterMock,
		},

		{
			"When status.modules contains invalid modules not in spec.module, expect removed and spec.module keep",
			[]string{ModuleShouldKeep},
			[]string{ModuleShouldKeep, InvalidModulePrefix + ModuleToBeRemoved},
			[]string{ModuleShouldKeep},
			moduleStillExistsInClusterMock,
		},
		{
			"When status.modules contains invalid modules in spec.module, expect keep",
			[]string{InvalidModulePrefix + ModuleShouldKeep},
			[]string{InvalidModulePrefix + ModuleShouldKeep, ModuleToBeRemoved},
			[]string{InvalidModulePrefix + ModuleShouldKeep, ModuleToBeRemoved},
			moduleStillExistsInClusterMock,
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			kyma := testutils.NewTestKyma("test-kyma")
			configureModuleInKyma(kyma, testCase.ModulesInKymaSpec, testCase.ModulesInKymaStatus)
			sync.DeleteNoLongerExistingModuleStatus(context.TODO(), kyma, testCase.getModule, nil)
			var modulesInFinalModuleStatus []string
			for _, moduleStatus := range kyma.Status.Modules {
				modulesInFinalModuleStatus = append(modulesInFinalModuleStatus, moduleStatus.Name)
			}
			sort.Strings(testCase.ExpectedModulesInKymaStatus)
			sort.Strings(modulesInFinalModuleStatus)
			assert.Equal(t, testCase.ExpectedModulesInKymaStatus, modulesInFinalModuleStatus)
		})
	}
}

func configureModuleInKyma(
	kyma *v1beta2.Kyma,
	modulesInKymaSpec, modulesInKymaStatus []string,
) {
	for _, moduleName := range modulesInKymaSpec {
		module := v1beta2.Module{
			Name: moduleName,
		}
		kyma.Spec.Modules = append(kyma.Spec.Modules, module)
	}
	for _, moduleName := range modulesInKymaStatus {
		manifest := &v1beta2.TrackingObject{}
		if strings.Contains(moduleName, InvalidModulePrefix) {
			manifest = nil
		}
		module := v1beta2.ModuleStatus{
			Name:     moduleName,
			Manifest: manifest,
		}
		kyma.Status.Modules = append(kyma.Status.Modules, module)
	}
}

func Test_NeedToUpdate(t *testing.T) {
	type args struct {
		manifestInCluter *v1beta2.Manifest
		manifestObj      *v1beta2.Manifest
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should return true when manifest in cluster is nil",
			args: args{
				manifestInCluter: nil,
				manifestObj: &v1beta2.Manifest{
					Status: shared.Status{
						State: shared.StateReady,
					},
				},
			},
			want: true,
		},
		{
			name: "should return true when manifests channels are different",
			args: args{
				manifestInCluter: &v1beta2.Manifest{
					Status: shared.Status{
						State: shared.StateReady,
					},
					Spec: v1beta2.ManifestSpec{
						Version: "1.0.0",
					},
					ObjectMeta: apimetav1.ObjectMeta{
						Labels: map[string]string{
							shared.ChannelLabel: "fast",
						},
						Generation: 0,
					},
				},
				manifestObj: &v1beta2.Manifest{
					Status: shared.Status{
						State: shared.StateReady,
					},
					Spec: v1beta2.ManifestSpec{
						Version: "1.0.0",
					},
					ObjectMeta: apimetav1.ObjectMeta{
						Labels: map[string]string{
							shared.ChannelLabel: "regular",
						},
						Generation: 0,
					},
				},
			},
			want: true,
		},
		{
			name: "should return true when manifests versions are different",
			args: args{
				manifestInCluter: &v1beta2.Manifest{
					Status: shared.Status{
						State: shared.StateReady,
					},
					Spec: v1beta2.ManifestSpec{
						Version: "1.0.0",
					},
					ObjectMeta: apimetav1.ObjectMeta{
						Labels: map[string]string{
							shared.ChannelLabel: "regular",
						},
						Generation: 0,
					},
				},
				manifestObj: &v1beta2.Manifest{
					Status: shared.Status{
						State: shared.StateReady,
					},
					Spec: v1beta2.ManifestSpec{
						Version: "1.0.1",
					},
					ObjectMeta: apimetav1.ObjectMeta{
						Labels: map[string]string{
							shared.ChannelLabel: "regular",
						},
						Generation: 0,
					},
				},
			},
			want: true,
		},
		{
			name: "should return false when manifests are the same",
			args: args{
				manifestInCluter: &v1beta2.Manifest{
					Status: shared.Status{
						State: shared.StateReady,
					},
					Spec: v1beta2.ManifestSpec{
						Version: "1.0.0",
					},
					ObjectMeta: apimetav1.ObjectMeta{
						Labels: map[string]string{
							shared.ChannelLabel: "regular",
						},
						Generation: 0,
					},
				},
				manifestObj: &v1beta2.Manifest{
					Status: shared.Status{
						State: shared.StateReady,
					},
					Spec: v1beta2.ManifestSpec{
						Version: "1.0.0",
					},
					ObjectMeta: apimetav1.ObjectMeta{
						Labels: map[string]string{
							shared.ChannelLabel: "regular",
						},
						Generation: 0,
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, sync.NeedToUpdate(tt.args.manifestInCluter, tt.args.manifestObj),
				"NeedToUpdate(%v, %v)", tt.args.manifestInCluter, tt.args.manifestObj)
		})
	}
}

func TestRunner_DoUpdateWithStrategy(t *testing.T) {
	type args struct {
		isEnabledModule bool
		manifestObj     *v1beta2.Manifest
	}
	scheme := machineryruntime.NewScheme()
	err := v1beta2.AddToScheme(scheme)
	require.NoError(t, err, "error adding v1beta2 to scheme")
	currentManifest := &v1beta2.Manifest{
		ObjectMeta: apimetav1.ObjectMeta{
			Name:       "manifest",
			Namespace:  "kcp-system",
			Generation: 0,
			Labels: map[string]string{
				shared.ChannelLabel: "regular",
			},
		},
		Spec: v1beta2.ManifestSpec{
			Version: "1.0.0",
			Install: v1beta2.InstallInfo{
				Name: "first",
			},
		},
		Status: shared.Status{
			State: shared.StateReady,
		},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(currentManifest).Build()
	tests := []struct {
		name             string
		args             args
		expectedManifest *v1beta2.Manifest
	}{
		{
			name: "should not patch manifest if manifest is not updated",
			args: args{
				isEnabledModule: true,
				manifestObj: &v1beta2.Manifest{
					ObjectMeta: apimetav1.ObjectMeta{
						Name:       "manifest",
						Namespace:  "kcp-system",
						Generation: 0,
						Labels: map[string]string{
							shared.ChannelLabel: "regular",
						},
					},
					Spec: v1beta2.ManifestSpec{
						Version: "1.0.0",
						Install: v1beta2.InstallInfo{
							Name: "first",
						},
					},
					Status: shared.Status{
						State: shared.StateReady,
						Synced: []shared.Resource{
							{
								Name: "test",
							},
						},
					},
				},
			},
			expectedManifest: currentManifest,
		},
		{
			name: "should update manifest spec only if module is disabled and manifest is updated",
			args: args{
				isEnabledModule: false,
				manifestObj: &v1beta2.Manifest{
					ObjectMeta: apimetav1.ObjectMeta{
						Name:       "manifest",
						Namespace:  "kcp-system",
						Generation: 0,
						Labels: map[string]string{
							shared.ChannelLabel: "fast",
						},
					},
					Spec: v1beta2.ManifestSpec{
						Version: "1.0.0",
						Install: v1beta2.InstallInfo{
							Name: "test",
						},
					},
					Status: shared.Status{
						State: shared.StateReady,
					},
				},
			},
			expectedManifest: &v1beta2.Manifest{
				ObjectMeta: apimetav1.ObjectMeta{
					Name:       "manifest",
					Namespace:  "kcp-system",
					Generation: 0,
					Labels: map[string]string{
						shared.ChannelLabel: "regular",
					},
				},
				Spec: v1beta2.ManifestSpec{
					Version: "1.0.0",
					Install: v1beta2.InstallInfo{
						Name: "test",
					},
				},
				Status: shared.Status{
					State: shared.StateReady,
				},
			},
		},
	}

	for _, tt := range tests {
		testCase := tt
		owner := "test"
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.TODO()
			runner := &sync.Runner{
				Client: fakeClient,
			}

			err = runner.DoUpdateWithStrategy(ctx, owner, testCase.args.isEnabledModule,
				testCase.args.manifestObj)
			require.NoError(t, err, "error updating with strategy")

			var manifest v1beta2.Manifest
			err := runner.Get(ctx,
				client.ObjectKey{Name: testCase.args.manifestObj.Name, Namespace: testCase.args.manifestObj.Namespace},
				&manifest)
			require.NoError(t, err)
			testCase.expectedManifest.Generation = manifest.Generation
			testCase.expectedManifest.ResourceVersion = manifest.ResourceVersion
			require.Equal(t, testCase.expectedManifest, &manifest)
		})
	}
}
