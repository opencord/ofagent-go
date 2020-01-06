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

package settings

var debug bool
var debugMap = make(map[string]bool)

func SetDebug(on bool) {
	debug = on
}

func SetDeviceDebug(deviceId string, on bool) {
	debugMap[deviceId] = on
}
func GetDebug(deviceId string) bool {
	/*
		if debug {
			return true
		}
		return getDeviceDebug(deviceId)
	*/
	return debug
}
func getDeviceDebug(deviceId string) bool {
	//will return false if nothing has been added for that device
	return debugMap[deviceId]
}
