// Copyright (c) 2019 Zededa, Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/lf-edge/eve/api/go/attest"
	"github.com/lf-edge/eve/api/go/certs"
	"github.com/lf-edge/eve/api/go/config"
	"github.com/lf-edge/eve/api/go/logs"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	KB = 1024
	MB = 1024 * KB
)

// MaxSizes defines maximum sizes of objects storage
type MaxSizes struct {
	MaxLogSize         int
	MaxInfoSize        int
	MaxMetricSize      int
	MaxRequestsSize    int
	MaxAppLogsSize     int
	MaxFlowMessageSize int
}

// ChunkReader provides ability to request reader for the data for every available chunk
// device managers stores the data in separate chunks (e.g. files/slices/messages)
// we need readers for every chunk to be separated to be able to process data before present
type ChunkReader interface {
	// Next will return reader for the next chunk and the size of the chunk
	// in case of no next chunk available, will return io.EOF
	Next() (io.Reader, int64, error)
}

type BigData interface {
	Get(index int) ([]byte, error)
	Reader() (ChunkReader, error)
	Write(b []byte) (int, error)
}

type DeviceStorage struct {
	Cert        *x509.Certificate
	Info        BigData
	Metrics     BigData
	Logs        BigData
	Requests    BigData
	FlowMessage BigData
	Certs       BigData
	AppLogs     map[uuid.UUID]BigData
	CurrentLog  int
	Config      []byte
	AttestCerts []byte
	StorageKeys []byte
	Serial      string
	Onboard     *x509.Certificate
	Options     []byte // stores json representation of DeviceOptions
}

type FullCertsEntry struct {
	*logs.LogEntry
	Image      string `json:"image,omitempty"`      // SW image the log got emitted from
	EveVersion string `json:"eveVersion,omitempty"` // EVE software version
}

type FullLogEntry struct {
	*logs.LogEntry
	Image      string `json:"image,omitempty"`      // SW image the log got emitted from
	EveVersion string `json:"eveVersion,omitempty"` // EVE software version
}

type Zcerts struct {
	Certs []*certs.ZCert `json:"certs,omitempty"` // EVE device certs
}

// ApiRequest stores information about requests from EVE
type ApiRequest struct {
	Timestamp time.Time `json:"timestamp"`
	UUID      uuid.UUID `json:"uuid,omitempty"`
	ClientIP  string    `json:"client-ip"`
	Forwarded string    `json:"forwarded,omitempty"`
	Method    string    `json:"method"`
	URL       string    `json:"url"`
}

// PCRValue stores one single PCR value from TPM, from a particular hash bank
type PCRValue struct {
	Index uint32 `json:"index"`
	Value string `json:"value"` // may contain '*' to allow any value in template
}

// PCRTemplate stores template with EVE version, Firmware version, GPSInfo and set of PCRValues
type PCRTemplate struct {
	EveVersion      string      `json:"eveVersion"`
	FirmwareVersion string      `json:"firmwareVersion"`
	PCRValues       []*PCRValue `json:"PCRValues"`
}

// GlobalOptions configure controller behaviour for attestation requests
type GlobalOptions struct {
	EnforceTemplateAttestation bool           `json:"enforceTemplateAttestation"`
	PCRTemplates               []*PCRTemplate `json:"PCRTemplates"`
}

// DeviceOptions stores received nonce, PCRTemplate structure received from device
// and IntegrityToken generated by controller
type DeviceOptions struct {
	Nonce               string                     `json:"nonce"`
	IntegrityToken      string                     `json:"integrityToken"`
	ReceivedPCRTemplate *PCRTemplate               `json:"receivedPCRTemplate"`
	Attested            bool                       `json:"attested"`
	EventLog            []*attest.TpmEventLogEntry `json:"eventLog,omitempty"`
}

// Bytes convenience to convert to json bytes
func (f FullLogEntry) Json() ([]byte, error) {
	return protojson.Marshal(f)
}

func (d *DeviceStorage) AddLogs(b []byte) error {
	// what if the device was not initialized yet?
	if d.Logs == nil {
		return errors.New("AddLog: Logs struct not yet initialized")
	}
	_, err := d.Logs.Write(b)
	return err
}
func (d *DeviceStorage) AddAppLog(instanceID uuid.UUID, b []byte) error {
	// what if the device was not initialized yet?
	if d.AppLogs == nil {
		return fmt.Errorf("AddAppLog: AppLogs struct not yet initialized")
	}
	if _, ok := d.AppLogs[instanceID]; !ok {
		return fmt.Errorf("AddAppLog: AppLogs for instance %s not yet initialized", instanceID)
	}
	_, err := d.AppLogs[instanceID].Write(b)
	return err
}
func (d *DeviceStorage) AddInfo(b []byte) error {
	// what if the device was not initialized yet?
	if d.Info == nil {
		return errors.New("AddInfo: Info struct not yet initialized")
	}
	_, err := d.Info.Write(b)
	return err
}
func (d *DeviceStorage) AddMetrics(b []byte) error {
	// what if the device was not initialized yet?
	if d.Metrics == nil {
		return errors.New("AddMetrics: Metrics struct not yet initialized")
	}
	_, err := d.Metrics.Write(b)
	return err
}

func (d *DeviceStorage) AddRequest(b []byte) error {
	// what if the device was not initialized yet?
	if d.Requests == nil {
		return errors.New("AddRequest: Requests struct not yet initialized")
	}
	_, err := d.Requests.Write(b)
	return err
}

func (d *DeviceStorage) AddFlowRecord(b []byte) error {
	// what if the device was not initialized yet?
	if d.FlowMessage == nil {
		return errors.New("AddFlowRecord: FlowMessage struct not yet initialized")
	}
	_, err := d.FlowMessage.Write(b)
	return err
}

func CreateBaseConfig(u uuid.UUID) []byte {
	conf := &config.EdgeDevConfig{
		Id: &config.UUIDandVersion{
			Uuid:    u.String(),
			Version: "4",
		},
	}
	// we ignore the error because it is tightly controlled
	// we probably should handle it, but then we have to do it with everything downstream
	// eventually
	b, _ := proto.Marshal(conf)
	return b
}

func CreateBaseDeviceOptions(_ uuid.UUID) []byte {
	conf := &DeviceOptions{}
	b, _ := json.Marshal(conf)
	return b
}

func CreateBaseGlobalOptions() []byte {
	conf := &GlobalOptions{}
	b, _ := json.Marshal(conf)
	return b
}
