package asn

import (
	"fmt"
	"path"
	"reflect"
	"runtime"

	"github.com/free5gc/aper/logger"
)

type berByteData struct {
	bytes      []byte
	byteOffset uint64
	bitsOffset uint
}

func berTrace(level int, s string) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		logger.AperLog.Debugln(s)
	} else {
		logger.AperLog.Debugf("%s (%s:%d)\n", s, path.Base(file), line)
	}
}

func berByteLog(numBits uint64, byteOffset uint64, bitsOffset uint, value interface{}) string {
	if reflect.TypeOf(value).Kind() == reflect.Uint64 {
		return fmt.Sprintf("  [BER got %2d bits, byteOffset(after): %d, bitsOffset(after): %d, value: 0x%0x]",
			numBits, byteOffset, bitsOffset, reflect.ValueOf(value).Uint())
	}
	return fmt.Sprintf("  [BER got %2d bits, byteOffset(after): %d, bitsOffset(after): %d, value: 0x%0x]",
		numBits, byteOffset, bitsOffset, reflect.ValueOf(value).Bytes())
}

// BerGetBitString is to get BitString with desire size from source byte array with bit offset
func BerGetBitString(srcBytes []byte, bitsOffset uint, numBits uint) (dstBytes []byte, err error) {
	bitsLeft := uint(len(srcBytes))*8 - bitsOffset
	if numBits > bitsLeft {
		err = fmt.Errorf("Get bits overflow, requireBits: %d, leftBits: %d", numBits, bitsLeft)
		return
	}
	byteLen := (bitsOffset + numBits + 7) >> 3
	numBitsByteLen := (numBits + 7) >> 3
	dstBytes = make([]byte, numBitsByteLen)
	numBitsMask := byte(0xff)
	if modEight := numBits & 0x7; modEight != 0 {
		numBitsMask <<= uint8(8 - (modEight))
	}
	for i := 1; i < int(byteLen); i++ {
		dstBytes[i-1] = srcBytes[i-1]<<bitsOffset | srcBytes[i]>>(8-bitsOffset)
	}
	if byteLen == numBitsByteLen {
		dstBytes[byteLen-1] = srcBytes[byteLen-1] << bitsOffset
	}
	dstBytes[numBitsByteLen-1] &= numBitsMask
	return
}

// GetFewBits is to get Value with desire few bits from source byte with bit offset
// func GetFewBits(srcByte byte, bitsOffset uint, numBits uint) (value uint64, err error) {

// 	if numBits == 0 {
// 		value = 0
// 		return
// 	}
// 	bitsLeft := 8 - bitsOffset
// 	if bitsLeft < numBits {
// 		err = fmt.Errorf("Get bits overflow, requireBits: %d, leftBits: %d", numBits, bitsLeft)
// 		return
// 	}
// 	if bitsOffset == 0 {
// 		value = uint64(srcByte >> (8 - numBits))
// 	} else {
// 		value = uint64((srcByte << bitsOffset) >> (8 - numBits))
// 	}
// 	return
// }

// BerGetBitsValue is to get Value with desire bits from source byte array with bit offset
func BerGetBitsValue(srcBytes []byte, bitsOffset uint, numBits uint) (value uint64, err error) {
	var dstBytes []byte
	dstBytes, err = BerGetBitString(srcBytes, bitsOffset, numBits)
	if err != nil {
		return
	}
	for i, j := 0, numBits; j >= 8; i, j = i+1, j-8 {
		value <<= 8
		value |= uint64(uint(dstBytes[i]))
	}
	if numBitsOff := (numBits & 0x7); numBitsOff != 0 {
		var mask uint = (1 << numBitsOff) - 1
		value <<= numBitsOff
		value |= uint64(uint(dstBytes[len(dstBytes)-1]>>(8-numBitsOff)) & mask)
	}
	return
}

func (bd *berByteData) bitCarry() {
	bd.byteOffset += uint64(bd.bitsOffset >> 3)
	bd.bitsOffset = bd.bitsOffset & 0x07
}

