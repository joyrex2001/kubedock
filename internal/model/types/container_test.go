package types

import (
	"reflect"
	"sort"
	"strconv"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNew(t *testing.T) {
	c := &Container{}
	if c.Finished.Format("2006-01-02T15:04:05Z") != "0001-01-01T00:00:00Z" {
		t.Errorf("failed initial finished timestamp")
	}
}

func TestGetEnvVar(t *testing.T) {
	tests := []struct {
		in  *Container
		out []corev1.EnvVar
	}{
		{
			in: &Container{Env: []string{
				"rc738",
				"rc743=Penguin Adventure",
				"rc768=Space Manbow",
			}},
			out: []corev1.EnvVar{
				{Name: "rc743", Value: "Penguin Adventure"},
				{Name: "rc768", Value: "Space Manbow"},
			},
		},
		{
			in: &Container{Env: []string{
				"NO_EQUALS=123",
				"EQUALS_AT_END=456=",
				"EQUALS_IN_MIDDLE=abc123=aabbcc",
				"MULTIPLE_EQUALS=abc123==aa=bb=cc==",
			}},
			out: []corev1.EnvVar{
				{Name: "NO_EQUALS", Value: "123"},
				{Name: "EQUALS_AT_END", Value: "456="},
				{Name: "EQUALS_IN_MIDDLE", Value: "abc123=aabbcc"},
				{Name: "MULTIPLE_EQUALS", Value: "abc123==aa=bb=cc=="},
			},
		},
	}
	for i, tst := range tests {
		res := tst.in.GetEnvVar()
		if !reflect.DeepEqual(res, tst.out) {
			t.Errorf("failed test %d - expected %v, but got %v", i, tst.out, res)
		}
	}
}

func TestGetResourceRequirements(t *testing.T) {
	mem64, _ := resource.ParseQuantity("64Mi")

	tests := []struct {
		in        *Container
		resources corev1.ResourceRequirements
		reqlim    map[string]string
		err       bool
	}{
		{ // 0
			in:     &Container{Labels: map[string]string{}},
			reqlim: map[string]string{},
			err:    false,
		},
		{ // 1
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.request-cpu": "500m",
			}},
			reqlim: map[string]string{"reqcpu": "500m"},
			err:    false,
		},
		{ // 2
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.request-cpu": "500m,2000m",
			}},
			reqlim: map[string]string{"reqcpu": "500m", "limcpu": "2"},
			err:    false,
		},
		{ // 3
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.request-cpu": ",2000m",
			}},
			reqlim: map[string]string{"reqcpu": "2", "limcpu": "2"},
			err:    false,
		},
		{ // 4
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.request-cpu": "joyrex",
			}},
			reqlim: map[string]string{},
			err:    true,
		},
		{ // 5
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.request-memory": "500Mi",
			}},
			reqlim: map[string]string{"reqmem": "500Mi"},
			err:    false,
		},
		{ // 6
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.request-memory": "500Mi,2000Mi",
			}},
			reqlim: map[string]string{"reqmem": "500Mi", "limmem": "2000Mi"},
			err:    false,
		},
		{ // 7
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.request-memory": ",2000Mi",
			}},
			reqlim: map[string]string{"reqmem": "2000Mi", "limmem": "2000Mi"},
			err:    false,
		},
		{ // 8
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.request-memory": "joyrex",
			}},
			reqlim: map[string]string{},
			err:    true,
		},
		{ // 9
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.request-memory": "500Mi,2000Mi,2500Mi",
			}},
			reqlim: map[string]string{},
			err:    true,
		},
		{ // 10
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.request-memory": "500Mi,joyrex",
			}},
			reqlim: map[string]string{},
			err:    true,
		},
		{ // 11
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.request-memory": " , 2000Mi",
			}},
			reqlim: map[string]string{"reqmem": "2000Mi", "limmem": "2000Mi"},
			err:    false,
		},
		{ // 12
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.request-cpu": ",1000000000n",
			}},
			reqlim: map[string]string{"reqcpu": "1", "limcpu": "1"},
			err:    false,
		},
		{ // 13
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.request-memory": "209715200",
			}},
			reqlim: map[string]string{"reqmem": "209715200"},
			err:    false,
		},
		{ // 14
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.request-memory": "209715200",
			}},
			reqlim: map[string]string{"reqmem": "209715200", "limmem": mem64.String()},
			resources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					"memory": mem64,
				},
			},
			err: false,
		},
		{ // 15
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.request-memory": "209715200",
			}},
			reqlim: map[string]string{"reqmem": "209715200"},
			resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"memory": mem64,
				},
			},
			err: false,
		},
	}
	for i, tst := range tests {
		res, err := tst.in.GetResourceRequirements(tst.resources)
		if err != nil && !tst.err {
			t.Errorf("failed test %d - unexpected error: %s", i, err)
		}
		if err == nil && tst.err {
			t.Errorf("failed test %d - expected error, but succeeded without error", i)
		}

		reqlim := map[string]string{}
		if v, ok := res.Requests["cpu"]; ok {
			reqlim["reqcpu"] = v.String()
		}
		if v, ok := res.Requests["memory"]; ok {
			reqlim["reqmem"] = v.String()
		}
		if v, ok := res.Limits["cpu"]; ok {
			reqlim["limcpu"] = v.String()
		}
		if v, ok := res.Limits["memory"]; ok {
			reqlim["limmem"] = v.String()
		}

		if err == nil && !reflect.DeepEqual(reqlim, tst.reqlim) {
			t.Errorf("failed test %d - expected %v, but got %#v", i, tst.reqlim, reqlim)
		}
	}
}

