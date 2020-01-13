package ofagent

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/opencord/ofagent-go/internal/pkg/openflow"
	"github.com/opencord/voltha-lib-go/v2/pkg/log"
	"time"
)

func (o *OFAgent) refreshDeviceList(ctx context.Context) {
	// Create a ticker to update the list device every minute
	// TODO: Should be a configurable time
	tick := time.NewTicker(1 * time.Minute)

	o.updateDeviceList()
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-tick.C:
			log.Debug("GrpcClient entering device refresh pull")
			o.updateDeviceList()
		}
	}
}

func (o *OFAgent) updateDeviceList() {
	deviceList, err := o.volthaClient.ListLogicalDevices(context.Background(), &empty.Empty{})
	if err != nil {
		log.Errorw("OFAgent getDeviceList failed", log.Fields{"error": err})
		return
	}
	devices := deviceList.GetItems()

	var toAdd []string
	var toDel []string
	var deviceIDMap = make(map[string]string)
	log.Debugw("update-device-list",
		log.Fields{
			"device-count": len(devices),
			"devices":      fmt.Sprintf("%+v", devices)})
	for i := 0; i < len(devices); i++ {
		deviceID := devices[i].GetId()
		deviceIDMap[deviceID] = deviceID
		if o.clientMap[deviceID] == nil {
			toAdd = append(toAdd, deviceID)
		}
	}
	for key := range o.clientMap {
		if deviceIDMap[key] == "" {
			toDel = append(toDel, key)
		}
	}
	if log.GetDefaultLogLevel() <= log.DebugLevel {
		log.Debugw("GrpcClient refreshDeviceList", log.Fields{"ToAdd": toAdd, "ToDel": toDel})
	}
	for i := 0; i < len(toAdd); i++ {
		var client = o.addClient(toAdd[i])
		go client.Run()
	}
	for i := 0; i < len(toDel); i++ {
		o.clientMap[toDel[i]].End()
		o.mapLock.Lock()
		delete(o.clientMap, toDel[i])
		o.mapLock.Unlock()
	}
}
func (o *OFAgent) addClient(deviceID string) *openflow.OFClient {
	if log.GetDefaultLogLevel() <= log.DebugLevel {
		log.Debugw("GrpcClient addClient called ", log.Fields{"DeviceID": deviceID})
	}
	o.mapLock.Lock()
	var client *openflow.OFClient
	client = o.clientMap[deviceID]
	if client == nil {
		client = openflow.NewClient(o.volthaClient, o.packetOutChannel, o.OFControllerEndPoint, deviceID, true)
		go client.Run()
		o.clientMap[deviceID] = client
	}
	o.mapLock.Unlock()
	log.Debugw("Finished with addClient", log.Fields{"deviceID": deviceID})
	return client
}

//GetClient Returns a pointer to the OpenFlow client
func (o *OFAgent) GetClient(deviceID string) *openflow.OFClient {
	client := o.clientMap[deviceID]
	if client == nil {
		client = o.addClient(deviceID)
	}
	return client
}
