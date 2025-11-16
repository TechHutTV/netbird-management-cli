// geo-locations.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
)

// handleGeoLocationsCommand routes geo-location-related commands
func handleGeoLocationsCommand(client *Client, args []string) error {
	// Create a new flag set for the 'geo' command
	geoCmd := flag.NewFlagSet("geo", flag.ContinueOnError)
	geoCmd.SetOutput(os.Stderr)
	geoCmd.Usage = printGeoLocationUsage

	// Define the flags for the 'geo' command
	countriesFlag := geoCmd.Bool("countries", false, "List all country codes")
	citiesFlag := geoCmd.Bool("cities", false, "List cities in a country")

	// Filters
	countryFlag := geoCmd.String("country", "", "Country code (ISO 3166-1 alpha-2, e.g., DE, US)")

	// Output
	outputFlag := geoCmd.String("output", "table", "Output format: table or json")

	// If no flags are provided (just 'netbird-manage geo'), show usage
	if len(args) == 1 {
		printGeoLocationUsage()
		return nil
	}

	// Parse the flags (all args *after* 'geo')
	if err := geoCmd.Parse(args[1:]); err != nil {
		return nil
	}

	// Handle the flags in priority order

	// List countries
	if *countriesFlag {
		return client.listCountryCodes(*outputFlag)
	}

	// List cities
	if *citiesFlag {
		if *countryFlag == "" {
			return fmt.Errorf("--country is required when using --cities")
		}
		return client.listCitiesByCountry(*countryFlag, *outputFlag)
	}

	geoCmd.Usage()
	return nil
}

// listCountryCodes lists all country codes
func (c *Client) listCountryCodes(outputFormat string) error {
	resp, err := c.makeRequest("GET", "/locations/countries", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var countries []CountryCode
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
	fmt.Fprintln(w, "COUNTRY CODE")
	fmt.Fprintln(w, "------------")
	for _, country := range countries {
		fmt.Fprintf(w, "%s\n", country.Code)
	}
	w.Flush()

	return nil
}

// listCitiesByCountry lists cities in a specific country
func (c *Client) listCitiesByCountry(countryCode string, outputFormat string) error {
	endpoint := fmt.Sprintf("/locations/countries/%s/cities", countryCode)
	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var cities []City
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

// printGeoLocationUsage prints usage information for the geo command
func printGeoLocationUsage() {
	fmt.Println("Usage: netbird-manage geo [options]")
	fmt.Println("\nRetrieve geographic location data for posture checks")
	fmt.Println("\nOptions:")
	fmt.Println("  --countries                List all country codes (ISO 3166-1 alpha-2)")
	fmt.Println("  --cities                   List cities in a country (requires --country)")
	fmt.Println("  --country <code>           Country code (e.g., DE, US, FR)")
	fmt.Println("  --output <format>          Output format: table or json")
	fmt.Println("\nExamples:")
	fmt.Println("  netbird-manage geo --countries")
	fmt.Println("  netbird-manage geo --cities --country DE")
	fmt.Println("  netbird-manage geo --cities --country US --output json")
}
