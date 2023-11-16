package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime"

	"github.com/qedus/osmpbf"
)

func unique(intSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func main() {
	argsWithProg := os.Args

	f, err := os.Open(argsWithProg[1])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	d := osmpbf.NewDecoder(f)

	// use more memory from the start, it is faster
	d.SetBufferSize(osmpbf.MaxBlobSize)

	// start decoding with several goroutines, it is faster
	err = d.Start(runtime.GOMAXPROCS(-1))
	if err != nil {
		log.Fatal(err)
	}

	cityToZip := make(map[string]map[string][]string)
	zipToCity := make(map[string]map[string][]string)

	for {
		if v, err := d.Decode(); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		} else {
			switch v := v.(type) {
			case *osmpbf.Node:
				if city, ok := v.Tags["addr:city"]; ok {
					if country, ok := v.Tags["addr:country"]; ok {
						if cityToZip[country] == nil {
							cityToZip[country] = make(map[string][]string)
							zipToCity[country] = make(map[string][]string)
						}
						if zip, ok := v.Tags["postal_code"]; ok {
							cityToZip[country][city] = append(cityToZip[country][city], zip)
							zipToCity[country][zip] = append(zipToCity[country][zip], city)
						} else if zip, ok := v.Tags["addr:postcode"]; ok {
							cityToZip[country][city] = append(cityToZip[country][city], zip)
							zipToCity[country][zip] = append(zipToCity[country][zip], city)
						}
					}
				}
			}
		}
	}

	for country, cities := range cityToZip {
		for city, zips := range cities {
			cityToZip[country][city] = unique(zips)
		}
		file, _ := json.Marshal(cityToZip[country])
		_ = ioutil.WriteFile("city_to_zip-"+country+".json", file, 0644)
	}

	for country, zips := range zipToCity {
		for zip, cities := range zips {
			zipToCity[country][zip] = unique(cities)
		}
		file, _ := json.Marshal(zipToCity[country])
		_ = ioutil.WriteFile("zip_to_city-"+country+".json", file, 0644)
	}

	file, _ := json.Marshal(cityToZip)
	_ = ioutil.WriteFile("city_to_zip-global.json", file, 0644)

	file2, _ := json.Marshal(zipToCity)
	_ = ioutil.WriteFile("zip_to_city-global.json", file2, 0644)
}
