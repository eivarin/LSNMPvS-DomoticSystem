package packet

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"math/rand"
	"net"
	"sort"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/eivarin/LSNMPvS-DomoticSystem/CustomLogger"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types/CodableValues"
)

const (
	ErrorDecodingPacket = iota + 1
	ErrorIncorrectTag
	ErrorInvalidType
	ErrorDuplicateMessageId // por usar
	ErrorInvalidIID
	ErrorUnsupportedDataType // por usar
	ErrorInvalidDataType
	ErrorUnmatchedIIDValueList // por usar
	ErrorInvalidGroupIndexes
	ErrorChangingReadOnlyValue
	ErrorStructureDoesntExist
	ErrorObjectIdDoesntExist
	ErrorIndexOutOfRange
	ErrorValueOutOfRange

	fixedTag         = "kdk847ufh84jg87g"
	possibleChars    = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	possibleCharsLen = len(possibleChars)
	messageIdLength  = 16
)

type PacketErr int

func (e PacketErr) Compile(p LSNMPvS_Packet) (*LSNMPvS_Packet, error, bool) {
	return p.NewErrorResponsePacket([]int{int(e)}), fmt.Errorf(e.Error()), true
}

func (e PacketErr) Error() string {
	errorText := ""
	switch e {
	case ErrorInvalidIID:
		errorText = "refered invalid iid for the request type"
	case ErrorInvalidGroupIndexes:
		errorText = "refered invalid group indexes in iid"
	case ErrorChangingReadOnlyValue:
		errorText = "tried changing read-only value"
	case ErrorStructureDoesntExist:
		errorText = "refered structure doesn't exist"
	case ErrorObjectIdDoesntExist:
		errorText = "refered object id doesn't exist"
	case ErrorIndexOutOfRange:
		errorText = "refered index is out of range"
	case ErrorValueOutOfRange:
		errorText = "refered value is out of the allowed range for the object"
	case ErrorInvalidDataType:
		errorText = "incorred data type for object"
	case ErrorUnsupportedDataType:
		errorText = "unsupported data type"
	case ErrorUnmatchedIIDValueList:
		errorText = "unmatched iid and value list"
	case ErrorIncorrectTag:
		errorText = "incorrect packet tag"
	case ErrorInvalidType:
		errorText = "invalid packet type"
	case ErrorDuplicateMessageId:
		errorText = "duplicate message id"
	case ErrorDecodingPacket:
		errorText = "error decoding packet"
	}
	return errorText
}

func RandStringBytes() string {
	b := make([]byte, messageIdLength)
	for i := range b {
		b[i] = possibleChars[rand.Intn(len(possibleChars))]
	}
	return string(b)
}

type LSNMPvS_Packet struct {
	tag       string                      // tag
	pType     byte                        // packet type
	timestamp *types.CompleteCodableValue // timestamp
	messageId string                      // message id
	iidList   types.CodableList           // list of iid
	valueList types.CodableList           // list of values
	errorList []int                       // list of errors
}

func NewGetRequestPacket(iidList types.CodableList) *LSNMPvS_Packet {
	return &LSNMPvS_Packet{
		tag:       fixedTag,
		pType:     'G',
		timestamp: types.NewCodableTimestampNow(),
		messageId: RandStringBytes(),
		iidList:   iidList,
		valueList: types.CodableList{},
		errorList: []int{},
	}
}

func NewSetResponsePacket(iidList, valueList types.CodableList) *LSNMPvS_Packet {
	return &LSNMPvS_Packet{
		tag:       fixedTag,
		pType:     'S',
		timestamp: types.NewCodableTimestampNow(),
		messageId: RandStringBytes(),
		iidList:   iidList,
		valueList: valueList,
		errorList: []int{},
	}
}

func NewErrorDecodingPacket(pErr PacketErr) *LSNMPvS_Packet {
	return &LSNMPvS_Packet{
		tag:       fixedTag,
		pType:     'R',
		timestamp: types.NewCodableTimestampNow(),
		messageId: "ERROR",
		iidList:   types.CodableList{},
		valueList: types.CodableList{},
		errorList: []int{int(pErr)},
	}
}

func (p *LSNMPvS_Packet) NewErrorResponsePacket(errorList []int) *LSNMPvS_Packet {
	return &LSNMPvS_Packet{
		tag:       p.tag,
		pType:     'R',
		timestamp: p.timestamp,
		messageId: p.messageId,
		iidList:   p.iidList,
		valueList: p.valueList,
		errorList: errorList,
	}
}

func (p *LSNMPvS_Packet) NewResponsePacket(l []types.IdValuePair, uptime *types.CompleteCodableValue) *LSNMPvS_Packet {
	iidList := types.CodableList{}
	valueList := types.CodableList{}
	for _, v := range l {
		iidList.Append(v.IID)
		valueList.Append(v.Value)
	}
	return &LSNMPvS_Packet{
		tag:       p.tag,
		pType:     'R',
		timestamp: uptime,
		messageId: p.messageId,
		iidList:   iidList,
		valueList: valueList,
		errorList: []int{},
	}
}

