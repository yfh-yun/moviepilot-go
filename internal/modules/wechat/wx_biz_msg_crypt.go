package wechat

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
)

// 错误码定�?const (
	WXBizMsgCryptOK                     = 0
	WXBizMsgCryptValidateSignatureError = -40001
	WXBizMsgCryptParseXmlError          = -40002
	WXBizMsgCryptComputeSignatureError  = -40003
	WXBizMsgCryptIllegalAesKey          = -40004
	WXBizMsgCryptValidateCorpidError    = -40005
	WXBizMsgCryptEncryptAesError        = -40006
	WXBizMsgCryptDecryptAesError        = -40007
	WXBizMsgCryptIllegalBuffer          = -40008
	WXBizMsgCryptEncodeBase64Error      = -40009
	WXBizMsgCryptDecodeBase64Error      = -40010
	WXBizMsgCryptGenReturnXmlError      = -40011
)

// WXBizMsgCrypt 企业微信消息加解密类
type WXBizMsgCrypt struct {
	token          string
	encodingAESKey string
	receiveID      string
	aesKey         []byte
}

// NewWXBizMsgCrypt 创建新的WXBizMsgCrypt实例
func NewWXBizMsgCrypt(token, encodingAESKey, receiveID string) (*WXBizMsgCrypt, error) {
	// 解码AES Key
	aesKey, err := base64.StdEncoding.DecodeString(encodingAESKey + "=")
	if err != nil {
		return nil, fmt.Errorf("invalid encoding aes key: %v", err)
	}
	
	if len(aesKey) != 32 {
		return nil, fmt.Errorf("invalid aes key length, should be 32 bytes")
	}
	
	return &WXBizMsgCrypt{
		token:          token,
		encodingAESKey: encodingAESKey,
		receiveID:      receiveID,
		aesKey:         aesKey,
	}, nil
}

// VerifyURL 验证URL
func (w *WXBizMsgCrypt) VerifyURL(msgSignature, timestamp, nonce, echoStr string) (int, string) {
	// 验证签名
	sha1 := NewSHA1()
	ret, signature := sha1.getSHA1(w.token, timestamp, nonce, echoStr)
	if ret != WXBizMsgCryptOK {
		return ret, ""
	}
	
	if signature != msgSignature {
		return WXBizMsgCryptValidateSignatureError, ""
	}
	
	// 解密
	pc := NewPrpcrypt(w.aesKey)
	ret, replyEchoStr := pc.decrypt(echoStr, w.receiveID)
	if ret != WXBizMsgCryptOK {
		return ret, ""
	}
	
	return WXBizMsgCryptOK, string(replyEchoStr)
}

// EncryptMsg 加密消息
func (w *WXBizMsgCrypt) EncryptMsg(replyMsg, nonce string, timestamp *string) (int, string) {
	// 加密消息
	pc := NewPrpcrypt(w.aesKey)
	ret, encrypt := pc.encrypt(replyMsg, w.receiveID)
	if ret != WXBizMsgCryptOK {
		return ret, ""
	}
	
	// 生成时间�?	ts := ""
	if timestamp == nil {
		ts = fmt.Sprintf("%d", time.Now().Unix())
	} else {
		ts = *timestamp
	}
	
	// 生成签名
	sha1 := NewSHA1()
	ret, signature := sha1.getSHA1(w.token, ts, nonce, string(encrypt))
	if ret != WXBizMsgCryptOK {
		return ret, ""
	}
	
	// 生成XML
	xmlParse := NewXMLParse()
	xmlStr := xmlParse.generate(string(encrypt), signature, ts, nonce)
	
	return WXBizMsgCryptOK, xmlStr
}

// DecryptMsg 解密消息
func (w *WXBizMsgCrypt) DecryptMsg(postData, msgSignature, timestamp, nonce string) (int, string) {
	// 提取加密内容
	xmlParse := NewXMLParse()
	ret, encrypt := xmlParse.extract(postData)
	if ret != WXBizMsgCryptOK {
		return ret, ""
	}
	
	// 验证签名
	sha1 := NewSHA1()
	ret, signature := sha1.getSHA1(w.token, timestamp, nonce, encrypt)
	if ret != WXBizMsgCryptOK {
		return ret, ""
	}
	
	if signature != msgSignature {
		return WXBizMsgCryptValidateSignatureError, ""
	}
	
	// 解密
	pc := NewPrpcrypt(w.aesKey)
	ret, xmlContent := pc.decrypt(encrypt, w.receiveID)
	if ret != WXBizMsgCryptOK {
		return ret, ""
	}
	
	return WXBizMsgCryptOK, string(xmlContent)
}

// SHA1 SHA1计算�?type SHA1 struct{}

// NewSHA1 创建SHA1实例
func NewSHA1() *SHA1 {
	return &SHA1{}
}

// getSHA1 用SHA1算法生成安全签名
func (s *SHA1) getSHA1(token, timestamp, nonce, encrypt string) (int, string) {
	// 拼接排序
	strs := []string{token, timestamp, nonce, encrypt}
	sort.Strings(strs)
	
	// 计算SHA1
	h := sha1.New()
	h.Write([]byte(strings.Join(strs, "")))
	
	return WXBizMsgCryptOK, fmt.Sprintf("%x", h.Sum(nil))
}

// XMLParse XML解析�?type XMLParse struct {
	aesTextResponseTemplate string
}

