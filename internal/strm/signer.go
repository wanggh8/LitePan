package strm

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"strings"
)

// trimFilenameSegment 从 STRM 播放路径中剥离 /n/{filename} 尾段，仅保留核心鉴权凭证：
// /api/strm/play/{account_id}/{file_key}/t/{token} 或 /api/strm/path/.../t/{token}
// 签名完全基于核心凭证计算，不包含文件名。
func trimFilenameSegment(path string) string {
	if idx := strings.Index(path, "/n/"); idx != -1 {
		return path[:idx]
	}
	return path
}

func SignPath(path string, secret []byte) string {
	target := trimFilenameSegment(path)
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(target))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func VerifyPath(path, signature string, secret []byte) bool {
	if signature == "" {
		return false
	}
	expected := SignPath(path, secret)
	return subtle.ConstantTimeCompare([]byte(expected), []byte(signature)) == 1
}