func NewNotificationPacket(l []types.IdValuePair, uptime *types.CompleteCodableValue) *LSNMPvS_Packet {
	iidList := types.CodableList{}
	valueList := types.CodableList{}
	for _, v := range l {
		iidList.Append(v.IID)
		valueList.Append(v.Value)
	}
	return &LSNMPvS_Packet{
		tag:       fixedTag,
		pType:     'N',
		timestamp: uptime,
		messageId: RandStringBytes(),
		iidList:   iidList,
		valueList: valueList,
		errorList: []int{},
	}
}

func (p *LSNMPvS_Packet) AppendEntry(e types.IdValuePair) {
	p.iidList.Append(e.IID)
	p.valueList.Append(e.Value)
}

func (p *LSNMPvS_Packet) Encode() string {
	encoded := CodableValues.EncodeString(p.tag)
	encoded += string(p.pType)
	encoded += p.timestamp.Encode()
	encoded += CodableValues.EncodeString(p.messageId)
	encoded += p.iidList.Encode()
	encoded += p.valueList.Encode()
	encoded += CodableValues.EncodeInt(len(p.errorList))
	for _, v := range p.errorList {
		encoded += CodableValues.EncodeInt(v)
	}
	return Encrypt(encoded)
}

func Encrypt(plainText string) string {
	aes, err := aes.NewCipher([]byte(fixedTag))
	if err != nil {
		panic(err)
	}
	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		panic(err)
	}
	nonce := make([]byte, gcm.NonceSize())
	for i := range nonce {
		nonce[i] = byte(i)
	}
	cipherText := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return string(cipherText)
}

func (p *LSNMPvS_Packet) Decode(data string) (string, PacketErr) {
	data = Decrypt(data)
	var err error
	var rest string
	p.tag, rest, err = CodableValues.DecodeString(data)
	if err != nil {
		return "", ErrorDecodingPacket
	}
	p.pType = rest[0]
	p.timestamp = &types.CompleteCodableValue{}
	rest, err = p.timestamp.Decode(rest[1:])
	if err != nil {
		return "", ErrorDecodingPacket
	}
	p.messageId, rest, err = CodableValues.DecodeString(rest)
	if err != nil {
		return "", ErrorDecodingPacket
	}
	p.iidList = types.CodableList{}
	rest, err = p.iidList.Decode(rest)
	if err != nil {
		return "", ErrorDecodingPacket
	}
	p.valueList = types.CodableList{}
	rest, err = p.valueList.Decode(rest)
	if err != nil {
		return "", ErrorDecodingPacket
	}
	var length int
	length, rest, err = CodableValues.DecodeInt(rest)
	if err != nil {
		return "", ErrorDecodingPacket
	}
	p.errorList = make([]int, length)
	for i := 0; i < length; i++ {
		p.errorList[i], rest, err = CodableValues.DecodeInt(rest)
		if err != nil {
			return "", ErrorDecodingPacket
		}
	}
	return rest, 0
}

func Decrypt(cipherText string) string {
	aes, err := aes.NewCipher([]byte(fixedTag))
	if err != nil {
		panic(err)
	}
	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		panic(err)
	}
	nonceSize := gcm.NonceSize()
	nonce, cipherText := cipherText[:nonceSize], cipherText[nonceSize:]
	plainText, err := gcm.Open(nil, []byte(nonce), []byte(cipherText), nil)
	if err != nil {
		panic(err)
	}
	return string(plainText)
}


func (p *LSNMPvS_Packet) Equal(other *LSNMPvS_Packet) bool {
	if p.tag != other.tag {
		return false
	}
	if p.pType != other.pType {
		return false
	}
	if !p.timestamp.Equals(other.timestamp) {
		return false
	}
	if p.messageId != other.messageId {
		return false
	}
	if len(p.errorList) != len(other.errorList) {
		return false
	}
	for i := range p.errorList {
		if p.errorList[i] != other.errorList[i] {
			return false
		}
	}
	return p.iidList.Equals(other.iidList) && p.valueList.Equals(other.valueList)
}

func (p *LSNMPvS_Packet) String() string {
	return "Tag: " + p.tag + "\n" +
		"Type: " + string(p.pType) + "\n" +
		"Timestamp: " + p.timestamp.String() + "\n" +
		"Message ID: " + p.messageId + "\n" +
		"IID List: " + p.iidList.String() + "\n" +
		"Value List: " + p.valueList.String() + "\n" +
		"Error List: " + CodableValues.EncodeInt(len(p.errorList))
}

