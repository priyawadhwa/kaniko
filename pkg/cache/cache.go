/*
Copyright 2018 Google LLC

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

package cache

import (
	"fmt"

	"github.com/GoogleContainerTools/kaniko/pkg/config"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/authn/k8schain"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// CheckCacheForLayer checks the specified cache for a layer with the tag :cacheKey
// If no cache is specified, one is inferred from the destination provided
func CheckCacheForLayer(opts *config.KanikoOptions, cacheKey string) (v1.Image, error) {
	cache := opts.Cache
	if cache == "" {
		destination := opts.Destinations[0]
		destRef, err := name.NewTag(destination, name.WeakValidation)
		if err != nil {
			return nil, errors.Wrap(err, "getting tag for destination")
		}
		cache = fmt.Sprintf("%s/cache", destRef.Context())
	}
	cache = fmt.Sprintf("%s:%s", cache, cacheKey)
	logrus.Infof("Checking for cached layer %s...", cache)

	cacheRef, err := name.NewTag(cache, name.WeakValidation)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("getting reference for %s", cache))
	}
	k8sc, err := k8schain.NewNoClient()
	if err != nil {
		return nil, err
	}
	kc := authn.NewMultiKeychain(authn.DefaultKeychain, k8sc)
	img, err := remote.Image(cacheRef, remote.WithAuthFromKeychain(kc))
	if err != nil {
		return nil, err
	}
	_, err = img.Layers()
	return img, err
}
