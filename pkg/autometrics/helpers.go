package autometrics // import "github.com/autometrics-dev/autometrics-go/pkg/autometrics"

import (
	"errors"
	"fmt"
	"net"

	"github.com/oklog/ulid/v2"
)

// GetOutboundIP returns the preferred outbound ip of this machine.
//
// This function temporarily opens a TCP connexion to a Cloudflare-owned server
// to determine the external IP of the current process.
//
// This is a useful function to use for setting a `job` key when
// pushing metrics.
func GetOutboundIP() (net.IP, error) {
	conn, err := net.Dial("tcp", "1.1.1.1:80")
	if err != nil {
		return nil, fmt.Errorf("could not connect outside: %w", err)
	}
	defer conn.Close()

	localAddr, ok := conn.LocalAddr().(*net.TCPAddr)

	if ok {
		return localAddr.IP, nil
	}

	return nil, errors.New("No IP found.")
}

// DefaultJobName returns the default job name to use when pushing metrics to a collector.
//
// This function cannot fail.
func DefaultJobName() string {
	return ulid.Make().String()
}
