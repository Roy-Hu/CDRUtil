package util

import (
	"fmt"
	// "reflect"
	"testing"
	"strconv"

	"github.com/stretchr/testify/require"
)

func TestCdrFile(t *testing.T) {
	t.Parallel()

	fileOpeningTS := CdrHdrTimeStamp{4, 28, 17, 18, 1, 8, 0}
	timestampWhenLastCdrWasAppendedToFIle := CdrHdrTimeStamp{1, 2, 3, 4, 1, 6, 30}

	// timeNow := time.Now()
	cdrf := CdrFileHeader{
		FileLength:                            5,
		HeaderLength:                          6,
		HighReleaseIdentifier:                 2,
		HighVersionIdentifier:                 3,
		LowReleaseIdentifier:                  4,
		LowVersionIdentifier:                  5,
		FileOpeningTimestamp:                  fileOpeningTS,
		TimestampWhenLastCdrWasAppendedToFIle: timestampWhenLastCdrWasAppendedToFIle,
		NumberOfCdrsInFile:                    1,
		FileSequenceNumber:                    11,
		FileClosureTriggerReason:              4,
		//IpAddressOfNodeThatGeneratedFile      [20]byte(),
		LostCdrIndicator:          4,
		LengthOfCdrRouteingFilter: 4,
		CDRRouteingFilter:                     []byte("abcd"),
		LengthOfPrivateExtension: 5,
		PrivateExtension:                      []byte("fghjk"), // vendor specific
		HighReleaseIdentifierExtension: 2,
		LowReleaseIdentifierExtension:  3,
	}

	cdrHeader := CdrHeader {
		CdrLength                  :3,
		ReleaseIdentifier          :Rel6, // octet 3 bit 6..8
		VersionIdentifier          :3,                // otcet 3 bit 1..5
		DataRecordFormat           :UnalignedPackedEncodingRules,  // octet 4 bit 6..8
		TsNumber                   : TS32253,   // octet 4 bit 1..5
		ReleaseIdentifierExtension :4,
	}

	cdrFile1 := CDRFile{
		hdr: cdrf,
		cdrList: []CDR{{hdr:cdrHeader, cdrByte:[]byte("abc")},},
	}

	testCases := []struct {
		name  string
		in    CDRFile
	}{
		{"cdrfile1", cdrFile1},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fileName := "encoding"+strconv.Itoa(i)+".txt"
			tc.in.Encoding(fileName)
			newCdrFile := CDRFile{}
			newCdrFile.Decoding(fileName)

			fmt.Println("tc.in", tc.in)
			fmt.Println("newCdrFile", newCdrFile)

			require.Equal(t, tc.in, newCdrFile)
			// require.True(t, reflect.DeepEqual(tc.in, newCdrFile))
		})
	}
}