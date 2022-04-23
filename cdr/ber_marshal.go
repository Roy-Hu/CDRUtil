package cdr

import (
	"fmt"
	"reflect"
)

type berRawByteData struct {
	bytes []byte
}

func berRawBitLog(numBits uint64, byteLen int, bitsOffset uint, value interface{}) string {
	if reflect.TypeOf(value).Kind() == reflect.Uint64 {
		return fmt.Sprintf("  [BER put %2d bits, byteLen(after): %d, bitsOffset(after): %d, value: 0x%0x]",
			numBits, byteLen, bitsOffset, reflect.ValueOf(value).Uint())
	}
	return fmt.Sprintf("  [BER put %2d bits, byteLen(after): %d, bitsOffset(after): %d, value: 0x%0x]",
		numBits, byteLen, bitsOffset, reflect.ValueOf(value).Bytes())
}

func (bd *berRawByteData) appendIdentifier(class uint8, structed bool, value uint64) (err error) {
	var identifier uint8
	identifier |= class << 6
	if structed {
		identifier |= 0x20
	}
	if value < 31 {
		identifier |= byte(value)
		bd.bytes = append(bd.bytes, identifier)
		return
	}
	identifier |= 31
	bd.bytes = append(bd.bytes, identifier)

	encodeLen := 1
	tmp := value
	for tmp > 127 {
		encodeLen++
		tmp >>= 7
	}
	valueByte := make([]byte, encodeLen)
	for i := 0; i < encodeLen; i++ {
		valueByte[encodeLen-1-i] = byte(value) | 0x80
		value >>= 7
	}
	valueByte[encodeLen-1] &= 0x7f
	bd.bytes = append(bd.bytes, valueByte...)

	return
}

func (bd *berRawByteData) appendLength(value uint64, indefinite bool) (err error) {
	if indefinite {
		berTrace(2, "Length is indefinite format")
		bd.bytes = append(bd.bytes, 0x80)
		return
	}
	berTrace(2, fmt.Sprintf("Putting Length of Value : %d", value))
	if value <= 127 {
		bd.bytes = append(bd.bytes, byte(value))
		return
	}

	// long form
	encodeLen := 1
	tmp := value
	for tmp > 255 {
		encodeLen++
		tmp >>= 8
	}
	valueByte := make([]byte, encodeLen)
	for i := 0; i < encodeLen; i++ {
		valueByte[encodeLen-1-i] = byte(value)
		value >>= 8
	}
	// first byte of long form length bytes
	// bit 8 shall be one
	// bit 7-1 is the number of sequence bytes length
	// the value 127 is reserved for future use
	// therefore the length of length bytes is at most 126
	bd.bytes = append(bd.bytes, byte(encodeLen|0x80))
	bd.bytes = append(bd.bytes, valueByte...)

	return
}

func (bd *berRawByteData) appendBoolean(value bool) (err error) {
	// x.690 8.2
	bd.appendLength(1, false)
	if value {
		bd.bytes = append(bd.bytes, 1)
	} else {
		bd.bytes = append(bd.bytes, 0)
	}

	return
}

func (bd *berRawByteData) appendInteger(value int64) (err error) {
	// x.690 8.3
	encodeLen := 1

	tmp := value
	for tmp > 127 {
		encodeLen++
		tmp >>= 8
	}
	for tmp < -128 {
		encodeLen++
		tmp >>= 8
	}
	valueByte := make([]byte, encodeLen)
	for i := 0; i < encodeLen; i++ {
		valueByte[encodeLen-1-i] = byte(value)
		value >>= 8
	}
	bd.appendLength(uint64(encodeLen), false)
	bd.bytes = append(bd.bytes, valueByte...)

	return
}

func (bd *berRawByteData) appendEnumerated(value uint64) (err error) {
	// x.690 8.4
	encodeLen := 1

	tmp := value
	for tmp > 127 {
		encodeLen++
		tmp >>= 8
	}
	valueByte := make([]byte, encodeLen)
	for i := 0; i < encodeLen; i++ {
		valueByte[encodeLen-1-i] = byte(value)
		value >>= 8
	}
	bd.appendLength(uint64(encodeLen), false)
	bd.bytes = append(bd.bytes, valueByte...)

	return
}

func (bd *berRawByteData) appendReal() (err error) {
	// TODO
	// currently not used in chf-cdr 32.298 v16.9.0
	// x.690 8.5
	return fmt.Errorf("Encoding of type REAL is not implemented")
}

