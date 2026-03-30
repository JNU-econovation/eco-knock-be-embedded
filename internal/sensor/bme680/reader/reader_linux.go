//go:build linux

package reader

import (
	"eco-knock-be-embedded/internal/sensor/dto"
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	bme680config "eco-knock-be-embedded/internal/sensor/bme680/config"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/host/v3/sysfs"
)

const (
	regChipID    = 0xD0
	regVariantID = 0xF0
	regSoftReset = 0xE0

	regCoeff1 = 0x8A
	regCoeff2 = 0xE1
	regCoeff3 = 0x00

	regCtrlGas0 = 0x70
	regCtrlGas1 = 0x71
	regCtrlHum  = 0x72
	regCtrlMeas = 0x74
	regConfig   = 0x75

	regField0   = 0x1D
	regResHeat0 = 0x5A
	regGasWait0 = 0x64

	softResetCommand = 0xB6
	chipID           = 0x61

	variantGasLow  = 0x00
	variantGasHigh = 0x01

	os1x  = 1
	os2x  = 2
	os16x = 5

	filterOff = 0

	odrNone = 8

	sleepMode  = 0
	forcedMode = 1

	enableGasMeasureLow  = 0x01
	enableGasMeasureHigh = 0x02

	filterMask = 0x1C
	filterPos  = 2

	odr3Mask = 0x80
	odr3Pos  = 7

	odr20Mask = 0xE0
	odr20Pos  = 5

	ostMask = 0xE0
	ostPos  = 5

	ospMask = 0x1C
	ospPos  = 2

	oshMask = 0x07

	hctrlMask = 0x08
	hctrlPos  = 3

	runGasMask = 0x30
	runGasPos  = 4

	nbConvMask = 0x0F

	modeMask = 0x03

	newDataMask   = 0x80
	gasIndexMask  = 0x0F
	gasRangeMask  = 0x0F
	gasValidMask  = 0x20
	heatStabMask  = 0x10
	resHeatMask   = 0x30
	rangeErrMask  = 0xF0
	h1DataMask    = 0x0F
	fieldLength   = 17
	fieldRetries  = 5
	pollDelayUs   = 10_000
	resetDelayUs  = 10_000
	maxHeaterTemp = 400
)

var ErrSensorClosed = errors.New("bme680 sensor is closed")

type Sensor struct {
	mu      sync.Mutex
	bus     i2c.BusCloser
	dev     *i2c.Dev
	config  bme680config.Config
	calib   calibrationData
	variant byte
	closed  bool
}

type calibrationData struct {
	parH1 uint16
	parH2 uint16
	parH3 int8
	parH4 int8
	parH5 int8
	parH6 uint8
	parH7 int8

	parGH1 int8
	parGH2 int16
	parGH3 int8

	parT1 uint16
	parT2 int16
	parT3 int8

	parP1  uint16
	parP2  int16
	parP3  int8
	parP4  int16
	parP5  int16
	parP6  int8
	parP7  int8
	parP8  int16
	parP9  int16
	parP10 uint8

	tFine float64

	resHeatRange uint8
	resHeatVal   int8
	rangeSwErr   int8
}

func Open(cfg bme680config.Config) (*Sensor, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	busNumber, err := parseBusNumber(cfg.I2CDevice)
	if err != nil {
		return nil, err
	}

	bus, err := sysfs.NewI2C(busNumber)
	if err != nil {
		return nil, fmt.Errorf("open i2c bus %d: %w", busNumber, err)
	}

	sensor := &Sensor{
		bus:    bus,
		dev:    &i2c.Dev{Bus: bus, Addr: uint16(cfg.I2CAddress)},
		config: cfg,
	}

	if err := sensor.init(); err != nil {
		_ = bus.Close()
		return nil, err
	}

	return sensor, nil
}