func (bd *berByteData) getBitString(numBits uint) (dstBytes []byte, err error) {
	dstBytes, err = BerGetBitString(bd.bytes[bd.byteOffset:], bd.bitsOffset, numBits)
	if err != nil {
		return
	}
	bd.bitsOffset += numBits

	bd.bitCarry()
	berTrace(1, berByteLog(uint64(numBits), bd.byteOffset, bd.bitsOffset, dstBytes))
	return
}

func (bd *berByteData) getBitsValue(numBits uint) (value uint64, err error) {
	value, err = BerGetBitsValue(bd.bytes[bd.byteOffset:], bd.bitsOffset, numBits)
	if err != nil {
		return
	}
	bd.bitsOffset += numBits
	bd.bitCarry()
	berTrace(1, berByteLog(uint64(numBits), bd.byteOffset, bd.bitsOffset, value))
	return
}

func (bd *berByteData) parseAlignBits() error {
	if (bd.bitsOffset & 0x7) > 0 {
		alignBits := 8 - ((bd.bitsOffset) & 0x7)
		berTrace(2, fmt.Sprintf("Aligning %d bits", alignBits))
		if val, err := bd.getBitsValue(alignBits); err != nil {
			return err
		} else if val != 0 {
			return fmt.Errorf("Align Bit is not zero")
		}
	} else if bd.bitsOffset != 0 {
		bd.bitCarry()
	}
	return nil
}

func (bd *berByteData) parseConstraintValue(valueRange int64) (value uint64, err error) {
	berTrace(3, fmt.Sprintf("Getting Constraint Value with range %d", valueRange))

	var bytes uint
	if valueRange <= 255 {
		if valueRange < 0 {
			err = fmt.Errorf("Value range is negative")
			return
		}
		var i uint
		// 1 ~ 8 bits
		for i = 1; i <= 8; i++ {
			upper := 1 << i
			if int64(upper) >= valueRange {
				break
			}
		}
		value, err = bd.getBitsValue(i)
		return
	} else if valueRange == 256 {
		bytes = 1
	} else if valueRange <= 65536 {
		bytes = 2
	} else {
		err = fmt.Errorf("Constraint Value is large than 65536")
		return
	}
	if err = bd.parseAlignBits(); err != nil {
		return
	}
	value, err = bd.getBitsValue(bytes * 8)
	return value, err
}

func (bd *berByteData) parseSemiConstrainedWholeNumber(lb uint64) (value uint64, err error) {
	var repeat bool
	var length uint64
	if length, err = bd.parseLength(-1, &repeat); err != nil {
		return
	}
	if length > 8 || repeat {
		err = fmt.Errorf("Too long length: %d", length)
		return
	}
	if value, err = bd.getBitsValue(uint(length) * 8); err != nil {
		return
	}
	value += lb
	return
}

func (bd *berByteData) parseNormallySmallNonNegativeWholeNumber() (value uint64, err error) {
	var notSmallFlag uint64
	if notSmallFlag, err = bd.getBitsValue(1); err != nil {
		return
	}
	if notSmallFlag == 1 {
		if value, err = bd.parseSemiConstrainedWholeNumber(0); err != nil {
			return
		}
	} else {
		if value, err = bd.getBitsValue(6); err != nil {
			return
		}
	}
	return
}

func (bd *berByteData) parseLength(sizeRange int64, repeat *bool) (value uint64, err error) {
	*repeat = false
	if sizeRange <= 65536 && sizeRange > 0 {
		return bd.parseConstraintValue(sizeRange)
	}

	if err = bd.parseAlignBits(); err != nil {
		return
	}
	firstByte, err := bd.getBitsValue(8)
	if err != nil {
		return
	}
	if (firstByte & 128) == 0 { // #10.9.3.6
		value = firstByte & 0x7F
		return
	} else if (firstByte & 64) == 0 { // #10.9.3.7
		var secondByte uint64
		if secondByte, err = bd.getBitsValue(8); err != nil {
			return
		}
		value = ((firstByte & 63) << 8) | secondByte
		return
	}
	firstByte &= 63
	if firstByte < 1 || firstByte > 4 {
		err = fmt.Errorf("Parse Length Out of Constraint")
		return
	}
	*repeat = true
	value = 16384 * firstByte
	return value, err
}

