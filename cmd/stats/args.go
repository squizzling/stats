package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"

	"github.com/squizzling/stats/internal/emitters/procnetdev"
)

type Opts struct {
	Target    string             `short:"t" long:"target"                   description:"target statsd address" `
	Host      *string            `          long:"host"                     description:"local hostname"        `
	List      bool               `short:"l" long:"list"                     description:"List emitters"         `
	Disable   func(string) error `short:"d" long:"disable"                  description:"Disable emitter"       `
	Enable    func(string) error `short:"e" long:"enable"                   description:"Enable emitter"        `
	Interval  time.Duration      `short:"i" long:"interval" default:"1s"    description:"send interval"         `
	Verbose   bool               `short:"v" long:"verbose"                  description:"Enable verbose logging"`
	FakeStats bool               `short:"f" long:"fake-stats"               description:"Log stats only"        `
	procnetdev.ProcNetDevOpts

	positional  []string
	haveEnable  bool
	haveDisable bool
	selected    map[string]struct{}
}

func funcMakeEnableDisable(opts *Opts, enable bool) func(s string) error {
	return func(s string) error {
		if enable {
			opts.haveEnable = true
		} else if !enable {
			opts.haveDisable = true
		}
		opts.selected[s] = struct{}{}
		return nil
	}
}

func (opts *Opts) Get(name string) interface{} {
	if name == "procnetdev" {
		return &opts.ProcNetDevOpts
	}
	return nil
}

func (opts *Opts) Validate() []string {
	var errors []string

	if len(opts.positional) != 0 {
		errors = append(errors, "no positional arguments are allowed")
	}

	if opts.haveEnable && opts.haveDisable {
		errors = append(errors, "enable and disable are mutually exclusive")
	}

	if !opts.FakeStats {
		if opts.Host == nil {
			host, err := os.Hostname()
			if err != nil {
				errors = append(errors, fmt.Sprintf("unable to get hostname (%v), use --host", err))
			} else {
				opts.Host = &host
			}
		}

		if opts.Target == "" {
			errors = append(errors, "target is required when fake-stats is not enabled")
		}

		if strings.IndexByte(opts.Target, ':') == -1 {
			opts.Target = opts.Target + ":8125"
		}
	}

	return errors
}

func parseArgs(args []string) *Opts {
	opts := &Opts{}

	opts.selected = make(map[string]struct{})
	opts.Enable = funcMakeEnableDisable(opts, true)
	opts.Disable = funcMakeEnableDisable(opts, false)

	parser := flags.NewParser(opts, flags.HelpFlag|flags.PassDoubleDash)
	positional, err := parser.ParseArgs(args)
	if err != nil {
		if !isHelp(err) {
			parser.WriteHelp(os.Stderr)
			_, _ = fmt.Fprintf(os.Stderr, "\n\nerror parsing command line: %v\n", err)
			os.Exit(1)
		}
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}
	opts.positional = positional

	var errors []string

	errors = append(errors, opts.Validate()...)
	errors = append(errors, opts.ProcNetDevOpts.Validate()...)

	if len(errors) > 0 {
		parser.WriteHelp(os.Stderr)
		_, _ = fmt.Fprintf(os.Stderr, "\n\n")
		for _, err := range errors {
			_, _ = fmt.Fprintf(os.Stderr, "error parsing command line: %s\n", err)
		}
		os.Exit(1)
	}

	return opts
}

// isHelp is a helper to test the error from ParseArgs() to
// determine if the help message was written. It is safe to
// call without first checking that error is nil.
func isHelp(err error) bool {
	// This was copied from https://github.com/jessevdk/go-flags/blame/master/help.go#L499, as there has not been an
	// official release yet with this code. Renamed from WriteHelp to isHelp, as flags.ErrHelp is still returned when
	// flags.HelpFlag is set, flags.PrintError is clear, and -h/--help is passed on the command line, even though the
	// help is not displayed in such a situation.
	if err == nil { // No error
		return false
	}

	flagError, ok := err.(*flags.Error)
	if !ok { // Not a go-flag error
		return false
	}

	if flagError.Type != flags.ErrHelp { // Did not print the help message
		return false
	}

	return true
}
