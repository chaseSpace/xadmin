package namedotcom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/libdns/libdns"
)

// nameClient extends the namedotcom api and request handler to the provider..
type nameClient struct {
	client *nameDotCom
	mutex  sync.Mutex
}

// getClient initiates a new nameClient and assigns it to the provider..
func (p *Provider) getClient(ctx context.Context) error {
	newNameclient, err := NewNameDotComClient(ctx, p.Token, p.User, p.Server)
	if err != nil {
		return err
	}
	p.client = newNameclient
	return nil
}

// listAllRecords returns all records for the given zone .. GET /v4/domains/{ domainName }/records
func (p *Provider) listAllRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	var (
		records []libdns.Record

		/*** 'zone' args that are passed in using compliant zone formats have the FQDN '.' suffix qualifier
		and in order to use the zone arg as a domainName reference to name.com's api we must remove the '.' suffix.
		otherwise the api will not recognize the domain.. ***/
		unFQDNzone = strings.TrimSuffix(zone, ".")

		method  = "GET"
		body    io.Reader
		resp    = &listRecordsResponse{}
		reqPage = 1

		err error
	)

	if err = p.getClient(ctx); err != nil {
		return nil, err
	}

	// handle pagination, in case domain has more records then the default of 1000 per page
	for reqPage > 0 {
		if reqPage != 0 {
			endpoint := fmt.Sprintf("/v4/domains/%s/records?page=%d", unFQDNzone, reqPage)

			if body, err = p.client.doRequest(ctx, method, endpoint, nil); err != nil {
				return nil, fmt.Errorf("request failed:  %w", err)
			}

			if err = json.NewDecoder(body).Decode(resp); err != nil {
				return nil, fmt.Errorf("could not decode name.com's response:  %w", err)
			}

			for _, record := range resp.Records {
				records = append(records, record.toLibDNSRecord())
			}

			reqPage = int(resp.NextPage)
		}
	}

	return records, nil
}

// deleteRecord  DELETE /v4/domains/{ domainName }/records/{ record.ID }
func (p *Provider) deleteRecord(ctx context.Context, zone string, record libdns.Record) (libdns.Record, error) {
	id := recordID(record)
	if id == "" {
		var err error
		id, err = p.findRecordID(ctx, zone, record)
		if err != nil {
			return nil, err
		}
	}

	var (
		shouldDelete nameDotComRecord
		unFQDNzone   = strings.TrimSuffix(zone, ".")

		method   = "DELETE"
		endpoint = fmt.Sprintf("/v4/domains/%s/records/%s", unFQDNzone, id)
		body     io.Reader
		post     = &bytes.Buffer{}

		err error
	)

	shouldDelete.fromLibDNSRecord(record, unFQDNzone)
	shouldDelete.setIDFromString(id)

	p.mutex.Lock()
	defer p.mutex.Unlock()

	if err = p.getClient(ctx); err != nil {
		return nil, err
	}

	if err = json.NewEncoder(post).Encode(shouldDelete); err != nil {
		return nil, fmt.Errorf("could not encode form data for request:  %w", err)
	}

	if body, err = p.client.doRequest(ctx, method, endpoint, post); err != nil {
		return nil, fmt.Errorf("request to delete the record was not successful:  %w", err)
	}

	if err = json.NewDecoder(body).Decode(&shouldDelete); err != nil {
		return nil, fmt.Errorf("could not decode the response from name.com:  %w", err)
	}

	return shouldDelete.toLibDNSRecord(), nil
}

// upsertRecord  PUT || POST /v4/domains/{ domainName }/records/{ record.ID }
func (p *Provider) upsertRecord(ctx context.Context, zone string, canidateRecord libdns.Record) (libdns.Record, error) {
	id := recordID(canidateRecord)

	p.mutex.Lock()
	defer p.mutex.Unlock()

	var (
		shouldUpsert nameDotComRecord
		unFQDNzone   = strings.TrimSuffix(zone, ".")

		method   = "PUT"
		endpoint = fmt.Sprintf("/v4/domains/%s/records/%s", unFQDNzone, id)
		body     io.Reader
		post     = &bytes.Buffer{}

		err error
	)

	if id == "" {
		method = "POST"
		endpoint = fmt.Sprintf("/v4/domains/%s/records", unFQDNzone)
	}

	shouldUpsert.fromLibDNSRecord(canidateRecord, unFQDNzone)
	shouldUpsert.setIDFromString(id)

	if err = p.getClient(ctx); err != nil {
		return nil, err
	}

	if err = json.NewEncoder(post).Encode(shouldUpsert); err != nil {
		return nil, fmt.Errorf("could not encode the form data for the request:  %w", err)
	}

	if body, err = p.client.doRequest(ctx, method, endpoint, post); err != nil {
		if strings.Contains(err.Error(), "Duplicate Record") {
			err = fmt.Errorf("name.com will not allow an update to a record that has identical values to an existing record: %w", err)
		}

		return nil, fmt.Errorf("request to update the record was not successful:  %w", err)
	}

	if err = json.NewDecoder(body).Decode(&shouldUpsert); err != nil {
		return nil, fmt.Errorf("could not decode name.com's response:  %w", err)
	}

	return shouldUpsert.toLibDNSRecord(), nil
}
func (p *Provider) findRecordID(ctx context.Context, zone string, record libdns.Record) (string, error) {
	records, err := p.listAllRecords(ctx, zone)
	if err != nil {
		return "", err
	}

	rr := record.RR()
	targetHost := sanitizeName(rr.Name, zone)
	for _, rec := range records {
		switch typed := rec.(type) {
		case nameDotComRecord:
			if isMatchingRecord(rr, targetHost, typed) {
				return fmt.Sprint(typed.ID), nil
			}
		case *nameDotComRecord:
			if isMatchingRecord(rr, targetHost, *typed) {
				return fmt.Sprint(typed.ID), nil
			}
		}
	}

	return "", fmt.Errorf("name.com record %s (%s) not found", targetHost, rr.Type)
}

func isMatchingRecord(rr libdns.RR, host string, record nameDotComRecord) bool {
	if record.Type != rr.Type || record.Host != host {
		return false
	}
	return record.Answer == rr.Data
}