func TestGetNodeSelector(t *testing.T) {
	tests := []struct {
		in         *Container
		inNodeSel  map[string]string
		outNodeSel map[string]string
		err        bool
	}{
		{ // 0
			in:         &Container{Labels: map[string]string{}},
			inNodeSel:  map[string]string{},
			outNodeSel: map[string]string{},
			err:        false,
		},
		{ // 1
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.node-selector": "a=b",
			}},
			inNodeSel:  map[string]string{},
			outNodeSel: map[string]string{"a": "b"},
			err:        false,
		},
		{ // 2
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.node-selector": "a=b,c=d",
			}},
			inNodeSel:  map[string]string{},
			outNodeSel: map[string]string{"a": "b", "c": "d"},
			err:        false,
		},
		{ // 3
			in:         &Container{Labels: map[string]string{}},
			inNodeSel:  map[string]string{"z": "y"},
			outNodeSel: map[string]string{"z": "y"},
			err:        false,
		},
		{ // 4
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.node-selector": "a=b",
			}},
			inNodeSel:  map[string]string{"z": "y"},
			outNodeSel: map[string]string{"a": "b", "z": "y"},
			err:        false,
		},
		{ // 5
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.node-selector": "format-error",
			}},
			inNodeSel:  map[string]string{},
			outNodeSel: map[string]string{},
			err:        true,
		},
		{ // 6
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.node-selector": "a=b",
			}},
			inNodeSel:  nil,
			outNodeSel: map[string]string{"a": "b"},
			err:        false,
		},
	}
	for i, tst := range tests {
		res, err := tst.in.GetNodeSelector(tst.inNodeSel)
		if err != nil && !tst.err {
			t.Errorf("failed test %d - unexpected error: %s", i, err)
		}
		if err == nil && tst.err {
			t.Errorf("failed test %d - expected error, but succeeded without error", i)
		}
		if !reflect.DeepEqual(res, tst.outNodeSel) {
			t.Errorf("failed test %d - expected %s, but got %s", i, tst.outNodeSel, res)
		}
	}
}

func TestGetImagePullPolicy(t *testing.T) {
	tests := []struct {
		in     *Container
		policy corev1.PullPolicy
		err    bool
	}{
		{ // 0
			in:     &Container{Labels: map[string]string{}},
			policy: corev1.PullIfNotPresent,
			err:    false,
		},
		{ // 1
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.pull-policy": "always",
			}},
			policy: corev1.PullAlways,
			err:    false,
		},
		{ // 2
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.pull-policy": "something",
			}},
			policy: corev1.PullIfNotPresent,
			err:    true,
		},
	}
	for i, tst := range tests {
		res, err := tst.in.GetImagePullPolicy()
		if err != nil && !tst.err {
			t.Errorf("failed test %d - unexpected error: %s", i, err)
		}
		if err == nil && tst.err {
			t.Errorf("failed test %d - expected error, but succeeded without error", i)
		}
		if res != tst.policy {
			t.Errorf("failed test %d - expected %s, but got %s", i, tst.policy, res)
		}
	}
}

