// Package forward provides SSH forwarder
package forward

import (
	"io"
	"net"

	"github.com/Sirupsen/logrus"

	sshutils ".."
)

// Forwarder forwards traffic via SSH connection
type Forwarder struct {
	SSHUser     string
	SSHProto    string
	SSHAddr     string
	LocalProto  string
	LocalAddr   string
	RemoteProto string
	RemoteAddr  string
}

// Run starts SSH forwarding
func (f *Forwarder) Run() error {
	sshClient, err := sshutils.Dial(f.SSHUser, f.SSHProto, f.SSHAddr)
	if err != nil {
		return err
	}
	listener, err := net.Listen(f.LocalProto, f.LocalAddr)
	if err != nil {
		return err
	}
	defer listener.Close()
	for {
		remoteConn, err := sshClient.Dial(f.RemoteProto, f.RemoteAddr)
		if err != nil {
			return err
		}
		localConn, err := listener.Accept()
		if err != nil {
			return err
		}
		go copier(localConn, remoteConn)
	}
	return nil
}

func copier(localConn, remoteConn net.Conn) {
	var broker = func(to, from net.Conn) {
		_, err := io.Copy(to, from)
		if err != nil {
			logrus.Error(err)
		}
	}
	go broker(localConn, remoteConn)
	go broker(remoteConn, localConn)
}
