package runner

import (
	"os"
	"os/signal"
	"runtime"
	"time"
)

// Run executes the test
//
//	report, err := runner.Run(
//		"helloworld.Greeter.SayHello",
//		"localhost:50051",
//		WithProtoFile("greeter.proto", []string{}),
//		WithDataFromFile("data.json"),
//		WithInsecure(true),
//	)
func Run(call, host string, options ...Option) (*Report, error) {
	c, err := newConfig(call, host, options...)

	if err != nil {
		return nil, err
	}

	oldCPUs := runtime.NumCPU()

	runtime.GOMAXPROCS(c.cpus)

	reqr, err := newRequester(c)

	if err != nil {
		return nil, err
	}

	cancel := make(chan os.Signal, 1)
	signal.Notify(cancel, os.Interrupt)
	go func() {
		<-cancel
		reqr.Stop(ReasonCancel)
	}()

	if c.z > 0 {
		go func() {
			time.Sleep(c.z)
			reqr.Stop(ReasonTimeout)
		}()
	}

	rep, err := reqr.Run()

	// reset
	runtime.GOMAXPROCS(oldCPUs)

	return rep, err
}
