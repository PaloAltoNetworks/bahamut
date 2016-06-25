// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

// import (
// 	"testing"
//
// 	. "github.com/smartystreets/goconvey/convey"
// )
//
// func TestRedisInfo_newRedisInfo(t *testing.T) {
//
// 	Convey("Given I create have a new redis info without cluter name", t, func() {
//
// 		redisInfo := NewRedisInfo([]string{":1234"}, "pass", 42, "")
//
// 		Convey("Then the redis info should have the address set", func() {
// 			So(len(redisInfo.Addresses), ShouldEqual, 1)
// 			So(redisInfo.Addresses[0], ShouldEqual, ":1234")
// 		})
//
// 		Convey("Then the redis info should have the password set", func() {
// 			So(redisInfo.Password, ShouldEqual, "pass")
// 		})
//
// 		Convey("Then the redis info should have the db number set", func() {
// 			So(redisInfo.DBNumber, ShouldEqual, 42)
// 		})
//
// 		Convey("Then the redis info should not have the cluster name set", func() {
// 			So(redisInfo.ClusterName, ShouldEqual, "")
// 		})
//
// 		Convey("Then the redis info IsSentinelModeActive should return false", func() {
// 			So(redisInfo.IsSentinelModeActive(), ShouldBeFalse)
// 		})
//
// 	})
//
// 	Convey("Given I create have a new redis cluster info", t, func() {
//
// 		redisInfo := NewRedisInfo([]string{":1234", ":1235"}, "pass", 42, "mycluster")
//
// 		Convey("Then the redis info should have the addresses set", func() {
// 			So(len(redisInfo.Addresses), ShouldEqual, 2)
// 			So(redisInfo.Addresses[0], ShouldEqual, ":1234")
// 			So(redisInfo.Addresses[1], ShouldEqual, ":1235")
// 		})
//
// 		Convey("Then the redis info should have the password set", func() {
// 			So(redisInfo.Password, ShouldEqual, "pass")
// 		})
//
// 		Convey("Then the redis info should have the db number set", func() {
// 			So(redisInfo.DBNumber, ShouldEqual, 42)
// 		})
//
// 		Convey("Then the redis info should have the cluster name set", func() {
// 			So(redisInfo.ClusterName, ShouldEqual, "mycluster")
// 		})
//
// 		Convey("Then the redis info IsSentinelModeActive should return true", func() {
// 			So(redisInfo.IsSentinelModeActive(), ShouldBeTrue)
// 		})
// 	})
// }
//
// func TestRedisInfo_String(t *testing.T) {
//
// 	Convey("Given I create have a new redis info", t, func() {
//
// 		redisInfo := NewRedisInfo([]string{"127.0.0.1:1234"}, "pass", 42, "")
//
// 		Convey("Then the string representation should be correct", func() {
// 			So(redisInfo.String(), ShouldEqual, "<redis address: 127.0.0.1:1234 db: 42>")
// 		})
// 	})
//
// 	Convey("Given I create have a new redis info with cluster info", t, func() {
//
// 		redisInfo := NewRedisInfo([]string{"127.0.0.1:1234", "127.0.0.1:1235"}, "pass", 42, "mycluster")
//
// 		Convey("Then the string representation should be correct", func() {
// 			So(redisInfo.String(), ShouldEqual, "<redis clusterName: mycluster addresses: [127.0.0.1:1234 127.0.0.1:1235] db: 42>")
// 		})
// 	})
// }
//
// func TestRedisInfo_makeRedisClient(t *testing.T) {
//
// 	Convey("Given I create have a new redis info", t, func() {
//
// 		redisInfo := NewRedisInfo([]string{":1234"}, "pass", 42, "")
//
// 		Convey("When I make a redis Client", func() {
//
// 			c := redisInfo.makeRedisClient()
//
// 			Convey("Then the client should be correctly set", func() {
// 				So(c, ShouldBeNil)
// 			})
// 		})
// 	})
//
// 	Convey("Given I create have a new redis info with cluster info", t, func() {
//
// 		redisInfo := NewRedisInfo([]string{":1234", ":1235"}, "pass", 42, "mycluster")
//
// 		Convey("When I make a redis Client", func() {
//
// 			c := redisInfo.makeRedisClient()
//
// 			Convey("Then the client should be correctly set", func() {
// 				So(c, ShouldBeNil)
// 			})
// 		})
// 	})
// }
