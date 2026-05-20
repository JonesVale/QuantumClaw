package channeltype

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestChannelBaseURLs(t *testing.T) {
	Convey("channel base urls length matches QuantumDummy", t, func() {
		So(len(ChannelBaseURLs), ShouldEqual, QuantumDummy)
	})
	Convey("quantum channel base urls are empty by default", t, func() {
		for i := IonQ; i < QuantumDummy; i++ {
			So(ChannelBaseURLs[i], ShouldEqual, "")
		}
	})
}
