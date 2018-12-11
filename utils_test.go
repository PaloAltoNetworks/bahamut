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

	opentracing "github.com/opentracing/opentracing-go"
	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/elemental"
)

func TestUtils_RecoverFromPanic(t *testing.T) {

	Convey("Given I call a function that panics", t, func() {

		var err error
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer func() {
				err = handleRecoveredPanic(context.TODO(), recover(), false)
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
				handleRecoveredPanic(context.TODO(), recover(), true) // nolint
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
				err = handleRecoveredPanic(context.TODO(), recover(), true)
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

		Convey("When I call processError on standard error", func() {

			errIn := errors.New("boom")
			errOut := processError(ctx, errIn)

			Convey("Then errOut should be correct", func() {
				So(errOut, ShouldHaveSameTypeAs, elemental.Errors{})
				So(errOut.Code(), ShouldEqual, 500)
				So(errOut.Error(), ShouldEqual, "error 500 (bahamut): Internal Server Error: boom [trace: unknown]")
			})
		})

		Convey("When I call processError on elemental.Error error", func() {

			errIn := elemental.NewError("boom", "blang", "sub", http.StatusNotFound)
			errOut := processError(ctx, errIn)

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

			errOut := processError(ctx, errIn)

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

func TestTag_splitPtr(t *testing.T) {

	Convey("Given I have a tag a=b", t, func() {

		var k, v string
		t := "a=b"

		Convey("When I call Split", func() {

			e := splitPtr(t, &k, &v)

			Convey("Then e should be nil", func() {
				So(e, ShouldBeNil)
			})

			Convey("Then k should equal a", func() {
				So(k, ShouldEqual, "a")
			})

			Convey("Then v should equal b", func() {
				So(v, ShouldEqual, "b")
			})

		})
	})

	Convey("Given I have a tag a=b c", t, func() {

		var k, v string
		t := "a=b c"

		Convey("When I call Split", func() {

			e := splitPtr(t, &k, &v)

			Convey("Then e should be nil", func() {
				So(e, ShouldBeNil)
			})

			Convey("Then k should equal a", func() {
				So(k, ShouldEqual, "a")
			})

			Convey("Then v should equal b c", func() {
				So(v, ShouldEqual, "b c")
			})

		})
	})

	Convey("Given I have a tag a=b c=ddd", t, func() {

		var k, v string
		t := "a=b c=ddd"

		Convey("When I call Split", func() {

			e := splitPtr(t, &k, &v)

			Convey("Then e should be nil", func() {
				So(e, ShouldBeNil)
			})

			Convey("Then k should equal a", func() {
				So(k, ShouldEqual, "a")
			})

			Convey("Then v should equal b c=ddd", func() {
				So(v, ShouldEqual, "b c=ddd")
			})

		})
	})

	Convey("Given I have a tag a", t, func() {

		var k, v string
		t := "a"

		Convey("When I call Split", func() {

			e := splitPtr(t, &k, &v)

			Convey("Then e should not be nil", func() {
				So(e.Error(), ShouldEqual, "Invalid tag: invalid length 'a'")
			})

		})
	})

	Convey("Given I have a tag a", t, func() {

		var k, v string
		t := "a="

		Convey("When I call Split", func() {

			e := splitPtr(t, &k, &v)

			Convey("Then e should not be nil", func() {
				So(e.Error(), ShouldEqual, "Invalid tag: invalid length 'a='")
			})

		})
	})

	Convey("Given I have a tag a", t, func() {

		var k, v string
		t := "abc"

		Convey("When I call Split", func() {

			e := splitPtr(t, &k, &v)

			Convey("Then e should not be nil", func() {
				So(e.Error(), ShouldEqual, "Invalid tag: missing equal symbol 'abc'")
			})

		})
	})

	Convey("Given I have a tag a", t, func() {

		var k, v string
		t := "abc="

		Convey("When I call Split", func() {

			e := splitPtr(t, &k, &v)

			Convey("Then e should not be nil", func() {
				So(e.Error(), ShouldEqual, "Invalid tag: missing value 'abc='")
			})

		})
	})

	Convey("Given I have a tag a", t, func() {

		var k, v string
		t := "=abc"

		Convey("When I call Split", func() {

			e := splitPtr(t, &k, &v)

			Convey("Then e should not be nil", func() {
				So(e.Error(), ShouldEqual, "Invalid tag: missing key '=abc'")
			})

		})
	})
}
