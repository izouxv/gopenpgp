package crypto

import (
	"encoding/base64"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"runtime"

	"github.com/ProtonMail/gopenpgp/armor"
	"github.com/ProtonMail/gopenpgp/constants"
	"github.com/ProtonMail/gopenpgp/internal"

	"golang.org/x/crypto/openpgp/packet"
)

// ---- MODELS -----

// PlainTextMessage stores an unencrypted text message.
type CleartextMessage struct {
	// The content of the message
	Text string
	// If the decoded message was correctly signed. See constants.SIGNATURE* for all values.
 	Verified int
}

// BinaryMessage stores an unencrypted binary message.
type BinaryMessage struct {
	// The content of the message
	Data []byte
	// If the decoded message was correctly signed. See constants.SIGNATURE* for all values.
 	Verified int
}

// PGPMessage stores a PGP-encrypted message.
type PGPMessage struct {
	// The content of the message
	Data []byte
}

// PGPSignature stores a PGP-encoded detached signature.
type PGPSignature struct {
	// The content of the message
	Data []byte
}

// PGPSplitMessage contains a separate session key packet and symmetrically
// encrypted data packet.
type PGPSplitMessage struct {
	DataPacket []byte
	KeyPacket  []byte
}

// ---- GENERATORS -----

// NewCleartextMessage generates a new CleartextMessage ready for encryption,
// signature, or verification from the plaintext.
func NewCleartextMessage(text string) (*CleartextMessage) {
	return &CleartextMessage {
		Text: text,
		Verified: constants.SIGNATURE_NOT_SIGNED,
	}
}

// NewBinaryMessage generates a new BinaryMessage ready for encryption,
// signature, or verification from the unencrypted bianry data.
func NewBinaryMessage(data []byte) (*BinaryMessage) {
	return &BinaryMessage {
		Data: data,
		Verified: constants.SIGNATURE_NOT_SIGNED,
	}
}

// NewPGPMessage generates a new PGPMessage from the unarmored binary data.
func NewPGPMessage(data []byte) (*PGPMessage) {
	return &PGPMessage {
		Data: data,
	}
}

// NewPGPMessageFromArmored generates a new PGPMessage from an armored string ready for decryption.
func NewPGPMessageFromArmored(armored string) (*PGPMessage, error) {
	encryptedIO, err := internal.Unarmor(armored)
	if err != nil {
		return nil, err
	}

	message, err := ioutil.ReadAll(encryptedIO.Body)
	if err != nil {
		return nil, err
	}

	return &PGPMessage {
		Data: message,
	}, nil
}

// NewPGPSplitMessage generates a new PGPSplitMessage from the binary unarmored keypacket,
// datapacket, and encryption algorithm.
func NewPGPSplitMessage(keyPacket []byte, dataPacket []byte) (*PGPSplitMessage) {
	return &PGPSplitMessage {
		KeyPacket: keyPacket,
		DataPacket: dataPacket,
	}
}

// NewPGPSplitMessageFromArmored generates a new PGPSplitMessage by splitting an armored message into its
// session key packet and symmetrically encrypted data packet.
func NewPGPSplitMessageFromArmored (encrypted string) (*PGPSplitMessage, error) {
	message, err := NewPGPMessageFromArmored(encrypted)
	if err != nil {
		return nil, err
	}

	return message.SeparateKeyAndData(len(encrypted), -1)
}

// NewPGPSignature generates a new PGPSignature from the unarmored binary data.
func NewPGPSignature(data []byte) (*PGPSignature) {
	return &PGPSignature {
		Data: data,
	}
}

// NewPGPSignatureFromArmored generates a new PGPSignature from the armored string ready for verification.
func NewPGPSignatureFromArmored(armored string) (*PGPSignature, error) {
	encryptedIO, err := internal.Unarmor(armored)
	if err != nil {
		return nil, err
	}

	signature, err := ioutil.ReadAll(encryptedIO.Body)
	if err != nil {
		return nil, err
	}

	return &PGPSignature {
		Data: signature,
	}, nil
}

// ---- MODEL METHODS -----

// GetVerification returns the verification status of a message, to use after the KeyRing.Decrypt* or KeyRing.Verify*
// functions. The int value returned is to compare to constants.SIGNATURE*.
func (msg *CleartextMessage) GetVerification() int {
	return msg.Verified
}

// IsVerified returns true if the message is signed and the signature is valid.
// To use after the KeyRing.Decrypt* or KeyRing.Verify* functions.
func (msg *CleartextMessage) IsVerified() bool {
	return msg.Verified == constants.SIGNATURE_OK
}

// GetString returns the content of the message as a string
func (msg *CleartextMessage) GetString() string {
	return msg.Text
}

// NewReader returns a New io.Reader for the text of the message
func (msg *CleartextMessage) NewReader() io.Reader {
	return bytes.NewReader(bytes.NewBufferString(msg.GetString()).Bytes())
}

// GetVerification returns the verification status of a message, to use after the KeyRing.Decrypt* or KeyRing.Verify*
// functions. The int value returned is to compare to constants.SIGNATURE*.
func (msg *BinaryMessage) GetVerification() int {
	return msg.Verified
}

// IsVerified returns true if the message is signed and the signature is valid.
// To use after the KeyRing.Decrypt* or KeyRing.Verify* functions.
func (msg *BinaryMessage) IsVerified() bool {
	return msg.Verified == constants.SIGNATURE_OK
}

