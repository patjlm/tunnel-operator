package controllers

import (
	"context"
	b64 "encoding/base64"
	"errors"
	"math/rand"
	"os"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/go-logr/logr"
	tunnelv1alpha1 "github.com/patjlm/tunnel-operator/api/v1alpha1"
)

type Cloudflare struct {
	ctx       context.Context
	log       logr.Logger
	accountID string
	zoneName  string
	zoneID    string
	api       *cloudflare.API
}

func (c *Cloudflare) NewTunnelSecretB64() string {
	length := 32
	rand.Seed(time.Now().UnixNano())
	chars := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%^&*()-_=+")
	b := make([]rune, length)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return b64.StdEncoding.EncodeToString([]byte(string(b)))
}

func (c *Cloudflare) Api() (*cloudflare.API, error) {
	if c.api != nil {
		return c.api, nil
	}
	c.accountID = os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	token := os.Getenv("CLOUDFLARE_API_TOKEN")
	if token == "" {
		return nil, errors.New("missing environment variable CLOUDFLARE_API_TOKEN")
	}
	api, err := cloudflare.NewWithAPIToken(token)
	if err != nil {
		return api, err
	}
	if api.AccountID == "" {
		api.AccountID = c.accountID
	}
	c.api = api
	c.zoneName = os.Getenv("CLOUDFLARE_ZONE_NAME")
	c.zoneID, _ = api.ZoneIDByName(c.zoneName)
	return api, err
}

func (c *Cloudflare) CreateDNSRecord(recordType string, recordName string, content string, proxied bool) error {
	tpl := cloudflare.DNSRecord{Name: recordName, Type: "CNAME"}
	records, err := c.api.DNSRecords(context.Background(), c.zoneID, tpl)
	if err != nil {
		c.log.Error(err, "failed to retrieve CNAME DNS recods from zone "+c.zoneName)
		return err
	}
	if len(records) == 0 {
		c.log.Info("creating cloudflare CNAME record for " + recordName)
		_, err := c.api.CreateDNSRecord(c.ctx, c.zoneID, cloudflare.DNSRecord{
			Type:    "CNAME",
			Name:    recordName,
			Proxied: &proxied,
			Content: content,
		})
		if err != nil {
			c.log.Error(err, "failed to create CNAME record "+recordName)
			return err
		}
	}
	return nil
}

func (c *Cloudflare) CreateTunnelDNSRecord(recordName string, tunnel *tunnelv1alpha1.Tunnel) error {
	content := tunnel.Status.TunnelID + ".cfargotunnel.com"
	return c.CreateDNSRecord("CNAME", recordName, content, true)
}

func (c *Cloudflare) DeleteDNSRecords(recordType string, recordName string) error {
	tpl := cloudflare.DNSRecord{Type: "CNAME", Name: recordName}
	records, err := c.api.DNSRecords(c.ctx, c.zoneID, tpl)
	if err != nil {
		c.log.Error(err, "failed to list CNAME records matching "+recordName)
		return err
	}
	for _, record := range records {
		c.log.Info("deleting CNAME record " + recordName)
		if err := c.api.DeleteDNSRecord(c.ctx, c.zoneID, record.ID); err != nil {
			c.log.Error(err, "failed deleting CNAME record "+recordName)
			return err
		}
	}
	return nil
}

type TunnelConfig struct {
	Tunnel  string                          `yaml:"tunnel"`
	Ingress *[]tunnelv1alpha1.TunnelIngress `yaml:"ingress"`
}

func tunnelConfig(t *tunnelv1alpha1.Tunnel) *TunnelConfig {
	ingresses := []tunnelv1alpha1.TunnelIngress{}
	if t.Spec.Ingress != nil {
		ingresses = append(ingresses, *t.Spec.Ingress...)
	}
	defaultIngress := "http_status:404"
	ingresses = append(ingresses, tunnelv1alpha1.TunnelIngress{Service: &defaultIngress})
	config := &TunnelConfig{
		Tunnel:  t.Status.TunnelID,
		Ingress: &ingresses,
	}
	return config
}
