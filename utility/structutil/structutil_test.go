package structutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testUser struct {
	ID          uint      `gorm:"column:id"`
	Name        string    `gorm:"column:name"`
	Email       string    `gorm:"column:email"`
	Age         int       `gorm:"column:age"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	OtherFields any       `gorm:"column:other_fields"` // 其他缺省字段
	private     string    // 不导出的字段
}

func TestDiffUpdateMap(t *testing.T) {
	t.Parallel()
	o := &testUser{
		ID:      1,
		Name:    "Alice",
		Email:   "alice@old.com",
		Age:     20,
		private: "hidden",
	}

	n := &testUser{
		ID:      1,
		Name:    "Alice",
		Email:   "alice@new.com",
		Age:     25,
		private: "should_ignore",
	}

	// 不忽略任何字段
	diff := DiffUpdateMap(o, n)

	assert.Len(t, diff, 2)
	assert.Equal(t, "alice@new.com", diff["email"])
	assert.Equal(t, 25, diff["age"])
	assert.NotContains(t, diff, "id")
	assert.NotContains(t, diff, "name")
	assert.NotContains(t, diff, "other_fields")
	assert.NotContains(t, diff, "private")

	// 忽略 email 字段
	diff2 := DiffUpdateMap(o, n, "email")
	assert.Len(t, diff2, 1)
	assert.Equal(t, 25, diff2["age"])
	assert.NotContains(t, diff2, "email")

	// 忽略所有字段，结果应该是空
	diff3 := DiffUpdateMap(o, n, "email", "age")
	assert.Empty(t, diff3)

	// 值没变，不应出现在结果里
	n.Email = "alice@old.com"
	n.Age = 20
	diff4 := DiffUpdateMap(o, n)
	assert.Empty(t, diff4)
}
