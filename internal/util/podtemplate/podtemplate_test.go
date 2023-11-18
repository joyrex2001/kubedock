package podtemplate

import (
	"testing"
)

func TestPodFromFile(t *testing.T) {
	pod, err := PodFromFile("test/test_pod.yaml")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if pod == nil {
		t.Error("error unmarshalling pod")
	}
	if pod != nil && pod.Spec.ServiceAccountName != "kubedock" {
		t.Error("invalid serviceAccountName")
	}

	container := ContainerFromPod(pod)
	if container.Resources.Requests != nil {
		t.Error("unexpected resources in container template")
	}

	pod, err = PodFromFile("test/test_container.yaml")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	container = ContainerFromPod(pod)
	if container.Resources.Requests == nil {
		t.Error("expected resources in container template")
	} else {
		reqmem := container.Resources.Requests.Memory().String()
		if reqmem != "64Mi" {
			t.Errorf("unexpected value for request.memory %s, expected 64Mi", reqmem)
		}
	}

	pod, err = PodFromFile("test/notfound.yaml")
	if pod != nil {
		t.Error("unexpected pod object")
	}
	if err == nil {
		t.Error("expected an error when file is not available")
	}

	pod, err = PodFromFile("test/test_invalid_kind.yaml")
	if pod != nil {
		t.Error("unexpected pod object")
	}
	if err == nil {
		t.Error("expected an error when kind is not a pod")
	}

	pod, err = PodFromFile("test/test_invalid.yaml")
	if pod != nil {
		t.Error("unexpected pod object")
	}
	if err == nil {
		t.Error("expected an error when file is invalid yaml")
	}
}
