package remote

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
	v1extensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ErrNotSupportedAnnotation = errors.New("not supported annotation")
)

func PatchCRD(ctx context.Context, clnt client.Client, crd *v1extensions.CustomResourceDefinition) error {
	crdToApply := &v1extensions.CustomResourceDefinition{}
	crdToApply.SetGroupVersionKind(crd.GroupVersionKind())
	crdToApply.SetName(crd.Name)
	crdToApply.Spec = crd.Spec
	crdToApply.Spec.Conversion.Strategy = v1extensions.NoneConverter
	crdToApply.Spec.Conversion.Webhook = nil
	return clnt.Patch(ctx, crdToApply,
		client.Apply,
		client.ForceOwnership,
		client.FieldOwner(v1beta2.OperatorName))
}

type CrdType string

const (
	KCP CrdType = "KCP"
	SKR CrdType = "SKR"
)

func updateRemoteCRD(
	ctx context.Context, plural string, kyma *v1beta2.Kyma, runtimeClient Client, controlPlaneClient Client) (
	bool, *v1extensions.CustomResourceDefinition, *v1extensions.CustomResourceDefinition, error) {
	crd := &v1extensions.CustomResourceDefinition{}
	crdFromRuntime := &v1extensions.CustomResourceDefinition{}
	var err error
	err = controlPlaneClient.Get(
		ctx, client.ObjectKey{
			// this object name is derived from the plural and is the default kustomize value for crd namings, if the CRD
			// name changes, this also has to be adjusted here. We can think of making this configurable later
			Name: fmt.Sprintf("%s.%s", plural, v1beta2.GroupVersion.Group),
		}, crd,
	)

	if err != nil {
		return false, nil, nil, err
	}

	err = runtimeClient.Get(
		ctx, client.ObjectKey{
			Name: fmt.Sprintf("%s.%s", plural, v1beta2.GroupVersion.Group),
		}, crdFromRuntime,
	)

	if ShouldPatchRemoteCRD(crdFromRuntime, crd, kyma, err) {
		err = PatchCRD(ctx, runtimeClient, crd)
		if err != nil {
			return false, nil, nil, err
		}

		err = runtimeClient.Get(
			ctx, client.ObjectKey{
				Name: fmt.Sprintf("%s.%s", plural, v1beta2.GroupVersion.Group),
			}, crdFromRuntime,
		)
		if err != nil {
			return false, nil, nil, err
		}

		return true, crdFromRuntime, crd, nil
	}

	if err != nil {
		return false, nil, nil, err
	}

	return false, nil, nil, nil
}

func ShouldPatchRemoteCRD(
	runtimeCrd *v1extensions.CustomResourceDefinition, kcpCrd *v1extensions.CustomResourceDefinition,
	kyma *v1beta2.Kyma, err error) bool {
	if k8serrors.IsNotFound(err) {
		return true
	}
	kcpAnnotation, _ := getAnnotation(kcpCrd, KCP)
	skrAnnotation, _ := getAnnotation(runtimeCrd, SKR)

	latestGeneration := strconv.FormatInt(kcpCrd.Generation, 10)
	runtimeCRDGeneration := strconv.FormatInt(runtimeCrd.Generation, 10)
	return kyma.Annotations[kcpAnnotation] != latestGeneration ||
		kyma.Annotations[skrAnnotation] != runtimeCRDGeneration
}

func updateKymaAnnotations(kyma *v1beta2.Kyma, crd *v1extensions.CustomResourceDefinition, crdType CrdType) error {
	if kyma.Annotations == nil {
		kyma.Annotations = make(map[string]string)
	}
	annotation, err := getAnnotation(crd, crdType)
	if err != nil {
		return err
	}
	kyma.Annotations[annotation] = strconv.FormatInt(crd.Generation, 10)
	return nil
}

func getAnnotation(crd *v1extensions.CustomResourceDefinition, crdType CrdType) (string, error) {
	if crdType == SKR {
		if crd.Spec.Names.Kind == string(v1beta2.KymaKind) {
			return v1beta2.SkrKymaCRDGenerationAnnotation, nil
		} else if crd.Spec.Names.Kind == string(v1beta2.ModuleTemplateKind) {
			return v1beta2.SkrModuleTemplateCRDGenerationAnnotation, nil
		}
	}

	if crdType == KCP {
		if crd.Spec.Names.Kind == string(v1beta2.KymaKind) {
			return v1beta2.KcpKymaCRDGenerationAnnotation, nil
		} else if crd.Spec.Names.Kind == string(v1beta2.ModuleTemplateKind) {
			return v1beta2.KcpModuleTemplateCRDGenerationAnnotation, nil
		}
	}

	return "", ErrNotSupportedAnnotation
}

func SyncCrdsAndUpdateKymaAnnotations(ctx context.Context, kyma *v1beta2.Kyma,
	runtimeClient Client, controlPlaneClient Client) (bool, error) {
	kymaCrdUpdated, kymaSkrCrd, kymaKcpCrd, err := updateRemoteCRD(ctx, v1beta2.KymaKind.Plural(),
		kyma, runtimeClient, controlPlaneClient)
	if err != nil {
		return false, err
	}
	if kymaCrdUpdated {
		if err := updateKymaAnnotations(kyma, kymaKcpCrd, KCP); err != nil {
			return false, err
		}
		if err := updateKymaAnnotations(kyma, kymaSkrCrd, SKR); err != nil {
			return false, err
		}
	}

	moduleTemplateCrdUpdated, moduleTemplateSkrCrd, moduleTemplateKcpCrd, err := updateRemoteCRD(ctx,
		v1beta2.ModuleTemplateKind.Plural(), kyma, runtimeClient, controlPlaneClient)
	if err != nil {
		return false, err
	}
	if moduleTemplateCrdUpdated {
		if err := updateKymaAnnotations(kyma, moduleTemplateKcpCrd, KCP); err != nil {
			return false, err
		}
		if err := updateKymaAnnotations(kyma, moduleTemplateSkrCrd, SKR); err != nil {
			return false, err
		}
	}

	return kymaCrdUpdated || moduleTemplateCrdUpdated, nil
}

func ContainsLatestVersion(crdFromRuntime *v1extensions.CustomResourceDefinition, latestVersion string) bool {
	for _, version := range crdFromRuntime.Spec.Versions {
		if latestVersion == version.Name {
			return true
		}
	}
	return false
}
