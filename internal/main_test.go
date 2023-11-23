package internal

import (
	"strings"
	"testing"

	"github.com/joyrex2001/kubedock/internal/util/myip"
	"github.com/spf13/viper"
)

func TestGetKubedockURL(t *testing.T) {
	tests := []struct {
		listen string
		tls    bool
		res    string
		suc    bool
	}{
		{":1234", false, "http://{{IP}}:1234", true},
		{":1234", true, "https://{{IP}}:1234", true},
		{"1234", false, "", false},
	}

	ip, _ := myip.Get()
	for i, tst := range tests {
		viper.Set("server.listen-addr", tst.listen)
		viper.Set("server.tls-enable", tst.tls)
		res, err := getKubedockURL()
		if tst.suc && err != nil {
			t.Errorf("failed test %d - unexpected error %s", i, err)
		}
		if !tst.suc && err == nil {
			t.Errorf("failed test %d - expected error, but succeeded instead", i)
		}
		tst.res = strings.ReplaceAll(tst.res, "{{IP}}", ip)
		if err == nil && res != tst.res {
			t.Errorf("failed test %d - expected %s, but got %s", i, tst.res, res)
		}
	}
}
