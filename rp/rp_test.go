// Copyright © 2017, 2018 Genome Research Limited
// Author: Sendu Bala <sb10@sanger.ac.uk>.
//
//  This file is part of wr.
//
//  wr is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Lesser General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  wr is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Lesser General Public License for more details.
//
//  You should have received a copy of the GNU Lesser General Public License
//  along with wr. If not, see <http://www.gnu.org/licenses/>.

package rp

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func BenchmarkRP(b *testing.B) {
	delayBetween := 0 * time.Millisecond
	releaseTimeout := 200 * time.Millisecond
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		rp1 := New("l1", delayBetween, 5, releaseTimeout)
		rp2 := New("l2", delayBetween, 6, releaseTimeout)

		r11 := getRequest(rp1, 1)
		r12 := getRequest(rp1, 1)
		r13 := getRequest(rp1, 1)
		r14 := getRequest(rp1, 1)
		r15 := getRequest(rp1, 1)
		r16 := getRequest(rp1, 1)
		r21 := getRequest(rp2, 1)
		r22 := getRequest(rp2, 1)
		r23 := getRequest(rp2, 1)
		r24 := getRequest(rp2, 1)
		r25 := getRequest(rp2, 1)
		r26 := getRequest(rp2, 1)
		r27 := getRequest(rp2, 1)

		rp1.WaitUntilGranted(r11)
		rp1.WaitUntilGranted(r12)
		rp1.WaitUntilGranted(r13)
		rp1.WaitUntilGranted(r14)
		rp1.WaitUntilGranted(r15)
		rp1.Granted(r16)
		rp2.WaitUntilGranted(r21)
		rp2.WaitUntilGranted(r22)
		rp2.WaitUntilGranted(r23)
		rp2.WaitUntilGranted(r24)
		rp2.WaitUntilGranted(r25)
		rp2.WaitUntilGranted(r26)
		rp2.Granted(r27)

		rp1.Release(r11)
		rp1.Release(r12)
		rp1.Release(r13)
		rp1.Release(r14)
		rp1.Release(r15)
		rp2.Release(r21)
		rp2.Release(r22)
		rp2.Release(r23)
		rp2.Release(r24)
		rp2.Release(r25)
		rp2.Release(r26)

		r11 = getRequest(rp1, 1)
		r12 = getRequest(rp1, 1)
		r13 = getRequest(rp1, 1)
		r14 = getRequest(rp1, 1)
		r15 = getRequest(rp1, 1)
		r16 = getRequest(rp1, 1)
		r17 := getRequest(rp1, 1)
		r18 := getRequest(rp1, 1)
		r19 := getRequest(rp1, 1)
		r110 := getRequest(rp1, 1)

		rp1.Granted(r11)
		rp1.Granted(r12)
		rp1.Granted(r13)
		rp1.Granted(r14)
		rp1.Granted(r15)
		rp1.Granted(r16)
		rp1.Granted(r17)
		rp1.Granted(r18)
		rp1.Granted(r19)
		rp1.Granted(r110)
		rp1.Release(r11)
		rp1.Release(r12)
		rp1.Release(r13)
		rp1.Release(r14)
		rp1.Release(r15)
		rp1.Release(r16)
		rp1.Release(r17)
		rp1.Release(r18)
		rp1.Release(r19)
		rp1.Release(r110)
	}
}

func getRequest(rp *Protector, numTokens int) Receipt {
	r, err := rp.Request(numTokens)
	if err != nil {
		fmt.Printf("Request had an error: %s\n", err)
	}
	return r
}

