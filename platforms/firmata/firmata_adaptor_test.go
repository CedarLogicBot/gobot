package firmata

import (
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/hybridgroup/gobot"
)

var connect = func(a *FirmataAdaptor) []error {
	defaultInitTimeInterval = 0 * time.Second
	gobot.After(1*time.Millisecond, func() {
		// arduino uno r3 firmware response "StandardFirmata.ino"
		a.board.process([]byte{240, 121, 2, 3, 83, 0, 116, 0, 97, 0, 110, 0, 100,
			0, 97, 0, 114, 0, 100, 0, 70, 0, 105, 0, 114, 0, 109, 0, 97, 0, 116, 0,
			97, 0, 46, 0, 105, 0, 110, 0, 111, 0, 247})
		// arduino uno r3 capabilities response
		a.board.process([]byte{240, 108, 127, 127, 0, 1, 1, 1, 4, 14, 127, 0, 1,
			1, 1, 3, 8, 4, 14, 127, 0, 1, 1, 1, 4, 14, 127, 0, 1, 1, 1, 3, 8, 4, 14,
			127, 0, 1, 1, 1, 3, 8, 4, 14, 127, 0, 1, 1, 1, 4, 14, 127, 0, 1, 1, 1,
			4, 14, 127, 0, 1, 1, 1, 3, 8, 4, 14, 127, 0, 1, 1, 1, 3, 8, 4, 14, 127,
			0, 1, 1, 1, 3, 8, 4, 14, 127, 0, 1, 1, 1, 4, 14, 127, 0, 1, 1, 1, 4, 14,
			127, 0, 1, 1, 1, 2, 10, 127, 0, 1, 1, 1, 2, 10, 127, 0, 1, 1, 1, 2, 10,
			127, 0, 1, 1, 1, 2, 10, 127, 0, 1, 1, 1, 2, 10, 6, 1, 127, 0, 1, 1, 1,
			2, 10, 6, 1, 127, 247})
		// arduino uno r3 analog mapping response
		a.board.process([]byte{240, 106, 127, 127, 127, 127, 127, 127, 127, 127,
			127, 127, 127, 127, 127, 127, 0, 1, 2, 3, 4, 5, 247})
	})
	return a.Connect()
}

func initTestFirmataAdaptor() *FirmataAdaptor {
	a := NewFirmataAdaptor("board", "/dev/null")
	a.connect = func(port string) (io.ReadWriteCloser, error) {
		return &NullReadWriteCloser{}, nil
	}
	connect(a)
	return a
}
func TestFirmataAdaptor(t *testing.T) {
	a := initTestFirmataAdaptor()
	gobot.Assert(t, a.Name(), "board")
	gobot.Assert(t, a.Port(), "/dev/null")
}

func TestFirmataAdaptorFinalize(t *testing.T) {
	a := initTestFirmataAdaptor()
	gobot.Assert(t, len(a.Finalize()), 0)

	closeErr = errors.New("close error")
	a = initTestFirmataAdaptor()
	gobot.Assert(t, a.Finalize()[0], errors.New("close error"))
}

func TestFirmataAdaptorDisconnect(t *testing.T) {
	a := NewFirmataAdaptor("board", "/dev/null")
	gobot.Assert(t, a.Disconnect(), errors.New("no board connected"))
}

func TestFirmataAdaptorConnect(t *testing.T) {
	a := NewFirmataAdaptor("board", "/dev/null")
	a.connect = func(port string) (io.ReadWriteCloser, error) {
		return &NullReadWriteCloser{}, nil
	}
	gobot.Assert(t, len(connect(a)), 0)

	a = NewFirmataAdaptor("board", "/dev/null")
	a.connect = func(port string) (io.ReadWriteCloser, error) {
		return nil, errors.New("connect error")
	}
	gobot.Assert(t, a.Connect()[0], errors.New("connect error"))

	a = NewFirmataAdaptor("board", &NullReadWriteCloser{})
	gobot.Assert(t, len(connect(a)), 0)
}

func TestFirmataAdaptorServoWrite(t *testing.T) {
	a := initTestFirmataAdaptor()
	a.ServoWrite("1", 50)
}

func TestFirmataAdaptorPwmWrite(t *testing.T) {
	a := initTestFirmataAdaptor()
	a.PwmWrite("1", 50)
}

func TestFirmataAdaptorDigitalWrite(t *testing.T) {
	a := initTestFirmataAdaptor()
	a.DigitalWrite("1", 1)
}

func TestFirmataAdaptorDigitalRead(t *testing.T) {
	a := initTestFirmataAdaptor()
	pinNumber := "1"
	// -1 on no data
	val, _ := a.DigitalRead(pinNumber)
	gobot.Assert(t, val, -1)

	go func() {
		<-time.After(5 * time.Millisecond)
		gobot.Publish(a.board.events[fmt.Sprintf("digital_read_%v", pinNumber)],
			[]byte{0x01})
	}()
	val, _ = a.DigitalRead(pinNumber)
	gobot.Assert(t, val, 0x01)
}

func TestFirmataAdaptorAnalogRead(t *testing.T) {
	a := initTestFirmataAdaptor()
	pinNumber := "1"
	// -1 on no data
	val, _ := a.AnalogRead(pinNumber)
	gobot.Assert(t, val, -1)

	value := 133
	go func() {
		<-time.After(5 * time.Millisecond)
		gobot.Publish(a.board.events[fmt.Sprintf("analog_read_%v", pinNumber)],
			[]byte{
				byte(value >> 24),
				byte(value >> 16),
				byte(value >> 8),
				byte(value & 0xff),
			},
		)
	}()
	val, _ = a.AnalogRead(pinNumber)
	gobot.Assert(t, val, 133)
}

func TestFirmataAdaptorI2cStart(t *testing.T) {
	a := initTestFirmataAdaptor()
	a.I2cStart(0x00)
}
func TestFirmataAdaptorI2cRead(t *testing.T) {
	a := initTestFirmataAdaptor()
	// [] on no data
	data, _ := a.I2cRead(1)
	gobot.Assert(t, data, []byte{})

	i := []byte{100}
	i2cReply := map[string][]byte{}
	i2cReply["data"] = i
	go func() {
		<-time.After(5 * time.Millisecond)
		gobot.Publish(a.board.events["i2c_reply"], i2cReply)
	}()
	data, _ = a.I2cRead(1)
	gobot.Assert(t, data, i)
}
func TestFirmataAdaptorI2cWrite(t *testing.T) {
	a := initTestFirmataAdaptor()
	a.I2cWrite([]byte{0x00, 0x01})
}