func (bd *berByteData) parseBitString(extensed bool, lowerBoundPtr *int64, upperBoundPtr *int64) (BitString, error) {
	var lb, ub, sizeRange int64 = 0, -1, -1
	if !extensed {
		if lowerBoundPtr != nil {
			lb = *lowerBoundPtr
		}
		if upperBoundPtr != nil {
			ub = *upperBoundPtr
			sizeRange = ub - lb + 1
		}
	}
	if ub > 65535 {
		sizeRange = -1
	}
	// initailization
	bitString := BitString{[]byte{}, 0}
	// lowerbound == upperbound
	if sizeRange == 1 {
		sizes := uint64(ub+7) >> 3
		bitString.BitLength = uint64(ub)
		berTrace(2, fmt.Sprintf("Decoding BIT STRING size %d", ub))
		if sizes > 2 {
			if err := bd.parseAlignBits(); err != nil {
				return bitString, err
			}
			if (bd.byteOffset + sizes) > uint64(len(bd.bytes)) {
				err := fmt.Errorf("BER data out of range")
				return bitString, err
			}
			bitString.Bytes = bd.bytes[bd.byteOffset : bd.byteOffset+sizes]
			bd.byteOffset += sizes
			bd.bitsOffset = uint(ub & 0x7)
			if bd.bitsOffset > 0 {
				bd.byteOffset--
			}
			berTrace(1, berByteLog(uint64(ub), bd.byteOffset, bd.bitsOffset, bitString.Bytes))
		} else {
			if bytes, err := bd.getBitString(uint(ub)); err != nil {
				logger.AperLog.Warnf("PD BerGetBitString error: %+v", err)
				return bitString, err
			} else {
				bitString.Bytes = bytes
			}
		}
		berTrace(2, fmt.Sprintf("Decoded BIT STRING (length = %d): %0.8b", ub, bitString.Bytes))
		return bitString, nil
	}
	repeat := false
	for {
		var rawLength uint64
		if length, err := bd.parseLength(sizeRange, &repeat); err != nil {
			return bitString, err
		} else {
			rawLength = length
		}
		rawLength += uint64(lb)
		berTrace(2, fmt.Sprintf("Decoding BIT STRING size %d", rawLength))
		if rawLength == 0 {
			return bitString, nil
		}
		sizes := (rawLength + 7) >> 3
		if err := bd.parseAlignBits(); err != nil {
			return bitString, err
		}

		if (bd.byteOffset + sizes) > uint64(len(bd.bytes)) {
			err := fmt.Errorf("BER data out of range")
			return bitString, err
		}
		bitString.Bytes = append(bitString.Bytes, bd.bytes[bd.byteOffset:bd.byteOffset+sizes]...)
		bitString.BitLength += rawLength
		bd.byteOffset += sizes
		bd.bitsOffset = uint(rawLength & 0x7)
		if bd.bitsOffset != 0 {
			bd.byteOffset--
		}
		berTrace(1, berByteLog(rawLength, bd.byteOffset, bd.bitsOffset, bitString.Bytes))
		berTrace(2, fmt.Sprintf("Decoded BIT STRING (length = %d): %0.8b", rawLength, bitString.Bytes))

		if !repeat {
			// if err = bd.parseAlignBits(); err != nil {
			// 	return
			// }
			break
		}
	}
	return bitString, nil
}

