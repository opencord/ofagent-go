/*
   Copyright 2017 the original author or authors.

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
	"encoding/json"
	ofp "github.com/donNewtonAlpha/goloxi/of13"
	"github.com/opencord/voltha-protos/v2/go/openflow_13"
	"golang.org/x/net/context"
	"log"
)

func handleMeterModRequest(request *ofp.MeterMod, client *Client) {
	jsonRequest, _ := json.Marshal(request)
	log.Printf("handleMeterModRequest called with %s", jsonRequest)

	var meterModUpdate openflow_13.MeterModUpdate
	meterModUpdate.Id = client.DeviceId
	var meterMod openflow_13.OfpMeterMod
	meterMod.MeterId = request.MeterId
	meterMod.Flags = uint32(request.Flags)
	command := openflow_13.OfpMeterModCommand(request.Command)
	meterMod.Command = command
	var bands []*openflow_13.OfpMeterBandHeader
	ofpBands := request.GetMeters()
	for i := 0; i < len(ofpBands); i++ {
		ofpBand := ofpBands[i]
		bandType := ofpBand.GetType()
		var band openflow_13.OfpMeterBandHeader
		switch bandType {
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
	grpcClient := *getGrpcClient()
	meterModJS, _ := json.Marshal(meterModUpdate)
	log.Printf("METER MOD UPDATE %s", meterModJS)
	empty, err := grpcClient.UpdateLogicalDeviceMeterTable(context.Background(), &meterModUpdate)
	if err != nil {
		log.Printf("ERROR DOING METER MOD ADD %v", err)
	}
	emptyJs, _ := json.Marshal(empty)
	log.Printf("METER MOD RESPONSE %s", emptyJs)
}
