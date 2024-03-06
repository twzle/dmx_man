package artnet

import (
	"context"
	"fmt"

	"git.miem.hse.ru/hubman/dmx-executor/internal/models"
	"git.miem.hse.ru/hubman/dmx-executor/internal/config"
	"git.miem.hse.ru/hubman/dmx-executor/internal/device"
	"github.com/jsimonetti/go-artnet"
)

func NewArtNetDevice(ctx context.Context, conf config.ArtNetConfig) (device.Device, error) {
	log := artnet.NewDefaultLogger()
	dev := artnet.NewController(conf.Alias, conf.IP, log)
	dev.Start()
	
	newArtNet := &artnetDevice{alias: conf.Alias, dev: dev}
	newArtNet.SetUniverse(conf.Universe)

	return newArtNet, nil
}

type artnetDevice struct {
	alias    string
	universe [512]byte
	dev      *artnet.Controller
}

func (d *artnetDevice) GetAlias() string {
	return d.alias
}

func (d *artnetDevice) SetUniverse(universe []config.ChannelRange) {
	artNetUniverse := make(map[uint16]uint16, len(universe))
	for _, channelRange := range universe {
		artNetUniverse[channelRange.InitialIndex] = channelRange.Value
	}

	var channelValue uint16
	for idx, channel := range d.universe {
		value, ok := artNetUniverse[uint16(idx)] 
		if ok {
			channelValue = value
		}

		channel = byte(channelValue)
		d.universe[idx] = channel
	}
}

func (d *artnetDevice) SetValueToChannel(ctx context.Context, command models.SetChannel) error {
	if command.Channel < 1 || command.Channel >= 512 {
		return fmt.Errorf("channel number should be beetwen 1 and 511, but got: %v", command.Channel)
	}
	d.universe[command.Channel] = byte(command.Value)
	d.dev.SendDMXToAddress(d.universe, artnet.Address{Net: 0, SubUni: 0})

	return nil
}

func (d *artnetDevice) Blackout(ctx context.Context) error {
	d.universe = [512]byte{}
	d.dev.SendDMXToAddress(d.universe, artnet.Address{Net: 0, SubUni: 0})

	return nil
}