// NewXMLParse 创建XMLParse实例
func NewXMLParse() *XMLParse {
	return &XMLParse{
		aesTextResponseTemplate: `<xml>
<Encrypt><![CDATA[%(msg_encrypt)s]]></Encrypt>
<MsgSignature><![CDATA[%(msg_signaturet)s]]></MsgSignature>
<TimeStamp>%(timestamp)s</TimeStamp>
<Nonce><![CDATA[%(nonce)s]]></Nonce>
</xml>`,
	}
}

// extract 提取xml数据包中的加密消�?func (x *XMLParse) extract(xmlText string) (int, string) {
	var result struct {
		Encrypt string `xml:"Encrypt"`
	}
	
	err := xml.Unmarshal([]byte(xmlText), &result)
	if err != nil {
		return WXBizMsgCryptParseXmlError, ""
	}
	
	return WXBizMsgCryptOK, result.Encrypt
}

// generate 生成xml消息
func (x *XMLParse) generate(encrypt, signature, timestamp, nonce string) string {
	result := strings.Replace(x.aesTextResponseTemplate, "%(msg_encrypt)s", encrypt, -1)
	result = strings.Replace(result, "%(msg_signaturet)s", signature, -1)
	result = strings.Replace(result, "%(timestamp)s", timestamp, -1)
	result = strings.Replace(result, "%(nonce)s", nonce, -1)
	
	return result
}

// PKCS7Encoder PKCS7填充�?type PKCS7Encoder struct {
	blockSize int
}

// NewPKCS7Encoder 创建PKCS7Encoder实例
func NewPKCS7Encoder() *PKCS7Encoder {
	return &PKCS7Encoder{
		blockSize: 32,
	}
}

// encode 对需要加密的明文进行填充补位
func (p *PKCS7Encoder) encode(text []byte) []byte {
	textLength := len(text)
	// 计算需要填充的位数
	amountToPad := p.blockSize - (textLength % p.blockSize)
	if amountToPad == 0 {
		amountToPad = p.blockSize
	}
	
	// 获得补位所用的字符
	pad := bytes.Repeat([]byte{byte(amountToPad)}, amountToPad)
	return append(text, pad...)
}

// decode 删除解密后明文的补位字符
func (p *PKCS7Encoder) decode(decrypted []byte) []byte {
	if len(decrypted) == 0 {
		return decrypted
	}
	
	pad := int(decrypted[len(decrypted)-1])
	if pad < 1 || pad > 32 {
		pad = 0
	}
	
	if pad > len(decrypted) {
		return decrypted
	}
	
	return decrypted[:len(decrypted)-pad]
}

// Prpcrypt 加解密类
type Prpcrypt struct {
	key  []byte
	mode cipher.BlockMode
}

// NewPrpcrypt 创建Prpcrypt实例
func NewPrpcrypt(key []byte) *Prpcrypt {
	return &Prpcrypt{
		key: key,
	}
}

// encrypt 对明文进行加�?func (p *Prpcrypt) encrypt(text, receiveID string) (int, []byte) {
	// 16位随机字符串添加到明文开�?	randomStr := p.getRandomStr()
	
	// 构造消息体
	textBytes := []byte(text)
	msgLen := make([]byte, 4)
	binary.BigEndian.PutUint32(msgLen, uint32(len(textBytes)))
	
	// 拼接: random(16) + msg_len(4) + msg + receiveID
	plaintext := append(randomStr, msgLen...)
	plaintext = append(plaintext, textBytes...)
	plaintext = append(plaintext, []byte(receiveID)...)
	
	// 使用自定义的填充方式对明文进行补位填�?	pkcs7 := NewPKCS7Encoder()
	plaintext = pkcs7.encode(plaintext)
	
	// 加密
	block, err := aes.NewCipher(p.key)
	if err != nil {
		return WXBizMsgCryptEncryptAesError, nil
	}
	
	// CBC模式加密
	iv := p.key[:aes.BlockSize]
	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(plaintext))
	mode.CryptBlocks(ciphertext, plaintext)
	
	// 使用BASE64对加密后的字符串进行编码
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(ciphertext)))
	base64.StdEncoding.Encode(encoded, ciphertext)
	
	return WXBizMsgCryptOK, encoded
}

// decrypt 对密文进行解�?func (p *Prpcrypt) decrypt(text, receiveID string) (int, []byte) {
	// 使用BASE64对密文进行解�?	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(text)))
	n, err := base64.StdEncoding.Decode(decoded, []byte(text))
	if err != nil {
		return WXBizMsgCryptDecodeBase64Error, nil
	}
	decoded = decoded[:n]
	
	// 解密
	block, err := aes.NewCipher(p.key)
	if err != nil {
		return WXBizMsgCryptDecryptAesError, nil
	}
	
	// CBC模式解密
	iv := p.key[:aes.BlockSize]
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(decoded))
	mode.CryptBlocks(plaintext, decoded)
	
	// 去掉补位字符
	pkcs7 := NewPKCS7Encoder()
	plaintext = pkcs7.decode(plaintext)
	
	// 去除16位随机字符串
	if len(plaintext) <= 20 {
		return WXBizMsgCryptIllegalBuffer, nil
	}
	
	content := plaintext[16:]
	
	// 获取消息长度
	msgLen := binary.BigEndian.Uint32(content[:4])
	
	// 获取消息内容
	xmlContent := content[4 : 4+msgLen]
	
	// 获取receiveID
	fromReceiveID := content[4+msgLen:]
	
	if string(fromReceiveID) != receiveID {
		return WXBizMsgCryptValidateCorpidError, nil
	}
	
	return WXBizMsgCryptOK, xmlContent
}

// getRandomStr 随机生成16位字符串
func (p *Prpcrypt) getRandomStr() []byte {
	// 生成16位随机字符串
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 16)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return b
}
