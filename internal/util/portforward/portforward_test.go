package portforward

import (
	"net/url"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func TestGetURLScheme(t *testing.T) {
	tests := []struct {
		in  Request
		out *url.URL
	}{
		{
			in: Request{
				Pod:        corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "abc", Name: "def"}},
				RestConfig: &rest.Config{Host: "https://tst-cluster"},
			},
			out: &url.URL{
				Host:   "tst-cluster",
				Scheme: "https",
				Path:   "/api/v1/namespaces/abc/pods/def/portforward",
			},
		},
	}

	for i, tst := range tests {
		out := getURLScheme(tst.in)
		if !reflect.DeepEqual(out, tst.out) {
			t.Errorf("failed test %d - expected %v, but got %v", i, tst.out, out)
		}
	}
}
