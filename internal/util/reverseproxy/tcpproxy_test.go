package reverseproxy

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

func helloServer(host string, port int, stop chan struct{}) error {
	l, err := net.Listen("tcp", net.JoinHostPort(host, fmt.Sprintf("%d", port)))
	if err != nil {
		return err
	}
	defer l.Close()

	done := false
	go func() {
		<-stop
		done = true
		l.Close()
	}()

	for {
		if done {
			return nil
		}
		conn, err := l.Accept()
		if err != nil {
			if done {
				return nil
			}
			return err
		}
		conn.Write([]byte("Hello!\n"))
		conn.Close()
	}
}

func callServer(host string, port int) (string, error) {
	conn, err := net.Dial("tcp", net.JoinHostPort(host, fmt.Sprintf("%d", port)))
	if err != nil {
		return "", err
	}
	defer conn.Close()
	buf := bufio.NewReader(conn)
	return buf.ReadString('\n')
}

func TestProxyNormal(t *testing.T) {
	stopP := make(chan struct{}, 1)
	req := Request{
		LocalPort:  30390,
		RemoteIP:   "127.0.0.1",
		RemotePort: 30391,
		StopCh:     stopP,
		MaxRetry:   2,
	}

	if err := Proxy(req); err != nil {
		t.Errorf("unexpected error starting proxy: %s", err)
	}

	stopS := make(chan struct{}, 1)
	go func() {
		if err := helloServer("127.0.0.1", 30391, stopS); err != nil {
			t.Errorf("unexpected error running helloServer: %s", err)
		}
	}()

	<-time.After(time.Second)

	res, err := callServer("127.0.0.1", 30390)
	if err != nil {
		t.Errorf("unexpected error calling helloServer via proxy: %s", err)
	}

	if res != "Hello!\n" {
		t.Errorf("unexpected answer calling helloServer via proxy: %s", res)
	}

	stopP <- struct{}{}
	stopS <- struct{}{}
}

func TestProxyRefused(t *testing.T) {
	stopP := make(chan struct{}, 1)
	req := Request{
		LocalPort:  30490,
		RemoteIP:   "127.0.0.1",
		RemotePort: 30392,
		StopCh:     stopP,
		MaxRetry:   1,
	}

	if err := Proxy(req); err != nil {
		t.Errorf("unexpected error starting proxy: %s", err)
	}

	_, err := callServer("127.0.0.1", 30490)
	if err == nil {
		t.Errorf("expected error calling helloServer via proxy but didn't get any")
	}

	stopP <- struct{}{}
}

func TestProxyNotReady(t *testing.T) {
	stopP := make(chan struct{}, 1)
	req := Request{
		LocalPort:  30590,
		RemoteIP:   "127.0.0.1",
		RemotePort: 30393,
		StopCh:     stopP,
		MaxRetry:   2,
	}

	if err := Proxy(req); err != nil {
		t.Errorf("unexpected error starting proxy: %s", err)
	}

	<-time.After(time.Second)

	res, err := callServer("127.0.0.1", 30590)
	if err != io.EOF {
		t.Errorf("unexpected error calling helloServer via proxy: %s", err)
	}

	if res != "" {
		t.Errorf("unexpected answer calling helloServer via proxy: %s", res)
	}

	stopS := make(chan struct{}, 1)
	go func() {
		if err := helloServer("127.0.0.1", 30393, stopS); err != nil {
			t.Errorf("unexpected error running helloServer: %s", err)
		}
	}()

	<-time.After(time.Second)

	res, err = callServer("127.0.0.1", 30590)
	if err != nil {
		t.Errorf("unexpected error calling helloServer via proxy: %s", err)
	}

	if res != "Hello!\n" {
		t.Errorf("unexpected answer calling helloServer via proxy: %s", res)
	}

	stopP <- struct{}{}
	stopS <- struct{}{}
}

func TestProxyIPv6Normal(t *testing.T) {
	stopP := make(chan struct{}, 1)
	req := Request{
		LocalPort:  30690,
		RemoteIP:   "::1",
		RemotePort: 30691,
		StopCh:     stopP,
		MaxRetry:   2,
	}

	if err := Proxy(req); err != nil {
		t.Errorf("unexpected error starting proxy: %s", err)
	}

	stopS := make(chan struct{}, 1)
	go func() {
		if err := helloServer("::1", 30691, stopS); err != nil {
			t.Errorf("unexpected error running helloServer: %s", err)
		}
	}()

	<-time.After(time.Second)

	res, err := callServer("::1", 30690)
	if err != nil {
		t.Errorf("unexpected error calling helloServer via proxy: %s", err)
	}

	if res != "Hello!\n" {
		t.Errorf("unexpected answer calling helloServer via proxy: %s", res)
	}

	stopP <- struct{}{}
	stopS <- struct{}{}
}
