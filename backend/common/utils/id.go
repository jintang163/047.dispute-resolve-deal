package utils

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/sony/sonyflake"
	"github.com/google/uuid"
)

var (
	flake *sonyflake.Sonyflake
	once  sync.Once
)

func InitIDGenerator(machineID uint16) {
	once.Do(func() {
		flake = sonyflake.NewSonyflake(sonyflake.Settings{
			StartTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			MachineID: func() (uint16, error) {
				return machineID, nil
			},
		})
	})
}

func GenerateID() int64 {
	if flake == nil {
		InitIDGenerator(1)
	}
	id, err := flake.NextID()
	if err != nil {
		return time.Now().UnixNano() / 1e6
	}
	return int64(id)
}

func GenerateUUID() string {
	return uuid.New().String()
}

func GenerateCaseNo(orgCode string) string {
	now := time.Now()
	return fmt.Sprintf("JF%s%s%06d", orgCode, now.Format("20060102"), GenerateID()%1000000)
}

func GenerateApprovalNo() string {
	now := time.Now()
	return fmt.Sprintf("SP%s%08d", now.Format("20060102"), GenerateID()%100000000)
}

func GenerateConsultNo() string {
	now := time.Now()
	return fmt.Sprintf("ZX%s%08d", now.Format("20060102"), GenerateID()%100000000)
}

func GenerateConfirmNo(orgID int64) string {
	now := time.Now()
	return fmt.Sprintf("SF%04d%s%06d", orgID%10000, now.Format("20060102"), GenerateID()%1000000)
}

func GenerateTransferNo() string {
	now := time.Now()
	return fmt.Sprintf("ZB%s%08d", now.Format("20060102"), GenerateID()%100000000)
}

func GenerateIDStr() string {
	return strconv.FormatInt(GenerateID(), 10)
}

func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seed.Intn(len(charset))]
	}
	return string(b)
}

func GenerateRandomNumber(length int) string {
	const charset = "0123456789"
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seed.Intn(len(charset))]
	}
	return string(b)
}