func TestGetServiceAccountName(t *testing.T) {
	tests := []struct {
		in    *Container
		insa  string
		outsa string
	}{
		{ // 0
			in:    &Container{Labels: map[string]string{}},
			outsa: "default",
		},
		{ // 1
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.service-account": "default",
			}},
			outsa: "default",
		},
		{ // 2
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.service-account": "jtkirk",
			}},
			outsa: "jtkirk",
		},
		{ // 3
			in:    &Container{Labels: map[string]string{}},
			insa:  "jtkirk",
			outsa: "jtkirk",
		},
	}
	for i, tst := range tests {
		sa := tst.in.GetServiceAccountName(tst.insa)
		if sa != tst.outsa {
			t.Errorf("failed test %d - expected %s, but got %s", i, tst.outsa, sa)
		}
	}
}

func TestGetPodName(t *testing.T) {
	tests := []struct {
		in   *Container
		name string
	}{
		{ // 0
			in:   &Container{ShortID: "1234", Name: "", Labels: map[string]string{}},
			name: "kubedock-1234",
		},
		{ // 1
			in: &Container{ShortID: "1234", Name: "", Labels: map[string]string{
				"com.joyrex2001.kubedock.name-prefix": "space",
			}},
			name: "space-1234",
		},
		{ // 2
			in: &Container{ShortID: "1234", Name: "", Labels: map[string]string{
				"com.joyrex2001.kubedock.name-prefix": "s.pace",
			}},
			name: "space-1234",
		},
		{ // 3
			in: &Container{ShortID: "1234", Name: "exploration", Labels: map[string]string{
				"com.joyrex2001.kubedock.name-prefix": "s.pace",
			}},
			name: "space-exploration-1234",
		},
		{ // 4
			in: &Container{ShortID: "1234", Name: "exploration/909", Labels: map[string]string{
				"com.joyrex2001.kubedock.name-prefix": "s.pace",
			}},
			name: "space-exploration909-1234",
		},
		{ // 5
			in: &Container{ShortID: "1234", Name: "exploration_909", Labels: map[string]string{
				"com.joyrex2001.kubedock.name-prefix": "s_pace",
			}},
			name: "s-pace-exploration-909-1234",
		},
		{ // 6
			in: &Container{ShortID: "1234", Name: "MyContainer", Labels: map[string]string{
				"com.joyrex2001.kubedock.name-prefix": "space",
			}},
			name: "space-mycontainer-1234",
		},
	}
	for i, tst := range tests {
		name := tst.in.GetPodName()
		if name != tst.name {
			t.Errorf("failed test %d - expected %s, but got %s", i, tst.name, name)
		}
	}
}

