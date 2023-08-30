package types

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/joyrex2001/kubedock/internal/util/tar"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog"
)

// Container describes the details of a container.
type Container struct {
	ID             string
	ShortID        string
	Name           string
	Image          string
	Labels         map[string]string
	Entrypoint     []string
	Cmd            []string
	Env            []string
	Binds          []string
	PreArchives    []PreArchive
	HostIP         string
	ExposedPorts   map[string]interface{}
	ImagePorts     map[string]interface{}
	HostPorts      map[int]int
	MappedPorts    map[int]int
	Networks       map[string]interface{}
	NetworkAliases []string
	StopChannels   []chan struct{}
	AttachChannels []chan struct{}
	Running        bool
	Completed      bool
	Failed         bool
	Stopped        bool
	Killed         bool
	Created        time.Time
	Finished       time.Time
}

// PreArchive contains the path and contents of archives (tar) that need to be
// copied over to the container before it has been started.
type PreArchive struct {
	Path    string
	Archive []byte
}

const (
	// LabelRequestCPU is the label to be used to specify cpu request/limits
	LabelRequestCPU = "com.joyrex2001.kubedock.request-cpu"
	// LabelRequestMemory is the label to use to specify memory request/limits
	LabelRequestMemory = "com.joyrex2001.kubedock.request-memory"
	// LabelPullPolicy is the label to be used to configure the pull policy
	LabelPullPolicy = "com.joyrex2001.kubedock.pull-policy"
	// LabelServiceAccount is the label to be used to enforce a service account
	// other than 'default' for the created pods.
	LabelServiceAccount = "com.joyrex2001.kubedock.service-account"
	// LabelNamePrefix is the label to be used to enforce a prefix for the names used
	// for the container deployments.
	LabelNamePrefix = "com.joyrex2001.kubedock.name-prefix"
	// LabelRunasUser is the label to be used to enforce a specific user (uid) that
	// runs inside the container can also be enforced w
	LabelRunasUser = "com.joyrex2001.kubedock.runas-user"
)

// GetEnvVar will return the environment variables of the container
// as k8s EnvVars.
func (co *Container) GetEnvVar() []corev1.EnvVar {
	env := []corev1.EnvVar{}
	for _, e := range co.Env {
		key, value, found := strings.Cut(e, "=")
		if !found {
			klog.Errorf("could not parse env %s", e)
			continue
		}
		env = append(env, corev1.EnvVar{Name: key, Value: value})
	}
	return env
}

// GetImagePullPolicy will return the image pull policy that should be applied
// for this container.
func (co *Container) GetImagePullPolicy() (corev1.PullPolicy, error) {
	ps := map[string]corev1.PullPolicy{
		"default":      corev1.PullIfNotPresent,
		"notpresent":   corev1.PullIfNotPresent,
		"ifnotpresent": corev1.PullIfNotPresent,
		"always":       corev1.PullAlways,
		"allways":      corev1.PullAlways,
		"never":        corev1.PullNever,
	}
	p := co.Labels[LabelPullPolicy]
	if p != "" {
		if c, ok := ps[strings.ToLower(p)]; ok {
			return c, nil
		}
		return ps["default"], fmt.Errorf("invalid pull policy: %s", p)
	}
	return ps["default"], nil
}

// GetResourceRequirements will return a k8s request/limits configuration
// based on the LabelRequestCPU and LabelRequestMemory labels set on the
// container.
func (co *Container) GetResourceRequirements() (corev1.ResourceRequirements, error) {
	req := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{},
		Limits:   corev1.ResourceList{},
	}
	for typ, labl := range map[string]string{"cpu": LabelRequestCPU, "memory": LabelRequestMemory} {
		rls, ok := co.Labels[labl]
		if !ok {
			continue
		}

		var r, l string
		rl := strings.Split(strings.ReplaceAll(rls, " ", ""), ",")
		if len(rl) == 0 || len(rl) > 2 {
			return req, fmt.Errorf("invalid resource requirement: %s", rls)
		}
		r = rl[0]
		if len(rl) == 2 {
			l = rl[1]
		}
		if r == "" && l != "" {
			r = l
		}

		rq, err := resource.ParseQuantity(r)
		if err != nil {
			return req, err
		}
		req.Requests[corev1.ResourceName(typ)] = rq

		if l != "" {
			lt, err := resource.ParseQuantity(l)
			if err != nil {
				return req, err
			}
			req.Limits[corev1.ResourceName(typ)] = lt
		}
	}
	return req, nil
}

