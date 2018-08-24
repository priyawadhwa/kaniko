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
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func CheckCacheForLayer(opts *config.KanikoOptions, cacheKey string) (v1.Layer, error) {
	destination := opts.Destinations[0]
	destRef, err := name.NewTag(destination, name.WeakValidation)
	if err != nil {
		return nil, errors.Wrap(err, "getting tag for destination")
	}
	name := fmt.Sprintf("%s/cache:%s", destRef.Context(), cacheKey)
	logrus.Infof("checking %s for cached layer", name)
	return nil, nil
}
