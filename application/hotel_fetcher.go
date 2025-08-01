package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"hotelsdatapipeline/config/domain"
)

type HotelFetcher struct {
	repository   domain.HotelRepository
	client       *http.Client
	supplierURLs []string
}

func NewHotelFetcher(repository domain.HotelRepository, supplierURLs []string) *HotelFetcher {
	return &HotelFetcher{
		repository: repository,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		supplierURLs: supplierURLs,
	}
}

func (hf *HotelFetcher) FetchAndProcess() error {
	log.Println("Starting hotel data fetch from suppliers...")
	startTime := time.Now()

	hotelsBySupplier := make(map[string][]*domain.Hotel)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var fetchErrors []error

	for _, url := range hf.supplierURLs {
		wg.Add(1)
		go func(supplierURL string) {
			defer wg.Done()

			hotels, err := hf.fetchFromSupplier(supplierURL)
			if err != nil {
				log.Printf("Failed to fetch from %s: %v", supplierURL, err)
				mu.Lock()
				fetchErrors = append(fetchErrors, err)
				mu.Unlock()
				return
			}

			mu.Lock()
			hotelsBySupplier[supplierURL] = hotels
			mu.Unlock()

			log.Printf("Successfully fetched %d hotels from %s", len(hotels), supplierURL)
		}(url)
	}

	wg.Wait()

	if len(hotelsBySupplier) == 0 {
		return fmt.Errorf("no data fetched from any supplier")
	}

	mergedHotels := hf.mergeHotelsByID(hotelsBySupplier)

	if err := hf.storeHotels(mergedHotels); err != nil {
		return fmt.Errorf("failed to store hotels: %w", err)
	}

	duration := time.Since(startTime)
	log.Printf("Hotel data processing completed in %v. Processed %d hotels from %d suppliers",
		duration, len(mergedHotels), len(hotelsBySupplier))

	return nil
}

func (hf *HotelFetcher) fetchFromSupplier(url string) ([]*domain.Hotel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := hf.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var hotels []*domain.Hotel
	if err := json.NewDecoder(resp.Body).Decode(&hotels); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	for _, hotel := range hotels {
		hotel.CleanData()
	}

	return hotels, nil
}

func (hf *HotelFetcher) mergeHotelsByID(hotelsBySupplier map[string][]*domain.Hotel) map[string]*domain.Hotel {
	mergedHotels := make(map[string]*domain.Hotel)

	for supplierURL, hotels := range hotelsBySupplier {
		for _, hotel := range hotels {
			if hotel.HotelID == "" {
				log.Printf("Skipping hotel with empty ID from %s", supplierURL)
				continue
			}

			if existing, exists := mergedHotels[hotel.HotelID]; exists {
				existing.MergeWith(hotel)
				log.Printf("Merged hotel %s from %s", hotel.HotelID, supplierURL)
			} else {
				mergedHotels[hotel.HotelID] = hotel
				log.Printf("Added new hotel %s from %s", hotel.HotelID, supplierURL)
			}
		}
	}

	return mergedHotels
}

func (hf *HotelFetcher) storeHotels(hotels map[string]*domain.Hotel) error {
	hotelsByDestination := make(map[int][]*domain.Hotel)

	for _, hotel := range hotels {
		if err := hotel.Validate(); err != nil {
			log.Printf("Skipping invalid hotel %s: %v", hotel.HotelID, err)
			continue
		}

		if err := hf.repository.StoreHotelByID(hotel.HotelID, hotel); err != nil {
			log.Printf("Failed to store hotel %s: %v", hotel.HotelID, err)
		}

		hotelsByDestination[hotel.DestinationID] = append(hotelsByDestination[hotel.DestinationID], hotel)
	}

	for destinationID, destinationHotels := range hotelsByDestination {
		if err := hf.repository.StoreHotelsByDestinationID(destinationID, destinationHotels); err != nil {
			log.Printf("Failed to store hotels for destination %d: %v", destinationID, err)
		}
	}

	return nil
}
