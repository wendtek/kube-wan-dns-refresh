package aws

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	r53Types "github.com/aws/aws-sdk-go-v2/service/route53/types"

	"github.com/wendtek/kube-wan-dns-refresh/pkg/config"
)

type RecordZoneRelation struct {
	Type         string
	RecordName   string
	ZoneId       string
	ShouldUpsert bool
}

type Route53Client interface {
	ListHostedZones(context.Context, *route53.ListHostedZonesInput, ...func(*route53.Options)) (*route53.ListHostedZonesOutput, error)
	ListResourceRecordSets(context.Context, *route53.ListResourceRecordSetsInput,  ...func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error)
	ChangeResourceRecordSets(context.Context, *route53.ChangeResourceRecordSetsInput, ...func(*route53.Options)) (*route53.ChangeResourceRecordSetsOutput, error)
}

// SyncRecords will look at all the records in the config and ensure the state in Route53 matches
func SyncRecords(ctx context.Context, cfg *config.Config, ip string, r53Client Route53Client) error {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))

	// Get the hosted zones
	// TODO: Implement pagination logic
	resp, err := r53Client.ListHostedZones(context.Background(), &route53.ListHostedZonesInput{})
	if err != nil {
		return err
	}

	// Sort Hosted zones by descending length. This ensures if we have multiple matching zones,
	// we'll match the most specific zone for the record we are updating.
	sort.Slice(resp.HostedZones, func(i, j int) bool {
		return len(*resp.HostedZones[i].Name) > len(*resp.HostedZones[j].Name)
	})

	recordsToSync := []RecordZoneRelation{}

	// Match hosted zones to the records in the config
	for _, name := range cfg.Route53Records.A {
		matched := false
		for _, zone := range resp.HostedZones {
			if isCorrectZone(name, *zone.Name) {
				logger.Info(fmt.Sprintf("Matched name %s to zone %s", name, *zone.Name))
				recordsToSync = append(recordsToSync, RecordZoneRelation{
					Type:       "A",
					RecordName: name,
					ZoneId:     *zone.Id,
				})
				matched = true
				continue
			}
		}
		if !matched {
			logger.Info(fmt.Sprintf("No matching zone for %s", name))
		}
	}

	// For each record with a matching zone, check if the record is correct in the zone
	for i, record := range recordsToSync {
		records, err := r53Client.ListResourceRecordSets(context.Background(), &route53.ListResourceRecordSetsInput{
			HostedZoneId: &record.ZoneId,
		})
		if err != nil {
			return err
		}

		trimmedLocalName := strings.TrimSuffix(record.RecordName, ".")
		matched := false
		for _, r := range records.ResourceRecordSets {
			trimmedRemoteName := strings.TrimSuffix(*r.Name, ".")
			if trimmedLocalName == trimmedRemoteName && *r.ResourceRecords[0].Value == ip {
				matched = true
			}
		}
		// If record doesn't exist or isn't correct, create or update it
		if !matched {
			recordsToSync[i].ShouldUpsert = true
		}
	}

	// Divide recordsToUpdate by zone and update each zone as a batch update
	for _, zone := range resp.HostedZones {
		logger.Debug(fmt.Sprintf("Updating zone %s", *zone.Name))
		changes := []r53Types.Change{}
		for _, record := range recordsToSync {
			if record.ZoneId == *zone.Id && record.ShouldUpsert {
				ttl := int64(60)
				changes = append(changes, r53Types.Change{
					Action: r53Types.ChangeActionUpsert,
					ResourceRecordSet: &r53Types.ResourceRecordSet{
						Name: &record.RecordName,
						Type: r53Types.RRTypeA,
						TTL:  &ttl,
						ResourceRecords: []r53Types.ResourceRecord{
							{
								Value: &ip,
							},
						},
					},
				})
			}
		}
		if len(changes) > 0 && !cfg.DryRun {
			logger.Info(fmt.Sprintf("Updating %s with %v records", *zone.Name, len(changes)))
			_, err := r53Client.ChangeResourceRecordSets(context.Background(), &route53.ChangeResourceRecordSetsInput{
				HostedZoneId: zone.Id,
				ChangeBatch: &r53Types.ChangeBatch{
					Changes: changes,
				},
			})
			if err != nil {
				return err
			}
		} else if cfg.DryRun {
			logger.Info(fmt.Sprintf("Dry run: Would have updated %s with %v records", *zone.Name, len(changes)))
		}
	}

	logger.Info("Successfully synced Route53 records.")
	return nil
}

// isCorrectZone returns whether the given A record name may be created in the given zone.
// Correct match examples:
// - Name "example.com"     matches Zone "example.com."
// - Name "sub.example.com" matches Zone "example.com."
// - Name "sub.example.com" matches Zone "sub.example.com."
func isCorrectZone(name, zoneName string) bool {
	trimmedZoneName := strings.TrimSuffix(zoneName, ".")
	trimmedName := strings.TrimSuffix(name, ".")
	return strings.HasSuffix(trimmedName, trimmedZoneName)
}