func (bd *berByteData) parseOctetString(extensed bool, lowerBoundPtr *int64, upperBoundPtr *int64) (
	OctetString, error) {
	var lb, ub, sizeRange int64 = 0, -1, -1
	if !extensed {
		if lowerBoundPtr != nil {
			lb = *lowerBoundPtr
		}
		if upperBoundPtr != nil {
			ub = *upperBoundPtr
			sizeRange = ub - lb + 1
		}
	}
	if ub > 65535 {
		sizeRange = -1
	}
	// initailization
	octetString := OctetString("")
	// lowerbound == upperbound
	if sizeRange == 1 {
		berTrace(2, fmt.Sprintf("Decoding OCTET STRING size %d", ub))
		if ub > 2 {
			unsignedUB := uint64(ub)
			if err := bd.parseAlignBits(); err != nil {
				return octetString, err
			}
			if (int64(bd.byteOffset) + ub) > int64(len(bd.bytes)) {
				err := fmt.Errorf("per data out of range")
				return octetString, err
			}
			octetString = bd.bytes[bd.byteOffset : bd.byteOffset+unsignedUB]
			bd.byteOffset += uint64(ub)
			berTrace(1, berByteLog(8*unsignedUB, bd.byteOffset, bd.bitsOffset, octetString))
		} else {
			if octet, err := bd.getBitString(uint(ub * 8)); err != nil {
				return octetString, err
			} else {
				octetString = octet
			}
		}
		berTrace(2, fmt.Sprintf("Decoded OCTET STRING (length = %d): 0x%0x", ub, octetString))
		return octetString, nil
	}
	repeat := false
	for {
		var rawLength uint64
		if length, err := bd.parseLength(sizeRange, &repeat); err != nil {
			return octetString, err
		} else {
			rawLength = length
		}
		rawLength += uint64(lb)
		berTrace(2, fmt.Sprintf("Decoding OCTET STRING size %d", rawLength))
		if rawLength == 0 {
			return octetString, nil
		} else if err := bd.parseAlignBits(); err != nil {
			return octetString, err
		}
		if (rawLength + bd.byteOffset) > uint64(len(bd.bytes)) {
			err := fmt.Errorf("per data out of range ")
			return octetString, err
		}
		octetString = append(octetString, bd.bytes[bd.byteOffset:bd.byteOffset+rawLength]...)
		bd.byteOffset += rawLength
		berTrace(1, berByteLog(8*rawLength, bd.byteOffset, bd.bitsOffset, octetString))
		berTrace(2, fmt.Sprintf("Decoded OCTET STRING (length = %d): 0x%0x", rawLength, octetString))
		if !repeat {
			// if err = bd.parseAlignBits(); err != nil {
			// 	return
			// }
			break
		}
	}
	return octetString, nil
}

func (bd *berByteData) parseBool() (value bool, err error) {
	berTrace(3, "Decoding BOOLEAN Value")
	bit, err1 := bd.getBitsValue(1)
	if err1 != nil {
		err = err1
		return
	}
	if bit == 1 {
		value = true
		berTrace(2, "Decoded BOOLEAN Value : ture")
	} else {
		value = false
		berTrace(2, "Decoded BOOLEAN Value : false")
	}
	return
}

func (bd *berByteData) parseInteger(extensed bool, lowerBoundPtr *int64, upperBoundPtr *int64) (int64, error) {
	var lb, ub, valueRange int64 = 0, -1, 0
	if !extensed {
		if lowerBoundPtr == nil {
			berTrace(3, "Decoding INTEGER with Unconstraint Value")
			valueRange = -1
		} else {
			lb = *lowerBoundPtr
			if upperBoundPtr != nil {
				ub = *upperBoundPtr
				valueRange = ub - lb + 1
				berTrace(3, fmt.Sprintf("Decoding INTEGER with Value Range(%d..%d)", lb, ub))
			} else {
				berTrace(3, fmt.Sprintf("Decoding INTEGER with Semi-Constraint Range(%d..)", lb))
			}
		}
	} else {
		valueRange = -1
		berTrace(3, "Decoding INTEGER with Extensive Value")
	}
	var rawLength uint
	if valueRange == 1 {
		return ub, nil
	} else if valueRange <= 0 {
		// semi-constraint or unconstraint
		if err := bd.parseAlignBits(); err != nil {
			return int64(0), err
		}
		if bd.byteOffset >= uint64(len(bd.bytes)) {
			return int64(0), fmt.Errorf("per data out of range")
		}
		rawLength = uint(bd.bytes[bd.byteOffset])
		bd.byteOffset++
		berTrace(1, berByteLog(8, bd.byteOffset, bd.bitsOffset, uint64(rawLength)))
	} else if valueRange <= 65536 {
		rawValue, err := bd.parseConstraintValue(valueRange)
		if err != nil {
			return int64(0), err
		} else {
			return int64(rawValue) + lb, nil
		}
	} else {
		// valueRange > 65536
		var byteLen uint
		unsignedValueRange := uint64(valueRange - 1)
		for byteLen = 1; byteLen <= 127; byteLen++ {
			unsignedValueRange >>= 8
			if unsignedValueRange == 0 {
				break
			}
		}
		var i, upper uint
		// 1 ~ 8 bits
		for i = 1; i <= 8; i++ {
			upper = 1 << i
			if upper >= byteLen {
				break
			}
		}
		if tempLength, err := bd.getBitsValue(i); err != nil {
			return int64(0), err
		} else {
			rawLength = uint(tempLength)
		}
		rawLength++
		if err := bd.parseAlignBits(); err != nil {
			return int64(0), err
		}
	}
	berTrace(2, fmt.Sprintf("Decoding INTEGER Length with %d bytes", rawLength))

	if rawValue, err := bd.getBitsValue(rawLength * 8); err != nil {
		return int64(0), err
	} else if valueRange < 0 {
		signedBitMask := uint64(1 << (rawLength*8 - 1))
		valueMask := signedBitMask - 1
		// negative
		if rawValue&signedBitMask > 0 {
			return int64((^rawValue)&valueMask+1) * -1, nil
		}
		return int64(rawValue) + lb, nil
	} else {
		return int64(rawValue) + lb, nil
	}
}