func (sensor *Sensor) Read() (dto.SampleDTO, error) {
	sensor.mu.Lock()
	defer sensor.mu.Unlock()

	if sensor.closed {
		return dto.SampleDTO{}, ErrSensorClosed
	}

	if err := sensor.setOpMode(forcedMode); err != nil {
		return dto.SampleDTO{}, err
	}

	time.Sleep(sensor.measurementDelay())

	sample, err := sensor.readFieldData()
	if err != nil {
		return dto.SampleDTO{}, err
	}

	sample.MeasuredAt = time.Now()
	return sample, nil
}

func (sensor *Sensor) Close() error {
	sensor.mu.Lock()
	defer sensor.mu.Unlock()

	if sensor.closed {
		return nil
	}

	sensor.closed = true
	return sensor.bus.Close()
}

func (sensor *Sensor) init() error {
	if err := sensor.softReset(); err != nil {
		return err
	}

	id, err := sensor.readReg(regChipID)
	if err != nil {
		return fmt.Errorf("read chip id: %w", err)
	}
	if id != chipID {
		return fmt.Errorf("unexpected chip id 0x%02x", id)
	}

	variant, err := sensor.readReg(regVariantID)
	if err != nil {
		return fmt.Errorf("read variant id: %w", err)
	}
	sensor.variant = variant

	if err := sensor.loadCalibrationData(); err != nil {
		return err
	}

	if err := sensor.applySensorConfig(); err != nil {
		return err
	}

	if err := sensor.applyHeaterConfig(); err != nil {
		return err
	}

	return nil
}

func (sensor *Sensor) softReset() error {
	if err := sensor.writeReg(regSoftReset, softResetCommand); err != nil {
		return fmt.Errorf("soft reset: %w", err)
	}

	time.Sleep(resetDelayUs * time.Microsecond)
	return nil
}

func (sensor *Sensor) loadCalibrationData() error {
	coeff := make([]byte, 42)

	if err := sensor.readRegs(regCoeff1, coeff[:23]); err != nil {
		return fmt.Errorf("read coeff block 1: %w", err)
	}
	if err := sensor.readRegs(regCoeff2, coeff[23:37]); err != nil {
		return fmt.Errorf("read coeff block 2: %w", err)
	}
	if err := sensor.readRegs(regCoeff3, coeff[37:42]); err != nil {
		return fmt.Errorf("read coeff block 3: %w", err)
	}

	sensor.calib.parT1 = concatBytes(coeff[32], coeff[31])
	sensor.calib.parT2 = int16(concatBytes(coeff[1], coeff[0]))
	sensor.calib.parT3 = int8(coeff[2])

	sensor.calib.parP1 = concatBytes(coeff[5], coeff[4])
	sensor.calib.parP2 = int16(concatBytes(coeff[7], coeff[6]))
	sensor.calib.parP3 = int8(coeff[8])
	sensor.calib.parP4 = int16(concatBytes(coeff[11], coeff[10]))
	sensor.calib.parP5 = int16(concatBytes(coeff[13], coeff[12]))
	sensor.calib.parP6 = int8(coeff[15])
	sensor.calib.parP7 = int8(coeff[14])
	sensor.calib.parP8 = int16(concatBytes(coeff[19], coeff[18]))
	sensor.calib.parP9 = int16(concatBytes(coeff[21], coeff[20]))
	sensor.calib.parP10 = coeff[22]

	sensor.calib.parH1 = (uint16(coeff[25]) << 4) | uint16(coeff[24]&h1DataMask)
	sensor.calib.parH2 = (uint16(coeff[23]) << 4) | uint16(coeff[24]>>4)
	sensor.calib.parH3 = int8(coeff[26])
	sensor.calib.parH4 = int8(coeff[27])
	sensor.calib.parH5 = int8(coeff[28])
	sensor.calib.parH6 = coeff[29]
	sensor.calib.parH7 = int8(coeff[30])

	sensor.calib.parGH1 = int8(coeff[35])
	sensor.calib.parGH2 = int16(concatBytes(coeff[34], coeff[33]))
	sensor.calib.parGH3 = int8(coeff[36])

	sensor.calib.resHeatRange = (coeff[39] & resHeatMask) / 16
	sensor.calib.resHeatVal = int8(coeff[37])
	sensor.calib.rangeSwErr = int8(coeff[41]&rangeErrMask) / 16

	return nil
}