// GetBinary returns the binary content of the message as a []byte
func (msg *BinaryMessage) GetBinary() []byte {
	return msg.Data
}

// NewReader returns a New io.Reader for the bianry data of the message
func (msg *BinaryMessage) NewReader() io.Reader {
	return bytes.NewReader(msg.GetBinary())
}

// GetBinary returns the base-64 encoded binary content of the message as a string
func (msg *BinaryMessage) GetBase64() string {
	return base64.StdEncoding.EncodeToString(msg.Data)
}

// GetBinary returns the unarmored binary content of the message as a []byte
func (msg *PGPMessage) GetBinary() []byte {
	return msg.Data
}

// NewReader returns a New io.Reader for the unarmored bianry data of the message
func (msg *PGPMessage) NewReader() io.Reader {
	return bytes.NewReader(msg.GetBinary())
}

// GetArmored returns the armored message as a string
func (msg *PGPMessage) GetArmored() (string, error) {
	return armor.ArmorWithType(msg.Data, constants.PGPMessageHeader)
}

// GetDataPacket returns the unarmored binary datapacket as a []byte
func (msg *PGPSplitMessage) GetDataPacket() []byte {
	return msg.DataPacket
}

// GetKeyPacket returns the unarmored binary keypacket as a []byte
func (msg *PGPSplitMessage) GetKeyPacket() []byte {
	return msg.KeyPacket
}

// SeparateKeyAndData returns the first keypacket and the (hopefully unique) dataPacket (not verified)
func (msg *PGPMessage) SeparateKeyAndData(estimatedLength, garbageCollector int)(outSplit *PGPSplitMessage, err error) {
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	packets := packet.NewReader(bytes.NewReader(msg.Data))
	outSplit = &PGPSplitMessage{}
	gcCounter := 0

	// Store encrypted key and symmetrically encrypted packet separately
	var encryptedKey *packet.EncryptedKey
	var decryptErr error
	for {
		var p packet.Packet
		if p, err = packets.Next(); err == io.EOF {
			err = nil
			break
		}
		switch p := p.(type) {
		case *packet.EncryptedKey:
			if encryptedKey != nil && encryptedKey.Key != nil {
				break
			}
			encryptedKey = p

		case *packet.SymmetricallyEncrypted:
			// FIXME: add support for multiple keypackets
			var b bytes.Buffer
			// 2^16 is an estimation of the size difference between input and output, the size difference is most probably
			// 16 bytes at a maximum though.
			// We need to avoid triggering a grow from the system as this will allocate too much memory causing problems
			// in low-memory environments
			b.Grow(1<<16 + estimatedLength)
			// empty encoded length + start byte
			b.Write(make([]byte, 6))
			b.WriteByte(byte(1))
			actualLength := 1
			block := make([]byte, 128)
			for {
				n, err := p.Contents.Read(block)
				if err == io.EOF {
					break
				}
				b.Write(block[:n])
				actualLength += n
				gcCounter += n
				if gcCounter > garbageCollector && garbageCollector > 0 {
					runtime.GC()
					gcCounter = 0
				}
			}

			// quick encoding
			symEncryptedData := b.Bytes()
			if actualLength < 192 {
				symEncryptedData[4] = byte(210)
				symEncryptedData[5] = byte(actualLength)
				symEncryptedData = symEncryptedData[4:]
			} else if actualLength < 8384 {
				actualLength = actualLength - 192
				symEncryptedData[3] = byte(210)
				symEncryptedData[4] = 192 + byte(actualLength>>8)
				symEncryptedData[5] = byte(actualLength)
				symEncryptedData = symEncryptedData[3:]
			} else {
				symEncryptedData[0] = byte(210)
				symEncryptedData[1] = byte(255)
				symEncryptedData[2] = byte(actualLength >> 24)
				symEncryptedData[3] = byte(actualLength >> 16)
				symEncryptedData[4] = byte(actualLength >> 8)
				symEncryptedData[5] = byte(actualLength)
			}

			outSplit.DataPacket = symEncryptedData
		}
	}
	if decryptErr != nil {
		return nil, fmt.Errorf("gopenpgp: cannot decrypt encrypted key packet: %v", decryptErr)
	}
	if encryptedKey == nil {
		return nil, errors.New("gopenpgp: packets don't include an encrypted key packet")
	}


	var buf bytes.Buffer
	if err := encryptedKey.Serialize(&buf); err != nil {
		return nil, fmt.Errorf("gopenpgp: cannot serialize encrypted key: %v", err)
	}
	outSplit.KeyPacket = buf.Bytes()

	return outSplit, nil
}

// GetBinary returns the unarmored binary content of the signature as a []byte
func (msg *PGPSignature) GetBinary() []byte {
	return msg.Data
}

// GetBinary returns the base-64 encoded binary content of the signature as a string
func (msg *PGPSignature) GetArmored() (string, error) {
	return armor.ArmorWithType(msg.Data, constants.PGPSignatureHeader)
}

// ---- UTILS -----

// IsPGPMessage checks if data if has armored PGP message format.
func (pgp *GopenPGP) IsPGPMessage(data string) bool {
	re := regexp.MustCompile("^-----BEGIN " + constants.PGPMessageHeader + "-----(?s:.+)-----END " +
		constants.PGPMessageHeader + "-----");
	return re.MatchString(data);
}
