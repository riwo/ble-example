package main

import (
	"context"
	"os"
	"os/signal"

	api "github.com/muka/go-bluetooth/api"
	service "github.com/muka/go-bluetooth/api/service"
	agent "github.com/muka/go-bluetooth/bluez/profile/agent"
	gatt "github.com/muka/go-bluetooth/bluez/profile/gatt"
	hw "github.com/muka/go-bluetooth/hw"
	log "github.com/sirupsen/logrus"
)

func main() {

	log.SetLevel(log.DebugLevel)

	adapter, err := api.GetDefaultAdapter()
	if err != nil {
		log.Panicf("failed to get default adapter error: %v", err)
	}

	adapterID, err := adapter.GetAdapterID()
	if err != nil {
		log.Panicf("failed to get default adapter id error: %v", err)
	}

	btmgmt := hw.NewBtMgmt(adapterID)

	// Disable paring
	err = btmgmt.SetPairable(false)
	if err != nil {
		log.Panicf("failed to set pairable off adapter id: error: %v", btmgmt.BinPath, err)
	}

	// set LE mode
	btmgmt.SetPowered(false)
	btmgmt.SetBredr(false)
	btmgmt.SetLe(true)
	btmgmt.SetPrivacy(false)
	btmgmt.SetFastConnectable(true)
	//btmgmt.SetLinkLevelSecurity(false)
	btmgmt.SetBondable(false)
	btmgmt.SetPowered(true)

	options := service.AppOptions{
		AdapterID:         adapterID,
		AgentCaps:         agent.CapNoInputNoOutput,
		UUIDSuffix:        "-0000-1000-8000-00805F9B34FB",
		UUID:              "6789",
		AgentSetAsDefault: true,
	}

	app, err := service.NewApp(options)
	if err != nil {
		log.Panicf("failed to get default adapter id error: %v", err)
	}

	defer app.Close()

	app.SetName("example-test")

	if !app.Adapter().Properties.Powered {
		log.Infof("Powering adapter: %v", app.AdapterID())
		err = app.Adapter().SetPowered(true)
		if err != nil {
			log.Panicf("Failed to power the adapter: %s error: %v", app.Adapter().Properties.Name, err)
		}
	}

	svc, err := app.NewService("1000")
	if err != nil {
		log.Panicf("failed to get default adapter id error: %v", err)
	}

	// BSSID
	err = addReadChar(svc, "1001", stringToCharReadCallback(func() string {
		return "AA:BB:CC:DD:EE:FF"
	}))

	if err != nil {
		log.Panicf("failed to add bssid char error: %v", err)
	}

	// SSID
	err = addReadChar(svc, "1002", stringToCharReadCallback(func() string {
		return "example-test"
	}))

	if err != nil {
		log.Panicf("failed to add ssid char error: %v", err)
	}

	// PSK
	err = addReadChar(svc, "1003", stringToCharReadCallback(func() string {
		return "N0t$oSecure0verB!E"
	}))

	if err != nil {
		log.Panicf("failed to add psk char error: %v", err)
	}

	chars := svc.GetChars()
	log.Warnf("%v", chars)

	err = app.AddService(svc)
	if err != nil {
		log.Panicf("failed to add service error: %v", err)
	}

	err = app.Run()
	if err != nil {
		log.Panicf("failed to run app error: %v", err)
	}

	adv := app.GetAdvertisement()
	adv.MinInterval = 15
	adv.MaxInterval = 30
	adv.Discoverable = true

	cf, err := app.Advertise(0)
	if err != nil {
		log.Panicf("failed to start advertisement error: %v", err)
	}
	defer cf()

	ctx, stop := signal.NotifyContext(context.TODO(), os.Interrupt)
	defer stop()

	<-ctx.Done()
}

func stringToCharReadCallback(f func() string) service.CharReadCallback {
	return func(c *service.Char, options map[string]interface{}) ([]byte, error) {
		return []byte(f()), nil
	}
}

func addReadChar(svc *service.Service, uuid string, rcb service.CharReadCallback) error {

	char, err := svc.NewChar(uuid)
	if err != nil {
		return err
	}

	char.Properties.Flags = []string{
		gatt.FlagCharacteristicRead,
	}

	char.OnRead(rcb)

	char.OnWrite(service.CharWriteCallback(func(c *service.Char, value []byte) ([]byte, error) {
		return value, nil
	}))

	return svc.AddChar(char)
}
