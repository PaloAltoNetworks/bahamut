package bahamut

import (
	"testing"

	"github.com/aporeto-inc/elemental"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMock_execute(t *testing.T) {

	Convey("Given I have a mock", t, func() {

		m := &Mock{}

		Convey("When I set it as panic", func() {

			m.Panic = true
			a, err := m.execute(nil)

			Convey("Then the action should be correct", func() {
				So(a, ShouldEqual, mockActionDone)
			})

			Convey("Then calling should execute should return the panic requested error", func() {
				So(err.Error(), ShouldEqual, "Panic requested by mock")
			})
		})

		Convey("When I set a bad delay duration", func() {

			m.Delay = "not-good"
			a, err := m.execute(nil)

			Convey("Then the action should be correct", func() {
				So(a, ShouldEqual, mockActionDone)
			})

			Convey("Then calling should execute should return the panic requested error", func() {
				So(err.Error(), ShouldEqual, "mock: unable to parse duration: time: invalid duration not-good")
			})
		})

		Convey("When I set a bad code function", func() {

			m.Function = `not js code`
			a, err := m.execute(nil)

			Convey("Then the action should be correct", func() {
				So(a, ShouldEqual, mockActionDone)
			})

			Convey("Then err should be correct", func() {
				So(err.Error(), ShouldEqual, "mock: unable to parse function: (anonymous): Line 1:5 Unexpected identifier")
			})
		})

		Convey("When I set a code function with missig process func", func() {

			m.Function = `function not_process(ctx) {return null}`
			a, err := m.execute(nil)

			Convey("Then the action should be correct", func() {
				So(a, ShouldEqual, mockActionDone)
			})

			Convey("Then err should be correct", func() {
				So(err.Error(), ShouldEqual, "mock: unable to call 'process': ReferenceError: 'process' is not defined")
			})
		})

		Convey("When I set a code function that returns null", func() {

			m.Function = `function process(ctx) {return null}`
			a, err := m.execute(nil)

			Convey("Then the action should be correct", func() {
				So(a, ShouldEqual, mockActionContinue)
			})

			Convey("Then err should be correct", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When I set a code function that returns a missing code", func() {

			m.Function = `function process(ctx) { return {} }`
			a, err := m.execute(NewContext())

			Convey("Then the action should be correct", func() {
				So(a, ShouldEqual, mockActionDone)
			})

			Convey("Then err should be correct", func() {
				So(err.Error(), ShouldEqual, "mock: function returned undefined for 'code'")
			})
		})

		Convey("When I set a code function that returns a non int code", func() {

			m.Function = `function process(ctx) { return {code: "hey"} }`
			a, err := m.execute(NewContext())

			Convey("Then the action should be correct", func() {
				So(a, ShouldEqual, mockActionDone)
			})

			Convey("Then err should be correct", func() {
				So(err.Error(), ShouldEqual, "mock: function returned a string for 'code': hey")
			})
		})

		Convey("When I set a code function that returns undefined data", func() {

			m.Function = `function process(ctx) { return {code: 200} }`
			a, err := m.execute(NewContext())

			Convey("Then the action should be correct", func() {
				So(a, ShouldEqual, mockActionDone)
			})

			Convey("Then err should be correct", func() {
				So(err.Error(), ShouldEqual, "mock: function returned undefined for 'data'")
			})
		})

		Convey("When I set a code function that returns non string data", func() {

			m.Function = `function process(ctx) { return {code: 200, data: 1} }`
			a, err := m.execute(NewContext())

			Convey("Then the action should be correct", func() {
				So(a, ShouldEqual, mockActionDone)
			})

			Convey("Then err should be correct", func() {
				So(err.Error(), ShouldEqual, "mock: function returned a non string for 'data': 1")
			})
		})

		Convey("When I set a code function that returns data with invalid json", func() {

			m.Function = `function process(ctx) { return {code: 200, data: "{broken}"} }`
			a, err := m.execute(NewContext())

			Convey("Then the action should be correct", func() {
				So(a, ShouldEqual, mockActionDone)
			})

			Convey("Then err should be correct", func() {
				So(err.Error(), ShouldEqual, `mock: unable to decode provided data: ReadMapCB: expect " after }, but found b, error found in #2 byte of ...|{broken}|..., bigger context ...|{broken}|...`)
			})
		})

		Convey("When I set a code function that returns data array for operation RetrieveMany", func() {

			m.Function = `function process(ctx) { return {code: 333, data: JSON.stringify([{name: "toto1"}, {name: "toto2"}])} }`
			ctx := NewContext()
			ctx.Request = &elemental.Request{
				Operation: elemental.OperationRetrieveMany,
			}
			a, err := m.execute(ctx)

			Convey("Then the action should be correct", func() {
				So(a, ShouldEqual, mockActionDone)
			})

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the code should be correct", func() {
				So(ctx.StatusCode, ShouldEqual, 333)
			})

			Convey("Then the the output data of the context should be correct", func() {
				o := ctx.OutputData.([]interface{})
				So(len(o), ShouldEqual, 2)
				So(o[0].(map[string]interface{})["name"], ShouldEqual, "toto1")
				So(o[1].(map[string]interface{})["name"], ShouldEqual, "toto2")
			})
		})

		Convey("When I set a code function that returns data array for operation Retrieve", func() {

			m.Function = `function process(ctx) { return {code: 333, data: JSON.stringify({name: "toto1"})} }`
			ctx := NewContext()
			ctx.Request = &elemental.Request{
				Operation: elemental.OperationRetrieveMany,
			}
			a, err := m.execute(ctx)

			Convey("Then the action should be correct", func() {
				So(a, ShouldEqual, mockActionDone)
			})

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the code should be correct", func() {
				So(ctx.StatusCode, ShouldEqual, 333)
			})

			Convey("Then the the output data of the context should be correct", func() {
				So(ctx.OutputData.(map[string]interface{})["name"], ShouldEqual, "toto1")
			})
		})
	})
}
