package domain

import (
	"fmt"
	"sort"
	"strings"
)

type Hotel struct {
	HotelID           string    `json:"hotel_id"`
	DestinationID     int       `json:"destination_id"`
	HotelName         string    `json:"hotel_name"`
	Location          Location  `json:"location"`
	Details           string    `json:"details"`
	Amenities         Amenities `json:"amenities"`
	Images            Images    `json:"images"`
	BookingConditions []string  `json:"booking_conditions"`
}

type Location struct {
	Address string `json:"address"`
	Country string `json:"country"`
}

type Amenities struct {
	General []string `json:"general"`
	Room    []string `json:"room"`
}

type Image struct {
	Link    string `json:"link"`
	Caption string `json:"caption"`
}

type Images struct {
	Rooms []Image `json:"rooms"`
	Site  []Image `json:"site"`
}

type HotelRepository interface {
	StoreHotelByID(hotelID string, hotel *Hotel) error
	StoreHotelsByDestinationID(destinationID int, hotels []*Hotel) error
	GetHotelByID(hotelID string) (*Hotel, error)
	GetHotelsByDestinationID(destinationID int) ([]*Hotel, error)
	GetHotelsByIDRange(hotelIDs []string) ([]*Hotel, error)
}

func (h *Hotel) CleanData() {
	h.HotelID = strings.TrimSpace(h.HotelID)
	h.HotelName = strings.TrimSpace(h.HotelName)
	h.Location.Address = strings.TrimSpace(h.Location.Address)
	h.Location.Country = strings.TrimSpace(h.Location.Country)
	h.Details = strings.TrimSpace(h.Details)

	h.Amenities.General = cleanStringSlice(h.Amenities.General)
	h.Amenities.Room = cleanStringSlice(h.Amenities.Room)

	h.BookingConditions = cleanStringSlice(h.BookingConditions)

	h.Images.Rooms = cleanImages(h.Images.Rooms)
	h.Images.Site = cleanImages(h.Images.Site)
}

func (h *Hotel) MergeWith(other *Hotel) {
	if strings.TrimSpace(h.HotelName) == "" && strings.TrimSpace(other.HotelName) != "" {
		h.HotelName = other.HotelName
	}
	if strings.TrimSpace(h.Location.Address) == "" && strings.TrimSpace(other.Location.Address) != "" {
		h.Location.Address = other.Location.Address
	}
	if strings.TrimSpace(h.Location.Country) == "" && strings.TrimSpace(other.Location.Country) != "" {
		h.Location.Country = other.Location.Country
	}

	if len(strings.TrimSpace(other.Details)) > len(strings.TrimSpace(h.Details)) {
		h.Details = other.Details
	}

	h.Amenities.General = mergeStringSlices(h.Amenities.General, other.Amenities.General)
	h.Amenities.Room = mergeStringSlices(h.Amenities.Room, other.Amenities.Room)

	h.BookingConditions = mergeStringSlices(h.BookingConditions, other.BookingConditions)

	h.Images.Rooms = mergeImages(h.Images.Rooms, other.Images.Rooms)
	h.Images.Site = mergeImages(h.Images.Site, other.Images.Site)
}

func cleanStringSlice(slice []string) []string {
	var result []string
	for _, s := range slice {
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func cleanImages(images []Image) []Image {
	var result []Image
	for _, img := range images {
		img.Link = strings.TrimSpace(img.Link)
		img.Caption = strings.TrimSpace(img.Caption)
		if img.Link != "" {
			result = append(result, img)
		}
	}
	return result
}

func mergeStringSlices(a, b []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, s := range append(a, b...) {
		lower := strings.ToLower(strings.TrimSpace(s))
		if !seen[lower] && lower != "" {
			seen[lower] = true
			result = append(result, strings.TrimSpace(s))
		}
	}

	sort.Strings(result)
	return result
}

func mergeImages(a, b []Image) []Image {
	seen := make(map[string]bool)
	var result []Image

	for _, img := range append(a, b...) {
		link := strings.TrimSpace(img.Link)
		if !seen[link] && link != "" {
			seen[link] = true
			result = append(result, Image{
				Link:    link,
				Caption: strings.TrimSpace(img.Caption),
			})
		}
	}

	return result
}

func (h *Hotel) Validate() error {
	if h.HotelID == "" {
		return fmt.Errorf("hotel ID is required")
	}
	if h.DestinationID <= 0 {
		return fmt.Errorf("destination ID must be positive")
	}
	return nil
}