func (bd *berByteData) parseEnumerated(extensed bool, lowerBoundPtr *int64, upperBoundPtr *int64) (value uint64,
	err error) {
	if lowerBoundPtr == nil || upperBoundPtr == nil {
		err = fmt.Errorf("ENUMERATED value constraint is error")
		return
	}
	lb, ub := *lowerBoundPtr, *upperBoundPtr
	if lb < 0 || lb > ub {
		err = fmt.Errorf("ENUMERATED value constraint is error")
		return
	}

	if extensed {
		berTrace(2, fmt.Sprintf("Decoding ENUMERATED with Extensive Value of Range(%d..)", ub+1))
		if value, err = bd.parseNormallySmallNonNegativeWholeNumber(); err != nil {
			return
		}
		value += uint64(ub) + 1
	} else {
		berTrace(2, fmt.Sprintf("Decoding ENUMERATED with Value Range(%d..%d)", lb, ub))
		valueRange := ub - lb + 1
		if valueRange > 1 {
			value, err = bd.parseConstraintValue(valueRange)
		}
	}
	berTrace(2, fmt.Sprintf("Decoded ENUMERATED Value : %d", value))
	return
}

func (bd *berByteData) parseSequenceOf(sizeExtensed bool, params fieldParameters, sliceType reflect.Type) (
	reflect.Value, error) {
	var sliceContent reflect.Value
	var lb int64 = 0
	var sizeRange int64
	if params.sizeLowerBound != nil && *params.sizeLowerBound < 65536 {
		lb = *params.sizeLowerBound
	}
	if !sizeExtensed && params.sizeUpperBound != nil && *params.sizeUpperBound < 65536 {
		ub := *params.sizeUpperBound
		sizeRange = ub - lb + 1
		berTrace(3, fmt.Sprintf("Decoding Length of \"SEQUENCE OF\"  with Size Range(%d..%d)", lb, ub))
	} else {
		sizeRange = -1
		berTrace(3, fmt.Sprintf("Decoding Length of \"SEQUENCE OF\" with Semi-Constraint Range(%d..)", lb))
	}

	var numElements uint64
	if sizeRange > 1 {
		if numElementsTmp, err := bd.parseConstraintValue(sizeRange); err != nil {
			logger.AperLog.Warnf("Parse Constraint Value failed: %+v", err)
		} else {
			numElements = numElementsTmp
		}
		numElements += uint64(lb)
	} else if sizeRange == 1 {
		numElements += uint64(lb)
	} else {
		if err := bd.parseAlignBits(); err != nil {
			return sliceContent, err
		}
		if bd.byteOffset >= uint64(len(bd.bytes)) {
			err := fmt.Errorf("per data out of range")
			return sliceContent, err
		}
		numElements = uint64(bd.bytes[bd.byteOffset])
		bd.byteOffset++
		berTrace(1, berByteLog(8, bd.byteOffset, bd.bitsOffset, numElements))
	}
	berTrace(2, fmt.Sprintf("Decoding  \"SEQUENCE OF\" struct %s with len(%d)", sliceType.Elem().Name(), numElements))
	params.sizeUpperBound = nil
	params.sizeLowerBound = nil
	intNumElements := int(numElements)
	sliceContent = reflect.MakeSlice(sliceType, intNumElements, intNumElements)
	for i := 0; i < intNumElements; i++ {
		err := ParseField(sliceContent.Index(i), bd, params)
		if err != nil {
			return sliceContent, err
		}
	}
	return sliceContent, nil
}

