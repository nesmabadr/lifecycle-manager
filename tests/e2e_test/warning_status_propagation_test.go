package e2e_test

import (
	"context"

	"github.com/kyma-project/lifecycle-manager/api/shared"
	"github.com/kyma-project/lifecycle-manager/api/v1beta2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kyma-project/lifecycle-manager/pkg/testutils"
)

var _ = Describe("Warning Status Propagation", Ordered, func() {
	kyma := NewKymaWithSyncLabel("kyma-sample", "kcp-system", v1beta2.DefaultChannel,
		v1beta2.SyncStrategyLocalSecret)
	module := NewTemplateOperator(v1beta2.DefaultChannel)
	moduleCR := NewTestModuleCR(remoteNamespace)

	InitEmptyKymaBeforeAll(kyma)
	CleanupKymaAfterAll(kyma)

	Context("Given an SKR Cluster", func() {
		It("When a Kyma Module is enabled", func() {
			Eventually(EnableModule).
				WithContext(ctx).
				WithArguments(runtimeClient, defaultRemoteKymaName, remoteNamespace, module).
				Should(Succeed())
		})

		It("Then Module CR exists", func() {
			Eventually(ModuleCRExists).
				WithContext(ctx).
				WithArguments(runtimeClient, moduleCR).
				Should(Succeed())
		})

		It("And resource is defined in Manifest CR", func() {
			Eventually(func(g Gomega, ctx context.Context) {
				resource, err := GetManifestResource(ctx, controlPlaneClient,
					kyma.GetName(), kyma.GetNamespace(), module.Name)
				Expect(err).ToNot(HaveOccurred())
				Expect(resource.GetName()).To(Equal(moduleCR.GetName()))
				Expect(resource.GetNamespace()).To(Equal(moduleCR.GetNamespace()))
				Expect(resource.GroupVersionKind().Version).To(Equal(moduleCR.GroupVersionKind().Version))
				Expect(resource.GroupVersionKind().Group).To(Equal(moduleCR.GroupVersionKind().Group))
				Expect(resource.GroupVersionKind().Kind).To(Equal(moduleCR.GroupVersionKind().Kind))
			}).WithContext(ctx).Should(Succeed())
		})

		It("And the Module in KCP Kyma CR is in \"Warning\" State", func() {
			Eventually(CheckModuleState).
				WithContext(ctx).
				WithArguments(controlPlaneClient, kyma.GetName(), kyma.GetNamespace(),
					module.Name, shared.StateWarning).
				Should(Succeed())
		})

		It("Then KCP kyma CR is in \"Warning\" State", func() {
			Eventually(KymaIsInState).
				WithContext(ctx).
				WithArguments(kyma.GetName(), kyma.GetNamespace(), controlPlaneClient, shared.StateWarning).
				Should(Succeed())
		})
	})

	It("When Kyma Module is disabled", func() {
		Eventually(DisableModule).
			WithContext(ctx).
			WithArguments(runtimeClient, defaultRemoteKymaName, remoteNamespace, module.Name).
			Should(Succeed())

		By("Then no Module in KCP Kyma CR spec")
		Eventually(NotContainsModuleInSpec, Timeout, Interval).
			WithContext(ctx).
			WithArguments(runtimeClient, defaultRemoteKymaName, remoteNamespace, module.Name).
			Should(Succeed())

		By("And KCP Kyma CR is in \"Ready\" State")
		Eventually(KymaIsInState).
			WithContext(ctx).
			WithArguments(kyma.GetName(), kyma.GetNamespace(), controlPlaneClient, shared.StateReady).
			Should(Succeed())
	})

	It("When Kyma Module is re-enabled", func() {
		Eventually(EnableModule).
			WithContext(ctx).
			WithArguments(runtimeClient, defaultRemoteKymaName, remoteNamespace, module).
			Should(Succeed())

		By("Then Module is in KCP Kyma CR spec")
		Eventually(ContainsModuleInSpec, Timeout, Interval).
			WithContext(ctx).
			WithArguments(runtimeClient, defaultRemoteKymaName, remoteNamespace, module.Name).
			Should(Succeed())

		By("And KCP Kyma CR is in \"Warning\" State")
		Eventually(KymaIsInState).
			WithContext(ctx).
			WithArguments(kyma.GetName(), kyma.GetNamespace(), controlPlaneClient, shared.StateWarning).
			Should(Succeed())
	})

	It("When the Module is disabled with an existing finalizer", func() {
		Eventually(SetFinalizer).
			WithContext(ctx).
			WithArguments("sample-yaml", "kyma-system", "operator.kyma-project.io", "v1alpha1", "Sample",
				[]string{"sample.kyma-project.io/finalizer", "blocking-finalizer"}, runtimeClient).
			Should(Succeed())

		Eventually(DisableModule).
			WithContext(ctx).
			WithArguments(runtimeClient, defaultRemoteKymaName, remoteNamespace, module.Name).
			Should(Succeed())
	})

	It("Then there is no Module in KCP Kyma CR spec", func() {
		Eventually(NotContainsModuleInSpec, Timeout, Interval).
			WithContext(ctx).
			WithArguments(runtimeClient, defaultRemoteKymaName, remoteNamespace, module.Name).
			Should(Succeed())
	})

	It("And KCP Kyma CR is in \"Warning\" State", func() {
		Eventually(KymaIsInState).
			WithContext(ctx).
			WithArguments(kyma.GetName(), kyma.GetNamespace(), controlPlaneClient, shared.StateWarning).
			Should(Succeed())
		By("And the Module Manifest CR is in a \"Warning\" State")
		Eventually(CheckManifestIsInState).
			WithContext(ctx).
			WithArguments(kyma.GetName(), kyma.GetNamespace(), module.Name, controlPlaneClient, shared.StateWarning).
			Should(Succeed())
	})

	It("When the blocking finalizers from the Module CR get removed", func() {
		Eventually(SetFinalizer).
			WithContext(ctx).
			WithArguments("sample-yaml", "kyma-system", "operator.kyma-project.io", "v1alpha1", "Sample",
				[]string{}, runtimeClient).
			Should(Succeed())

		By("Then the Module CR is removed")
		Eventually(CheckIfNotExists).
			WithContext(ctx).
			WithArguments("sample-yaml", "kyma-system",
				"operator.kyma-project.io", "v1alpha1", "Sample", runtimeClient).
			Should(Succeed())

		By("And the Manifest CR is removed")
		Eventually(ManifestExists).
			WithContext(ctx).
			WithArguments(controlPlaneClient, kyma.GetName(), kyma.GetNamespace(), module.Name).
			Should(Equal(ErrNotFound))

		By("And KCP Kyma CR is in \"Ready\" State")
		Eventually(KymaIsInState).
			WithContext(ctx).
			WithArguments(kyma.GetName(), kyma.GetNamespace(), controlPlaneClient, shared.StateReady).
			Should(Succeed())
	})
})