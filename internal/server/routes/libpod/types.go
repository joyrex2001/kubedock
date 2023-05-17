package libpod

// ContainerCreateRequest represents the json structure that
// is used for the /libpod/container/create post endpoint.
type ContainerCreateRequest struct {
	Name         string                      `json:"name"`
	Image        string                      `json:"image"`
	Labels       map[string]string           `json:"Labels"`
	Entrypoint   []string                    `json:"Entrypoint"`
	Command      []string                    `json:"Command"`
	Env          []string                    `json:"Env"`
	User         string                      `json:"User"`
	PortMappings []PortMapping               `json:"portmappings"`
	Network      map[string]NetworksProperty `json:"Networks"`
	Mounts       []Mount                     `json:"mounts"`
}

// PortMapping describes how to map a port into the container.
type PortMapping struct {
	ContainerPort int    `json:"container_port"`
	HostIP        string `json:"host_ip"`
	HostPort      int    `json:"host_port"`
	Protocol      string `json:"protocol"`
	Range         int    `json:"range"`
}

// NetworksProperty describes the container networks.
type NetworksProperty struct {
	Aliases []string `json:"aliases"`
}

// Mount describes how volumes should be mounted.
type Mount struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Type        string `json:"type"`
}