func TestGetRunasUser(t *testing.T) {
	tests := []struct {
		in    *Container
		outsc corev1.PodSecurityContext
		insc  *corev1.PodSecurityContext
		err   bool
	}{
		{ // 0
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.runas-user": "",
			}},
			outsc: corev1.PodSecurityContext{},
			err:   false,
		},
		{ // 1
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.runas-user": "1000",
			}},
			outsc: corev1.PodSecurityContext{RunAsUser: makeIntPointer(1000)},
			err:   false,
		},
		{ // 2
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.runas-user": "0",
			}},
			outsc: corev1.PodSecurityContext{RunAsUser: makeIntPointer(0)},
			err:   false,
		},
		{ // 3
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.runas-user": "9999999999999999999999999999999",
			}},
			outsc: corev1.PodSecurityContext{},
			err:   true,
		},
		{ // 4
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.runas-user": "abc",
			}},
			outsc: corev1.PodSecurityContext{},
			err:   true,
		},
		{ // 5
			in:    &Container{},
			outsc: corev1.PodSecurityContext{},
			err:   false,
		},
		{ // 6
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.runas-user": "",
			}},
			insc:  &corev1.PodSecurityContext{RunAsUser: makeIntPointer(1000)},
			outsc: corev1.PodSecurityContext{RunAsUser: makeIntPointer(1000)},
			err:   false,
		},
		{ // 7
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.runas-user": "1000",
			}},
			insc:  &corev1.PodSecurityContext{RunAsUser: makeIntPointer(500)},
			outsc: corev1.PodSecurityContext{RunAsUser: makeIntPointer(1000)},
			err:   false,
		},
	}
	for i, tst := range tests {
		res, err := tst.in.GetPodSecurityContext(tst.insc)
		if err != nil && !tst.err {
			t.Errorf("failed test %d - unexpected error: %s", i, err)
		}
		if err == nil && tst.err {
			t.Errorf("failed test %d - expected error, but succeeded without error", i)
		}
		if res == nil && tst.outsc.RunAsUser != nil {
			t.Errorf("failed test %d - expected %d, but got nil", i, *tst.outsc.RunAsUser)
		}
		if res != nil && res.RunAsUser != nil && tst.outsc.RunAsUser == nil {
			t.Errorf("failed test %d - expected nil, but got %d", i, *res.RunAsUser)
		}
		if res != nil && res.RunAsUser != nil && tst.outsc.RunAsUser != nil && *res.RunAsUser != *tst.outsc.RunAsUser {
			t.Errorf("failed test %d - expected %d, but got %d", i, *tst.outsc.RunAsUser, *res.RunAsUser)
		}
	}
}

func TestMapPort(t *testing.T) {
	in := &Container{}
	if in.MappedPorts != nil {
		t.Errorf("mapped ports to be expect nil when no mapping done")
	}
	in.MapPort(808, 1808)
	in.MapPort(909, 1909)
	if !reflect.DeepEqual(in.MappedPorts, map[int]int{808: 1808, 909: 1909}) {
		t.Errorf("port mapping failed")
	}
}

func TestGetTCPPorts(t *testing.T) {
	tests := []struct {
		in  *Container
		out []int
	}{
		{
			in: &Container{ExposedPorts: map[string]interface{}{
				"sh101":     0,
				"303/tcp":   0,
				"606/udp":   0,
				"tr808/tcp": 0,
				"909/tcp":   0,
			}, ImagePorts: map[string]interface{}{
				"sh101":     0,
				"303/tcp":   0,
				"606/udp":   0,
				"tr808/tcp": 0,
				"909/tcp":   0,
			}},
			out: []int{303, 909},
		},
		{
			in:  &Container{},
			out: []int{},
		},
	}
	for i, tst := range tests {
		res := tst.in.GetContainerTCPPorts()
		sort.Ints(res)
		if !reflect.DeepEqual(res, tst.out) {
			t.Errorf("failed test container %d - expected %v, but got %v", i, tst.out, res)
		}
		res = tst.in.GetImageTCPPorts()
		sort.Ints(res)
		if !reflect.DeepEqual(res, tst.out) {
			t.Errorf("failed test image %d - expected %v, but got %v", i, tst.out, res)
		}
	}
}

func TestAddHostPort(t *testing.T) {
	tests := []struct {
		src string
		dst string
		out map[int]int
		suc bool
	}{
		{
			src: "303",
			dst: "606/tcp",
			out: map[int]int{303: 606},
			suc: true,
		},
		{
			src: "",
			dst: "606/tcp",
			out: map[int]int{-606: 606},
			suc: true,
		},
		{
			src: "303",
			dst: "606",
			out: map[int]int{303: 606},
			suc: true,
		},
		{
			src: "three-o-three",
			dst: "606/tcp",
			suc: false,
		},
	}
	for i, tst := range tests {
		in := &Container{}
		err := in.AddHostPort(tst.src, tst.dst)
		if err != nil && tst.suc {
			t.Errorf("failed test %d - unexpected error: %s", i, err)
		}
		if err == nil && !tst.suc {
			t.Errorf("failed test %d - expected error, but succeeded instead", i)
		}
		if !reflect.DeepEqual(in.HostPorts, tst.out) {
			t.Errorf("failed test %d - expected %v, but got %v", i, tst.out, in.HostPorts)
		}
	}
}

