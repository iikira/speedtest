package speedtestclient

import (
	"encoding/xml"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcstable"
	"github.com/iikira/speedtest/speedtestutil/xmlhelper"
	"io"
	"strconv"
	"unsafe"
)

type (
	speedtestConfigClient struct {
		XMLName   xml.Name `xml:"client"`
		IP        string   `xml:"ip,attr"`
		Lat       float64  `xml:"lat,attr"`
		Lon       float64  `xml:"lon,attr"`
		ISP       string   `xml:"isp,attr"`
		ISPID     int      `xml:"ispid,attr"`
		Carrier   string   `xml:"carrier,attr"`
		CarrierID int      `xml:"carrierid,attr"`
		LatestVer string   `xml:"latestver,attr"`
	}

	speedtestConfigServer struct {
		XMLName xml.Name `xml:"server"`
		ID      int      `xml:"id,attr"`
		Name    string   `xml:"name,attr"`
		Sponsor string   `xml:"sponsor,attr"`
		Lat     float64  `xml:"lat,attr"`
		Lon     float64  `xml:"lon,attr"`
		Host    string   `xml:"host,attr"`
	}

	speedtestConfigSettings struct {
		XMLName xml.Name                 `xml:"settings"`
		Client  speedtestConfigClient    `xml:"client"`
		Servers []*speedtestConfigServer `xml:"servers>server"`
	}

	LocalInfo struct {
		_         xml.Name
		IP        string
		Lat       float64
		Lon       float64
		ISP       string
		ISPID     int
		Carrier   string
		CarrierID int
		LatestVer string
	}

	SpeedtestServer struct {
		_       xml.Name
		ID      int
		Name    string
		Sponsor string
		Lat     float64
		Lon     float64
		Host    string
	}

	SpeedtestServerList []*SpeedtestServer
)

func (sc *SpeedtestClient) GetLocalInfoAndServerList() (li *LocalInfo, servList SpeedtestServerList, err error) {
	sc.lazyInit()
	u := sc.genURL("/api/android/config.php", map[string]interface{}{
		"pt":                    1,
		"gaidOptOut":            "false",
		"gaid":                  "93521673-7cf5-411a-984d-278941c1a5a6",
		"sim_operator_name":     "China Unicom",
		"manufacturer":          "OnePlus",
		"network_operator":      "46001",
		"generalizedSize":       2,
		"dct":                   0,
		"ypixels":               2135,
		"fingerprint":           "OnePlus/OnePlus6T/OnePlus6T:9/PKQ1.180716.001/1904161800:user/release-keys",
		"model":                 "ONEPLUS A6010",
		"network_operator_name": "CHN-UNICOM",
		"brand":                 "OnePlus",
		"xdpi":                  "403.411",
		"hardware":              "qcom",
		"product":               "OnePlus6T",
		"orientation":           1,
		"sim_operator":          46002,
		"appversion":            "4.2.4",
		"carriers":              46002,
		"android_api":           28,
		"xpixels":               1080,
		"appversion_extended":   "4.2.4.47568",
		"build_id":              "PKQ1.180716.001",
		"imei":                  "869386056656016",
		"coord_src":             1,
		"ni":                    13,
		"ydpi":                  "409.903",
		"generalizedDpi":        420,
		"device":                "OnePlus6T",
	})

	resp, err := sc.hc.Req("GET", u.String(), nil, nil)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return
	}

	settings := speedtestConfigSettings{}
	err = xmlhelper.UnmarshalXMLData(resp.Body, &settings)
	if err != nil {
		return
	}

	li = (*LocalInfo)(unsafe.Pointer(&settings.Client))
	servList = *(*SpeedtestServerList)(unsafe.Pointer(&settings.Servers))
	return
}

func (sc *SpeedtestClient) GetAllServerList() (servList SpeedtestServerList, err error) {
	sc.lazyInit()
	u := sc.genURL("/speedtest-servers-static.php", map[string]interface{}{
		"x": "111",
	})

	resp, err := sc.hc.Req("GET", u.String(), nil, nil)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return
	}

	settings := speedtestConfigSettings{}
	err = xmlhelper.UnmarshalXMLData(resp.Body, &settings)
	if err != nil {
		return
	}

	servList = *(*SpeedtestServerList)(unsafe.Pointer(&settings.Servers))
	return
}

func (server *SpeedtestServer) String() string {
	return fmt.Sprintf("ID: %d, Name: %s, Sponsor: %s, Latitude: %f, Longtitude: %f, Host: %s", server.ID, server.Name, server.Sponsor, server.Lat, server.Lon, server.Host)
}

func (li *LocalInfo) PrintTo(w io.Writer) {
	fmt.Fprintf(w, "IP: %s, Lat: %f, Lon: %f, ISP: %s, ISPID: %d, Carrier: %s, CarrierID: %d, LatestVer: %s\n", li.IP, li.Lat, li.Lon, li.ISP, li.ISPID, li.Carrier, li.CarrierID, li.LatestVer)
}

func (servList SpeedtestServerList) PrintTo(w io.Writer) {
	table := pcstable.NewTable(w)
	table.SetHeader([]string{"ID", "NAME", "SPONSOR", "LATITUDE", "LONGTITUDE", "HOST"})
	for _, v := range servList {
		table.Append([]string{strconv.Itoa(v.ID), v.Name, v.Sponsor, strconv.FormatFloat(v.Lat, 'f', -1, 64), strconv.FormatFloat(v.Lon, 'f', -1, 64), v.Host})
	}
	table.Render()
	return
}

func (servList SpeedtestServerList) FindByID(id int) (server *SpeedtestServer) {
	for _, v := range servList {
		if v == nil {
			continue
		}
		if v.ID == id {
			return v
		}
	}
	return nil
}