// GetServiceAccountName will return the service account to be used for containers
// that are deployed.
func (co *Container) GetServiceAccountName(current string) string {
	if current == "" {
		current = "default"
	}
	if sa, ok := co.Labels[LabelServiceAccount]; ok {
		return sa
	}
	return current
}

// GetPodName will return a human friendly name that can be used for the
// the container deployments.
func (co *Container) GetPodName() string {
	name := co.Name
	if prefix, ok := co.Labels[LabelNamePrefix]; ok {
		name = prefix + "-" + co.Name
	} else {
		name = "kubedock-" + co.Name
	}
	name = strings.ReplaceAll(name, "_", "-")
	re := regexp.MustCompile("[^A-Za-z0-9-]")
	name = re.ReplaceAllString(name, "")
	if len(name) > 32 {
		name = name[:32]
	}
	name = name + "-" + co.ShortID
	name = strings.ReplaceAll(name, "--", "-")
	re = regexp.MustCompile("^[^A-Za-z0-9]+")
	name = re.ReplaceAllString(name, "")
	name = strings.ToLower(name)
	return name
}

// GetPodSecurityContext will create a security context for the Pod that implements
// the relenvant features of the Docker API. Right now this only covers the ability
// to specify the numeric user a container should run as.
func (co *Container) GetPodSecurityContext(context *corev1.PodSecurityContext) (*corev1.PodSecurityContext, error) {
	user, ok := co.Labels[LabelRunasUser]
	if !ok || user == "" {
		if context == nil || context.RunAsUser == nil {
			klog.Warningf("user not set, will run as user defined in image")
		}
		return context, nil
	}

	if context == nil {
		context = &corev1.PodSecurityContext{}
	}

	parsed, err := strconv.ParseInt(user, 10, 64)
	if err != nil {
		return context, fmt.Errorf("failed to parse %s to Int64", user)
	}

	context.RunAsUser = &parsed

	return context, nil
}

// MapPort will map a pod port to a local port.
func (co *Container) MapPort(pod, local int) {
	if co.MappedPorts == nil {
		co.MappedPorts = map[int]int{}
	}
	co.MappedPorts[pod] = local
}

// AddHostPort will add a predefined port mapping.
func (co *Container) AddHostPort(src string, dst string) error {
	var err error
	var sp, dp int

	dp, err = co.getTCPPort(dst)
	if err != nil {
		return err
	}

	if src != "" && src != "0" {
		sp, err = strconv.Atoi(src)
		if err != nil {
			return fmt.Errorf("could not parse exposed port %s: %w", dst, err)
		}
	} else {
		sp = -dp
	}

	if co.HostPorts == nil {
		co.HostPorts = map[int]int{}
	}
	co.HostPorts[sp] = dp

	return nil
}

// GetContainerTCPPorts will return a list of all ports that are
// exposed by this container.
func (co *Container) GetContainerTCPPorts() []int {
	return co.getTCPPorts(co.ExposedPorts)
}

// GetImageTCPPorts will return a list of all ports that are
// exposed by the image.
func (co *Container) GetImageTCPPorts() []int {
	return co.getTCPPorts(co.ImagePorts)
}

// GetServicePorts will return a list of ports and their mapping as they
// should be applied on a k8s service.
func (co *Container) GetServicePorts() map[int]int {
	ports := map[int]int{}
	for _, pp := range co.GetImageTCPPorts() {
		ports[pp] = pp
	}
	for _, pp := range co.GetContainerTCPPorts() {
		ports[pp] = pp
	}
	add := func(prts map[int]int) {
		for src, dst := range prts {
			if src < 0 {
				src = dst
			}
			ports[src] = dst
		}
	}
	add(co.HostPorts)
	add(co.MappedPorts)
	return ports
}

// getTCPPorts will return a list of all tcp ports in given map.
func (co *Container) getTCPPorts(ports map[string]interface{}) []int {
	res := []int{}
	if ports == nil {
		return res
	}
	for p := range ports {
		pp, err := co.getTCPPort(p)
		if err != nil {
			klog.Errorf("could not parse exposed port %s", p)
			continue
		}
		res = append(res, pp)
	}
	return res
}

// getTCPPort will convert a "9000/tcp" string to the port. If "/tcp" is
// missing, it will add it as a default.
func (co *Container) getTCPPort(p string) (int, error) {
	f := strings.Split(p, "/")
	if len(f) == 0 || len(f) > 2 {
		return 0, fmt.Errorf("could not parse exposed port %s", p)
	}
	pp, err := strconv.Atoi(f[0])
	if err != nil {
		return 0, fmt.Errorf("could not parse exposed port %s: %w", p, err)
	}
	if len(f) == 2 && f[1] != "tcp" {
		return 0, fmt.Errorf("unsupported protocol %s for port: %d - only tcp is supported", f[1], pp)
	}
	return pp, nil
}

