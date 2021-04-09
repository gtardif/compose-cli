/*
   Copyright 2020 Docker Compose CLI authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package compose

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/compose-spec/compose-go/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/docker/cli/cli"
	"github.com/docker/compose-cli/api/client"
	"github.com/docker/compose-cli/api/compose"
	"github.com/docker/compose-cli/api/errdefs"
	"github.com/docker/compose-cli/api/progress"
)

type buildOptions struct {
	*projectOptions
	composeOptions
	quiet    bool
	pull     bool
	progress string
	args     []string
	noCache  bool
	memory   string
}

func buildCommand(p *projectOptions) *cobra.Command {
	opts := buildOptions{
		projectOptions: p,
	}
	cmd := &cobra.Command{
		Use:   "build [SERVICE...]",
		Short: "Build or rebuild services",
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.memory != "" {
				fmt.Println("WARNING --memory is ignored as not supported in buildkit.")
			}
			if opts.quiet {
				devnull, err := os.Open(os.DevNull)
				if err != nil {
					return err
				}
				os.Stdout = devnull
			}
			return runBuild(cmd.Context(), opts, args)
		},
	}
	cmd.Flags().BoolVarP(&opts.quiet, "quiet", "q", false, "Don't print anything to STDOUT")
	cmd.Flags().BoolVar(&opts.pull, "pull", false, "Always attempt to pull a newer version of the image.")
	cmd.Flags().StringVar(&opts.progress, "progress", "auto", `Set type of progress output ("auto", "plain", "tty")`)
	cmd.Flags().StringArrayVar(&opts.args, "build-arg", []string{}, "Set build-time variables for services.")
	cmd.Flags().Bool("parallel", true, "Build images in parallel. DEPRECATED")
	cmd.Flags().MarkHidden("parallel") //nolint:errcheck
	cmd.Flags().Bool("compress", true, "Compress the build context using gzip. DEPRECATED")
	cmd.Flags().MarkHidden("compress") //nolint:errcheck
	cmd.Flags().Bool("force-rm", true, "Always remove intermediate containers. DEPRECATED")
	cmd.Flags().MarkHidden("force-rm") //nolint:errcheck
	cmd.Flags().BoolVar(&opts.noCache, "no-cache", false, "Do not use cache when building the image")
	cmd.Flags().Bool("no-rm", false, "Do not remove intermediate containers after a successful build. DEPRECATED")
	cmd.Flags().MarkHidden("no-rm") //nolint:errcheck
	cmd.Flags().StringVarP(&opts.memory, "memory", "m", "", "Set memory limit for the build container. Not supported on buildkit yet.")
	cmd.Flags().MarkHidden("memory") //nolint:errcheck

	return cmd
}

func runBuild(ctx context.Context, opts buildOptions, services []string) error {
	ctx, cancel := context.WithCancel(ctx)
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-s
		cancel()
	}()
	c, err := client.New(ctx)
	if err != nil {
		return err
	}

	project, err := opts.toProject(services)
	if err != nil {
		return err
	}

	_, err = progress.Run(ctx, func(ctx context.Context) (string, error) {
		return "", c.ComposeService().Build(ctx, project, compose.BuildOptions{
			Pull:     opts.pull,
			Progress: opts.progress,
			Args:     types.NewMapping(opts.args),
			NoCache:  opts.noCache,
		})
	})

	if err != nil {
		if errdefs.IsErrCanceled(err) || errors.Is(ctx.Err(), context.Canceled) {
			return cli.StatusError{StatusCode: 130}
		}
	}

	return err
}