func (sensor *Sensor) applySensorConfig() error {
	if err := sensor.setOpMode(sleepMode); err != nil {
		return err
	}

	if err := sensor.writeReg(regCtrlHum, setBitsPos0(0, oshMask, os16x)); err != nil {
		return fmt.Errorf("write ctrl_hum: %w", err)
	}

	ctrlMeas, err := sensor.readReg(regCtrlMeas)
	if err != nil {
		return fmt.Errorf("read ctrl_meas: %w", err)
	}
	ctrlMeas = setBits(ctrlMeas, ostMask, ostPos, os2x)
	ctrlMeas = setBits(ctrlMeas, ospMask, ospPos, os1x)
	ctrlMeas &= ^byte(modeMask)
	if err := sensor.writeReg(regCtrlMeas, ctrlMeas); err != nil {
		return fmt.Errorf("write ctrl_meas: %w", err)
	}

	config, err := sensor.readReg(regConfig)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	config = setBits(config, filterMask, filterPos, filterOff)
	config = setBits(config, odr20Mask, odr20Pos, 0)
	config = setBits(config, odr3Mask, odr3Pos, 1)
	if err := sensor.writeReg(regConfig, config); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

func (sensor *Sensor) applyHeaterConfig() error {
	if err := sensor.setOpMode(sleepMode); err != nil {
		return err
	}

	resHeat := sensor.calcResHeat(sensor.config.HeaterTempC)
	if err := sensor.writeReg(regResHeat0, resHeat); err != nil {
		return fmt.Errorf("write res_heat0: %w", err)
	}

	gasWait := calcGasWait(uint16(sensor.config.HeaterDuration / time.Millisecond))
	if err := sensor.writeReg(regGasWait0, gasWait); err != nil {
		return fmt.Errorf("write gas_wait0: %w", err)
	}

	ctrlGas0, err := sensor.readReg(regCtrlGas0)
	if err != nil {
		return fmt.Errorf("read ctrl_gas_0: %w", err)
	}
	ctrlGas0 = setBits(ctrlGas0, hctrlMask, hctrlPos, 0)
	if err := sensor.writeReg(regCtrlGas0, ctrlGas0); err != nil {
		return fmt.Errorf("write ctrl_gas_0: %w", err)
	}

	ctrlGas1, err := sensor.readReg(regCtrlGas1)
	if err != nil {
		return fmt.Errorf("read ctrl_gas_1: %w", err)
	}
	ctrlGas1 = setBitsPos0(ctrlGas1, nbConvMask, 0)
	ctrlGas1 = setBits(ctrlGas1, runGasMask, runGasPos, sensor.runGasValue())
	if err := sensor.writeReg(regCtrlGas1, ctrlGas1); err != nil {
		return fmt.Errorf("write ctrl_gas_1: %w", err)
	}

	return nil
}

func (sensor *Sensor) setOpMode(mode byte) error {
	for {
		ctrlMeas, err := sensor.readReg(regCtrlMeas)
		if err != nil {
			return fmt.Errorf("read ctrl_meas: %w", err)
		}

		current := ctrlMeas & modeMask
		if current == sleepMode {
			if mode == sleepMode {
				return nil
			}

			ctrlMeas = (ctrlMeas & ^byte(modeMask)) | (mode & modeMask)
			if err := sensor.writeReg(regCtrlMeas, ctrlMeas); err != nil {
				return fmt.Errorf("write ctrl_meas: %w", err)
			}
			return nil
		}

		ctrlMeas &= ^byte(modeMask)
		if err := sensor.writeReg(regCtrlMeas, ctrlMeas); err != nil {
			return fmt.Errorf("write ctrl_meas sleep: %w", err)
		}

		time.Sleep(pollDelayUs * time.Microsecond)
	}
}

func (sensor *Sensor) measurementDelay() time.Duration {
	osToMeasCycles := [...]uint32{0, 1, 2, 4, 8, 16}
	measCycles := osToMeasCycles[os2x] + osToMeasCycles[os1x] + osToMeasCycles[os16x]
	durationUs := measCycles*1963 + 477*4 + 477*5 + 1000
	return time.Duration(durationUs)*time.Microsecond + sensor.config.HeaterDuration
}

func (sensor *Sensor) readFieldData() (dto.SampleDTO, error) {
	var field [fieldLength]byte

	for i := 0; i < fieldRetries; i++ {
		if err := sensor.readRegs(regField0, field[:]); err != nil {
			return dto.SampleDTO{}, fmt.Errorf("read field data: %w", err)
		}

		status := field[0] & newDataMask
		gasIndex := field[0] & gasIndexMask
		_ = gasIndex

		adcTemp := (uint32(field[5]) * 4096) | (uint32(field[6]) * 16) | (uint32(field[7]) / 16)
		adcHum := (uint16(field[8]) * 256) | uint16(field[9])
		adcGasLow := (uint16(field[13]) * 4) | (uint16(field[14]) / 64)
		adcGasHigh := (uint16(field[15]) * 4) | (uint16(field[16]) / 64)
		gasRangeLow := field[14] & gasRangeMask
		gasRangeHigh := field[16] & gasRangeMask

		if sensor.variant == variantGasHigh {
			status |= field[16] & gasValidMask
			status |= field[16] & heatStabMask
		} else {
			status |= field[14] & gasValidMask
			status |= field[14] & heatStabMask
		}

		if status&newDataMask == 0 {
			time.Sleep(pollDelayUs * time.Microsecond)
			continue
		}

		temperature := sensor.calcTemperature(adcTemp)
		humidity := sensor.calcHumidity(adcHum)

		var gasResistance float64
		if sensor.variant == variantGasHigh {
			gasResistance = calcGasResistanceHigh(adcGasHigh, gasRangeHigh)
		} else {
			gasResistance = sensor.calcGasResistanceLow(adcGasLow, gasRangeLow)
		}

		return dto.SampleDTO{
			TemperatureC:     temperature,
			HumidityRH:       humidity,
			GasResistanceOhm: gasResistance,
			Status:           status,
			GasValid:         status&gasValidMask != 0,
			HeatStable:       status&heatStabMask != 0,
		}, nil
	}

	return dto.SampleDTO{}, errors.New("bme680 returned no new data")
}

func (sensor *Sensor) calcTemperature(adc uint32) float64 {
	var1 := ((float64(adc) / 16384.0) - (float64(sensor.calib.parT1) / 1024.0)) * float64(sensor.calib.parT2)
	var2 := ((float64(adc)/131072.0 - float64(sensor.calib.parT1)/8192.0) *
		(float64(adc)/131072.0 - float64(sensor.calib.parT1)/8192.0)) * (float64(sensor.calib.parT3) * 16.0)
	sensor.calib.tFine = var1 + var2
	return sensor.calib.tFine / 5120.0
}

func (sensor *Sensor) calcHumidity(adc uint16) float64 {
	tempComp := sensor.calib.tFine / 5120.0
	var1 := float64(adc) - ((float64(sensor.calib.parH1) * 16.0) + (float64(sensor.calib.parH3)/2.0)*tempComp)
	var2 := var1 * ((float64(sensor.calib.parH2) / 262144.0) *
		(1.0 + (float64(sensor.calib.parH4)/16384.0)*tempComp + (float64(sensor.calib.parH5)/1048576.0)*tempComp*tempComp))
	var3 := float64(sensor.calib.parH6) / 16384.0
	var4 := float64(sensor.calib.parH7) / 2097152.0
	humidity := var2 + ((var3 + var4*tempComp) * var2 * var2)

	switch {
	case humidity > 100:
		return 100
	case humidity < 0:
		return 0
	default:
		return humidity
	}
}

func (sensor *Sensor) calcGasResistanceLow(adc uint16, gasRange byte) float64 {
	lookupK1 := [...]float64{0, 0, 0, 0, 0, -1, 0, -0.8, 0, 0, -0.2, -0.5, 0, -1, 0, 0}
	lookupK2 := [...]float64{0, 0, 0, 0, 0.1, 0.7, 0, -0.8, -0.1, 0, 0, 0, 0, 0, 0, 0}

	var1 := 1340.0 + 5.0*float64(sensor.calib.rangeSwErr)
	var2 := var1 * (1.0 + lookupK1[gasRange]/100.0)
	var3 := 1.0 + lookupK2[gasRange]/100.0
	gasRangeFactor := math.Exp2(float64(gasRange))

	return 1.0 / (var3 * 0.000000125 * gasRangeFactor * (((float64(adc) - 512.0) / var2) + 1.0))
}

func calcGasResistanceHigh(adc uint16, gasRange byte) float64 {
	var1 := float64(uint32(262144) >> gasRange)
	var2 := float64((int32(adc)-512)*3 + 4096)
	return 1000000.0 * var1 / var2
}

func (sensor *Sensor) calcResHeat(temp uint16) byte {
	if temp > maxHeaterTemp {
		temp = maxHeaterTemp
	}

	var1 := float64(sensor.calib.parGH1)/16.0 + 49.0
	var2 := float64(sensor.calib.parGH2)/32768.0*0.0005 + 0.00235
	var3 := float64(sensor.calib.parGH3) / 1024.0
	var4 := var1 * (1.0 + var2*float64(temp))
	var5 := var4 + var3*float64(sensor.config.AmbientTempC)

	resHeat := 3.4 * ((var5 * (4.0 / (4.0 + float64(sensor.calib.resHeatRange))) * (1.0 / (1.0 + float64(sensor.calib.resHeatVal)*0.002))) - 25.0)
	if resHeat < 0 {
		return 0
	}
	if resHeat > math.MaxUint8 {
		return math.MaxUint8
	}

	return byte(resHeat)
}

func (sensor *Sensor) runGasValue() byte {
	if sensor.variant == variantGasHigh {
		return enableGasMeasureHigh
	}
	return enableGasMeasureLow
}

func (sensor *Sensor) readReg(reg byte) (byte, error) {
	var data [1]byte
	if err := sensor.readRegs(reg, data[:]); err != nil {
		return 0, err
	}
	return data[0], nil
}

func (sensor *Sensor) readRegs(reg byte, out []byte) error {
	return sensor.dev.Tx([]byte{reg}, out)
}

func (sensor *Sensor) writeReg(reg, value byte) error {
	return sensor.dev.Tx([]byte{reg, value}, nil)
}

func parseBusNumber(ref string) (int, error) {
	base := strings.TrimSpace(ref)
	if base == "" {
		return 0, errors.New("I2C device not configured")
	}

	base = strings.ToLower(filepath.Base(base))
	switch {
	case strings.HasPrefix(base, "i2c-"):
		base = strings.TrimPrefix(base, "i2c-")
	case strings.HasPrefix(base, "i2c"):
		base = strings.TrimPrefix(base, "i2c")
	}

	number, err := strconv.Atoi(base)
	if err != nil {
		return 0, fmt.Errorf("invalid I2C bus reference %q", ref)
	}

	return number, nil
}

func calcGasWait(durationMs uint16) byte {
	if durationMs >= 0xFC0 {
		return 0xFF
	}

	factor := uint16(0)
	for durationMs > 0x3F {
		durationMs /= 4
		factor++
	}

	return byte(durationMs + factor*64)
}

func concatBytes(msb, lsb byte) uint16 {
	return (uint16(msb) << 8) | uint16(lsb)
}

func setBits(reg, mask byte, pos uint8, value byte) byte {
	return (reg &^ mask) | ((value << pos) & mask)
}

func setBitsPos0(reg, mask, value byte) byte {
	return (reg &^ mask) | (value & mask)
}