// GetVolumes will return a map of volumes that should be mounted on the
// target container. The key is the target location, and the value is the
// local location.
func (co *Container) GetVolumes() map[string]string {
	mounts := map[string]string{}
	for _, bind := range co.Binds {
		f := strings.Split(bind, ":")
		mounts[f[1]] = f[0]
	}
	return mounts
}

// GetVolumeFolders will return a map of volumes that are pointing to a
// folder and should be mounted on the target container. The key
// is the target location, and the value is the local location.
func (co *Container) GetVolumeFolders() map[string]string {
	mounts := map[string]string{}
	for dst, src := range co.GetVolumes() {
		if info, err := os.Stat(src); err == nil && info.IsDir() {
			mounts[dst] = src
		}
	}
	return mounts
}

// GetVolumeFiles will return a map of volumes that are pointing to a
// single file and should be mounted on the target container. The key
// is the target location, and the value is the local location.
func (co *Container) GetVolumeFiles() map[string]string {
	mounts := map[string]string{}
	for dst, src := range co.GetVolumes() {
		if info, err := os.Stat(src); err == nil && !info.IsDir() {
			mounts[dst] = src
		}
	}
	return mounts
}

// GetPreArchiveFiles will return all single files from the pre-archives as
// a map with the filename as key, and the actual file contents as value.
func (co *Container) GetPreArchiveFiles() map[string][]byte {
	files := map[string][]byte{}
	for _, pa := range co.PreArchives {
		fls, err := tar.GetTargetFileNames(pa.Path, bytes.NewReader(pa.Archive))
		if err != nil {
			klog.Errorf("error determining pre archive filenames: %s", err)
			continue
		}
		if len(fls) != 1 {
			continue
		}
		var dat bytes.Buffer
		if err := tar.UnpackFile(pa.Path, fls[0], bytes.NewReader(pa.Archive), io.Writer(&dat)); err != nil {
			klog.Errorf("error extracting %s from archive: %s", fls[0], err)
			continue
		}
		files[fls[0]] = dat.Bytes()
	}
	return files
}

// HasVolumes will return true if the container has volumes configured.
func (co *Container) HasVolumes() bool {
	return len(co.Binds) > 0
}

// AddStopChannel will add channels that should be notified when
// SignalStop is called.
func (co *Container) AddStopChannel(stop chan struct{}) {
	if co.StopChannels == nil {
		co.StopChannels = []chan struct{}{}
	}
	co.StopChannels = append(co.StopChannels, stop)
}

// SignalStop will signal all stop channels.
func (co *Container) SignalStop() {
	for _, stop := range co.StopChannels {
		stop <- struct{}{}
		close(stop)
	}
	co.StopChannels = []chan struct{}{}
}

// AddAttachChannel will add channels that should be notified when
// SignalDetach is called.
func (co *Container) AddAttachChannel(stop chan struct{}) {
	if co.AttachChannels == nil {
		co.AttachChannels = []chan struct{}{}
	}
	co.AttachChannels = append(co.AttachChannels, stop)
}

// SignalDetach will signal all stop channels.
func (co *Container) SignalDetach() {
	for _, stop := range co.AttachChannels {
		stop <- struct{}{}
		close(stop)
	}
	co.AttachChannels = []chan struct{}{}
}

// ConnectNetwork will attach a network to the container.
func (co *Container) ConnectNetwork(id string) {
	if co.Networks == nil {
		co.Networks = map[string]interface{}{}
	}
	co.Networks[id] = nil
}

// DisconnectNetwork will detach a network from the container.
func (co *Container) DisconnectNetwork(id string) error {
	if _, ok := co.Networks[id]; !ok {
		return fmt.Errorf("container is not connected to network %s", id)
	}
	delete(co.Networks, id)
	return nil
}

// Match will match given type with given key value pair.
func (co *Container) Match(typ string, key string, val string) bool {
	if typ == "name" {
		return co.Name == key
	}
	if typ != "label" {
		return true
	}
	v, ok := co.Labels[key]
	if !ok {
		return false
	}
	return v == val
}

// StateString returns a string that describes the state.
func (co *Container) StateString() string {
	if co.Running {
		return "Up"
	}
	if co.Stopped || co.Killed {
		return "Dead"
	}
	if co.Failed {
		return "Dead"
	}
	if co.Completed {
		return "Exited"
	}
	return "Created"
}

// StatusString returns a string that describes the status.
func (co *Container) StatusString() string {
	if co.Running {
		return "healthy"
	}
	return "unhealthy"
}