func (p *LSNMPvS_Packet) RenderPacketWithLipgloss(width int) string {
	var rendered string
	tWidth := width - 4
	TitleStyle := lipgloss.NewStyle().Width(tWidth).Align(lipgloss.Center).Border(lipgloss.RoundedBorder(), false, false, true).BorderForeground(lipgloss.ANSIColor(208)).Foreground(lipgloss.ANSIColor(208))
	// TableStyle := table.New().Width(tWidth).Border(lipgloss.RoundedBorder()).BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99")))
	Headers := []string{"Tag", "Type", "Timestamp", "Message ID"}
	Values := []string{p.tag, string(p.pType), p.timestamp.String(), p.messageId}
	rendered = TitleStyle.Render("Headers")
	rendered = lipgloss.JoinVertical(lipgloss.Center, rendered, renderTableWithLipGloss(Headers, Values, tWidth), TitleStyle.Render("IID Values Pairs"))
	if len(p.iidList) > 0 {
		IID := []string{}
		Value := []string{}
		leng := len(p.iidList)
		for i := 1; i <= leng; i++ {
			IID = append(IID, p.iidList[i].String())
			if i > len(p.valueList) {
				Value = append(Value, "")
			} else {
				Value = append(Value, p.valueList[i].String())
			}
		}
		rendered = lipgloss.JoinVertical(lipgloss.Center, rendered, renderTableWithLipGloss(IID, Value, tWidth))
	}
	rendered = lipgloss.JoinVertical(lipgloss.Center, rendered, TitleStyle.Render("Errors"))
	if len(p.errorList) > 0 {
		Errors := []string{}
		for _, v := range p.errorList {
			Errors = append(Errors, fmt.Sprintf("Error code %d: %v", v, PacketErr(v)))
		}
		rendered = lipgloss.JoinVertical(lipgloss.Center, rendered, renderTableWithLipGloss([]string{"Errors"}, Errors, tWidth))
	}
	rendered = lipgloss.NewStyle().Width(width - 2).Align(lipgloss.Center).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.ANSIColor(27)).Render(rendered)
	return rendered
}

func renderTableWithLipGloss(headers []string, values []string, width int) string {
	t := table.New().Border(lipgloss.RoundedBorder()).BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).Width(width).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == 0:
				return lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
			case row%2 == 0:
				return lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
			default:
				return lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
			}
		}).Headers(headers...).
		Rows(values)
	return t.Render()
}

func (p *LSNMPvS_Packet) VerifyAndGetType() (byte, PacketErr) {
	if p.tag != fixedTag {
		return 0, ErrorIncorrectTag
	}
	if p.pType != 'G' && p.pType != 'S' && p.pType != 'R' && p.pType != 'N' {
		return 0, ErrorInvalidType
	}
	return p.pType, 0
}

func (p *LSNMPvS_Packet) GetIidValuePairList() []types.IdValuePair {
	var idValuePairList []types.IdValuePair
	for i := range p.iidList {
		idValuePairList = append(idValuePairList, types.IdValuePair{
			IID:   p.iidList[i],
			Value: p.valueList[i],
		})
	}
	return idValuePairList
}

func (p *LSNMPvS_Packet) GetUncompressedIdValuePairList(structureLengths map[int]map[int]int) ([]types.IdValuePair, bool) {
	var idValuePairList []types.IdValuePair
	var iList []int
	for i := range p.iidList {
		iList = append(iList, i)
	}
	sort.Ints(iList)
	for _, i := range iList {
		v, ok := p.valueList[i]
		if !ok {
			v = nil
		}
		compressedIID := p.iidList[i].Value.(*CodableValues.IID)
		uncompressedList, res := compressedIID.GenListOfIIDs(structureLengths[compressedIID.Structure][compressedIID.Object])
		if !res {
			return nil, false
		}
		for _, rawIID := range uncompressedList {
			if rawIID.FirstIndex == nil {
				idValuePairList = append(idValuePairList, types.IdValuePair{
					IID:   types.NewCodableIID(rawIID.Structure, rawIID.Object, []int{}),
					Value: v,
				})
				continue
			}
			iid := types.NewCodableIID(rawIID.Structure, rawIID.Object, []int{*rawIID.FirstIndex})
			idValuePairList = append(idValuePairList, types.IdValuePair{
				IID:   iid,
				Value: v,
			})
		}
	}
	return idValuePairList, true
}

func (p *LSNMPvS_Packet) SendPacketTo(remAddr net.UDPAddr) error {
	encoded := p.Encode()
	conn, err := net.DialUDP("udp", nil, &remAddr)
	if err != nil {
		return err
	}
	_, err = conn.Write([]byte(encoded))
	return err
}

func (p *LSNMPvS_Packet) GetMessageID() string {
	return p.messageId
}

func (p *LSNMPvS_Packet) VerifyIfPacketIsDuplicate(other *LSNMPvS_Packet) bool {
	// if other.timestamp.Value.(*CodableValues.Timestamp).Ts.Sub(p.timestamp.Value.(*CodableValues.Timestamp).Ts) < 10 * time.Second {
	if other.pType == p.pType && other.messageId == p.messageId && p.pType != 'N' {
		return true
	} else {
		return false
	}
}

func (p *LSNMPvS_Packet) TryLogErrors(logger *CustomLogger.CustomLogger) bool {
	if len(p.errorList) > 0 {
		for _, v := range p.errorList {
			logger.LogError(fmt.Sprintf("Error code %d: %v", v, PacketErr(v)), "Request")
		}
		return true
	}
	return false
}
