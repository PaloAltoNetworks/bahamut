// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"

	"github.com/opentracing/opentracing-go"

	"github.com/aporeto-inc/elemental"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUtils_RecoverFromPanic(t *testing.T) {

	Convey("Given I call a function that panics", t, func() {

		var err error
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer func() {
				err = handleRecoveredPanic(context.TODO(), recover(), elemental.NewResponse(), true)
				wg.Done()
			}()
			panic("this is a panic!")
		}()

		wg.Wait()

		Convey("Then err should not be nil", func() {
			So(err, ShouldNotBeNil)
		})
	})

	Convey("Given I call a function that panics and I don't want to recover", t, func() {

		f := func() {
			defer func() {
				handleRecoveredPanic(context.TODO(), recover(), elemental.NewResponse(), false) // nolint
			}()
			func() { panic("this is a panic!") }()
		}

		Convey("Then err should not be nil", func() {
			So(f, ShouldPanic)
		})
	})

	Convey("Given I call a function that doesn't panic", t, func() {

		var err error
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer func() {
				err = handleRecoveredPanic(context.TODO(), recover(), elemental.NewResponse(), true)
				wg.Done()
			}()
			func() {}()
		}()

		wg.Wait()

		Convey("Then err should be nil", func() {
			So(err, ShouldBeNil)
		})
	})
}

func TestUtils_processError(t *testing.T) {

	Convey("Given I have an error and response with a span", t, func() {

		_, ctx := opentracing.StartSpanFromContext(context.Background(), "test")
		resp := elemental.NewResponse()

		Convey("When I call processError on standard error", func() {

			errIn := errors.New("boom")
			errOut := processError(ctx, errIn, resp)

			Convey("Then errOut should be correct", func() {
				So(errOut, ShouldHaveSameTypeAs, elemental.Errors{})
				So(errOut.Code(), ShouldEqual, 500)
				So(errOut.Error(), ShouldEqual, "error 500 (bahamut): Internal Server Error: boom [trace: unknown]")
			})
		})

		Convey("When I call processError on elemental.Error error", func() {

			errIn := elemental.NewError("boom", "blang", "sub", http.StatusNotFound)
			errOut := processError(ctx, errIn, resp)

			Convey("Then errOut should be correct", func() {
				So(errOut, ShouldHaveSameTypeAs, elemental.Errors{})
				So(errOut.Code(), ShouldEqual, http.StatusNotFound)
				So(errOut.Error(), ShouldEqual, "error 404 (sub): boom: blang [trace: unknown]")
			})
		})

		Convey("When I call processError on elemental.Errors error", func() {

			errIn := elemental.NewErrors(
				elemental.NewError("boom", "blang", "sub", http.StatusNotFound),
				elemental.NewError("clash", "klong", "sub", http.StatusMovedPermanently),
				errors.New("kaboom"),
			)

			errOut := processError(ctx, errIn, resp)

			Convey("Then errOut should be correct", func() {
				So(errOut, ShouldHaveSameTypeAs, elemental.Errors{})
				So(errOut.At(0).Code, ShouldEqual, http.StatusNotFound)
				So(errOut.At(1).Code, ShouldEqual, http.StatusMovedPermanently)
				So(errOut.At(2).Code, ShouldEqual, http.StatusInternalServerError)
				So(errOut.Error(), ShouldEqual, "error 404 (sub): boom: blang [trace: unknown], error 301 (sub): clash: klong [trace: unknown], error 500 (bahamut): Internal Server Error: kaboom [trace: unknown]")
			})
		})
	})
}

func TestUtils_claimsToMap(t *testing.T) {

	Convey("Given I have a claims list", t, func() {

		claims := []string{"a=b", "c=d"}

		Convey("When I call claimsToMap", func() {

			out := claimsToMap(claims)

			Convey("Then the maps should be correct", func() {
				So(len(out), ShouldEqual, 2)
				So(out["a"], ShouldEqual, "b")
				So(out["c"], ShouldEqual, "d")
			})
		})
	})

	Convey("Given I have a claims with bad claims", t, func() {

		claims := []string{"a=b", "c"}

		Convey("When I call claimsToMap", func() {

			Convey("Then it should should panic", func() {
				So(func() { claimsToMap(claims) }, ShouldPanic)
			})
		})
	})
}
