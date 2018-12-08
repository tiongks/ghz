package ghz

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/bojand/hri"
	"github.com/pkg/errors"
)

// RunConfig represents the request Configs
type RunConfig struct {
	// call settings
	call        string
	host        string
	proto       string
	importPaths []string
	protoset    string

	// securit settings
	cert     string
	cname    string
	insecure bool

	// test
	n   int
	c   int
	qps int

	// timeouts
	z             time.Duration
	timeout       time.Duration
	dialTimeout   time.Duration
	keepaliveTime time.Duration

	// data
	data     []byte
	binary   bool
	metadata []byte

	// misc
	name string
	cpus int
}

// Option controls some aspect of run
type Option func(*RunConfig) error

// WithCertificate specifies the certificate options for the run
func WithCertificate(cert string, cname string) Option {
	return func(o *RunConfig) error {
		cert = strings.TrimSpace(cert)
		if cert == "" {
			return errors.Errorf(fmt.Sprintf("cert required"))
		}

		o.cert = cert
		o.cname = cname

		return nil
	}
}

// WithInsecure specifies that this run should be done using insecure mode
func WithInsecure(o *RunConfig) error {
	o.insecure = true
	return nil
}

// WithN specifies the N (number of total requests) Config
func WithN(n uint) Option {
	return func(o *RunConfig) error {
		o.n = int(n)

		return nil
	}
}

// WithC specifies the N (number of concurrent requests) Config
func WithC(c uint) Option {
	return func(o *RunConfig) error {
		o.c = int(c)

		return nil
	}
}

// WithQPS specifies the QPS (queries per second) limit Config
func WithQPS(qps uint) Option {
	return func(o *RunConfig) error {
		o.qps = int(qps)

		return nil
	}
}

// WithZ specifies the Z (total test duration) Config
func WithZ(z time.Duration) Option {
	return func(o *RunConfig) error {
		o.z = z

		return nil
	}
}

// WithTimeout specifies the timeout for each request
func WithTimeout(timeout time.Duration) Option {
	return func(o *RunConfig) error {
		o.timeout = timeout

		return nil
	}
}

// WithDialTimeout specifies the inital connection dial timeout
func WithDialTimeout(dt time.Duration) Option {
	return func(o *RunConfig) error {
		o.dialTimeout = dt

		return nil
	}
}

// WithKeepalive specifies the keepalive timeout
func WithKeepalive(k time.Duration) Option {
	return func(o *RunConfig) error {
		o.keepaliveTime = k

		return nil
	}
}

// WithBinaryData specifies the binary data
func WithBinaryData(data []byte) Option {
	return func(o *RunConfig) error {
		o.data = data
		o.binary = true

		return nil
	}
}

// WithDataFromJSON loads JSON data from string
func WithDataFromJSON(data string) Option {
	return func(o *RunConfig) error {
		o.data = []byte(data)
		o.binary = false

		return nil
	}
}

// WithData specifies data as generic data that can be serailized to JSON
func WithData(data interface{}) Option {
	return func(o *RunConfig) error {
		dataJSON, err := json.Marshal(data)

		if err != nil {
			return err
		}

		o.data = dataJSON
		o.binary = false

		return nil
	}
}

// WithMetadataFromJSON specifies the metadata to be read from JSON string
func WithMetadataFromJSON(md string) Option {
	return func(o *RunConfig) error {
		o.metadata = []byte(md)

		return nil
	}
}

// WithMetadata specifies the metadata to be used as a map
func WithMetadata(md *map[string]string) Option {
	return func(o *RunConfig) error {
		mdJSON, err := json.Marshal(md)
		if err != nil {
			return err
		}

		o.metadata = mdJSON

		return nil
	}
}

// WithName sets the name of the test run
func WithName(name string) Option {
	return func(o *RunConfig) error {
		name = strings.TrimSpace(name)
		if name == "" {
			return errors.Errorf(fmt.Sprintf("name required"))
		}

		o.name = name

		return nil
	}
}

// WithCPUs specifies the number of CPU's to be used
func WithCPUs(c uint) Option {
	return func(o *RunConfig) error {
		o.cpus = int(c)

		return nil
	}
}

// WithProtoFile specified proto file path and optionally import paths
// We will automatically add the proto file path's directory and the current directory
func WithProtoFile(proto string, importPaths []string) Option {
	return func(o *RunConfig) error {
		proto = strings.TrimSpace(proto)
		if proto == "" {
			return errors.Errorf(fmt.Sprintf("proto path required"))
		}

		if filepath.Ext(proto) != ".proto" {
			return errors.Errorf(fmt.Sprintf("proto: must have .proto extension"))
		}

		o.proto = proto

		dir := filepath.Dir(proto)
		if dir != "." {
			o.importPaths = append(o.importPaths, dir)
		}

		o.importPaths = append(o.importPaths, ".")

		if len(importPaths) > 0 {
			o.importPaths = append(o.importPaths, importPaths...)
		}

		return nil
	}
}

// WithProtoset specified protoset file path
func WithProtoset(protoset string) Option {
	return func(o *RunConfig) error {
		protoset = strings.TrimSpace(protoset)
		if protoset == "" {
			return errors.Errorf(fmt.Sprintf("protoset path required"))
		}

		o.protoset = protoset

		return nil
	}
}

func newConfig(call, host string, options ...Option) (*RunConfig, error) {
	call = strings.TrimSpace(call)
	host = strings.TrimSpace(host)

	c := &RunConfig{
		call:        call,
		host:        host,
		n:           200,
		c:           50,
		timeout:     time.Duration(20 * time.Second),
		dialTimeout: time.Duration(10 * time.Second),
		cpus:        runtime.GOMAXPROCS(-1),
		name:        hri.Random(),
	}

	for _, option := range options {
		err := option(c)

		if err != nil {
			return nil, err
		}
	}

	if c.call == "" {
		return nil, errors.New("Call required")
	}

	if c.host == "" {
		return nil, errors.New("Host required")
	}

	if c.proto == "" && c.protoset == "" {
		return nil, errors.New("Must provide proto or protoset")
	}

	return c, nil
}

func minValue(v int, min int) error {
	if v < min {
		return errors.Errorf(fmt.Sprintf("must be at least %d", min))
	}

	return nil
}