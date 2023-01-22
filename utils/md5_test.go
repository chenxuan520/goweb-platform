package utils_test

import (
	"gitlab.dian.org.cn/helper/miniapp-platform/utils"
	"testing"
)

func TestMD5(t *testing.T) {
	t.Log(utils.MD5("123456"))
}
