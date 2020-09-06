package iio

import (
	"bytes"
	"context"
	"os/exec"
	"time"
)

func Execute(command string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1 * time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)
	outputBuffer := &bytes.Buffer{}
	cmd.Stdout = outputBuffer
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	return outputBuffer.Bytes(), nil
}