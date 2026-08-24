package main

import (
	"context"
	"slices"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestSliceFlagsKeepSeparators(t *testing.T) {
	tests := []struct {
		name string
		path []string
		flag string
		args []string
		want []string
	}{
		{
			name: "lambda env",
			path: []string{"lambda"},
			flag: "env",
			args: []string{"eru-cli", "lambda", "--env", "JAVA_OPTS=-Xmx1g,-Xms1g", "true"},
			want: []string{"JAVA_OPTS=-Xmx1g,-Xms1g"},
		},
		{
			name: "workload deploy after-create",
			path: []string{"workload", "deploy"},
			flag: "after-create",
			args: []string{"eru-cli", "workload", "deploy", "--after-create", "echo a,b", "spec.yaml"},
			want: []string{"echo a,b"},
		},
		{
			name: "node add label",
			path: []string{"node", "add"},
			flag: "label",
			args: []string{"eru-cli", "node", "add", "--label", "a=1,2", "dev"},
			want: []string{"a=1,2"},
		},
		{
			name: "repeated flags still accumulate",
			path: []string{"lambda"},
			flag: "env",
			args: []string{"eru-cli", "lambda", "--env", "A=1", "--env", "B=2", "true"},
			want: []string{"A=1", "B=2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newApp()
			var got []string
			lookupCommand(t, app, tt.path).Action = func(_ context.Context, cmd *cli.Command) error {
				got = cmd.StringSlice(tt.flag)
				return nil
			}
			if err := app.Run(t.Context(), tt.args); err != nil {
				t.Fatalf("run: %v", err)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func lookupCommand(t *testing.T, app *cli.Command, path []string) *cli.Command {
	t.Helper()
	cmd := app
	for _, name := range path {
		idx := slices.IndexFunc(cmd.Commands, func(c *cli.Command) bool { return c.Name == name })
		if idx < 0 {
			t.Fatalf("command %v not found under %q", path, cmd.Name)
		}
		cmd = cmd.Commands[idx]
	}
	return cmd
}