func (bd *berRawByteData) appendBitString(bytes []byte, bitLen uint64) (err error) {
	// x.690 8.6
	unusedBitLen := 8 - bitLen%8
	bd.appendLength(uint64(len(bytes)+1), false)
	bd.bytes = append(bd.bytes, byte(unusedBitLen))
	bd.bytes = append(bd.bytes, bytes...)

	return
}

func (bd *berRawByteData) appendOctetString(bytes []byte) (err error) {
	// x.690 8.7
	bd.appendLength(uint64(len(bytes)), false)
	bd.bytes = append(bd.bytes, bytes...)

	return
}

func (bd *berRawByteData) appendNull() (err error) {
	// x.690 8.8
	bd.appendLength(0, false)

	return
}

func (bd *berRawByteData) appendOpenType() (err error) {
	// x.690 8.15

	return
}

func (bd *berRawByteData) appendInstanceOf() (err error) {
	// TODO
	// currently not used in chf-cdr 32.298 v16.9.0
	// x.690 8.16
	return fmt.Errorf("Encoding of type INSTANCE OF is not implemented")
}

func (bd *berRawByteData) appendEmbeddedPdvType() (err error) {
	// TODO
	// currently not used in chf-cdr 32.298 v16.9.0
	// x.690 8.17

	return fmt.Errorf("Encoding of type EMDEDDED PDV is not implemented")
}

func (bd *berRawByteData) appendExternalType() (err error) {
	// TODO
	// currently not used in chf-cdr 32.298 v16.9.0
	// x.690 8.18
	return fmt.Errorf("Encoding of type EXTERNAL is not implemented")
}

func (bd *berRawByteData) appendObjectIdentifier() (err error) {
	// x.690 8.19

	return fmt.Errorf("Encoding of type OBJECT IDENTIFIER is not implemented")
}

func (bd *berRawByteData) appendRelativeOid() (err error) {
	// TODO
	// currently not used in chf-cdr 32.298 v16.9.0
	// x.690 8.20
	return fmt.Errorf("Encoding of type RELATIVE-OID is not implemented")
}

func (bd *berRawByteData) appendOidIri() (err error) {
	// TODO
	// currently not used in chf-cdr 32.298 v16.9.0
	// x.690 8.21

	return fmt.Errorf("Encoding of type OID-IRI is not implemented")
}

func (bd *berRawByteData) appendRelativeOidIri() (err error) {
	// TODO
	// currently not used in chf-cdr 32.298 v16.9.0
	// x.690 8.22

	return fmt.Errorf("Encoding of type RELATIVE-OID-IRI is not implemented")
}

func (bd *berRawByteData) appendUtf8String(str string) (err error) {
	// x.690 8.23
	bd.appendOctetString([]byte(str))
	return
}

func (bd *berRawByteData) appendNumericString() (err error) {
	// TODO
	// currently not used in chf-cdr 32.298 v16.9.0
	// x.690 8.23

	return fmt.Errorf("Encoding of type NumericString is not implemented")
}

func (bd *berRawByteData) appendPrintableStirng() (err error) {
	// TODO
	// currently not used in chf-cdr 32.298 v16.9.0
	// x.690 8.23

	return fmt.Errorf("Encoding of type PrintableString is not implemented")
}

func (bd *berRawByteData) appendT61String() (err error) {
	// TODO
	// currently not used in chf-cdr 32.298 v16.9.0
	// x.690 8.23

	return fmt.Errorf("Encoding of type TeletexString(T61String) is not implemented")
}

func (bd *berRawByteData) appendVideotexString() (err error) {
	// TODO
	// currently not used in chf-cdr 32.298 v16.9.0
	// x.690 8.23

	return fmt.Errorf("Encoding of type VideotexString is not implemented")
}

func (bd *berRawByteData) appendIA5String(str string) (err error) {
	// x.690 8.23
	bd.appendOctetString([]byte(str))

	return
}

func (bd *berRawByteData) appendGraphicStirng(str string) (err error) {
	// x.690 8.23
	bd.appendOctetString([]byte(str))

	return
}

func (bd *berRawByteData) appendIso646String() (err error) {
	// TODO
	// currently not used in chf-cdr 32.298 v16.9.0
	// x.690 8.23

	return fmt.Errorf("Encoding of type VisibleString(ISO646String) is not implemented")
}

func (bd *berRawByteData) appendGeneralString() (err error) {
	// TODO
	// currently not used in chf-cdr 32.298 v16.9.0
	// x.690 8.23

	return fmt.Errorf("Encoding of type GeneralString is not implemented")
}

