package image

import (
	"fmt"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/registry/generic"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util"
	"k8s.io/kubernetes/pkg/util/validation/field"

	"github.com/openshift/origin/pkg/image/api"
	"github.com/openshift/origin/pkg/image/api/validation"
)

// imageStrategy implements behavior for Images.
type imageStrategy struct {
	runtime.ObjectTyper
	kapi.NameGenerator
}

// Strategy is the default logic that applies when creating and updating
// Image objects via the REST API.
var Strategy = imageStrategy{kapi.Scheme, kapi.SimpleNameGenerator}

// NamespaceScoped is false for images.
func (imageStrategy) NamespaceScoped() bool {
	return false
}

// PrepareForCreate clears fields that are not allowed to be set by end users on creation.
// It extracts the latest information from the manifest (if available) and sets that onto the object.
func (imageStrategy) PrepareForCreate(obj runtime.Object) {
	newImage := obj.(*api.Image)
	// ignore errors, change in place
	if err := api.ImageWithMetadata(newImage); err != nil {
		util.HandleError(fmt.Errorf("Unable to update image metadata for %q: %v", newImage.Name, err))
	}
}

// Validate validates a new image.
func (imageStrategy) Validate(ctx kapi.Context, obj runtime.Object) field.ErrorList {
	image := obj.(*api.Image)
	return validation.ValidateImage(image)
}

// AllowCreateOnUpdate is false for images.
func (imageStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (imageStrategy) AllowUnconditionalUpdate() bool {
	return false
}

// Canonicalize normalizes the object after validation.
func (imageStrategy) Canonicalize(obj runtime.Object) {
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update.
// It extracts the latest info from the manifest and sets that on the object.
func (imageStrategy) PrepareForUpdate(obj, old runtime.Object) {
	newImage := obj.(*api.Image)
	oldImage := old.(*api.Image)

	// image metadata cannot be altered
	newImage.DockerImageReference = oldImage.DockerImageReference
	newImage.DockerImageMetadata = oldImage.DockerImageMetadata
	newImage.DockerImageManifest = oldImage.DockerImageManifest
	newImage.DockerImageMetadataVersion = oldImage.DockerImageMetadataVersion
	newImage.DockerImageLayers = oldImage.DockerImageLayers

	if err := api.ImageWithMetadata(newImage); err != nil {
		util.HandleError(fmt.Errorf("Unable to update image metadata for %q: %v", newImage.Name, err))
	}
}

// ValidateUpdate is the default update validation for an end user.
func (imageStrategy) ValidateUpdate(ctx kapi.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateImageUpdate(old.(*api.Image), obj.(*api.Image))
}

// MatchImage returns a generic matcher for a given label and field selector.
func MatchImage(label labels.Selector, field fields.Selector) generic.Matcher {
	return generic.MatcherFunc(func(obj runtime.Object) (bool, error) {
		image, ok := obj.(*api.Image)
		if !ok {
			return false, fmt.Errorf("not an image")
		}
		fields := api.ImageToSelectableFields(image)
		return label.Matches(labels.Set(image.Labels)) && field.Matches(fields), nil
	})
}