func (bd *berByteData) getChoiceIndex(extensed bool, upperBoundPtr *int64) (present int, err error) {
	if extensed {
		err = fmt.Errorf("Unsupport value of CHOICE type is in Extensed")
	} else if upperBoundPtr == nil {
		err = fmt.Errorf("The upper bound of CHIOCE is missing")
	} else if ub := *upperBoundPtr; ub < 0 {
		err = fmt.Errorf("The upper bound of CHIOCE is negative")
	} else if rawChoice, err1 := bd.parseConstraintValue(ub + 1); err1 != nil {
		err = err1
	} else {
		berTrace(2, fmt.Sprintf("Decoded Present index of CHOICE is %d + 1", rawChoice))
		present = int(rawChoice) + 1
	}
	return
}

func berGetReferenceFieldValue(v reflect.Value) (value int64, err error) {
	fieldType := v.Type()
	switch v.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64:
		value = v.Int()
	case reflect.Struct:
		if fieldType.Field(0).Name == "Present" {
			present := int(v.Field(0).Int())
			if present == 0 {
				err = fmt.Errorf("ReferenceField Value present is 0(present's field number)")
			} else if present >= fieldType.NumField() {
				err = fmt.Errorf("Present is bigger than number of struct field")
			} else {
				value, err = berGetReferenceFieldValue(v.Field(present))
			}
		} else {
			value, err = berGetReferenceFieldValue(v.Field(0))
		}
	default:
		err = fmt.Errorf("OpenType reference only support INTEGER")
	}
	return
}

func (bd *berByteData) parseOpenType(v reflect.Value, params fieldParameters) error {
	bdOpenType := &berByteData{[]byte(""), 0, 0}
	repeat := false
	for {
		var rawLength uint64
		if rawLengthTmp, err := bd.parseLength(-1, &repeat); err != nil {
			return err
		} else {
			rawLength = rawLengthTmp
		}
		if rawLength == 0 {
			break
		} else if err := bd.parseAlignBits(); err != nil {
			return err
		}
		if (rawLength + bd.byteOffset) > uint64(len(bd.bytes)) {
			return fmt.Errorf("per data out of range ")
		}
		bdOpenType.bytes = append(bdOpenType.bytes, bd.bytes[bd.byteOffset:bd.byteOffset+rawLength]...)
		bd.byteOffset += rawLength

		if !repeat {
			if err := bd.parseAlignBits(); err != nil {
				return err
			}
			break
		}
	}
	berTrace(2, fmt.Sprintf("Decoding OpenType %s with (len = %d byte)", v.Type().String(), len(bdOpenType.bytes)))
	err := ParseField(v, bdOpenType, params)
	berTrace(2, fmt.Sprintf("Decoded OpenType %s", v.Type().String()))
	return err
}