func (bd *berRawByteData) appendBmpString() (err error) {
	// TODO
	// currently not used in chf-cdr 32.298 v16.9.0
	// x.690 8.23

	return fmt.Errorf("Encoding of type BMPString is not implemented")
}

func (bd *berRawByteData) appendCharacterString() (err error) {
	// TODO
	// currently not used in chf-cdr 32.298 v16.9.0
	// x.690 8.24
	return fmt.Errorf("Encoding of type CHARACTER STRING is not implemented")
}

func (bd *berRawByteData) appendGeneralizedTime() (err error) {
	// TODO
	// currently not used in chf-cdr 32.298 v16.9.0
	// x.690 8.25

	return fmt.Errorf("Encoding of type GeneralizedTime is not implemented")
}

func (bd *berRawByteData) appendUniversalTime() (err error) {
	// TODO
	// currently not used in chf-cdr 32.298 v16.9.0
	// x.690 8.25

	return fmt.Errorf("Encoding of type UniversalTime is not implemented")
}

func (bd *berRawByteData) appendObjectDescriptor() (err error) {
	// TODO
	// currently not used in chf-cdr 32.298 v16.9.0
	// x.690 8.25
	return fmt.Errorf("Encoding of type ObjectDescriptor is not implemented")
}

func (bd *berRawByteData) appendTime() (err error) {
	// TODO
	// currently not used in chf-cdr 32.298 v16.9.0
	// x.690 8.26.1

	return fmt.Errorf("Encoding of type TIME is not implemented")
}

func (bd *berRawByteData) appendDate() (err error) {
	// TODO
	// currently not used in chf-cdr 32.298 v16.9.0
	// x.690 8.26.2

	return fmt.Errorf("Encoding of type DATE is not implemented")
}

func (bd *berRawByteData) appendTimeOfDay() (err error) {
	// TODO
	// currently not used in chf-cdr 32.298 v16.9.0
	// x.690 8.26.3

	return fmt.Errorf("Encoding of type TIME-OF-DAY is not implemented")
}

func (bd *berRawByteData) appendDateTime() (err error) {
	// TODO
	// currently not used in chf-cdr 32.298 v16.9.0
	// x.690 8.26.4

	return fmt.Errorf("Encoding of type DATE-TIME is not implemented")
}

func (bd *berRawByteData) appendDuration() (err error) {
	// TODO
	// currently not used in chf-cdr 32.298 v16.9.0
	// x.690 8.26.5

	return fmt.Errorf("Encoding of type DURATION is not implemented")
}

