package pmbus

import (
	"fmt"
	"math"

	"github.com/karalabe/hid"
	"go.uber.org/zap"
)

const (
	pmbusClearFaults      = 0x03
	pmbusPage             = 0x00
	pmbusReadVin          = 0x88
	pmbusReadVOut         = 0x8b
	pmbusReadIOut         = 0x8c
	pmbusReadTemperature1 = 0x8d
	pmbusReadTemperature2 = 0x8e
	pmbusReadFanSpeed1    = 0x90
	pmbusReadPOut         = 0x96
	pmbusMfrSpecific30    = 0xee
)

type pmbusDevice struct {
	dev      *hid.Device
	logger   *zap.Logger
	lastPage int
}

func newPmbusDevice(logger *zap.Logger, vid, pid uint16) *pmbusDevice {
	dis := hid.Enumerate(vid, pid)
	if len(dis) != 1 {
		return nil
	}
	di := dis[0]
	dev, err := di.Open()
	if err != nil {
		logger.Error("open", zap.Error(err))
		return nil
	}
	//logger.Info("found device", zap.String("path", di.Path))
	return &pmbusDevice{
		dev:      dev,
		logger:   logger,
		lastPage: -1,
	}
}

func (pm *pmbusDevice) write(b []byte) {
	send := make([]byte, 64)
	copy(send, b)
	n, err := pm.dev.Write(send)
	if err != nil {
		pm.logger.Error("write", zap.Error(err))
		return
	}
	if n != 64 {
		pm.logger.Error("write", zap.Error(fmt.Errorf("only wrote %d bytes", n)))
		return
	}
}

func (pm *pmbusDevice) read() []byte {
	buf := make([]byte, 64)
	n, err := pm.dev.Read(buf)
	if err != nil {
		pm.logger.Error("read", zap.Error(err))
	}
	return buf[:n]
}

func (pm *pmbusDevice) selectPage(page byte) bool {
	if pm.lastPage == int(page) {
		return true
	}
	buffer := []byte{2, pmbusPage, page}
	pm.write(buffer)
	buffer = pm.read()
	if buffer[0] != 2 || buffer[2] != page {
		pm.logger.Error("switch page failed", zap.Int("page", int(page)), zap.ByteString("result", trimZero(buffer)))
		return false
	}
	pm.lastPage = int(page)
	return true
}

func (pm *pmbusDevice) execWriteAddress(address byte, command byte, data ...byte) []byte {
	buffer := append([]byte{address, command}, data...)
	pm.write(buffer)
	return pm.read()
}

func (pm *pmbusDevice) execWrite(command byte, data ...byte) []byte {
	return pm.execWriteAddress(2, command, data...)
}

func (pm *pmbusDevice) execWriteToPage(page byte, command byte, data ...byte) []byte {
	if !pm.selectPage(page) {
		return nil
	}
	return pm.execWrite(command, data...)
}

func (pm *pmbusDevice) execReadFromPage(page byte, command byte, data ...byte) []byte {
	if !pm.selectPage(page) {
		return nil
	}
	return pm.execRead(command, data...)

}

func (pm *pmbusDevice) execRead(command byte, data ...byte) []byte {
	buffer := append([]byte{3, command}, data...)
	pm.write(buffer)
	return pm.read()
}

func (pm *pmbusDevice) close() {
	_ = pm.dev.Close()
}

func linearToFloat64(b []byte) float64 {
	return float64((int32(b[1])<<29)>>21|int32(b[0])) * math.Pow(2, float64(int8(b[1])>>3))
}

func trimZero(b []byte) []byte {
	for i := len(b) - 1; i > 0; i-- {
		if b[i] == 0 {
			b = b[:i]
		} else {
			break
		}
	}
	return b
}
