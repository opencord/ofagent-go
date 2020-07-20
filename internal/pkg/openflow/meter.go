/*
Copyright 2020 the original author or authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/
package openflow

import (
	"encoding/binary"
	"encoding/json"
	"unsafe"

	ofp "github.com/opencord/goloxi/of13"
	"github.com/opencord/voltha-lib-go/v3/pkg/log"
	"github.com/opencord/voltha-protos/v3/go/openflow_13"
	"golang.org/x/net/context"
)

func (ofc *OFConnection) handleMeterModRequest(ctx context.Context, request *ofp.MeterMod) {
	if logger.V(log.DebugLevel) {
		js, _ := json.Marshal(request)
		logger.Debugw(ctx, "handleMeterModRequest called",
			log.Fields{
				"device-id": ofc.DeviceID,
				"request":   js})
	}

	volthaClient := ofc.VolthaClient.Get()
	if volthaClient == nil {
		logger.Errorw(ctx, "no-voltha-connection",
			log.Fields{"device-id": ofc.DeviceID})
		return
	}

	meterModUpdate := openflow_13.MeterModUpdate{Id: ofc.DeviceID}
	meterMod := openflow_13.OfpMeterMod{
		MeterId: request.MeterId,
		Flags:   uint32(request.Flags),
		Command: openflow_13.OfpMeterModCommand(request.Command),
	}
	var bands []*openflow_13.OfpMeterBandHeader
	for _, ofpBand := range request.GetMeters() {
		var band openflow_13.OfpMeterBandHeader
		switch ofpBand.GetType() {
		case ofp.OFPMBTDrop:
			ofpDrop := ofpBand.(ofp.IMeterBandDrop)
			band.Type = openflow_13.OfpMeterBandType_OFPMBT_DROP
			band.Rate = ofpDrop.GetRate()
			band.BurstSize = ofpDrop.GetBurstSize()
		case ofp.OFPMBTDSCPRemark:
			ofpDscpRemark := ofpBand.(ofp.IMeterBandDscpRemark)
			var dscpRemark openflow_13.OfpMeterBandDscpRemark
			band.Type = openflow_13.OfpMeterBandType_OFPMBT_DSCP_REMARK
			band.BurstSize = ofpDscpRemark.GetBurstSize()
			band.Rate = ofpDscpRemark.GetRate()
			dscpRemark.PrecLevel = uint32(ofpDscpRemark.GetPrecLevel())
			/*
				var meterBandHeaderDscp openflow_13.OfpMeterBandHeader_DscpRemark
				meterBandHeaderDscp.DscpRemark = &dscpRemark
				band.Data = &meterBandHeaderDscp

			*/
		case ofp.OFPMBTExperimenter:
			ofpExperimenter := ofpBand.(ofp.IMeterBandExperimenter)
			var experimenter openflow_13.OfpMeterBandExperimenter
			experimenter.Experimenter = ofpExperimenter.GetExperimenter()
			band.Type = openflow_13.OfpMeterBandType_OFPMBT_EXPERIMENTER
			band.BurstSize = ofpExperimenter.GetBurstSize()
			band.Rate = ofpExperimenter.GetRate()
			/*
				var meterBandHeaderExperimenter openflow_13.OfpMeterBandHeader_Experimenter
				meterBandHeaderExperimenter.Experimenter = &experimenter
				band.Data = &meterBandHeaderExperimenter

			*/
		}
		bands = append(bands, &band)
	}
	meterMod.Bands = bands
	meterModUpdate.MeterMod = &meterMod
	if logger.V(log.DebugLevel) {
		meterModJS, _ := json.Marshal(meterModUpdate)
		logger.Debugw(ctx, "handleMeterModUpdate sending request",
			log.Fields{
				"device-id":         ofc.DeviceID,
				"meter-mod-request": meterModJS})
	}
	if _, err := volthaClient.UpdateLogicalDeviceMeterTable(context.Background(), &meterModUpdate); err != nil {
		logger.Errorw(ctx, "Error calling UpdateLogicalDeviceMeterTable",
			log.Fields{
				"device-id": ofc.DeviceID,
				"error":     err})
		message := ofp.NewMeterModFailedErrorMsg()
		message.SetXid(request.Xid)
		//OF 1.3
		message.SetVersion(4)
		//FIXME Hardcoding to Invalid Meter for now.
		message.SetCode(ofp.OFPMMFCInvalidMeter)
		meterByteId := make([]byte, 4)
		binary.BigEndian.PutUint32(meterByteId, request.MeterId)
		message.SetData(meterByteId)
		message.Length = uint16(unsafe.Sizeof(*message))
		if err := ofc.SendMessage(ctx, message); err != nil {
			logger.Errorw(ctx, "Error reporting failure of MeterMod to controller",
				log.Fields{
					"device-id": ofc.DeviceID,
					"error":     err})
		}
	}
}
