package tunnel

import (
	"fmt"
	"strings"

	"github.com/RefalFalah/syrva/internal/host"
	"github.com/RefalFalah/syrva/internal/ssh"
)

func Args(h host.Host, t host.Tunnel) ([]string, error) {
	if err := host.ValidateTunnel(t); err != nil {
		return nil, err
	}
	args, err := ssh.BaseArgs(h)
	if err != nil {
		return nil, err
	}
	remote := strings.Trim(t.RemoteHost, "[]")
	if strings.Contains(remote, ":") {
		remote = "[" + remote + "]"
	}
	// An explicit loopback bind cannot expose the service via GatewayPorts.
	forward := fmt.Sprintf("127.0.0.1:%d:%s:%d", t.LocalPort, remote, t.RemotePort)
	args = append(args, "-N", "-o", "ExitOnForwardFailure=yes", "-L", forward)
	return append(args, "--", ssh.Destination(h)), nil
}
