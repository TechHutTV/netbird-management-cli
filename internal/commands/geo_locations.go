// geo_locations.go
package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"netbird-manage/internal/models"
)

// HandleGeoLocationsCommand routes geo-location-related commands
func (s *Service) HandleGeoLocationsCommand(args []string) error {
	// Create a new flag set for the 'geo' command
	geoCmd := flag.NewFlagSet("geo", flag.ContinueOnError)
	geoCmd.SetOutput(os.Stderr)
	geoCmd.Usage = PrintGeoLocationUsage

	// Define the flags for the 'geo' command
	countriesFlag := geoCmd.Bool("countries", false, "List all country codes")
	citiesFlag := geoCmd.Bool("cities", false, "List cities in a country")

	// Filters
	countryFlag := geoCmd.String("country", "", "Country code (ISO 3166-1 alpha-2, e.g., DE, US)")

	// Output
	outputFlag := geoCmd.String("output", "table", "Output format: table or json")

	// If no flags are provided (just 'netbird-manage geo'), show usage
	if len(args) == 1 {
		PrintGeoLocationUsage()
		return nil
	}

	// Parse the flags (all args *after* 'geo')
	if err := geoCmd.Parse(args[1:]); err != nil {
		return nil
	}

	// Handle the flags in priority order

	// List countries
	if *countriesFlag {
		return s.listCountryCodes(*outputFlag)
	}

	// List cities
	if *citiesFlag {
		if *countryFlag == "" {
			return fmt.Errorf("--country is required when using --cities")
		}
		return s.listCitiesByCountry(*countryFlag, *outputFlag)
	}

	geoCmd.Usage()
	return nil
}

// listCountryCodes lists all country codes
func (s *Service) listCountryCodes(outputFormat string) error {
	resp, err := s.Client.MakeRequest("GET", "/locations/countries", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var countries []models.CountryCode
	if err := json.NewDecoder(resp.Body).Decode(&countries); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// JSON output
	if outputFormat == "json" {
		output, err := json.MarshalIndent(countries, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	// Table output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "CODE\tCOUNTRY NAME")
	fmt.Fprintln(w, "----\t------------")
	for _, country := range countries {
		fmt.Fprintf(w, "%s\t%s\n", country.Code, country.Name)
	}
	w.Flush()

	return nil
}

// listCitiesByCountry lists cities in a specific country
func (s *Service) listCitiesByCountry(countryCode string, outputFormat string) error {
	endpoint := fmt.Sprintf("/locations/countries/%s/cities", countryCode)
	resp, err := s.Client.MakeRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var cities []models.City
	if err := json.NewDecoder(resp.Body).Decode(&cities); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// JSON output
	if outputFormat == "json" {
		output, err := json.MarshalIndent(cities, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	// Table output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "GEONAME ID\tCITY NAME")
	fmt.Fprintln(w, "----------\t---------")
	for _, city := range cities {
		fmt.Fprintf(w, "%d\t%s\n", city.GeonameID, city.CityName)
	}
	w.Flush()

	return nil
}
