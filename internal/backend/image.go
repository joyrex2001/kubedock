package backend

import (
	"github.com/joyrex2001/kubedock/internal/util/image"
)

// GetImageExposedPorts will inspect the image in the registry and return the
// configured exposed ports from the image, or will return an error if failed.
func (in *instance) GetImageExposedPorts(img string) (map[string]struct{}, error) {
	cfg, err := image.InspectConfig("docker://" + img)
	if err != nil {
		return nil, err
	}
	return cfg.Config.ExposedPorts, nil
}
