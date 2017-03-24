package main

import (
	"flag"
	"fmt"
	"github.com/Sirupsen/logrus"
	"net/http"
	"net/url"
	"path/filepath"

	"./sshutils"
	"./sshutils/forward"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	logrus.Warnf("remove unix sockets before running this test")
	target := flag.String("target", "tcp://suda@tmp01:22", "ssh target addr")
	local := flag.String("local", "unix:///tmp/tmp.sock", "local socket addr")
	remote := flag.String("remote", "tcp://tmp01:8080", "remote addr addr")

	reverse := flag.Bool("reverse", false, "reverse-direction test with dummy http server")
	reverseRemote := flag.String("reverse-remote", "unix:///tmp/test-14242.sock", "remote addr addr for reverse test")

	flag.Parse()

	targetURL, err := url.Parse(*target)
	if err != nil {
		logrus.Fatal(err)
	}
	localURL, err := url.Parse(*local)
	if err != nil {
		logrus.Fatal(err)
	}
	remoteURL, err := url.Parse(*remote)
	if err != nil {
		logrus.Fatal(err)
	}
	reverseRemoteURL, err := url.Parse(*reverseRemote)
	if err != nil {
		logrus.Fatal(err)
	}

	if !*reverse {
		logrus.Infof("remote %q should be accessible as %q on this host",
			*remote, *local)
		f := &forward.Forwarder{
			targetURL.User.Username(),
			targetURL.Scheme, filepath.Join(targetURL.Host, targetURL.Path),
			localURL.Scheme, filepath.Join(localURL.Host, localURL.Path),
			remoteURL.Scheme, filepath.Join(remoteURL.Host, remoteURL.Path),
		}
		if err := f.Run(); err != nil {
			logrus.Fatal(err)
		}
	} else {
		logrus.Infof("built-in dummy http server on this local host should be accessible as %q on the remote host",
			*reverseRemote)
		logrus.Warnf("Unused: %q", *local)
		logrus.Warnf("Unused: %q", *remote)

		conn, err := sshutils.Dial(targetURL.User.Username(),
			targetURL.Scheme, filepath.Join(targetURL.Host, targetURL.Path))
		if err != nil {
			logrus.Fatal(err)
		}

		defer conn.Close()
		l, err := conn.Listen(reverseRemoteURL.Scheme,
			filepath.Join(reverseRemoteURL.Host, reverseRemoteURL.Path))
		if err != nil {
			logrus.Fatal(err)
		}
		defer l.Close()
		http.Serve(l, http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			fmt.Fprintf(resp, "Hello world from built-in dummy http server\n")
		}))
	}
}
