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

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestJob_RunJob(t *testing.T) {

	Convey("Given I have a context and a job func to run", t, func() {

		var called int
		l := &sync.Mutex{}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		j := func() error {
			l.Lock()
			called++
			l.Unlock()
			return nil
		}

		Convey("When I call RunJob", func() {

			interrupted, err := RunJob(ctx, j)

			Convey("Then interrupted should be false", func() {
				l.Lock()
				defer l.Unlock()
				So(interrupted, ShouldBeFalse)
				So(err, ShouldBeNil)
				So(called, ShouldEqual, 1)

			})
		})
	})

	Convey("Given I have a context and a job func to run that returns an error", t, func() {

		var called int

		l := &sync.Mutex{}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		j := func() error {
			l.Lock()
			called++
			l.Unlock()
			return errors.New("oops")
		}

		Convey("When I call RunJob", func() {

			interrupted, err := RunJob(ctx, j)

			Convey("Then interrupted should be false", func() {
				l.Lock()
				defer l.Unlock()
				So(interrupted, ShouldBeFalse)
				So(err, ShouldNotBeNil)
				So(called, ShouldEqual, 1)
			})
		})
	})

	Convey("Given I have a context and a job func to run that I cancel", t, func() {

		var called int
		l := &sync.Mutex{}
		l2 := &sync.Mutex{}

		ctx, cancel := context.WithCancel(context.Background())

		j := func() error {
			time.Sleep(300 * time.Millisecond)
			l.Lock()
			called++
			l.Unlock()
			return errors.New("oops")
		}

		Convey("When I call RunJob", func() {

			var interrupted bool
			var err error

			go func() {
				l2.Lock()
				interrupted, err = RunJob(ctx, j)
				l2.Unlock()
			}()
			time.Sleep(30 * time.Millisecond)
			cancel()

			Convey("Then interrupted should be false", func() {
				l.Lock()
				l2.Lock()
				defer l.Unlock()
				defer l2.Unlock()
				So(interrupted, ShouldBeTrue)
				So(err, ShouldBeNil)
				So(called, ShouldEqual, 0)
			})
		})
	})
}
