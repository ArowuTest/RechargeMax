package utils

import (
	"fmt"
	"strings"
)

// NetworkPrefix represents a network provider and its associated prefixes
type NetworkPrefix struct {
	Network  string
	Prefixes []string
}

// Nigerian network prefixes (without country code)
var nigerianNetworkPrefixes = []NetworkPrefix{
	{
		Network: "MTN",
		Prefixes: []string{
			"0803", "0703", "0903", "0806", "0813",
			"0810", "0814", "0816", "0906",
		},
	},
	{
		Network: "Airtel",
		Prefixes: []string{
			"0802", "0708", "0902", "0808", "0701",
			"0812", "0901", "0907",
		},
	},
	{
		Network: "Glo",
		Prefixes: []string{
			"0805", "0705", "0905", "0807", "0815",
			"0811",
		},
	},
	{
		Network: "9mobile",
		Prefixes: []string{
			"0809", "0818", "0909", "0817", "0908",
		},
	},
}

// Nigerian network prefixes with country code (234)
var nigerianNetworkPrefixesWithCC = []NetworkPrefix{
	{
		Network: "MTN",
		Prefixes: []string{
			"2348031", "2348032", "2348033", "2348034", "2348035", "2348036", "2348037", "2348038", "2348039",
			"2347031", "2347032", "2347033", "2347034", "2347035", "2347036", "2347037", "2347038", "2347039",
			"2349031", "2349032", "2349033", "2349034", "2349035", "2349036", "2349037", "2349038", "2349039",
			"2348061", "2348062", "2348063", "2348064", "2348065", "2348066", "2348067", "2348068", "2348069",
			"2348131", "2348132", "2348133", "2348134", "2348135", "2348136", "2348137", "2348138", "2348139",
			"2348101", "2348102", "2348103", "2348104", "2348105", "2348106", "2348107", "2348108", "2348109",
			"2348141", "2348142", "2348143", "2348144", "2348145", "2348146", "2348147", "2348148", "2348149",
			"2348161", "2348162", "2348163", "2348164", "2348165", "2348166", "2348167", "2348168", "2348169",
			"2349061", "2349062", "2349063", "2349064", "2349065", "2349066", "2349067", "2349068", "2349069",
		},
	},
	{
		Network: "Airtel",
		Prefixes: []string{
			"2348021", "2348022", "2348023", "2348024", "2348025", "2348026", "2348027", "2348028", "2348029",
			"2347081", "2347082", "2347083", "2347084", "2347085", "2347086", "2347087", "2347088", "2347089",
			"2349021", "2349022", "2349023", "2349024", "2349025", "2349026", "2349027", "2349028", "2349029",
			"2348081", "2348082", "2348083", "2348084", "2348085", "2348086", "2348087", "2348088", "2348089",
			"2347011", "2347012", "2347013", "2347014", "2347015", "2347016", "2347017", "2347018", "2347019",
			"2348121", "2348122", "2348123", "2348124", "2348125", "2348126", "2348127", "2348128", "2348129",
			"2349011", "2349012", "2349013", "2349014", "2349015", "2349016", "2349017", "2349018", "2349019",
			"2349071", "2349072", "2349073", "2349074", "2349075", "2349076", "2349077", "2349078", "2349079",
		},
	},
	{
		Network: "Glo",
		Prefixes: []string{
			"2348051", "2348052", "2348053", "2348054", "2348055", "2348056", "2348057", "2348058", "2348059",
			"2347051", "2347052", "2347053", "2347054", "2347055", "2347056", "2347057", "2347058", "2347059",
			"2349051", "2349052", "2349053", "2349054", "2349055", "2349056", "2349057", "2349058", "2349059",
			"2348071", "2348072", "2348073", "2348074", "2348075", "2348076", "2348077", "2348078", "2348079",
			"2348151", "2348152", "2348153", "2348154", "2348155", "2348156", "2348157", "2348158", "2348159",
			"2348111", "2348112", "2348113", "2348114", "2348115", "2348116", "2348117", "2348118", "2348119",
		},
	},
	{
		Network: "9mobile",
		Prefixes: []string{
			"2348091", "2348092", "2348093", "2348094", "2348095", "2348096", "2348097", "2348098", "2348099",
			"2348181", "2348182", "2348183", "2348184", "2348185", "2348186", "2348187", "2348188", "2348189",
			"2349091", "2349092", "2349093", "2349094", "2349095", "2349096", "2349097", "2349098", "2349099",
			"2348171", "2348172", "2348173", "2348174", "2348175", "2348176", "2348177", "2348178", "2348179",
			"2349081", "2349082", "2349083", "2349084", "2349085", "2349086", "2349087", "2349088", "2349089",
		},
	},
}

// NormalizePhoneNumberTo234 normalizes a phone number to the format 234XXXXXXXXXX
func NormalizePhoneNumberTo234(phoneNumber string) string {
	// Remove all non-numeric characters
	phoneNumber = strings.ReplaceAll(phoneNumber, " ", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "(", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, ")", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "+", "")
	
	// If starts with 0, replace with 234
	if strings.HasPrefix(phoneNumber, "0") {
		phoneNumber = "234" + phoneNumber[1:]
	}
	
	// If doesn't start with 234, add it
	if !strings.HasPrefix(phoneNumber, "234") {
		phoneNumber = "234" + phoneNumber
	}
	
	return phoneNumber
}

// DetectNetwork detects the network provider from a phone number
func DetectNetwork(phoneNumber string) (string, error) {
	normalized := NormalizePhoneNumberTo234(phoneNumber)
	
	// Check against prefixes with country code
	for _, networkPrefix := range nigerianNetworkPrefixesWithCC {
		for _, prefix := range networkPrefix.Prefixes {
			if strings.HasPrefix(normalized, prefix) {
				return networkPrefix.Network, nil
			}
		}
	}
	
	return "", fmt.Errorf("unable to detect network for phone number: %s", phoneNumber)
}

// ValidatePhoneNetwork validates if a phone number belongs to the specified network
func ValidatePhoneNetwork(phoneNumber string, network string) (bool, error) {
	detectedNetwork, err := DetectNetwork(phoneNumber)
	if err != nil {
		return false, err
	}
	
	// Case-insensitive comparison
	return strings.EqualFold(detectedNetwork, network), nil
}

// GetNetworkPrefixes returns all prefixes for a given network
func GetNetworkPrefixes(network string) []string {
	for _, networkPrefix := range nigerianNetworkPrefixes {
		if strings.EqualFold(networkPrefix.Network, network) {
			return networkPrefix.Prefixes
		}
	}
	return []string{}
}

// IsValidNigerianPhoneNumber checks if a phone number is a valid Nigerian number
func IsValidNigerianPhoneNumber(phoneNumber string) bool {
	normalized := NormalizePhoneNumberTo234(phoneNumber)
	
	// Must be 13 characters (234 + 10 digits)
	if len(normalized) != 13 {
		return false
	}
	
	// Must start with 234
	if !strings.HasPrefix(normalized, "234") {
		return false
	}
	
	// Must be able to detect network
	_, err := DetectNetwork(phoneNumber)
	return err == nil
}
