package golightify

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"io"
	"strings"
)

const (
	LightifyCommand_ListAllLights    = 0x13
	LightifyCommand_ListAllGroups    = 0x1e
	LightifyCommand_LightDetails     = 0x68
	LightifyCommand_GroupDetails     = 0x26
	LightifyCommand_LightBrightness  = 0x31
	LightifyCommand_LightOnOff       = 0x32
	LightifyCommand_LightTemperature = 0x33
	LightifyCommand_LightColor       = 0x36

	LightifyMessageHeader_DataLength = 6
)

type LightifyMessage struct {
	Header LightifyMessageHeader
	Data   LightifyRequest
}

type LightifyMessageHeader struct {
	Length   uint16
	Unknown1 uint8
	Command  uint8
	Id       uint32
}

type LightifyString16 [16]byte

func (s LightifyString16) MarshalJSON() ([]byte, error) {
	return json.Marshal(strings.TrimRight(string(s[:]), "\x00"))
}

type LightifyRGB struct {
	Red   byte
	Green byte
	Blue  byte
}

type LightifyLightId [8]byte
type LightifyGroupId uint16

func (s LightifyLightId) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(s[:]))
}

type LightifyFirmware [4]byte

type LightifyRequest interface {
	Command() uint8
	NewResponse() interface{}
}

// ****************************** List all lights ******************************
type LightifyRequest_ListAllLightsReq struct {
	AllDetails uint8
}

type LightifyRequest_ListAllLightsResLight struct {
	U1       uint16
	Id       LightifyLightId
	Firmware LightifyFirmware
	U2       [2]byte
	Groups   uint16
	On       byte
	Bri      byte
	Temp     uint16
	Color    LightifyRGB
	U3       byte
	Name     LightifyString16
	U4       [8]byte
}

type LightifyRequest_ListAllLightsRes struct {
	U1         uint8
	LightCount uint16
	Lights     []LightifyRequest_ListAllLightsResLight
}

func (msg *LightifyRequest_ListAllLightsReq) Command() uint8 {
	return LightifyCommand_ListAllLights
}

func (msg *LightifyRequest_ListAllLightsReq) NewResponse() interface{} {
	return &LightifyRequest_ListAllLightsRes{}
}
func (response *LightifyRequest_ListAllLightsRes) LightifyDeserialize(r io.Reader) error {
	f := make([]interface{}, 0, 2)
	f = append(f, &response.U1)
	f = append(f, &response.LightCount)

	for i := 0; i < len(f); i++ {
		err := binary.Read(r, binary.LittleEndian, f[i])
		if err != nil {
			return err
		}
	}
	response.Lights = make([]LightifyRequest_ListAllLightsResLight, response.LightCount, response.LightCount)
	err := binary.Read(r, binary.LittleEndian, &response.Lights)
	if err != nil {
		return err
	}

	return nil
}

// ****************************** List all groups ******************************
type LightifyRequest_ListAllGroupsReq struct {
	U1 uint8
	U2 uint8
}
type LightifyComponent_ListAllGroupsResGroup struct {
	Id   LightifyGroupId
	Name LightifyString16
}

type LightifyRequest_ListAllGroupsRes struct {
	U1         uint8
	GroupCount uint16
	Groups     []LightifyComponent_ListAllGroupsResGroup
}

func (r *LightifyRequest_ListAllGroupsReq) Command() uint8 {
	return LightifyCommand_ListAllGroups
}

func (msg *LightifyRequest_ListAllGroupsReq) NewResponse() interface{} {
	return &LightifyRequest_ListAllGroupsRes{}
}

func (response *LightifyRequest_ListAllGroupsRes) LightifyDeserialize(r io.Reader) error {
	f := make([]interface{}, 0, 2)
	f = append(f, &response.U1)
	f = append(f, &response.GroupCount)

	for i := 0; i < len(f); i++ {
		err := binary.Read(r, binary.LittleEndian, f[i])
		if err != nil {
			return err
		}
	}

	response.Groups = make([]LightifyComponent_ListAllGroupsResGroup, response.GroupCount, response.GroupCount)
	err := binary.Read(r, binary.LittleEndian, &response.Groups)
	if err != nil {
		return err
	}

	return nil
}

