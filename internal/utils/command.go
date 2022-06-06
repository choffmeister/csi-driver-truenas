package utils

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

func Command(name string, args ...string) (string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	Debug.Printf("Executing command %s %v\n", name, args)
	cmd := exec.CommandContext(ctx, name, args...)
	outputBytes, err := cmd.CombinedOutput()
	output := string(outputBytes)
	Debug.Printf("Executed command %s %v: %s", name, args, output)
	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if !ok {
			return output, 0, err
		}
		return output, exitError.ExitCode(), fmt.Errorf("%w\n%s", exitError, output)
	}
	return output, 0, nil
}
