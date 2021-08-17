package service

import "testing"

func TestWMTSServiceGetCapabilities(t *testing.T) {

}

func TestWMTSServiceGetTile(t *testing.T) {

}

func TestWMTSServiceGetFeatureInfo(t *testing.T) {

}

func TestWMTSRestService(t *testing.T) {

}

func TestWMTSCapabilities(t *testing.T) {

	service := make(map[string]string)
	service["url"] = "http://flywave.net"
	service["title"] = "flywave"
	service["abstract"] = ""

	service["serviceprovider.providername"] = "flywave"
	service["serviceprovider.providersite.type"] = "wms"
	service["serviceprovider.providersite.href"] = "http://flywave.net"
	service["serviceprovider.servicecontact.individualname"] = "test"
	service["serviceprovider.servicecontact.positionname"] = "test"
	service["serviceprovider.servicecontact.contactinfo.text"] = "test"
	service["serviceprovider.servicecontact.contactinfo.phone.voice"] = "1288766553"
	service["serviceprovider.servicecontact.contactinfo.phone.facsimile"] = "1288766553"
	service["serviceprovider.servicecontact.contactinfo.address.deliverypoint"] = "1288766553"
	service["serviceprovider.servicecontact.contactinfo.address.city"] = "test"
	service["serviceprovider.servicecontact.contactinfo.address.administrativearea"] = "1288766553"
	service["serviceprovider.servicecontact.contactinfo.address.postalcode"] = "1288766553"
	service["serviceprovider.servicecontact.contactinfo.address.country"] = "test"
	service["serviceprovider.servicecontact.contactinfo.address.electronicmailaddress"] = "test"
	service["serviceprovider.servicecontact.contactinfo.onlineresource.type"] = "test"
	service["serviceprovider.servicecontact.contactinfo.onlineresource.href"] = "http://flywave.net"
	service["serviceprovider.servicecontact.contactinfo.hoursofservice"] = "test"
	service["serviceprovider.servicecontact.contactinfo.contactinstructions"] = "test"
	service["serviceprovider.servicecontact.role"] = "test"

}

func TestTileMatrixSet(t *testing.T) {

}
