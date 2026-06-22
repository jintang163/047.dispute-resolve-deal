package amap

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/logger"
)

type AmapClient struct {
	webKey        string
	webServiceKey string
	securityCode  string
	defaultCity   string
	httpClient    *http.Client
}

var amapClient *AmapClient

func InitAmapClient() {
	cfg := config.GetConfig().Amap
	amapClient = &AmapClient{
		webKey:        cfg.WebKey,
		webServiceKey: cfg.WebServiceKey,
		securityCode:  cfg.SecurityCode,
		defaultCity:   cfg.DefaultCity,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func GetAmapClient() *AmapClient {
	if amapClient == nil {
		InitAmapClient()
	}
	return amapClient
}

type RoutePlanRequest struct {
	Origin      string
	Destination string
	Waypoints   []string
	Strategy    int
}

type RoutePlanResponse struct {
	Status    int    `json:"status"`
	Info      string `json:"info"`
	Infocode  string `json:"infocode"`
	Route     Route  `json:"route"`
}

type Route struct {
	Origin      string      `json:"origin"`
	Destination string      `json:"destination"`
	Paths       []Path      `json:"paths"`
	Waypoints   []Waypoint  `json:"waypoints"`
}

type Path struct {
	Distance string `json:"distance"`
	Duration string `json:"duration"`
	Steps    []Step `json:"steps"`
	Polyline string `json:"polyline"`
}

type Step struct {
	Instruction string `json:"instruction"`
	Distance    string `json:"distance"`
	Duration    string `json:"duration"`
	Polyline    string `json:"polyline"`
	Action      string `json:"action"`
	Road        string `json:"road"`
}

type Waypoint struct {
	Location string `json:"location"`
	City     string `json:"city"`
	District string `json:"district"`
	Shipment int    `json:"shipment"`
}

type GeocodeResponse struct {
	Status   int    `json:"status"`
	Info     string `json:"info"`
	Infocode string `json:"infocode"`
	Geocodes []Geocode `json:"geocodes"`
}

type Geocode struct {
	FormattedAddress string `json:"formatted_address"`
	Location         string `json:"location"`
	Level            string `json:"level"`
	City             string `json:"city"`
	District         string `json:"district"`
	Township         string `json:"township"`
	Adcode           string `json:"adcode"`
}

type RegeocodeResponse struct {
	Status    string  `json:"status"`
	Info      string  `json:"info"`
	Infocode  string  `json:"infocode"`
	Regeocode Regeocode `json:"regeocode"`
}

type Regeocode struct {
	FormattedAddress string `json:"formatted_address"`
	AddressComponent AddressComponent `json:"addressComponent"`
	POIs             []POI            `json:"pois"`
}

type AddressComponent struct {
	Country  string `json:"country"`
	Province string `json:"province"`
	City     string `json:"city"`
	District string `json:"district"`
	Township string `json:"township"`
	Street   string `json:"street"`
	Number   string `json:"number"`
	Adcode   string `json:"adcode"`
}

type POI struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Location string `json:"location"`
	Address  string `json:"address"`
	Distance string `json:"distance"`
}

type DrivingRouteResult struct {
	TotalDistance float64
	TotalDuration int
	Polyline      string
	Steps         []RouteStep
	OrderedPoints []OrderedRoutePoint
}

type RouteStep struct {
	Instruction string
	Distance    float64
	Duration    int
	Polyline    string
}

type OrderedRoutePoint struct {
	ID        int64
	SortOrder int
	Longitude float64
	Latitude  float64
	Distance  float64
	Duration  int
}

func (c *AmapClient) PlanDrivingRoute(originLng, originLat float64, points []OrderedRoutePoint, strategy int) (*DrivingRouteResult, error) {
	if len(points) == 0 {
		return nil, fmt.Errorf("points cannot be empty")
	}

	origin := fmt.Sprintf("%.6f,%.6f", originLng, originLat)
	destination := fmt.Sprintf("%.6f,%.6f", points[len(points)-1].Longitude, points[len(points)-1].Latitude)

	var waypoints string
	if len(points) > 1 {
		wpList := make([]string, len(points)-1)
		for i := 0; i < len(points)-1; i++ {
			wpList[i] = fmt.Sprintf("%.6f,%.6f", points[i].Longitude, points[i].Latitude)
		}
		for i, wp := range wpList {
			if i > 0 {
				waypoints += ";"
			}
			waypoints += wp
		}
	}

	params := url.Values{}
	params.Set("key", c.webServiceKey)
	params.Set("origin", origin)
	params.Set("destination", destination)
	params.Set("strategy", strconv.Itoa(strategy))
	if waypoints != "" {
		params.Set("waypoints", waypoints)
		params.Set("avoidpolygons", "")
		params.Set("driving_style", "1")
	}

	apiURL := fmt.Sprintf("https://restapi.amap.com/v3/direction/driving?%s", params.Encode())

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	if c.securityCode != "" {
		req.Header.Set("Referer", c.securityCode)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Error("Amap driving route request failed", logger.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var routeResp RoutePlanResponse
	if err := json.Unmarshal(body, &routeResp); err != nil {
		logger.Error("Amap response parse failed", logger.Error(err))
		return nil, err
	}

	if routeResp.Status != 1 {
		return nil, fmt.Errorf("amap api error: %s", routeResp.Info)
	}

	if len(routeResp.Route.Paths) == 0 {
		return nil, fmt.Errorf("no route found")
	}

	path := routeResp.Route.Paths[0]
	totalDistance, _ := strconv.ParseFloat(path.Distance, 64)
	totalDuration, _ := strconv.Atoi(path.Duration)

	result := &DrivingRouteResult{
		TotalDistance: totalDistance,
		TotalDuration: totalDuration,
		Polyline:      path.Polyline,
		OrderedPoints: make([]OrderedRoutePoint, 0, len(points)),
	}

	for _, step := range path.Steps {
		dist, _ := strconv.ParseFloat(step.Distance, 64)
		dur, _ := strconv.Atoi(step.Duration)
		result.Steps = append(result.Steps, RouteStep{
			Instruction: step.Instruction,
			Distance:    dist,
			Duration:    dur,
			Polyline:    step.Polyline,
		})
	}

	for i, p := range points {
		p.SortOrder = i + 1
		result.OrderedPoints = append(result.OrderedPoints, p)
	}

	return result, nil
}

func (c *AmapClient) PlanOptimalRoute(originLng, originLat float64, points []OrderedRoutePoint) (*DrivingRouteResult, error) {
	if len(points) <= 2 {
		return c.PlanDrivingRoute(originLng, originLat, points, 10)
	}

	return c.PlanDrivingRoute(originLng, originLat, points, 10)
}

func (c *AmapClient) GetAddressByLocation(lng, lat float64) (string, *Regeocode, error) {
	location := fmt.Sprintf("%.6f,%.6f", lng, lat)

	params := url.Values{}
	params.Set("key", c.webServiceKey)
	params.Set("location", location)
	params.Set("radius", "1000")
	params.Set("extensions", "base")

	apiURL := fmt.Sprintf("https://restapi.amap.com/v3/geocode/regeo?%s", params.Encode())

	resp, err := c.httpClient.Get(apiURL)
	if err != nil {
		logger.Error("Amap regeocode request failed", logger.Error(err))
		return "", nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}

	var regeocodeResp RegeocodeResponse
	if err := json.Unmarshal(body, &regeocodeResp); err != nil {
		logger.Error("Amap regeocode parse failed", logger.Error(err))
		return "", nil, err
	}

	if regeocodeResp.Status != "1" {
		return "", nil, fmt.Errorf("amap api error: %s", regeocodeResp.Info)
	}

	return regeocodeResp.Regeocode.FormattedAddress, &regeocodeResp.Regeocode, nil
}

func (c *AmapClient) GetLocationByAddress(address string, city string) (float64, float64, error) {
	if city == "" {
		city = c.defaultCity
	}

	params := url.Values{}
	params.Set("key", c.webServiceKey)
	params.Set("address", address)
	params.Set("city", city)

	apiURL := fmt.Sprintf("https://restapi.amap.com/v3/geocode/geo?%s", params.Encode())

	resp, err := c.httpClient.Get(apiURL)
	if err != nil {
		logger.Error("Amap geocode request failed", logger.Error(err))
		return 0, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}

	var geocodeResp GeocodeResponse
	if err := json.Unmarshal(body, &geocodeResp); err != nil {
		logger.Error("Amap geocode parse failed", logger.Error(err))
		return 0, 0, err
	}

	if geocodeResp.Status != 1 || len(geocodeResp.Geocodes) == 0 {
		return 0, 0, fmt.Errorf("address not found")
	}

	geocode := geocodeResp.Geocodes[0]
	loc := geocode.Location
	parts := splitLocation(loc)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid location format")
	}

	lng, _ := strconv.ParseFloat(parts[0], 64)
	lat, _ := strconv.ParseFloat(parts[1], 64)

	return lng, lat, nil
}

func splitLocation(loc string) []string {
	return strings.Split(loc, ",")
}

func CalculateDistance(lng1, lat1, lng2, lat2 float64) float64 {
	const earthRadius = 6371000.0

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLng := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}
