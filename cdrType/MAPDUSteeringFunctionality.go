package cdrType

import "github.com/free5gc/CDRUtil/asn"
// Need to import "gofree5gc/lib/aper" if it uses "aper"

const (	/* Enum Type */
	MAPDUSteeringFunctionalityPresentMPTCP	asn.Enumerated = 0
	MAPDUSteeringFunctionalityPresentATSSSLL	asn.Enumerated = 1
)

type MAPDUSteeringFunctionality struct {
	Value	asn.Enumerated 
}
