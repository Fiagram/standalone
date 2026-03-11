package dao_database_test

import (
	"math/rand"
	"strings"
	"time"

	dao_database "github.com/Fiagram/standalone/internal/dao/database"
)

func RandomString(length uint) string {
	g := rand.New(rand.NewSource(time.Now().UnixNano()))
	const alphabet = "qazwsxedcrfvtgbyhnujmikolp"
	var sb strings.Builder
	k := len(alphabet)
	for i := 0; i < int(length); i++ {
		c := alphabet[g.Intn(k)]
		sb.WriteByte(c)
	}
	return strings.TrimSpace(sb.String())
}

func RandomVnPersonName() string {
	lastnames := []string{
		"Nguyễn", "Vũ", "Trần", "Huỳnh", "Lê", "Phạm",
		"Phan", "Hoàng", "Phùng", "Tô", "Mai", "Trương"}
	middles := []string{
		"Đỗ", "Đức", "Mạnh", "Thị", "Uyển", "Lâm",
		"Văn", "Hàn", "Thùy", "Anh", "Duy", "Khánh",
	}
	firstnames := []string{
		"Thế", "Tuấn", "Trung", "Hùng", "Dũng", "Tân",
		"Hà", "Trí", "Hiếu", "Thái", "Tiến", "Ngọc",
	}
	g := rand.New(rand.NewSource(time.Now().UnixNano()))
	out := make([]string, 4)
	out = append(out, lastnames[g.Intn(len(lastnames))])
	out = append(out, middles[g.Intn(len(middles))])
	out = append(out, middles[g.Intn(len(middles))])
	out = append(out, firstnames[g.Intn(len(firstnames))])
	return strings.TrimSpace(strings.Join(out, " "))
}

func RandomGmailAddress() string {
	return RandomString(50) + "@gmail.com"
}

func RandomVnPhoneNum() string {
	g := rand.New(rand.NewSource(time.Now().UnixNano()))
	const nums = "0123456789"
	var sb strings.Builder
	k := len(nums)
	for i := range 9 {
		c := nums[rand.Intn(k)]
		for c == '0' && i == 0 {
			c = nums[g.Intn(k)]
		}
		sb.WriteByte(c)
	}
	return strings.TrimSpace("+84 " + sb.String())
}

func RandomAccount() dao_database.Account {
	return dao_database.Account{
		Username:    RandomString(50),
		Fullname:    RandomVnPersonName(),
		Email:       RandomGmailAddress(),
		PhoneNumber: RandomVnPhoneNum(),
		RoleId:      2,
	}
}