func TestGetServicePorts(t *testing.T) {
	tests := []struct {
		in  *Container
		out map[int]int
	}{
		{
			in: &Container{ExposedPorts: map[string]interface{}{
				"303/tcp": 0,
				"909/tcp": 0,
			}, ImagePorts: map[string]interface{}{
				"606/tcp": 0,
			}, HostPorts: map[int]int{
				202: 202,
			}},
			out: map[int]int{202: 202, 303: 303, 606: 606, 909: 909},
		},
		{
			in: &Container{ExposedPorts: map[string]interface{}{
				"303/tcp": 0,
				"909/tcp": 0,
			}, ImagePorts: map[string]interface{}{
				"303/tcp": 0,
			}, HostPorts: map[int]int{
				-202: 202,
			}},
			out: map[int]int{202: 202, 303: 303, 909: 909},
		},
		{
			in: &Container{ExposedPorts: map[string]interface{}{
				"303/tcp": 0,
				"909/tcp": 0,
			}, ImagePorts: map[string]interface{}{
				"303/tcp": 0,
			}, HostPorts: map[int]int{
				-202: 202,
			}, MappedPorts: map[int]int{
				606: 808,
			}},
			out: map[int]int{202: 202, 303: 303, 606: 808, 909: 909},
		},
	}
	for i, tst := range tests {
		res := tst.in.GetServicePorts()
		if !reflect.DeepEqual(res, tst.out) {
			t.Errorf("failed test %d - expected %v, but got %v", i, tst.out, res)
		}
	}
}

func TestStop(t *testing.T) {
	tainr := &Container{}
	res := 0
	stop := make(chan struct{}, 1)
	done := make(chan struct{}, 1)
	tainr.AddStopChannel(stop)
	go func(res *int, in chan struct{}) {
		<-in
		*res = 1
		done <- struct{}{}
	}(&res, stop)
	tainr.SignalStop()
	select {
	case <-done:
	case <-time.After(1 * time.Second):
	}
	if res != 1 {
		t.Errorf("failed stop channels")
	}
	if len(tainr.StopChannels) != 0 {
		t.Errorf("expected stop channels to be erased")
	}
}

func TestDetach(t *testing.T) {
	tainr := &Container{}
	res := 0
	stop := make(chan struct{}, 1)
	done := make(chan struct{}, 1)
	tainr.AddAttachChannel(stop)
	go func(res *int, in chan struct{}) {
		<-in
		*res = 1
		done <- struct{}{}
	}(&res, stop)
	tainr.SignalDetach()
	select {
	case <-done:
	case <-time.After(1 * time.Second):
	}
	if res != 1 {
		t.Errorf("failed attach channels")
	}
	if len(tainr.AttachChannels) != 0 {
		t.Errorf("expected attach channels to be erased")
	}
}

func TestVolumes(t *testing.T) {
	tests := []struct {
		in      *Container
		all     map[string]string
		folders map[string]string
		files   map[string]string
		vol     bool
		sock    bool
	}{
		{
			in: &Container{
				Binds: []string{
					"container_test.go:/tmp/container_test.go:ro",
					"../types:/tmp/types:ro",
					"/var/run/docker.sock:/var/run/docker.sock:rw",
				},
				Mounts: []Mount{{
					Source:   "/abc",
					Target:   "def",
					ReadOnly: false,
					Type:     "bind",
				}},
			},
			all: map[string]string{
				"/tmp/container_test.go": "container_test.go",
				"/tmp/types":             "../types",
				"/var/run/docker.sock":   "/var/run/docker.sock",
				"def":                    "/abc",
			},
			files: map[string]string{
				"/tmp/container_test.go": "container_test.go",
			},
			folders: map[string]string{
				"/tmp/types": "../types",
			},
			vol:  true,
			sock: true,
		},
		{
			in:      &Container{Binds: []string{}},
			all:     map[string]string{},
			files:   map[string]string{},
			folders: map[string]string{},
			vol:     false,
			sock:    false,
		},
	}
	for i, tst := range tests {
		res := tst.in.GetVolumes()
		if !reflect.DeepEqual(res, tst.all) {
			t.Errorf("failed test %d all - expected %v, but got %v", i, tst.all, res)
		}
		res = tst.in.GetVolumeFiles()
		if !reflect.DeepEqual(res, tst.files) {
			t.Errorf("failed test %d files - expected %v, but got %v", i, tst.files, res)
		}
		res = tst.in.GetVolumeFolders()
		if !reflect.DeepEqual(res, tst.folders) {
			t.Errorf("failed test %d folders - expected %v, but got %v", i, tst.folders, res)
		}
		if tst.in.HasVolumes() != tst.vol {
			t.Errorf("failed test %d volumes- expected %t, but got %t", i, tst.in.HasVolumes(), tst.vol)
		}
		if tst.in.HasDockerSockBinding() != tst.sock {
			t.Errorf("failed test %d sock - expected %t, but got %t", i, tst.in.HasDockerSockBinding(), tst.sock)
		}
	}
}

