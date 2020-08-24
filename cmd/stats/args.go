package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
)

type Opts struct {
	Target        string             `short:"t" long:"target"   required:"true" description:"target statsd address"`
	Host          *string            `          long:"host"                     description:"local hostname"       `
	List          bool               `short:"l" long:"list"                     description:"List emitters"        `
	Disable       func(string) error `short:"d" long:"disable"                  description:"Disable emitter"      `
	Enable        func(string) error `short:"e" long:"enable"                   description:"Enable emitter"       `
	Interval      time.Duration      `short:"i" long:"interval" default:"1s"    description:"send interval"        `
	enableDisable *bool
	selected      map[string]struct{}
}

func funcMakeEnableDisable(opts *Opts, enable bool) func(s string) error {
	return func(s string) error {
		if opts.enableDisable == nil {
			opts.enableDisable = &enable
		}
		if *opts.enableDisable == enable {
			opts.selected[s] = struct{}{}
		} else {
			return errors.New("enable and disable are mutually exclusive")
		}
		return nil
	}
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

	if len(positional) != 0 {
		// Near as I can tell there's no way to say no positional arguments allowed.
		parser.WriteHelp(os.Stderr)
		_, _ = fmt.Fprintf(os.Stderr, "\n\nno positional arguments allowed\n")
		os.Exit(1)
	}

	if opts.Host == nil {
		host, err := os.Hostname()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "\n\nUnable to get hostname: %v\nUse --host to manually specify\n", err)
			os.Exit(1)
		}
		opts.Host = &host
	}

	if strings.IndexByte(opts.Target, ':') == -1 {
		opts.Target = opts.Target + ":8125"
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