func (bd *berRawByteData) makeField(v reflect.Value, params fieldParameters) error {
	if !v.IsValid() {
		return fmt.Errorf("ber: cannot marshal nil value")
	}
	// If the field is an interface{} then recurse into it.
	if v.Kind() == reflect.Interface && v.Type().NumMethod() == 0 {
		return bd.makeField(v.Elem(), params)
	}
	if v.Kind() == reflect.Ptr {
		return bd.makeField(v.Elem(), params)
	}
	fieldType := v.Type()

	// We deal with the structures defined in this package first.
	switch fieldType {
	case BitStringType:
		if params.explicitTag {
			bd.appendIdentifier(0, false, 3)
		}
		err := bd.appendBitString(v.Field(0).Bytes(), v.Field(1).Uint())
		return err
	case ObjectIdentifierType:
		err := fmt.Errorf("Unsupport ObjectIdenfier type")
		return err
	case OctetStringType:
		if params.explicitTag {
			bd.appendIdentifier(0, false, 4)
		}
		err := bd.appendOctetString(v.Bytes())
		return err
	case EnumeratedType:
		if params.explicitTag {
			bd.appendIdentifier(0, false, 10)
		}
		err := bd.appendEnumerated(v.Uint())
		return err
	}
	switch val := v; val.Kind() {
	case reflect.Bool:
		if params.explicitTag {
			bd.appendIdentifier(0, false, 1)
		}
		err := bd.appendBoolean(v.Bool())
		return err
	case reflect.Int, reflect.Int32, reflect.Int64:
		if params.explicitTag {
			bd.appendIdentifier(0, false, 2)
		}
		err := bd.appendInteger(v.Int())
		return err

	case reflect.Struct:
		structType := fieldType
		if structType.Field(0).Name == "Value" {
			// Non struct type
			fmt.Println("Non struct type")
			if err := bd.makeField(val.Field(0), params); err != nil {
				return err
			}
		} else if structType.Field(0).Name == "List" {
			// List Type: SEQUENCE/SET OF
			fmt.Println("List type")
			if err := bd.makeField(val.Field(0), params); err != nil {
				return err
			}
		} else if structType.Field(0).Name == "Present" {
			fmt.Println("Chioce type")
			// Open type or CHOICE type
			present := int(v.Field(0).Int())
			tempParams := parseFieldParameters(structType.Field(present).Tag.Get("ber"))
			if present == 0 {
				return fmt.Errorf("CHOICE or OpenType present is 0(present's field number)")
			} else if present >= structType.NumField() {
				return fmt.Errorf("Present is bigger than number of struct field")
			} else if params.openType {
				// TODO openType
				return fmt.Errorf("Open Type is not implemented")
			} else {
				// identifier, length
				if params.tagNumber != nil {
					bd.appendLength(0x80, true)
				}
				bd.appendIdentifier(2, tempParams.structType != "", *tempParams.tagNumber)
				// content
				if err := bd.makeField(val.Field(present), tempParams); err != nil {
					return err
				}
				if params.tagNumber != nil {
					bd.appendIdentifier(0, false, 0)
					bd.appendLength(0, false)
				}
			}
			return nil
		} else {
			// Struct type: SEQUENCE, SET
			fmt.Println("Struct type")
			if params.explicitTag {
				if params.structType == "Sequence" {
					bd.appendIdentifier(0, true, 0x10)
				} else if params.structType == "Set" {
					bd.appendIdentifier(0, true, 0x11)
				}
			}
			bd.appendLength(0x80, true)
			for i := 0; i < structType.NumField(); i++ {
				tempParams := parseFieldParameters(structType.Field(i).Tag.Get("ber"))
				if tempParams.optional {
					if v.Field(i).IsNil() {
						berTrace(3, fmt.Sprintf("Field \"%s\" in %s is OPTIONAL and not present", structType.Field(i).Name, structType))
						continue
					} else {
						berTrace(3, fmt.Sprintf("Field \"%s\" in %s is OPTIONAL and present", structType.Field(i).Name, structType))
					}
				}
				// identifier
				bd.appendIdentifier(2, tempParams.structType != "", *tempParams.tagNumber)
				// for open type reference
				if tempParams.openType {
					// TODO
					return fmt.Errorf("Open Type is not implemented")
				}
				// content
				if err := bd.makeField(val.Field(i), tempParams); err != nil {
					fmt.Errorf("iterate subtype error")
					return err
				}
			}
			// end of contents
			bd.appendIdentifier(0, false, 0)
			bd.appendLength(0, false)
			return nil
		}
		return nil
	case reflect.Slice:
		if params.explicitTag {
			if params.structType == "SEQUENCE" {
				bd.appendIdentifier(0, true, 16)
			} else if params.structType == "SET" {
				bd.appendIdentifier(0, true, 17)
			}
			bd.appendLength(0x80, true)
		}
		for i := 0; i < v.Len(); i++ {
			if err := bd.makeField(v.Index(i), params); err != nil {
				return err
			}
		}
		if params.explicitTag {
			bd.appendIdentifier(0, false, 0)
			bd.appendLength(0, false)
		}
		return nil
	case reflect.String:
		if params.explicitTag {
			switch params.stringType {
			case "Utf8String":
				bd.appendIdentifier(0, false, 12)
			case "IA5String":
				bd.appendIdentifier(0, false, 22)
			case "GraphicString":
				bd.appendIdentifier(0, false, 25)
			}
		}
		printableString := v.String()
		berTrace(2, fmt.Sprintf("Encoding PrintableString : \"%s\" using Octet String decoding method", printableString))
		err := bd.appendOctetString([]byte(printableString))
		return err
	}
	return fmt.Errorf("unsupported: " + v.Type().String())
}

// Marshal returns the ASN.1 encoding of val.
func BerMarshal(val interface{}) ([]byte, error) {
	return BerMarshalWithParams(val, "")
}

// MarshalWithParams allows field parameters to be specified for the
// top-level element. The form of the params is the same as the field tags.
func BerMarshalWithParams(val interface{}, params string) ([]byte, error) {
	bd := &berRawByteData{[]byte("")}
	err := bd.makeField(reflect.ValueOf(val), parseFieldParameters(params))
	if err != nil {
		return bd.bytes, err
	} else if len(bd.bytes) == 0 {
		bd.bytes = make([]byte, 1)
	}
	return bd.bytes, nil
}
