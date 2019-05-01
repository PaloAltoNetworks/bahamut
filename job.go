// Copyright 2019 Aporeto Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bahamut

import "context"

// Job is the type of function that can be run as a Job.
type Job func() error

// RunJob runs a Job can than be canceled at any time according to the context.
func RunJob(ctx context.Context, job Job) (bool, error) {

	out := make(chan error)

	go func() { out <- job() }()

	select {
	case <-ctx.Done():
		return true, nil
	case err := <-out:
		return false, err
	}
}
