package reverseproxy

import (
	"bufio"
	"fmt"
	"net"
	"testing"
	"time"
)

func helloServer(host string, port int, stop chan struct{}) error {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
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
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
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

func TestProxyTimeOut(t *testing.T) {
	stopP := make(chan struct{}, 1)
	req := Request{
		LocalPort:  30590,
		RemoteIP:   "1.0.0.1",
		RemotePort: 30393,
		StopCh:     stopP,
		TimeOut:    2 * time.Second,
	}

	if err := Proxy(req); err != nil {
		t.Errorf("unexpected error starting proxy: %s", err)
	}

	done := false
	go func() {
		_, err := callServer("127.0.0.1", 30590)
		if err == nil {
			t.Errorf("expected error calling helloServer via proxy but didn't get any")
		}
		done = true
	}()

	select {
	case <-time.After(1 * time.Second):
	}

	if done {
		t.Errorf("expected timeout calling helloServer, but request succeeded")
	}

	select {
	case <-time.After(2 * time.Second):
	}

	if !done {
		t.Errorf("expected timeout to be met in helloServer, but was not done")
	}

	stopP <- struct{}{}
}