// ParseField is the main parsing function. Given a byte slice and an offset
// into the array, it will try to parse a suitable ASN.1 value out and store it
// in the given Value. TODO : ObjectIdenfier, handle extension Field
func ParseField(v reflect.Value, bd *berByteData, params fieldParameters) error {
	fieldType := v.Type()

	// If we have run out of data return error.
	if bd.byteOffset == uint64(len(bd.bytes)) {
		return fmt.Errorf("sequence truncated")
	}
	if v.Kind() == reflect.Ptr {
		ptr := reflect.New(fieldType.Elem())
		v.Set(ptr)
		return ParseField(v.Elem(), bd, params)
	}
	sizeExtensible := false
	valueExtensible := false

	// We deal with the structures defined in this package first.
	switch fieldType {
	case BitStringType:
		bitString, err1 := bd.parseBitString(sizeExtensible, params.sizeLowerBound, params.sizeUpperBound)

		if err1 != nil {
			return err1
		}
		v.Set(reflect.ValueOf(bitString))
		return nil
	case ObjectIdentifierType:
		return fmt.Errorf("Unsupport ObjectIdenfier type")
	case OctetStringType:
		if octetString, err := bd.parseOctetString(sizeExtensible, params.sizeLowerBound, params.sizeUpperBound); err != nil {
			return err
		} else {
			v.Set(reflect.ValueOf(octetString))
			return nil
		}
	case EnumeratedType:
		if parsedEnum, err := bd.parseEnumerated(valueExtensible, params.valueLowerBound,
			params.valueUpperBound); err != nil {
			return err
		} else {
			v.SetUint(parsedEnum)
			return nil
		}
	}
	switch val := v; val.Kind() {
	case reflect.Bool:
		if parsedBool, err := bd.parseBool(); err != nil {
			return err
		} else {
			val.SetBool(parsedBool)
			return nil
		}
	case reflect.Int, reflect.Int32, reflect.Int64:
		if parsedInt, err := bd.parseInteger(valueExtensible, params.valueLowerBound, params.valueUpperBound); err != nil {
			return err
		} else {
			val.SetInt(parsedInt)
			berTrace(2, fmt.Sprintf("Decoded INTEGER Value: %d", parsedInt))
			return nil
		}
	case reflect.Struct:

		structType := fieldType
		var structParams []fieldParameters
		var optionalCount uint
		var optionalPresents uint64

		// pass tag for optional
		for i := 0; i < structType.NumField(); i++ {
			if structType.Field(i).PkgPath != "" {
				return fmt.Errorf("struct contains unexported fields : " + structType.Field(i).PkgPath)
			}
			tempParams := parseFieldParameters(structType.Field(i).Tag.Get("aper"))
			// for optional flag
			if tempParams.optional {
				optionalCount++
			}
			structParams = append(structParams, tempParams)
		}

		if optionalCount > 0 {
			if optionalPresentsTmp, err := bd.getBitsValue(optionalCount); err != nil {
				return err
			} else {
				optionalPresents = optionalPresentsTmp
			}
			berTrace(2, fmt.Sprintf("optionalPresents is %0b", optionalPresents))
		}

		// CHOICE or OpenType
		if structType.NumField() > 0 && structType.Field(0).Name == "Present" {
			var present int = 0
			if params.openType {
				if params.referenceFieldValue == nil {
					return fmt.Errorf("OpenType reference value is empty")
				}
				refValue := *params.referenceFieldValue

				for j, param := range structParams {
					if j == 0 {
						continue
					}
					if param.referenceFieldValue != nil && *param.referenceFieldValue == refValue {
						present = j
						break
					}
				}
				if present == 0 {
					return fmt.Errorf("OpenType reference value does not match any field")
				} else if present >= structType.NumField() {
					return fmt.Errorf("OpenType Present is bigger than number of struct field")
				} else {
					val.Field(0).SetInt(int64(present))
					berTrace(2, fmt.Sprintf("Decoded Present index of OpenType is %d ", present))
					return bd.parseOpenType(val.Field(present), structParams[present])
				}
			} else {
				if presentTmp, err := bd.getChoiceIndex(valueExtensible, params.valueUpperBound); err != nil {
					logger.AperLog.Errorf("bd.getChoiceIndex Error")
				} else {
					present = presentTmp
				}
				val.Field(0).SetInt(int64(present))
				if present == 0 {
					return fmt.Errorf("CHOICE present is 0(present's field number)")
				} else if present >= structType.NumField() {
					return fmt.Errorf("CHOICE Present is bigger than number of struct field")
				} else {
					return ParseField(val.Field(present), bd, structParams[present])
				}
			}
		}

		for i := 0; i < structType.NumField(); i++ {
			if structParams[i].optional && optionalCount > 0 {
				optionalCount--
				if optionalPresents&(1<<optionalCount) == 0 {
					berTrace(3, fmt.Sprintf("Field \"%s\" in %s is OPTIONAL and not present", structType.Field(i).Name, structType))
					continue
				} else {
					berTrace(3, fmt.Sprintf("Field \"%s\" in %s is OPTIONAL and present", structType.Field(i).Name, structType))
				}
			}
			// for open type reference
			if structParams[i].openType {
				fieldName := structParams[i].referenceFieldName
				var index int
				for index = 0; index < i; index++ {
					if structType.Field(index).Name == fieldName {
						break
					}
				}
				if index == i {
					return fmt.Errorf("Open type is not reference to the other field in the struct")
				}
				structParams[i].referenceFieldValue = new(int64)
				if referenceFieldValue, err := berGetReferenceFieldValue(val.Field(index)); err != nil {
					return err
				} else {
					*structParams[i].referenceFieldValue = referenceFieldValue
				}
			}
			if err := ParseField(val.Field(i), bd, structParams[i]); err != nil {
				return err
			}
		}
		return nil
	case reflect.Slice:
		sliceType := fieldType
		if newSlice, err := bd.parseSequenceOf(sizeExtensible, params, sliceType); err != nil {
			return err
		} else {
			val.Set(newSlice)
			return nil
		}
	case reflect.String:
		berTrace(2, "Decoding PrintableString using Octet String decoding method")

		if octetString, err := bd.parseOctetString(sizeExtensible, params.sizeLowerBound, params.sizeUpperBound); err != nil {
			return err
		} else {
			printableString := string(octetString)
			val.SetString(printableString)
			berTrace(2, fmt.Sprintf("Decoded PrintableString : \"%s\"", printableString))
			return nil
		}
	}
	return fmt.Errorf("unsupported: " + v.Type().String())
}

