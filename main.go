package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	"gopkg.in/yaml.v3"
)

type Record struct {
	Kind    string   `yaml:"kind"`
	Name    string   `yaml:"name"`
	Rrdatas []string `yaml:"rrdatas"`
	TTL     int64    `yaml:"ttl"`
	Type    string   `yaml:"type"`
}

func StringsToCtyStrings(s []string) []cty.Value {
	var ctyStrings []cty.Value
	for _, str := range s {
		ctyStrings = append(ctyStrings, cty.StringVal(str))
	}
	return ctyStrings
}

func (r Record) ToTerraformBlock(zoneName string) *hclwrite.Block {
	name := strings.ReplaceAll(fmt.Sprintf("_%s%s", r.Name, strings.ToLower(r.Type)), ".", "-")
	name = strings.ReplaceAll(name, "@", "_at_")
	name = strings.ReplaceAll(name, "*", "_ast_")
	block := hclwrite.NewBlock("resource", []string{"yandex_dns_recordset", name})

	body := block.Body()
	body.SetAttributeRaw("zone_id", hclwrite.TokensForIdentifier(fmt.Sprintf("yandex_dns_zone.%s.id", zoneName)))
	body.SetAttributeValue("name", cty.StringVal(r.Name))
	body.SetAttributeValue("type", cty.StringVal(r.Type))
	body.SetAttributeValue("ttl", cty.NumberIntVal(r.TTL))
	body.SetAttributeValue("data", cty.TupleVal(StringsToCtyStrings(r.Rrdatas)))
	return block
}

type Zone []Record

// Domain returns the domain name of the zone.
// Iterate through the records and find the record with shortest name
func (z Zone) Domain() string {
	shortest := z[0].Name
	for _, r := range z {
		if len(r.Name) < len(shortest) {
			shortest = r.Name
		}
	}
	return shortest
}

func (z Zone) ToTerraformBlock(zoneName string) *hclwrite.Block {
	zone := hclwrite.NewBlock("resource",
		[]string{"yandex_dns_zone", zoneName})
	domain := z.Domain()

	zoneBody := zone.Body()
	zoneBody.SetAttributeValue("name", cty.StringVal(zoneName))
	zoneBody.SetAttributeValue("description", cty.StringVal(fmt.Sprintf("Domain zone for %s", domain)))
	zoneBody.SetAttributeValue("zone", cty.StringVal(domain))
	zoneBody.SetAttributeValue("public", cty.True)

	return zone
}

func readZone(filename string) (Zone, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var records []Record

	decoder := yaml.NewDecoder(bytes.NewBuffer(buf))
	for {
		var d Record
		if err := decoder.Decode(&d); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("document decode failed: %w", err)
		}
		records = append(records, d)
	}

	return Zone(records), nil
}

func main() {

	dnsDir := flag.String("dns-dir", "./dns", "Directory with zone files")
	tfDir := flag.String("tf-dir", "./tf", "Directory to put terraform files")
	skipTypes := flag.String("skip-types", "ns,soa", "Comma separated list of record types to skip for root domain")

	flag.Parse()

	types := strings.Split(strings.ToLower(*skipTypes), ",")
	typesMap := make(map[string]bool)
	for _, t := range types {
		typesMap[t] = true
	}

	files, err := ioutil.ReadDir(*dnsDir)
	if err != nil {
		log.Fatal(err)
	}

	zones := make([]string, 0, len(files))

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		zones = append(zones, f.Name())
	}

	for _, zoneName := range zones {

		zone, err := readZone(path.Join(*dnsDir, zoneName))
		if err != nil {
			log.Fatal(err)
		}

		// create new empty hcl file object
		hclFile := hclwrite.NewEmptyFile()

		tfFile, err := os.Create(path.Join(*tfDir, fmt.Sprintf("%s.tf", zoneName)))
		if err != nil {
			fmt.Println(err)
			return
		}

		domain := zone.Domain()

		rootBody := hclFile.Body()

		rootBody.AppendBlock(zone.ToTerraformBlock(zoneName))

		rootBody.AppendNewline()
		for _, record := range zone {
			if typesMap[strings.ToLower(record.Type)] && record.Name == domain {
				continue
			}
			rootBody.AppendBlock(record.ToTerraformBlock(zoneName))
			rootBody.AppendNewline()
		}

		tfFile.Write(hclFile.Bytes())
		tfFile.Close()
	}
}
