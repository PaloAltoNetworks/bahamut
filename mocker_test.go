package bahamut

import (
	"testing"

	"github.com/aporeto-inc/elemental"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMocker_newMocker(t *testing.T) {

	Convey("Given I call newMocker", t, func() {

		m := newMocker()

		Convey("Then the mocker should be correctly initialized", func() {
			So(m.registry[elemental.OperationRetrieve], ShouldResemble, map[string]*Mock{})
			So(m.registry[elemental.OperationRetrieveMany], ShouldResemble, map[string]*Mock{})
			So(m.registry[elemental.OperationCreate], ShouldResemble, map[string]*Mock{})
			So(m.registry[elemental.OperationUpdate], ShouldResemble, map[string]*Mock{})
			So(m.registry[elemental.OperationDelete], ShouldResemble, map[string]*Mock{})
			So(m.registry[elemental.OperationInfo], ShouldResemble, map[string]*Mock{})
			So(m.registry[elemental.OperationPatch], ShouldResemble, map[string]*Mock{})
			So(&m.Mutex, ShouldNotBeNil)
		})
	})
}

func TestMocker_installation(t *testing.T) {

	Convey("Given I have a mocker", t, func() {

		m := newMocker()

		Convey("When I try to install a mock with a bad function", func() {

			err := m.installMock(&Mock{
				Function: "not-good-js",
			})

			Convey("Then err should be correct", func() {
				So(err.Error(), ShouldEqual, "invalid function: ReferenceError: 'not' is not defined")
			})
		})

		Convey("When I try to install a mock with missing operation", func() {

			err := m.installMock(&Mock{})

			Convey("Then err should be correct", func() {
				So(err.Error(), ShouldEqual, "invalid empty operation")
			})
		})

		Convey("When I try to install a mock with missing identity", func() {

			err := m.installMock(&Mock{
				Operation: elemental.OperationCreate,
			})

			Convey("Then err should be correct", func() {
				So(err.Error(), ShouldEqual, "invalid empty identity name")
			})
		})

		Convey("When I try to install a mock with invalid identity", func() {

			err := m.installMock(&Mock{
				Operation: elemental.Operation("not good"),
			})

			Convey("Then err should be correct", func() {
				So(err.Error(), ShouldEqual, "invalid operation: 'not good'. Must be one of 'create', 'update', 'delete', 'retrieve-many', 'retrieve', 'patch' or 'info'")
			})
		})

		Convey("When I try to install a valid mock", func() {

			m1 := &Mock{
				Operation:    elemental.OperationCreate,
				IdentityName: "toto",
			}
			err := m.installMock(m1)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("When I try to use get with valid op and identity", func() {

				m2 := m.get(elemental.OperationCreate, "toto")

				Convey("Then I should retrieve the same mock", func() {
					So(m1, ShouldEqual, m2)
				})

				Convey("When I unregister it", func() {

					err := m.uninstallMock(elemental.OperationCreate, "toto")

					Convey("Then err should be nil", func() {
						So(err, ShouldBeNil)
					})
				})

				Convey("When I unregister it using a bad operation", func() {

					err := m.uninstallMock(elemental.Operation("plif"), "toto")

					Convey("Then err should not be nil", func() {
						So(err.Error(), ShouldEqual, "invalid operation: 'plif'. Must be one of 'create', 'update', 'delete', 'retrieve-many', 'retrieve', 'patch' or 'info'")
					})
				})

				Convey("When I unregister it using a wrong operation", func() {

					err := m.uninstallMock(elemental.OperationRetrieve, "toto")

					Convey("Then err should not be nil", func() {
						So(err.Error(), ShouldEqual, "no mock installed for operation 'retrieve' and identity 'toto'")
					})
				})

				Convey("When I unregister it using a good operation but a bad identity name", func() {

					err := m.uninstallMock(elemental.OperationCreate, "titi")

					Convey("Then err should not be nil", func() {
						So(err.Error(), ShouldEqual, "no mock installed for operation 'create' and identity 'titi'")
					})
				})
			})

			Convey("When I try to use get with valid op and invalid identity", func() {

				m2 := m.get(elemental.OperationCreate, "not-toto")

				Convey("Then it should return nil", func() {
					So(m2, ShouldBeNil)
				})
			})

			Convey("When I try to use get with invalid op and valid identity", func() {

				m2 := m.get(elemental.OperationRetrieve, "toto")

				Convey("Then it should return nil", func() {
					So(m2, ShouldBeNil)
				})
			})
		})
	})
}
