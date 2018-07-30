package local

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/GoogleContainerTools/kaniko/pkg/commands"
	"github.com/GoogleContainerTools/kaniko/pkg/constants"
	"github.com/GoogleContainerTools/kaniko/pkg/dockerfile"
	"github.com/GoogleContainerTools/kaniko/pkg/executor"
	"github.com/GoogleContainerTools/kaniko/pkg/snapshot"
	"github.com/GoogleContainerTools/kaniko/pkg/util"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/sirupsen/logrus"
)

func DoBuild(k executor.KanikoBuildArgs) (v1.Image, error) {
	// Parse dockerfile and unpack base image to root
	d, err := ioutil.ReadFile(k.DockerfilePath)
	if err != nil {
		return nil, err
	}

	stages, err := dockerfile.Parse(d)
	if err != nil {
		return nil, err
	}

	for index, stage := range stages {
		sourceImage, err := util.RetrieveSourceImage(index, k.Args, stages)
		if err != nil {
			return nil, err
		}
		imageConfig, err := sourceImage.ConfigFile()
		if sourceImage == empty.Image {
			imageConfig.Config.Env = constants.ScratchEnvVars
		}
		hasher, err := executor.GetHasher(constants.SnapshotModeFull)
		if err != nil {
			return nil, err
		}
		l := snapshot.NewLayeredMap(hasher)
		snapshotter := snapshot.NewSnapshotter(l, k.SrcContext)
		for _, cmd := range stage.Commands {
			dockerCommand, err := commands.GetCommand(cmd, k.SrcContext)
			if err != nil {
				return nil, err
			}
			if err := dockerCommand.ExecuteCommandLocally(&imageConfig.Config, nil); err != nil {
				return nil, err
			}
			// If there are files to add to the image, then add them here
			if snapshotFiles := dockerCommand.LocalFilesToSnapshot(); snapshotFiles != nil {
				contents, err := snapshotter.TakeLocalSnapshot(snapshotFiles)
				if err != nil {
					return nil, err
				}
				if contents == nil {
					logrus.Info("No files were changed, appending empty layer to config.")
					continue
				}
				// Append the layer to the image
				opener := func() (io.ReadCloser, error) {
					return ioutil.NopCloser(bytes.NewReader(contents)), nil
				}
				layer, err := tarball.LayerFromOpener(opener)
				if err != nil {
					return nil, err
				}
				sourceImage, err = mutate.Append(sourceImage,
					mutate.Addendum{
						Layer: layer,
						History: v1.History{
							Author:    constants.Author,
							CreatedBy: dockerCommand.CreatedBy(),
						},
					},
				)
				if err != nil {
					return nil, err
				}
			}
		}
		return mutate.Config(sourceImage, imageConfig.Config)
	}
	return nil, nil
}