func TestRP(t *testing.T) {
	Convey("You can make a new Protector", t, func() {
		delayInt := 50
		delayBetween := time.Duration(delayInt) * time.Millisecond
		maxSimultaneous := 3
		releaseTimeout := time.Duration(delayInt*5) * time.Millisecond
		halfDelay := time.Duration(delayInt/2) * time.Millisecond
		oneFiftyPercentDelay := time.Duration(delayInt+(delayInt/2)) * time.Millisecond
		doubleDelay := time.Duration(delayInt*2) * time.Millisecond

		rp := New("irods", delayBetween, maxSimultaneous, releaseTimeout)
		So(rp, ShouldNotBeNil)
		begin := time.Now()

		Convey("Request() returns immediately, but there is a delay between each granting and once all tokens have been granted", func() {
			grantedCh := make(chan time.Time, maxSimultaneous)
			for i := 1; i <= maxSimultaneous; i++ {
				r, err := rp.Request(1)
				So(err, ShouldBeNil)

				go func(r Receipt) {
					rp.WaitUntilGranted(r)
					grantedCh <- time.Now()
				}(r)
			}

			So(time.Now(), ShouldHappenBefore, begin.Add(halfDelay))

			r, err := rp.Request(1)
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(rp.WaitUntilGranted(r), ShouldBeTrue)
			So(time.Now(), ShouldHappenOnOrBetween, begin.Add(releaseTimeout), begin.Add(releaseTimeout).Add(halfDelay))
			rp.Release(r)

			for i := 0; i < maxSimultaneous; i++ {
				So(<-grantedCh, ShouldHappenOnOrBetween, begin.Add(time.Duration(delayInt*i)*time.Millisecond), begin.Add(time.Duration(delayInt*i)*time.Millisecond).Add(halfDelay))
			}
		})

		Convey("You can't Request more tokens than max", func() {
			r, err := rp.Request(maxSimultaneous + 1)
			So(string(r), ShouldBeBlank)
			So(err, ShouldNotBeNil)
			rperr, ok := err.(Error)
			So(ok, ShouldBeTrue)
			So(rperr.Err, ShouldEqual, ErrOverMaximumTokens)
		})

		Convey("You can't do anything with an invalid receipt", func() {
			r, err := rp.Request(1)
			So(err, ShouldBeNil)

			badR := Receipt("invalid")
			So(rp.WaitUntilGranted(badR), ShouldBeFalse)
			So(rp.WaitUntilGranted(r), ShouldBeTrue)

			// Touch() and Release() don't return anything; the most we can do
			// is confirm we don't crash
			rp.Touch(badR)
			rp.Release(badR)
			rp.Touch(r)
			rp.Release(r)
		})

		Convey("You can't do anything with a Shutdown() Protector", func() {
			r, err := rp.Request(1)
			So(err, ShouldBeNil)
			So(rp.WaitUntilGranted(r), ShouldBeTrue)
			r2, err := rp.Request(1)
			So(err, ShouldBeNil)

			rp.Shutdown()

			So(rp.WaitUntilGranted(r2), ShouldBeFalse)
			r3, err := rp.Request(1)
			So(string(r3), ShouldBeBlank)
			So(err, ShouldNotBeNil)
			rperr, ok := err.(Error)
			So(ok, ShouldBeTrue)
			So(rperr.Err, ShouldEqual, ErrShutDown)
		})

		Convey("WaitUntilGranted can time out and cancel the request", func() {
			r, err := rp.Request(maxSimultaneous)
			So(err, ShouldBeNil)
			So(rp.WaitUntilGranted(r), ShouldBeTrue)

			r2, err := rp.Request(1)
			So(err, ShouldBeNil)
			So(rp.WaitUntilGranted(r2, oneFiftyPercentDelay), ShouldBeFalse)
			So(time.Now(), ShouldHappenOnOrBetween, begin.Add(oneFiftyPercentDelay), begin.Add(doubleDelay))

			So(rp.WaitUntilGranted(r), ShouldBeTrue)
			rp.Release(r)
			So(rp.WaitUntilGranted(r), ShouldBeFalse)
			So(rp.WaitUntilGranted(r2), ShouldBeFalse)
		})

		Convey("You can request the maximum tokens in a single request", func() {
			r, err := rp.Request(maxSimultaneous)
			So(err, ShouldBeNil)

			So(rp.WaitUntilGranted(r), ShouldBeTrue)
			So(time.Now(), ShouldHappenBefore, begin.Add(halfDelay))

			r2, err := rp.Request(1)
			So(err, ShouldBeNil)
			So(r2, ShouldNotBeNil)
			So(time.Now(), ShouldHappenBefore, begin.Add(halfDelay))

			Convey("Subsequent requests must wait until it is released", func() {
				go func() {
					<-time.After(oneFiftyPercentDelay)
					rp.Release(r)
				}()

				So(rp.WaitUntilGranted(r2), ShouldBeTrue)
				So(time.Now(), ShouldHappenOnOrBetween, begin.Add(oneFiftyPercentDelay), begin.Add(doubleDelay))
				rp.Release(r2)
			})

			Convey("Or until it times out", func() {
				So(rp.WaitUntilGranted(r2), ShouldBeTrue)
				So(time.Now(), ShouldHappenOnOrBetween, begin.Add(releaseTimeout), begin.Add(releaseTimeout).Add(halfDelay))
				rp.Release(r2)
			})

			Convey("Touch() delays the time out", func() {
				go func() {
					<-time.After(oneFiftyPercentDelay)
					rp.Touch(r)
				}()

				So(rp.WaitUntilGranted(r2), ShouldBeTrue)
				So(time.Now(), ShouldHappenOnOrBetween, begin.Add(releaseTimeout).Add(oneFiftyPercentDelay), begin.Add(releaseTimeout).Add(doubleDelay))
				rp.Release(r2)
			})
		})

		Convey("You can Touch multiple requests at once to delay all their timeouts", func() {
			r, err := rp.Request(1)
			So(err, ShouldBeNil)
			So(rp.WaitUntilGranted(r), ShouldBeTrue)

			r2, err := rp.Request(1)
			So(err, ShouldBeNil)
			So(rp.WaitUntilGranted(r2), ShouldBeTrue)

			go func() {
				<-time.After(oneFiftyPercentDelay)
				rp.Touch(r, r2)
			}()

			granted, keepChecking := rp.Granted(r)
			So(granted, ShouldBeTrue)
			So(keepChecking, ShouldBeFalse)
			granted, keepChecking = rp.Granted(r2)
			So(granted, ShouldBeTrue)
			So(keepChecking, ShouldBeFalse)

			<-time.After(releaseTimeout)

			granted, keepChecking = rp.Granted(r)
			So(granted, ShouldBeTrue)
			So(keepChecking, ShouldBeFalse)
			granted, keepChecking = rp.Granted(r2)
			So(granted, ShouldBeTrue)
			So(keepChecking, ShouldBeFalse)

			<-time.After(oneFiftyPercentDelay)
			<-time.After(halfDelay)

			granted, keepChecking = rp.Granted(r)
			So(granted, ShouldBeFalse)
			So(keepChecking, ShouldBeFalse)
			granted, keepChecking = rp.Granted(r2)
			So(granted, ShouldBeFalse)
			So(keepChecking, ShouldBeFalse)
		})

		Convey("You can release after a delay", func() {
			r, err := rp.Request(maxSimultaneous)
			So(err, ShouldBeNil)

			So(rp.WaitUntilGranted(r), ShouldBeTrue)
			So(time.Now(), ShouldHappenBefore, begin.Add(halfDelay))
			rp.ReleaseAfter(r, oneFiftyPercentDelay)

			r2, err := rp.Request(1)
			So(err, ShouldBeNil)
			So(r2, ShouldNotBeNil)
			So(time.Now(), ShouldHappenBefore, begin.Add(halfDelay))

			So(rp.WaitUntilGranted(r2), ShouldBeTrue)
			So(time.Now(), ShouldHappenOnOrBetween, begin.Add(oneFiftyPercentDelay), begin.Add(oneFiftyPercentDelay).Add(halfDelay))
			rp.Release(r2)

			Convey("Once released, the Request methods do nothing", func() {
				rp.Release(r2)
				rp.Touch(r2)
				So(rp.WaitUntilGranted(r2), ShouldBeFalse)
				So(time.Now(), ShouldHappenOnOrBetween, begin.Add(oneFiftyPercentDelay), begin.Add(oneFiftyPercentDelay).Add(halfDelay))
			})
		})

		Convey("Period use of Granted() is an alternative to WaitUntilGranted()", func() {
			r, err := rp.Request(maxSimultaneous)
			So(err, ShouldBeNil)

			So(rp.WaitUntilGranted(r), ShouldBeTrue)
			So(time.Now(), ShouldHappenBefore, begin.Add(halfDelay))
			rp.ReleaseAfter(r, oneFiftyPercentDelay)

			r2, err := rp.Request(1)
			So(err, ShouldBeNil)
			So(r2, ShouldNotBeNil)
			So(time.Now(), ShouldHappenBefore, begin.Add(halfDelay))

			granted, keepChecking := rp.Granted(r2)
			So(granted, ShouldBeFalse)
			So(keepChecking, ShouldBeTrue)

			<-time.After(halfDelay)
			granted, keepChecking = rp.Granted(r2)
			So(granted, ShouldBeFalse)
			So(keepChecking, ShouldBeTrue)

			<-time.After(oneFiftyPercentDelay)
			granted, keepChecking = rp.Granted(r2)
			So(granted, ShouldBeTrue)
			So(keepChecking, ShouldBeFalse)

			rp.Release(r2)
			granted, keepChecking = rp.Granted(r2)
			So(granted, ShouldBeFalse)
			So(keepChecking, ShouldBeFalse)
		})

		Convey("Releasing Request()s in less than delay time lets you request continuously", func() {
			grantedCh := make(chan time.Time, maxSimultaneous)
			for i := 1; i <= maxSimultaneous*3; i++ {
				r, err := rp.Request(1)
				So(err, ShouldBeNil)

				go func(r Receipt) {
					rp.WaitUntilGranted(r)
					grantedCh <- time.Now()
					<-time.After(halfDelay)
					rp.Release(r)
				}(r)
			}

			So(time.Now(), ShouldHappenBefore, begin.Add(halfDelay))

			for i := 0; i < maxSimultaneous*3; i++ {
				So(<-grantedCh, ShouldHappenOnOrBetween, begin.Add(time.Duration(delayInt*i)*time.Millisecond), begin.Add(time.Duration(delayInt*i)*time.Millisecond).Add(halfDelay))
			}
		})

		Convey("Releasing Request()s immediately with no delay time lets you request continuously with no delay", func() {
			rp = New("irods", 0*time.Second, maxSimultaneous, releaseTimeout)
			So(rp, ShouldNotBeNil)

			grantedCh := make(chan time.Time, maxSimultaneous)
			for i := 1; i <= maxSimultaneous*3; i++ {
				r, err := rp.Request(1)
				So(err, ShouldBeNil)

				go func(r Receipt) {
					rp.WaitUntilGranted(r)
					grantedCh <- time.Now()
					rp.Release(r)
				}(r)
			}

			So(time.Now(), ShouldHappenBefore, begin.Add(halfDelay))

			for i := 0; i < maxSimultaneous*3; i++ {
				So(<-grantedCh, ShouldHappenBefore, begin.Add(halfDelay))
			}
		})

		Convey("AvailabilityCallbacks are obeyed", func() {
			cbCalls := 0
			tooBusyFor := 2
			cb := func() int {
				cbCalls++
				if cbCalls <= tooBusyFor {
					return maxSimultaneous - 1
				}
				return maxSimultaneous
			}
			rp.SetAvailabilityCallback(cb)

			r, err := rp.Request(maxSimultaneous)
			So(err, ShouldBeNil)

			So(rp.WaitUntilGranted(r), ShouldBeTrue)
			So(time.Now(), ShouldHappenOnOrBetween, begin.Add(time.Duration(delayInt*tooBusyFor)*time.Millisecond), begin.Add(time.Duration(delayInt*tooBusyFor)*time.Millisecond).Add(halfDelay))
		})
	})
}
