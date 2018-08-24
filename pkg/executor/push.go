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

package executor

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/GoogleContainerTools/kaniko/pkg/config"
	"github.com/GoogleContainerTools/kaniko/pkg/constants"
	"github.com/GoogleContainerTools/kaniko/pkg/version"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/authn/k8schain"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type withUserAgent struct {
	t http.RoundTripper
}

func (w *withUserAgent) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("User-Agent", fmt.Sprintf("kaniko/%s", version.Version()))
	return w.t.RoundTrip(r)
}

func PushLayerToCache(opts *config.KanikoOptions, cacheKey string, layer v1.Layer, createdBy string) error {
	logrus.Infof("Trying to push layer to cache now")
	destination := opts.Destinations[0]
	destRef, err := name.NewTag(destination, name.WeakValidation)
	if err != nil {
		return errors.Wrap(err, "getting tag for destination")
	}
	cacheName := fmt.Sprintf("%s/cache:%s", destRef.Context(), cacheKey)
	logrus.Infof("pushing layer %s to cache", cacheName)
	empty := empty.Image
	empty, err = mutate.Append(empty,
		mutate.Addendum{
			Layer: layer,
			History: v1.History{
				Author:    constants.Author,
				CreatedBy: createdBy,
			},
		},
	)
	if err != nil {
		return errors.Wrap(err, "appending layer onto empty image")
	}
	return DoPush(empty, &config.KanikoOptions{
		Destinations: []string{cacheName},
	})
}

// DoPush is responsible for pushing image to the destinations specified in opts
func DoPush(image v1.Image, opts *config.KanikoOptions) error {
	if opts.NoPush {
		logrus.Info("Skipping push to container registry due to --no-push flag")
		return nil
	}
	// continue pushing unless an error occurs
	for _, destination := range opts.Destinations {
		// Push the image
		destRef, err := name.NewTag(destination, name.WeakValidation)
		if err != nil {
			return errors.Wrap(err, "getting tag for destination")
		}

		if opts.DockerInsecureSkipTLSVerify {
			newReg, err := name.NewInsecureRegistry(destRef.Repository.Registry.Name(), name.WeakValidation)
			if err != nil {
				return errors.Wrap(err, "getting new insecure registry")
			}
			destRef.Repository.Registry = newReg
		}

		if opts.TarPath != "" {
			return tarball.WriteToFile(opts.TarPath, destRef, image, nil)
		}

		k8sc, err := k8schain.NewNoClient()
		if err != nil {
			return errors.Wrap(err, "getting k8schain client")
		}
		kc := authn.NewMultiKeychain(authn.DefaultKeychain, k8sc)
		pushAuth, err := kc.Resolve(destRef.Context().Registry)
		if err != nil {
			return errors.Wrap(err, "resolving pushAuth")
		}

		// Create a transport to set our user-agent.
		tr := http.DefaultTransport
		if opts.DockerInsecureSkipTLSVerify {
			tr.(*http.Transport).TLSClientConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		rt := &withUserAgent{t: tr}

		if err := remote.Write(destRef, image, pushAuth, rt, remote.WriteOptions{}); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to push to destination %s", destination))
		}
	}
	return nil
}
