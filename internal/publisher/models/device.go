package models

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type Device struct {
	Name        string     `json:"name"`
	UniqueId    string     `json:"unique_id"`
	DeviceClass string     `json:"device_class"`
	Device      DeviceInfo `json:"device"`
}

func (d Device) MarshalJSON() ([]byte, error) {
	type Alias Device
	alias := &struct {
		Alias
		AttributesTopic string `json:"json_attributes_topic"`
	}{
		Alias:           (Alias)(d),
		AttributesTopic: d.GetAttributesTopic(),
	}

	return json.Marshal(alias)
}

var uniqueIdRegex = regexp.MustCompile("[^a-zA-Z0-9]+")

func NewDevice(name string, serial string, model string, manufacturer string) Device {
	uniqueId := uniqueIdRegex.ReplaceAllString(fmt.Sprintf("%s_%s", name, serial), "_")
	uniqueId = strings.ToLower(uniqueId)

	newDevice := Device{
		Name:        name,
		UniqueId:    uniqueId,
		DeviceClass: "device_tracker",
		Device: DeviceInfo{
			Identifiers:  []string{uniqueId},
			Name:         name,
			Model:        model,
			Manufacturer: manufacturer,
		},
	}

	return newDevice
}

func (d *Device) GetConfigTopic() string {
	topic := fmt.Sprintf("homeassistant/device_tracker/%s/config", d.UniqueId)

	return topic
}

func (d *Device) GetAttributesTopic() string {
	topic := fmt.Sprintf("findmy2mqtt/%s/attributes", d.UniqueId)

	return topic
}
