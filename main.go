package main

import (
	"crypto/sha256"
	"embed"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/gosimple/slug"
)

type DnsEntry struct {
	Host       string
	HostSlug   string
	TTL        string
	RecordType string
	Value      string
}

type Entry struct {
	DnsEntry DnsEntry
	Zone     string
	ZoneSlug string
	Hash     string
}

//go:embed templates/*.tml
var res embed.FS

func main() {
	csvFile, err := os.Open("dns.csv")
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()

	output, err := os.Create("12-dns.tf")
	if err != nil {
		panic(err)
	}

	output.WriteString("# DON'T CHANGE THIS FILE MANUALLY. IT IS GENERATED BY THE COLONIZER TOOL.\n\n")
	output.WriteString("# https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/dns_managed_zone\n")

	zoneTemplate := template.Must(template.ParseFS(res, "templates/zone.tml"))
	entryTemplate := template.Must(template.ParseFS(res, "templates/entry.tml"))

	var zoneCount, entryCount int
	zones := make(Set)
	entry := Entry{}
	dnsEntry := DnsEntry{}
	reader := csv.NewReader(csvFile)
	reader.Read()
	for {
		record, err := reader.Read()
		if err != nil {
			break
		}

		dnsEntry.Host = record[0]
		dnsEntry.HostSlug = slug.Make(record[0])
		dnsEntry.TTL = record[1]
		dnsEntry.RecordType = record[2]
		dnsEntry.Value = record[3]

		if dnsEntry.TTL == "" {
			dnsEntry.TTL = "300"
		}

		entry.DnsEntry = dnsEntry
		entry.Zone = getZone(dnsEntry.Host)
		entry.ZoneSlug = slug.Make(entry.Zone)
		entry.Hash = structHash(dnsEntry)

		if !zones.Contains(entry.Zone) {
			err = zoneTemplate.Execute(output, entry)
			if err != nil {
				panic(err)
			}
			zones.Add(entry.Zone)
			zoneCount++
		}

		err = entryTemplate.Execute(output, entry)
		if err != nil {
			panic(err)
		}
		entryCount++
	}

	fmt.Println("Total zones:", zoneCount)
	fmt.Println("Total entries:", entryCount)
}

func structHash(s DnsEntry) string {
	dataBytes := []byte(fmt.Sprintf("%#v", s))
	hash := sha256.Sum256(dataBytes)
	return hex.EncodeToString(hash[:])
}

func getZone(host string) string {
	domainParts := strings.Split(host, ".")
	secondLevelDomain := domainParts[len(domainParts)-3]
	topLevelDomain := domainParts[len(domainParts)-2]
	return fmt.Sprintf("%s.%s.", secondLevelDomain, topLevelDomain)
}
