package microsoft

import "encoding/xml"

type xmlString struct {
	XMLName   xml.Name `xml:"string"`
	Namespace string   `xml:"xmlns,attr"`
	Value     string   `xml:",innerxml"`
}

func newXMLString(value string) *xmlString {
	return &xmlString{
		Namespace: "http://schemas.microsoft.com/2003/10/Serialization/",
		Value:     value,
	}
}

type xmlArrayOfStrings struct {
	XMLName           xml.Name `xml:"ArrayOfstring"`
	Namespace         string   `xml:"xmlns,attr"`
	InstanceNamespace string   `xml:"xmlns:i,attr"`
	Strings           []string `xml:"string"`
}

func newXMLArrayOfStrings(values []string) *xmlArrayOfStrings {
	return &xmlArrayOfStrings{
		Namespace:         "http://schemas.microsoft.com/2003/10/Serialization/Arrays",
		InstanceNamespace: "http://www.w3.org/2001/XMLSchema-instance",
		Strings:           values,
	}
}
