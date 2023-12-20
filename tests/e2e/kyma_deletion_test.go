package e2e_test

import (
	"os/exec"

	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/lifecycle-manager/api/shared"
	"github.com/kyma-project/lifecycle-manager/api/v1beta2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kyma-project/lifecycle-manager/pkg/testutils"
)

var _ = Describe("KCP Kyma CR Deletion After SKR Cluster Removal", Ordered,
	func() {
		DescribeTable("KCP Kyma CR Deletion After SKR Cluster Removal",
			func(propagation apimetav1.DeletionPropagation) {
				kyma := NewKymaWithSyncLabel("kyma-sample", "kcp-system", v1beta2.DefaultChannel,
					v1beta2.SyncStrategyLocalSecret)
				module := NewTemplateOperator(v1beta2.DefaultChannel)
				moduleCR := NewTestModuleCR(remoteNamespace)

				InitEmptyKymaBeforeAll(kyma)

				Context("Given SKR Cluster", func() {
					It("When Kyma Module is enabled on SKR Cluster", func() {
						Eventually(EnableModule).
							WithContext(ctx).
							WithArguments(runtimeClient, defaultRemoteKymaName, remoteNamespace, module).
							Should(Succeed())
						Eventually(ModuleCRExists).
							WithContext(ctx).
							WithArguments(runtimeClient, moduleCR).
							Should(Succeed())

						By("And finalizer is added to Module CR")
						Expect(AddFinalizerToModuleCR(ctx, runtimeClient, moduleCR, moduleCRFinalizer)).
							Should(Succeed())

						By("And Kyma Module is disabled")
						Eventually(DisableModule).
							WithContext(ctx).
							WithArguments(runtimeClient, defaultRemoteKymaName, remoteNamespace, module.Name).
							Should(Succeed())
					})

					It("Then Module CR stays in \"Deleting\" State", func() {
						Eventually(ModuleCRIsInExpectedState).
							WithContext(ctx).
							WithArguments(runtimeClient, moduleCR, shared.StateDeleting).
							Should(BeTrue())
						Consistently(ModuleCRIsInExpectedState).
							WithContext(ctx).
							WithArguments(runtimeClient, moduleCR, shared.StateDeleting).
							Should(BeTrue())
					})

					It("When KCP Kyma CR is deleted", func() {
						Eventually(DeleteKyma).
							WithContext(ctx).
							WithArguments(controlPlaneClient, kyma, apimetav1.DeletePropagationForeground).
							Should(Succeed())
					})

					It("Then KCP Kyma CR still exists", func() {
						Eventually(KymaExists).
							WithContext(ctx).
							WithArguments(controlPlaneClient, kyma.GetName(), kyma.GetNamespace()).
							Should(Equal(ErrDeletionTimestampFound))
					})

					It("When SKR Cluster is removed", func() {
						cmd := exec.Command("k3d", "cluster", "rm", "skr")
						out, err := cmd.CombinedOutput()
						Expect(err).NotTo(HaveOccurred())
						GinkgoWriter.Printf(string(out))
					})

					It("Then KCP Kyma CR is in \"Error\" State", func() {
						Eventually(KymaIsInState).
							WithContext(ctx).
							WithArguments(kyma.GetName(), kyma.GetNamespace(), controlPlaneClient, shared.StateError).
							Should(Succeed())
					})

					It("When Kubeconfig Secret is deleted", func() {
						Eventually(DeleteKymaSecret).
							WithContext(ctx).
							WithArguments(kyma.GetName(), kyma.GetNamespace(), controlPlaneClient).
							Should(Succeed())
					})

					It("Then KCP Kyma CR is deleted", func() {
						Eventually(KymaDeleted).
							WithContext(ctx).
							WithArguments(kyma.GetName(), kyma.GetNamespace(), controlPlaneClient).
							Should(Succeed())
					})
				})
			},
			Entry("Test Background Propagation Deletion", apimetav1.DeletePropagationBackground),
			Entry("Test Foreground Propagation Deletion", apimetav1.DeletePropagationForeground),
		)
	})
