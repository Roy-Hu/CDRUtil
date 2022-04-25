package util

import (
	"time"
)

type CDRFile struct {
	hdr CdrFileHeader
	cdrList []CDR
}

type CDR struct {
	hdr CdrHeader
	cdrByte []byte
}

type CdrFileHeader struct {
	FileLength                            uint32
	HeaderLength                          uint32
	HighReleaseIdentifier                 uint8 // octet 9 bit 6..8
	HighVersionIdentifier                 uint8 // octet 9 bit 1..5
	LowReleaseIdentifier                  uint8 // octet 10 bit 6..8
	LowVersionIdentifier                  uint8 // octet 10 bit 1..5
	FileOpeningTimestamp                  *time.Time
	TimestampWhenLastCdrWasAppendedToFIle *time.Time
	NumberOfCdrsInFile                    uint32
	FileSequenceNumber                    uint32
	FileClosureTriggerReason              FileClosureTriggerReasonType
	IpAddressOfNodeThatGeneratedFile      [20]byte // ip address in ipv6 format
	LostCdrIndicator                      uint8
	LengthOfCdrRouteingFilter             uint16
	CDRRouteingFilter                     []byte // vendor specific
	LengthOfPrivateExtension              uint16
	PrivateExtension                      []byte // vendor specific
	HighReleaseIdentifierExtension        uint8
	LowReleaseIdentifierExtension         uint8
}

type CdrHeader struct {
	CdrLength                  uint16
	ReleaseIdentifier          ReleaseIdentifierType // octet 3 bit 6..8
	VersionIdentifier          uint8                 // otcet 3 bit 1..5
	DataRecordFormat           DataRecordFormatType  // octet 4 bit 6..8
	TsNumber                   TsNumberIdentifier    // octet 4 bit 1..5
	ReleaseIdentifierExtension uint8
}

type FileClosureTriggerReasonType uint8

const (
	NormalClosure                     FileClosureTriggerReasonType = 0
	FileSizeLimitReached              FileClosureTriggerReasonType = 1
	FileOpentimeLimitedReached        FileClosureTriggerReasonType = 2
	MaximumNumberOfCdrsInFileReached  FileClosureTriggerReasonType = 3
	FileClosedByManualIntervention    FileClosureTriggerReasonType = 4
	CdrReleaseVersionOrEncodingChange FileClosureTriggerReasonType = 5
	AbnormalFileClosure               FileClosureTriggerReasonType = 128
	FileSystemError                   FileClosureTriggerReasonType = 129
	FileSystemStorageExhausted        FileClosureTriggerReasonType = 130
	FileIntegrityError                FileClosureTriggerReasonType = 131
)

type ReleaseIdentifierType uint8

const ()

type DataRecordFormatType uint8

const ()

type TsNumberIdentifier uint8

const ()