func TestConnectNetwork(t *testing.T) {
	var err error
	in := &Container{}
	in.ConnectNetwork("1234")
	if in.Networks == nil {
		t.Errorf("networks to be expect populated when container is connected")
	}
	if _, ok := in.Networks["1234"]; !ok {
		t.Errorf("network 1234 expected to be connected")
	}
	err = in.DisconnectNetwork("1234")
	if err != nil {
		t.Errorf("unexpected error on delete %s", err)
	}
	if _, ok := in.Networks["1234"]; ok {
		t.Errorf("network 1234 expected to be disconnected")
	}
	err = in.DisconnectNetwork("1234")
	if err == nil {
		t.Errorf("expected error on delete non existing network, but got none")
	}
	err = in.DisconnectNetwork("bridge")
	if err == nil {
		t.Errorf("expected error on delete bridge, but got none")
	}
}

func TestMatch(t *testing.T) {
	tests := []struct {
		name   string
		labels map[string]string
		typ    string
		key    string
		val    string
		match  bool
	}{
		{
			labels: map[string]string{},
			typ:    "label",
			key:    "some",
			val:    "thing",
			match:  false,
		},
		{
			labels: map[string]string{"some": "thing"},
			typ:    "label",
			key:    "some",
			val:    "thing",
			match:  true,
		},
		{
			labels: map[string]string{"some": "what"},
			typ:    "label",
			key:    "some",
			val:    "thing",
			match:  false,
		},
		{
			labels: map[string]string{"some": "what"},
			typ:    "magic",
			key:    "some",
			val:    "thing",
			match:  true,
		},
		{
			labels: map[string]string{"some": "what"},
			typ:    "name",
			key:    "something",
			val:    "",
			match:  false,
		},
		{
			name:   "testymctestface",
			labels: map[string]string{"some": "what"},
			typ:    "name",
			key:    "testymctestface",
			val:    "",
			match:  true,
		},
	}
	for i, tst := range tests {
		in := &Container{Labels: tst.labels, Name: tst.name}
		if in.Match(tst.typ, tst.key, tst.val) != tst.match {
			t.Errorf("failed test %d - match %v", i, tst.match)
		}
	}
}

func makeIntPointer(x int64) *int64 {
	return &x
}

func ptrToString(v *int64) string {
	if v == nil {
		return "nil"
	}
	return strconv.FormatInt(*v, 10)
}

func TestGetActiveDeadlineSeconds(t *testing.T) {
	tests := []struct {
		in       *Container
		deadline *int64
		err      bool
	}{
		{ // 0
			in:       &Container{Labels: map[string]string{}},
			deadline: nil,
			err:      false,
		},
		{ // 1
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.active-deadline-seconds": "42",
			}},
			deadline: makeIntPointer(42),
			err:      false,
		},
		{ // 2
			in: &Container{Labels: map[string]string{
				"com.joyrex2001.kubedock.active-deadline-seconds": "foo",
			}},
			deadline: nil,
			err:      true,
		},
	}

	for i, tst := range tests {
		res, err := tst.in.GetActiveDeadlineSeconds()
		if err != nil && !tst.err {
			t.Errorf("failed test %d - unexpected error: %s", i, err)
		}
		if err == nil && tst.err {
			t.Errorf("failed test %d - expected error, but succeeded without error", i)
		}
		if !reflect.DeepEqual(tst.deadline, res) {
			t.Errorf("failed test %d - expected %s, but got %s", i, ptrToString(tst.deadline), ptrToString(res))
		}
	}
}