// Unmarshal parses the ABER-encoded ASN.1 data structure b
// and uses the reflect package to fill in an arbitrary value pointed at by value.
// Because Unmarshal uses the reflect package, the structs
// being written to must use upper case field names.
//
// An ASN.1 INTEGER can be written to an int, int32, int64,
// If the encoded value does not fit in the Go type,
// Unmarshal returns a parse error.
//
// An ASN.1 BIT STRING can be written to a BitString.
//
// An ASN.1 OCTET STRING can be written to a []byte.
//
// An ASN.1 OBJECT IDENTIFIER can be written to an
// ObjectIdentifier.
//
// An ASN.1 ENUMERATED can be written to an Enumerated.
//
// Any of the above ASN.1 values can be written to an interface{}.
// The value stored in the interface has the corresponding Go type.
// For integers, that type is int64.
//
// An ASN.1 SEQUENCE OF x can be written
// to a slice if an x can be written to the slice's element type.
//
// An ASN.1 SEQUENCE can be written to a struct
// if each of the elements in the sequence can be
// written to the corresponding element in the struct.
//
// The following tags on struct fields have special meaning to Unmarshal:
//
//	optional        	OPTIONAL tag in SEQUENCE
//	sizeExt             specifies that size  is extensible
//	valueExt            specifies that value is extensible
//	sizeLB		        set the minimum value of size constraint
//	sizeUB              set the maximum value of value constraint
//	valueLB		        set the minimum value of size constraint
//	valueUB             set the maximum value of value constraint
//	default             sets the default value
//	openType            specifies the open Type
//  referenceFieldName	the string of the reference field for this type (only if openType used)
//  referenceFieldValue	the corresponding value of the reference field for this type (only if openType used)
//
// Other ASN.1 types are not supported; if it encounters them,
// Unmarshal returns a parse error.
func Unmarshal(b []byte, value interface{}) error {
	return UnmarshalWithParams(b, value, "")
}

// UnmarshalWithParams allows field parameters to be specified for the
// top-level element. The form of the params is the same as the field tags.
func UnmarshalWithParams(b []byte, value interface{}, params string) error {
	v := reflect.ValueOf(value).Elem()
	bd := &berByteData{b, 0, 0}
	return ParseField(v, bd, parseFieldParameters(params))
}