// ******************************  Light details  ******************************
type LightifyRequest_LightDetailsReq struct {
	Id LightifyLightId
}

type LightifyComponent_LightDetailsResProperties struct {
	U2    byte
	On    byte
	Bri   byte
	Temp  uint16
	Color LightifyRGB
	U3    byte
	U4    [3]byte
}

type LightifyRequest_LightDetailsRes struct {
	U1         byte
	LigtCount  uint16
	Id         LightifyLightId
	Offline    int8
	Properties interface{}
}

func (r *LightifyRequest_LightDetailsReq) Command() uint8 {
	return LightifyCommand_LightDetails
}

func (msg *LightifyRequest_LightDetailsReq) NewResponse() interface{} {
	return &LightifyRequest_LightDetailsRes{}
}

func (response *LightifyRequest_LightDetailsRes) LightifyDeserialize(r io.Reader) error {
	f := make([]interface{}, 0, 4)
	f = append(f, &response.U1)
	f = append(f, &response.LigtCount)
	f = append(f, &response.Id)
	f = append(f, &response.Offline)

	for i := 0; i < len(f); i++ {
		err := binary.Read(r, binary.LittleEndian, f[i])
		if err != nil {
			return err
		}
	}

	if response.Offline == 0 {
		properties := &LightifyComponent_LightDetailsResProperties{}

		err := binary.Read(r, binary.LittleEndian, properties)
		if err != nil {
			return err
		}
		response.Properties = *properties
	}

	return nil
}

// ******************************  Group details  ******************************
type LightifyRequest_GroupDetailsReq struct {
	Id LightifyGroupId
}

type LightifyRequest_GroupDetailsRes struct {
	U1         byte
	Id         LightifyGroupId
	Name       LightifyString16
	LightCount uint8
	Lights     []LightifyLightId
}

func (r *LightifyRequest_GroupDetailsReq) Command() uint8 {
	return LightifyCommand_GroupDetails
}

func (msg *LightifyRequest_GroupDetailsReq) NewResponse() interface{} {
	return &LightifyRequest_GroupDetailsRes{}
}

func (response *LightifyRequest_GroupDetailsRes) LightifyDeserialize(r io.Reader) error {
	f := make([]interface{}, 0, 4)
	f = append(f, &response.U1)
	f = append(f, &response.Id)
	f = append(f, &response.Name)
	f = append(f, &response.LightCount)

	for i := 0; i < len(f); i++ {
		err := binary.Read(r, binary.LittleEndian, f[i])
		if err != nil {
			return err
		}
	}

	response.Lights = make([]LightifyLightId, response.LightCount, response.LightCount)
	err := binary.Read(r, binary.LittleEndian, &response.Lights)
	if err != nil {
		return err
	}

	return nil
}

// ******************************   Light on/off  ******************************
type LightifyRequest_LightOnOffReq struct {
	Id LightifyLightId
	On byte
}

func (r *LightifyRequest_LightOnOffReq) Command() uint8 {
	return LightifyCommand_LightOnOff
}

func (msg *LightifyRequest_LightOnOffReq) NewResponse() interface{} {
	return nil
}

// ******************************Light temperature******************************
type LightifyRequest_LightTemperatureReq struct {
	Id   LightifyLightId
	Temp uint16
	Time uint16
}

func (r *LightifyRequest_LightTemperatureReq) Command() uint8 {
	return LightifyCommand_LightTemperature
}

func (msg *LightifyRequest_LightTemperatureReq) NewResponse() interface{} {
	return nil
}

// ******************************Light brightness ******************************
type LightifyRequest_LightBrightnessReq struct {
	Id   LightifyLightId
	Bri  byte
	Time uint16
}

func (r *LightifyRequest_LightBrightnessReq) Command() uint8 {
	return LightifyCommand_LightBrightness
}

func (msg *LightifyRequest_LightBrightnessReq) NewResponse() interface{} {
	return nil
}

// ******************************   Light color   ******************************
type LightifyRequest_LightColorReq struct {
	Id    LightifyLightId
	Color LightifyRGB
	U1    byte
	Time  uint16
}

func (r *LightifyRequest_LightColorReq) Command() uint8 {
	return LightifyCommand_LightColor
}

func (msg *LightifyRequest_LightColorReq) NewResponse() interface{} {
	return nil
}
