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

package win

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/icmd"

	. "github.com/docker/compose-cli/utils/e2e"
)

func TestLocalComposeBuild(t *testing.T) {
	c := NewParallelE2eCLI(t, binDir)

	t.Run("build named and unnamed images", func(t *testing.T) {
		// ensure local test run does not reuse previously build image
		c.RunDockerOrExitError("rmi", "custom-aspnet")

		res := c.RunDockerCmd("compose", "--project-directory", "../compose/fixtures/build-test-win", "build")
		defer c.RunDockerCmd("rmi", "custom-aspnet")

		res.Assert(t, icmd.Expected{Out: "COPY aspnetapp/. ./aspnetapp/"})
		c.RunDockerCmd("image", "inspect", "custom-aspnet")
	})
}
