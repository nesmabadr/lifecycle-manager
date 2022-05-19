/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/kyma-project/kyma-operator/operator/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeModuleTemplates implements ModuleTemplateInterface
type FakeModuleTemplates struct {
	Fake *FakeOperatorV1alpha1
	ns   string
}

var moduletemplatesResource = schema.GroupVersionResource{Group: "operator.kyma-project.io", Version: "v1alpha1", Resource: "moduletemplates"}

var moduletemplatesKind = schema.GroupVersionKind{Group: "operator.kyma-project.io", Version: "v1alpha1", Kind: "ModuleTemplate"}

// Get takes name of the moduleTemplate, and returns the corresponding moduleTemplate object, and an error if there is any.
func (c *FakeModuleTemplates) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.ModuleTemplate, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(moduletemplatesResource, c.ns, name), &v1alpha1.ModuleTemplate{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ModuleTemplate), err
}

// List takes label and field selectors, and returns the list of ModuleTemplates that match those selectors.
func (c *FakeModuleTemplates) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ModuleTemplateList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(moduletemplatesResource, moduletemplatesKind, c.ns, opts), &v1alpha1.ModuleTemplateList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ModuleTemplateList{ListMeta: obj.(*v1alpha1.ModuleTemplateList).ListMeta}
	for _, item := range obj.(*v1alpha1.ModuleTemplateList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested moduleTemplates.
func (c *FakeModuleTemplates) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(moduletemplatesResource, c.ns, opts))

}

// Create takes the representation of a moduleTemplate and creates it.  Returns the server's representation of the moduleTemplate, and an error, if there is any.
func (c *FakeModuleTemplates) Create(ctx context.Context, moduleTemplate *v1alpha1.ModuleTemplate, opts v1.CreateOptions) (result *v1alpha1.ModuleTemplate, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(moduletemplatesResource, c.ns, moduleTemplate), &v1alpha1.ModuleTemplate{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ModuleTemplate), err
}

// Update takes the representation of a moduleTemplate and updates it. Returns the server's representation of the moduleTemplate, and an error, if there is any.
func (c *FakeModuleTemplates) Update(ctx context.Context, moduleTemplate *v1alpha1.ModuleTemplate, opts v1.UpdateOptions) (result *v1alpha1.ModuleTemplate, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(moduletemplatesResource, c.ns, moduleTemplate), &v1alpha1.ModuleTemplate{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ModuleTemplate), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeModuleTemplates) UpdateStatus(ctx context.Context, moduleTemplate *v1alpha1.ModuleTemplate, opts v1.UpdateOptions) (*v1alpha1.ModuleTemplate, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(moduletemplatesResource, "status", c.ns, moduleTemplate), &v1alpha1.ModuleTemplate{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ModuleTemplate), err
}

// Delete takes name of the moduleTemplate and deletes it. Returns an error if one occurs.
func (c *FakeModuleTemplates) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(moduletemplatesResource, c.ns, name, opts), &v1alpha1.ModuleTemplate{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeModuleTemplates) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(moduletemplatesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.ModuleTemplateList{})
	return err
}

// Patch applies the patch and returns the patched moduleTemplate.
func (c *FakeModuleTemplates) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ModuleTemplate, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(moduletemplatesResource, c.ns, name, pt, data, subresources...), &v1alpha1.ModuleTemplate{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ModuleTemplate), err
}
