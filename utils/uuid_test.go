package utils_test

import (
	"gitlab.dian.org.cn/helper/miniapp-platform/utils"
	"testing"
)

func TestUUID(t *testing.T) {
	uuid, err := utils.UUID()
	if err != nil {
		t.Error(err)
	}
	t.Log(uuid)
}
